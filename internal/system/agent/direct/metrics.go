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

package direct

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	"github.com/edgexfoundry/edgex-go/internal/system"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/concurrent"
	"github.com/edgexfoundry/edgex-go/internal/system/executor"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
)

// clientFactory defines contract for creating/retrieving a general client.
type clientFactory interface {
	New(serviceName string) (general.GeneralClient, error)
}

// metrics contains references to dependencies required to handle the metrics via direct service use case.
type metrics struct {
	clientFactory clientFactory
}

// NewMetrics is a factory function that returns an initialized metrics receiver struct.
func NewMetrics(clientFactory clientFactory) *metrics {
	return &metrics{
		clientFactory: clientFactory,
	}
}

// metricsViaDirectService calls a service's metrics endpoint directly, interprets the response, and returns a Result.
func (m *metrics) metricsViaDirectService(ctx context.Context, serviceName string) system.Result {
	client, err := m.clientFactory.New(serviceName)
	if err != nil {
		return system.Failure(serviceName, executor.Metrics, ExecutorType, err.Error())
	}

	result, err := client.FetchMetrics(ctx)
	if err != nil {
		return system.Failure(serviceName, executor.Metrics, ExecutorType, err.Error())
	}

	var s telemetry.SystemUsage
	if err := json.NewDecoder(bytes.NewBuffer([]byte(result))).Decode(&s); err != nil {
		return system.Failure(
			serviceName,
			executor.Metrics,
			ExecutorType,
			fmt.Sprintf("error decoding telemetry.SystemUsage: %s", err.Error()))
	}

	return system.MetricsSuccess(serviceName, ExecutorType, s.CpuBusyAvg, int64(s.Memory.Sys), []byte(result))
}

// Get implements the Metrics interface to obtain metrics directly from one or more services concurrently.
func (m *metrics) Get(ctx context.Context, services []string) []interface{} {
	var closures []concurrent.Closure
	for index := range services {
		closures = append(
			closures,
			func(serviceName string) concurrent.Closure {
				return func() interface{} {
					return m.metricsViaDirectService(ctx, serviceName)
				}
			}(services[index]),
		)
	}
	return concurrent.ExecuteAndAggregateResults(closures)
}
