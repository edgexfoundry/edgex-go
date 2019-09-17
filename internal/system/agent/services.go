/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package agent

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/response"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

func getConfig(
	services []string,
	ctx context.Context,
	loggingClient logger.LoggingClient,
	genClients *GeneralClients,
	configClients ConfigurationClients,
	registryClient registry.Client,
	serviceProtocol string) (interface{}, error) {

	result := struct {
		Configuration map[string]interface{} `json:"configuration"`
	}{
		Configuration: map[string]interface{}{},
	}

	// Loop through requested actions, along with (any) respectively-supplied parameters.
	for _, service := range services {
		client, ok := genClients.Get(service)
		// Check whether SMA does _not_ know of ServiceKey ("service") as being one for one of its ready-made list of clients.
		if !ok {
			loggingClient.Info(fmt.Sprintf("service %s not known to SMA as being in the ready-made list of clients", service))

			if registryClient == nil {
				result.Configuration[service] = "registry is required to obtain configuration."
				continue
			}

			// Service unknown to SMA, so ask the Registry whether `service` is available.
			err := registryClient.IsServiceAvailable(service)
			if err != nil {
				result.Configuration[service] = fmt.Sprintf(err.Error())
				loggingClient.Error(err.Error())
				continue
			}

			loggingClient.Info(fmt.Sprintf("Registry responded with %s service available", service))

			// Since service is unknown to SMA, ask the Registry for a ServiceEndpoint associated with `service`
			e, err := registryClient.GetServiceEndpoint(service)
			if err != nil {
				result.Configuration[service] = fmt.Sprintf("on attempting to get ServiceEndpoint for service %s, got error: %v", service, err.Error())
				loggingClient.Error(err.Error())
				continue
			}

			// This code will evolve to take into account a manifest-like functionality in future. So
			// rather than assume that the runtime bool flag useRegistry has been initialized to true,
			// given that the flow has reached this point, having already called functions on the Registry,
			// such as RegistryClient.IsServiceAvailable(service), we test for its truthiness. I expect
			// this code to be refactored as we evolve toward a manifest-like functionality in future.
			configClients[e.ServiceId] = config.ClientInfo{
				Protocol: serviceProtocol,
				Host:     e.Host,
				Port:     e.Port,
			}

			params := types.EndpointParams{
				ServiceKey:  e.ServiceId,
				Path:        "/",
				UseRegistry: registryClient != nil,
				Url:         configClients[e.ServiceId].Url() + clients.ApiConfigRoute,
				Interval:    internal.ClientMonitorDefault,
			}
			// Add the service key to the map where the value is the respective GeneralClient
			client = general.NewGeneralClient(params, startup.Endpoint{RegistryClient: &registryClient})
			genClients.Set(e.ServiceId, client)
		}

		responseJSON, err := client.FetchConfiguration(ctx)
		if err != nil {
			result.Configuration[service] = fmt.Sprintf(err.Error())
			loggingClient.Error(err.Error())
			continue
		}

		result.Configuration[service] = response.Process(responseJSON, loggingClient)
	}
	return result, nil
}

func getHealth(services []string, registryClient registry.Client) (map[string]interface{}, error) {
	health := make(map[string]interface{})
	for _, service := range services {
		if registryClient == nil {
			health[service] = "registry is required to obtain service health status."
			continue
		}

		// the registry service returns nil for a healthy service
		if err := registryClient.IsServiceAvailable(service); err != nil {
			health[service] = err.Error()
			continue
		}

		health[service] = true
	}
	return health, nil
}
