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
	"errors"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/system"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/response"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
)

func TestMetricsGetWithNoServices(t *testing.T) {
	executor := NewStub(map[string]stubCall{})
	sut := NewMetrics(executor.CommandExecutor, logger.NewMockClient(), "executorPathDoesNotMatter")

	result := sut.Get(context.Background(), []string{})

	var emptyResult []interface{}
	assert.Equal(t, result, emptyResult)
	assert.Equal(t, executor.Called, 0)
}

func TestMetricsGetWithServices(t *testing.T) {
	const (
		service1Name   = "service1Name"
		service2Name   = "service2Name"
		executorPath   = "executorPath"
		service1Result = "[{\"result\":\"foo\"}]"
		service2Result = "[{\"result\":\"bar\"}]"
	)

	lc := logger.NewMockClient()
	expectedError := errors.New("expectedError")
	tests := []struct {
		name           string
		services       []string
		expectedResult []interface{}
		executorCalls  map[string]stubCall
	}{
		{
			"one service with no error",
			[]string{service1Name},
			[]interface{}{response.Process(service1Result, lc)},
			map[string]stubCall{
				service1Name: {[]string{executorPath, service1Name, system.Metrics}, service1Result, nil},
			},
		},
		{
			"one service with error",
			[]string{service1Name},
			[]interface{}{system.Failure(service1Name, system.Metrics, UnknownExecutorType, expectedError.Error())},
			map[string]stubCall{
				service1Name: {[]string{executorPath, service1Name, system.Metrics}, "", expectedError},
			},
		},
		{
			"two services with no errors",
			[]string{service1Name, service2Name},
			[]interface{}{
				response.Process(service1Result, lc),
				response.Process(service2Result, lc),
			},
			map[string]stubCall{
				service1Name: {[]string{executorPath, service1Name, system.Metrics}, service1Result, nil},
				service2Name: {[]string{executorPath, service2Name, system.Metrics}, service2Result, nil},
			},
		},
		{
			"two services with first returning error",
			[]string{service1Name, service2Name},
			[]interface{}{
				system.Failure(service1Name, system.Metrics, UnknownExecutorType, expectedError.Error()),
				response.Process(service2Result, lc),
			},
			map[string]stubCall{
				service1Name: {[]string{executorPath, service1Name, system.Metrics}, "", expectedError},
				service2Name: {[]string{executorPath, service2Name, system.Metrics}, service2Result, nil},
			},
		},
		{
			"two services with second returning error",
			[]string{service1Name, service2Name},
			[]interface{}{
				response.Process(service1Result, lc),
				system.Failure(service2Name, system.Metrics, UnknownExecutorType, expectedError.Error()),
			},
			map[string]stubCall{
				service1Name: {[]string{executorPath, service1Name, system.Metrics}, service1Result, nil},
				service2Name: {[]string{executorPath, service2Name, system.Metrics}, "", expectedError},
			},
		},
		{
			"two services with both returning errors",
			[]string{service1Name, service2Name},
			[]interface{}{
				system.Failure(service1Name, system.Metrics, UnknownExecutorType, expectedError.Error()),
				system.Failure(service2Name, system.Metrics, UnknownExecutorType, expectedError.Error()),
			},
			map[string]stubCall{
				service1Name: {[]string{executorPath, service1Name, system.Metrics}, "", expectedError},
				service2Name: {[]string{executorPath, service2Name, system.Metrics}, "", expectedError},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := NewStub(test.executorCalls)
			sut := NewMetrics(executor.CommandExecutor, lc, executorPath)

			result := sut.Get(context.Background(), test.services)

			if assert.Equal(t, len(test.executorCalls), executor.Called) {
				var expectedArgs []string
				for key := range test.executorCalls {
					expectedArgs = append(expectedArgs, argsToString(test.executorCalls[key].expectedArgs))
				}

				var capturedArgs []string
				for key := range executor.capturedArgs {
					capturedArgs = append(capturedArgs, argsToString(executor.capturedArgs[key]))
				}

				assertArgsAreEqualInAnyOrder(t, expectedArgs, capturedArgs)
			}

			assertResultsAreEqualInAnyOrder(t, test.expectedResult, result)
		})
	}
}
