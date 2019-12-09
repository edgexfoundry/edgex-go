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

package executor

import (
	"github.com/edgexfoundry/edgex-go/internal/system"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/response"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// operations contains references to dependencies required to handle an operation via executor use case.
type operations struct {
	executor      interfaces.CommandExecutor
	loggingClient logger.LoggingClient
	executorPath  string
}

// NewOperations is a factory function that returns an initialized operations receiver struct.
func NewOperations(
	executor interfaces.CommandExecutor,
	loggingClient logger.LoggingClient,
	executorPath string) *operations {

	return &operations{
		executor:      executor,
		loggingClient: loggingClient,
		executorPath:  executorPath,
	}
}

// operationViaExecutor delegates a start/stop/restart operation request to the configuration-defined executor.
func (e operations) Do(services []string, operation string) []interface{} {
	var result []interface{}
	for _, serviceName := range services {
		r, err := e.executor(e.executorPath, serviceName, operation)
		if err != nil {
			result = append(result, system.Failure(serviceName, operation, UnknownExecutorType, err.Error()))
			continue
		}
		result = append(result, response.Process(r, e.loggingClient))
	}
	return result
}
