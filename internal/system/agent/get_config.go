package agent

import (
	"context"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal"
	conf "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	responses "github.com/edgexfoundry/go-mod-core-contracts/responses/configuration"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

type getConf struct {
	genClients      *GeneralClients
	configClients   ConfigurationClients
	registryClient  registry.Client
	loggingClient   logger.LoggingClient
	serviceProtocol string
}

func NewGetConfig(genClients *GeneralClients,
	configClients ConfigurationClients,
	registryClient registry.Client,
	loggingClient logger.LoggingClient,
	serviceProtocol string) *getConf {
	return &getConf{
		genClients:      genClients,
		configClients:   configClients,
		registryClient:  registryClient,
		loggingClient:   loggingClient,
		serviceProtocol: serviceProtocol,
	}
}

func (c getConf) GetConfig(serviceName string, ctx context.Context) (string, error) {
	var result string
	client, ok := c.genClients.Get(serviceName)
	if !ok {
		if c.registryClient == nil {
			return "", fmt.Errorf("registryClient not initialized; required to handle unknown service: %s", serviceName)
		}

		// Service unknown to SMA, so ask the Registry whether `serviceName` is available.
		if err := c.registryClient.IsServiceAvailable(serviceName); err != nil {
			return "", err
		}

		c.loggingClient.Info(fmt.Sprintf("Registry responded with %s available", serviceName))

		// Since serviceName is unknown to SMA, ask the Registry for a ServiceEndpoint associated with `serviceName`
		ep, err := c.registryClient.GetServiceEndpoint(serviceName)
		if err != nil {
			return "", fmt.Errorf("on attempting to get ServiceEndpoint for %s, got error: %v", serviceName, err.Error())
		}

		configClient := conf.ClientInfo{
			Protocol: c.serviceProtocol,
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
		client = general.NewGeneralClient(params, endpoint.Endpoint{RegistryClient: &c.registryClient})
		c.genClients.Set(ep.ServiceId, client)
	}

	result, err := client.FetchConfiguration(ctx)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (c getConf) createErrorResponse(description string) responses.SetConfigResponse {
	c.loggingClient.Error(description)
	return responses.SetConfigResponse{
		Success:     false,
		Description: description,
	}
}
