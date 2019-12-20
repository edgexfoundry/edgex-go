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
	"context"

	"github.com/edgexfoundry/edgex-go/internal/system"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/concurrent"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/response"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// metrics contains references to dependencies required to handle the metrics via external executor use case.
type metrics struct {
	executor      interfaces.CommandExecutor
	loggingClient logger.LoggingClient
	executorPath  string
}

// NewMetrics is a factory function that returns an initialized metrics receiver struct.
func NewMetrics(executor interfaces.CommandExecutor, lc logger.LoggingClient, executorPath string) *metrics {
	return &metrics{
		executor:      executor,
		loggingClient: lc,
		executorPath:  executorPath,
	}
}

// delegateToExecutor wraps executor execution and handles error response creation when necessary.
func (e metrics) delegateToExecutor(serviceName string) interface{} {
	r, err := e.executor(e.executorPath, serviceName, system.Metrics)
	if err != nil {
		return system.Failure(serviceName, system.Metrics, UnknownExecutorType, err.Error())
	}
	return response.Process(r, e.loggingClient)
}

// Get implements the Metrics interface to obtain metrics via executor for one or more services concurrently.
func (e metrics) Get(_ context.Context, services []string) []interface{} {
	var closures []concurrent.Closure
	for index := range services {
		closures = append(
			closures,
			func(serviceName string) concurrent.Closure {
				return func() interface{} {
					return e.delegateToExecutor(serviceName)
				}
			}(services[index]),
		)
	}
	return concurrent.ExecuteAndAggregateResults(closures)
}
