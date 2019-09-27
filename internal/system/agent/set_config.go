package agent

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	requests "github.com/edgexfoundry/go-mod-core-contracts/requests/configuration"
	responses "github.com/edgexfoundry/go-mod-core-contracts/responses/configuration"
	"github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
	"strings"
)

type setConf struct {
	loggingClient logger.LoggingClient
	configuration *ConfigurationStruct
}

func NewSetConfig(loggingClient logger.LoggingClient, configuration *ConfigurationStruct) *setConf {
	return &setConf{
		loggingClient: loggingClient,
		configuration: configuration,
	}
}

func (c setConf) SetConfig(service string, sc requests.SetConfigRequest) responses.SetConfigResponse {

	// The SMA will set configuration via Consul if EdgeX has been launched with the "--registry" flag.
	c.loggingClient.Info(fmt.Sprintf("the SMA has been requested to set (aka PUT/UPDATE) the config for: %s", service))
	c.loggingClient.Debug(fmt.Sprintf("key %s to use for config updated", sc.Key))
	c.loggingClient.Debug(fmt.Sprintf("value %s to use for config updated", sc.Value))

	// create a registryClient specific to the service and connect to the registry as if we are that service so
	// that we can update the service's corresponding key based on the request we received.
	var serviceSpecificRegistryClient registry.Client
	serviceSpecificRegistryClient, err := registry.NewRegistryClient(
		types.Config{
			Host:       c.configuration.Registry.Host,
			Port:       c.configuration.Registry.Port,
			Type:       c.configuration.Registry.Type,
			Stem:       internal.ConfigRegistryStemCore + internal.ConfigMajorVersion,
			ServiceKey: service,
		})
	if err != nil {
		return c.createErrorResponse("unable to create new registry client")
	}

	// Validate whether the key exists.
	key := strings.Replace(sc.Key, ".", "/", -1)
	exists, err := serviceSpecificRegistryClient.ConfigurationValueExists(key)
	switch {
	case err != nil:
		return c.createErrorResponse(err.Error())
	case !exists:
		return c.createErrorResponse("key does not exist")
	default:
		if err := serviceSpecificRegistryClient.PutConfigurationValue(key, []byte(sc.Value)); err != nil {
			return c.createErrorResponse("unable to update key")
		}

		return responses.SetConfigResponse{
			Success: true,
		}
	}
}

func (c setConf) createErrorResponse(description string) responses.SetConfigResponse {
	c.loggingClient.Error(description)
	return responses.SetConfigResponse{
		Success:     false,
		Description: description,
	}
}
