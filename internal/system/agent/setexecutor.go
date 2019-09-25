package agent

import (
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	requests "github.com/edgexfoundry/go-mod-core-contracts/requests/configuration"
	responses "github.com/edgexfoundry/go-mod-core-contracts/responses/configuration"
)

type setExecutor struct {
	executor      interfaces.SetExecutor
	loggingClient logger.LoggingClient
	configuration *ConfigurationStruct
}

func NewSetExecutor(executor interfaces.SetExecutor,
	loggingClient logger.LoggingClient,
	configuration *ConfigurationStruct) *setExecutor {
	return &setExecutor{
		executor:      executor,
		loggingClient: loggingClient,
		configuration: configuration,
	}
}

// Set provides the callout to the set config service executor.
func (se setExecutor) Set(services []string, sc requests.SetConfigRequest) interface{} {
	result := struct {
		Configuration map[string]responses.SetConfigResponse `json:"configuration"`
	}{
		Configuration: map[string]responses.SetConfigResponse{},
	}

	// Loop over services and accumulate the response (i.e. "result") to return to requester.
	for _, service := range services {
		result.Configuration[service] = se.executor(service, sc)
	}
	return result
}
