/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package direct

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	"github.com/edgexfoundry/edgex-go/internal/system"
	"github.com/edgexfoundry/edgex-go/internal/system/agent"
	"github.com/edgexfoundry/edgex-go/internal/system/executor"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

// metrics contains references to dependencies required to handle the metrics via direct service use case.
type metrics struct {
	loggingClient   logger.LoggingClient
	genClients      agent.GeneralClients
	configClients   agent.ConfigurationClients
	registryClient  registry.Client
	serviceProtocol string
}

// NewMetrics is a factory method that returns an initialized metrics receiver struct.
func NewMetrics(
	loggingClient logger.LoggingClient,
	genClients agent.GeneralClients,
	configClients agent.ConfigurationClients,
	registryClient registry.Client,
	serviceProtocol string) *metrics {

	return &metrics{
		loggingClient:   loggingClient,
		genClients:      genClients,
		configClients:   configClients,
		registryClient:  registryClient,
		serviceProtocol: serviceProtocol,
	}
}

// fetchMetrics provides a common implementation to gather metrics from a service defined in genClients,
// transform and normalize the cpuUsedPercent and memoryUsed fields provided by every executor (along with the raw
// result returned by the service), and return a corresponding MetricsSuccessResult.
func (m *metrics) fetchMetrics(serviceName string, ctx context.Context) (system.Result, error) {
	result, err := m.genClients[serviceName].FetchMetrics(ctx)
	if err != nil {
		return nil, err
	}

	var s telemetry.SystemUsage
	if err := json.NewDecoder(bytes.NewBuffer([]byte(result))).Decode(&s); err != nil {
		return nil, fmt.Errorf("error decoding telemetry.SystemUsage: %s", err.Error())
	}

	return system.MetricsSuccess(serviceName, ExecutorType, s.CpuBusyAvg, int64(s.Memory.Sys), []byte(result)), nil
}

// handleUnknownService fetches metrics from an unknown service (i.e. a service that does not have an entry in
// genClients).  It leverages the registry -- assuming it is enabled -- to query for the unknown service's
// endpoint settings.  If found, the service is added to genClients and the service's metrics are fetched and
// returned.
func (m *metrics) handleUnknownService(serviceName string, ctx context.Context) (system.Result, error) {
	m.loggingClient.Info(fmt.Sprintf("service %s not known to SMA as being in the ready-made list of clients", serviceName))

	if m.registryClient == nil {
		return nil, fmt.Errorf("registryClient not initialized; required to handle unknown service %s", serviceName)
	}

	// Service unknown to SMA, so ask the Registry whether `serviceName` is available.
	if err := m.registryClient.IsServiceAvailable(serviceName); err != nil {
		return nil, err
	}

	m.loggingClient.Info(fmt.Sprintf("Registry responded with %s serviceName available", serviceName))

	// Since serviceName is unknown to SMA, ask the Registry for a ServiceEndpoint associated with `serviceName`
	endpoint, err := m.registryClient.GetServiceEndpoint(serviceName)
	if err != nil {
		return nil, fmt.Errorf("on attempting to get ServiceEndpoint for serviceName %s, got error: %v", serviceName, err.Error())
	}

	// add the specified key to the map where the value will be the respective GeneralClient
	m.configClients[endpoint.ServiceId] = config.ClientInfo{
		Protocol: m.serviceProtocol,
		Host:     endpoint.Host,
		Port:     endpoint.Port,
	}

	params := types.EndpointParams{
		ServiceKey:  endpoint.ServiceId,
		Path:        "/",
		UseRegistry: true,
		Url:         m.configClients[endpoint.ServiceId].Url() + clients.ApiMetricsRoute,
		Interval:    internal.ClientMonitorDefault,
	}

	// Add the serviceName key to the map where the value is the respective GeneralClient
	m.genClients[endpoint.ServiceId] = general.NewGeneralClient(params, startup.Endpoint{RegistryClient: &m.registryClient})

	return m.fetchMetrics(endpoint.ServiceId, ctx)
}

// handleKnownService fetches metrics from a known service (i.e. a service that has an entry in genClients).
func (m *metrics) handleKnownService(serviceName string, ctx context.Context) (system.Result, error) {
	// Service is known to SMA, so no need to ask the Registry for a ServiceEndpoint associated with `serviceName`
	// Simply use one of the ready-made list of clients.
	m.loggingClient.Info(fmt.Sprintf("serviceName %s is known to SMA as being in the ready-made list of clients", serviceName))
	return m.fetchMetrics(serviceName, ctx)
}

// metricsViaDirectService calls a service's metrics endpoint directly, interprets the endpoint's response, and returns
// a Result value.
func (m *metrics) metricsViaDirectService(serviceName string, ctx context.Context) (system.Result, error) {
	if _, ok := m.genClients[serviceName]; ok {
		return m.handleKnownService(serviceName, ctx)
	}
	return m.handleUnknownService(serviceName, ctx)
}

func (m *metrics) Get(services []string, ctx context.Context) interface{} {
	var result []interface{}
	for _, service := range services {
		out, err := m.metricsViaDirectService(service, ctx)
		if err != nil {
			result = append(result, system.Failure(service, executor.Metrics, ExecutorType, err.Error()))
			continue
		}
		result = append(result, out)
	}
	return result
}
