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
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/config"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/config"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-configuration/v2/configuration"
	"github.com/edgexfoundry/go-mod-configuration/v2/pkg/types"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	requests "github.com/edgexfoundry/go-mod-core-contracts/v2/requests/configuration"
	responses "github.com/edgexfoundry/go-mod-core-contracts/v2/responses/configuration"
)

// executor contains references to dependencies required to execute a set configuration request.
type executor struct {
	loggingClient logger.LoggingClient
	configuration *config.ConfigurationStruct
	dic           *di.Container
}

// NewExecutor is a factory function that returns an initialized executor struct.
func NewExecutor(lc logger.LoggingClient, configuration *config.ConfigurationStruct, dic *di.Container) *executor {
	return &executor{
		loggingClient: lc,
		configuration: configuration,
		dic:           dic,
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

	secretProvider := bootstrapContainer.SecretProviderFrom(e.dic.Get)
	accessToken, err := secretProvider.GetAccessToken(e.configuration.Registry.Type, clients.SystemManagementAgentServiceKey)
	if err != nil {
		return createErrorResponse(fmt.Errorf("failed to get Configuration Provider of type `%s` access token: %w",
			e.configuration.Registry.Type, err).Error())
	}

	// based on registry config info, we create a config client specific to the service and connect to the config host
	// as if we are that service so that we can update the service's corresponding key based on the request we received.
	var serviceSpecificConfigClient configuration.Client
	serviceSpecificConfigClient, err = configuration.NewConfigurationClient(
		types.ServiceConfig{
			// using registry info to set up the service config
			Host:        e.configuration.Registry.Host,
			Port:        e.configuration.Registry.Port,
			Type:        e.configuration.Registry.Type,
			BasePath:    filepath.Join(internal.ConfigStemCore, bootstrapConfig.ConfigVersion, service),
			AccessToken: accessToken,
		})
	if err != nil {
		return createErrorResponse("unable to create new service configuration client based on registry info")
	}

	// Validate whether the key exists.
	key := strings.Replace(sc.Key, ".", "/", -1)
	exists, err := serviceSpecificConfigClient.ConfigurationValueExists(key)
	switch {
	case err != nil:
		return createErrorResponse(err.Error())
	case !exists:
		return createErrorResponse("key does not exist")
	default:
		if err := serviceSpecificConfigClient.PutConfigurationValue(key, []byte(sc.Value)); err != nil {
			return createErrorResponse("unable to update key")
		}

		return responses.SetConfigResponse{
			Success: true,
		}
	}
}