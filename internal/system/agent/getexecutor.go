package agent

import (
	"context"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/response"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

type getExecutor struct {
	executor        interfaces.GetExecutor
	genClients      *GeneralClients
	configClients   ConfigurationClients
	registryClient  registry.Client
	loggingClient   logger.LoggingClient
	serviceProtocol string
}

func NewGetExecutor(executor interfaces.GetExecutor,
	genClients *GeneralClients,
	configClients ConfigurationClients,
	registryClient registry.Client,
	loggingClient logger.LoggingClient,
	serviceProtocol string) *getExecutor {

	return &getExecutor{
		executor:        executor,
		genClients:      genClients,
		configClients:   configClients,
		registryClient:  registryClient,
		loggingClient:   loggingClient,
		serviceProtocol: serviceProtocol,
	}
}

// Get provides the callout to the config service executor.
func (ge getExecutor) Get(services []string, ctx context.Context) interface{} {
	result := struct {
		Configuration map[string]interface{} `json:"configuration"`
	}{
		Configuration: map[string]interface{}{},
	}
	for _, service := range services {
		config, err := ge.executor(service, ctx)
		if err != nil {
			ge.loggingClient.Error(fmt.Sprintf(err.Error()))
			result.Configuration[service] = fmt.Sprintf(err.Error())
			continue
		}
		result.Configuration[service] = response.Process(config, ge.loggingClient)
	}
	return result
}
