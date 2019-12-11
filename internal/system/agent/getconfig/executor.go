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

package getconfig

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	agentClients "github.com/edgexfoundry/edgex-go/internal/system/agent/clients"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

// executor contains references to dependencies required to execute a get configuration request.
type executor struct {
	genClients      *agentClients.General
	registryClient  registry.Client
	loggingClient   logger.LoggingClient
	serviceProtocol string
}

// NewExecutor is a factory function that returns an initialized executor struct.
func NewExecutor(
	genClients *agentClients.General,
	registryClient registry.Client,
	loggingClient logger.LoggingClient,
	serviceProtocol string) *executor {

	return &executor{
		genClients:      genClients,
		registryClient:  registryClient,
		loggingClient:   loggingClient,
		serviceProtocol: serviceProtocol,
	}
}

// Do fulfills the GetExecutor contract and implements the functionality to retrieve a service's configuration.
func (e executor) Do(ctx context.Context, serviceName string) (string, error) {
	var result string
	client, ok := e.genClients.Get(serviceName)
	if !ok {
		if e.registryClient == nil {
			return "", fmt.Errorf("registryClient not initialized; required to handle unknown service: %s", serviceName)
		}

		// Service unknown to SMA, so ask the Registry whether `serviceName` is available.
		if err := e.registryClient.IsServiceAvailable(serviceName); err != nil {
			return "", err
		}

		e.loggingClient.Info(fmt.Sprintf("Registry responded with %s available", serviceName))

		// Since serviceName is unknown to SMA, ask the Registry for a ServiceEndpoint associated with `serviceName`
		ep, err := e.registryClient.GetServiceEndpoint(serviceName)
		if err != nil {
			return "", fmt.Errorf("on attempting to get ServiceEndpoint for %s, got error: %v", serviceName, err.Error())
		}

		configClient := config.ClientInfo{
			Protocol: e.serviceProtocol,
			Host:     ep.Host,
			Port:     ep.Port,
		}
		params := types.EndpointParams{
			ServiceKey:  ep.ServiceId,
			Path:        "/",
			UseRegistry: true,
			Url:         configClient.Url() + clients.ApiConfigRoute,
			Interval:    internal.ClientMonitorDefault,
		}

		// Add the serviceName key to the map where the value is the respective GeneralClient
		client = general.NewGeneralClient(params, endpoint.Endpoint{RegistryClient: &e.registryClient})
		e.genClients.Set(ep.ServiceId, client)
	}

	result, err := client.FetchConfiguration(ctx)
	if err != nil {
		return "", err
	}
	return result, nil
}
