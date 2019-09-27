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
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
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
	genClients      *agent.GeneralClients
	registryClient  registry.Client
	serviceProtocol string
}

// NewMetrics is a factory function that returns an initialized metrics receiver struct.
func NewMetrics(
	loggingClient logger.LoggingClient,
	genClients *agent.GeneralClients,
	registryClient registry.Client,
	serviceProtocol string) *metrics {

	return &metrics{
		loggingClient:   loggingClient,
		genClients:      genClients,
		registryClient:  registryClient,
		serviceProtocol: serviceProtocol,
	}
}

// metricsViaDirectService calls a service's metrics endpoint directly, interprets the response, and returns a Result.
func (m *metrics) metricsViaDirectService(serviceName string, ctx context.Context) (system.Result, error) {
	client, ok := m.genClients.Get(serviceName)
	if !ok {
		if m.registryClient == nil {
			return nil, fmt.Errorf("registryClient not initialized; required to handle unknown service: %s", serviceName)
		}

		// Service unknown to SMA, so ask the Registry whether `serviceName` is available.
		if err := m.registryClient.IsServiceAvailable(serviceName); err != nil {
			return nil, err
		}

		m.loggingClient.Info(fmt.Sprintf("Registry responded with %s serviceName available", serviceName))

		// Since serviceName is unknown to SMA, ask the Registry for a ServiceEndpoint associated with `serviceName`
		e, err := m.registryClient.GetServiceEndpoint(serviceName)
		if err != nil {
			return nil, fmt.Errorf("on attempting to get ServiceEndpoint for serviceName %s, got error: %v", serviceName, err.Error())
		}

		configClient := config.ClientInfo{
			Protocol: m.serviceProtocol,
			Host:     e.Host,
			Port:     e.Port,
		}
		params := types.EndpointParams{
			ServiceKey:  e.ServiceId,
			Path:        "/",
			UseRegistry: true,
			Url:         configClient.Url() + clients.ApiMetricsRoute,
			Interval:    internal.ClientMonitorDefault,
		}

		// Add the serviceName key to the map where the value is the respective GeneralClient
		client = general.NewGeneralClient(params, endpoint.Endpoint{RegistryClient: &m.registryClient})
		m.genClients.Set(e.ServiceId, client)
	}

	result, err := client.FetchMetrics(ctx)
	if err != nil {
		return nil, err
	}

	var s telemetry.SystemUsage
	if err := json.NewDecoder(bytes.NewBuffer([]byte(result))).Decode(&s); err != nil {
		return nil, fmt.Errorf("error decoding telemetry.SystemUsage: %s", err.Error())
	}

	return system.MetricsSuccess(serviceName, ExecutorType, s.CpuBusyAvg, int64(s.Memory.Sys), []byte(result)), nil
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
