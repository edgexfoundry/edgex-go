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
	"github.com/edgexfoundry/edgex-go/internal/system/agent/concurrent"
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

// delegateToExecutor wraps executor execution and handles error response creation when necessary.
func (e operations) delegateToExecutor(serviceName, operation string) interface{} {
	r, err := e.executor(e.executorPath, serviceName, operation)
	if err != nil {
		return system.Failure(serviceName, operation, UnknownExecutorType, err.Error())
	}
	return response.Process(r, e.loggingClient)
}

// Do concurrently delegates a start/stop/restart operation request to the configuration-defined executor.
func (e operations) Do(services []string, operation string) []interface{} {
	var closures []concurrent.Closure
	for index := range services {
		closures = append(
			closures,
			func(serviceName string) concurrent.Closure {
				return func() interface{} {
					return e.delegateToExecutor(serviceName, operation)
				}
			}(services[index]),
		)
	}
	return concurrent.ExecuteAndAggregateResults(closures)
}
