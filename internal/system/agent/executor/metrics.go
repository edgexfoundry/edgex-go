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
func NewMetrics(executor interfaces.CommandExecutor, loggingClient logger.LoggingClient, executorPath string) *metrics {
	return &metrics{
		executor:      executor,
		loggingClient: loggingClient,
		executorPath:  executorPath,
	}
}

// Get delegates a metrics request to the configuration-defined executor.
func (e metrics) Get(services []string, ctx context.Context) interface{} {
	var result []interface{}
	for _, serviceName := range services {
		r, err := e.executor(e.executorPath, serviceName, system.Metrics, []string{})
		if err != nil {
			result = append(result, system.Failure(serviceName, system.Metrics, UnknownExecutorType, err.Error()))
			continue
		}
		result = append(result, response.Process(r, e.loggingClient))
	}
	return result
}
