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

package setconfig

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	requests "github.com/edgexfoundry/go-mod-core-contracts/requests/configuration"
	responses "github.com/edgexfoundry/go-mod-core-contracts/responses/configuration"

	"github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

// executor contains references to dependencies required to execute a set configuration request.
type executor struct {
	loggingClient logger.LoggingClient
	configuration *config.ConfigurationStruct
}

// NewExecutor is a factory function that returns an initialized executor struct.
func NewExecutor(lc logger.LoggingClient, configuration *config.ConfigurationStruct) *executor {
	return &executor{
		loggingClient: lc,
		configuration: configuration,
	}
}

// Do fulfills the SetExecutor contract and implements the functionality to set a service's configuration.
func (e executor) Do(service string, sc requests.SetConfigRequest) responses.SetConfigResponse {
	createErrorResponse := func(message string) responses.SetConfigResponse {
		e.loggingClient.Error(message)
		return responses.SetConfigResponse{
			Success:     false,
			Description: message,
		}
	}

	// The SMA will set configuration via Consul if EdgeX has been launched with the "--registry" flag.
	e.loggingClient.Info(fmt.Sprintf("the SMA has been requested to set (aka PUT/UPDATE) the config for: %s", service))
	e.loggingClient.Debug(fmt.Sprintf("key %s to use for config updated", sc.Key))
	e.loggingClient.Debug(fmt.Sprintf("value %s to use for config updated", sc.Value))

	// create a registryClient specific to the service and connect to the registry as if we are that service so
	// that we can update the service's corresponding key based on the request we received.
	var serviceSpecificRegistryClient registry.Client
	serviceSpecificRegistryClient, err := registry.NewRegistryClient(
		types.Config{
			Host:       e.configuration.Registry.Host,
			Port:       e.configuration.Registry.Port,
			Type:       e.configuration.Registry.Type,
			Stem:       internal.ConfigRegistryStemCore + internal.ConfigMajorVersion,
			ServiceKey: service,
		})
	if err != nil {
		return createErrorResponse("unable to create new registry client")
	}

	// Validate whether the key exists.
	key := strings.Replace(sc.Key, ".", "/", -1)
	exists, err := serviceSpecificRegistryClient.ConfigurationValueExists(key)
	switch {
	case err != nil:
		return createErrorResponse(err.Error())
	case !exists:
		return createErrorResponse("key does not exist")
	default:
		if err := serviceSpecificRegistryClient.PutConfigurationValue(key, []byte(sc.Value)); err != nil {
			return createErrorResponse("unable to update key")
		}

		return responses.SetConfigResponse{
			Success: true,
		}
	}
}
