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
	"errors"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/system"
	"github.com/stretchr/testify/assert"
)

const (
	serviceName                   = "serviceName"
	executableName                = "executableName"
	errorMessage                  = "errorMessage"
	invalidOperation              = "invalidOperation"
	metricsSuccessRawResult       = "metricsSuccessRawResult"
	jsonDecodeFailureErrorMessage = "EOF"
)

type executorStubCall struct {
	expectedArgs []string // expected arg value for specific executor call
	outBytes     []byte   // return value for specific executor call
	outError     error    // return value for specific executor call
}

type executorStub struct {
	Called         int                // number of times stub is called
	capturedArgs   [][]string         // captures arg values for each stub call
	perCallResults []executorStubCall // expected arg value and return values for each stub call
}

func newExecutor(results []executorStubCall) executorStub {
	return executorStub{
		perCallResults: results,
	}
}

func (e *executorStub) commandExecutor(arg ...string) ([]byte, error) {
	e.Called++
	e.capturedArgs = append(e.capturedArgs, arg)
	return e.perCallResults[e.Called-1].outBytes, e.perCallResults[e.Called-1].outError
}

func assertArgsAreEqual(t *testing.T, expected []string, actual []string) {
	assert.Equal(t, len(expected), len(actual))
	for key, expectedValue := range expected {
		assert.Equal(t, expectedValue, actual[key])
	}
}

func firstCommandCallFails(serviceName string, operation string) []executorStubCall {
	return []executorStubCall{
		{[]string{operation, serviceName}, []byte(nil), errors.New(errorMessage)},
	}
}

func secondCommandCallFails(serviceName string, operation string) []executorStubCall {
	return []executorStubCall{
		{[]string{operation, serviceName}, []byte(nil), nil},
		{[]string{inspect, serviceName}, []byte(nil), errors.New(errorMessage)},
	}
}

func secondCommandCallSucceeds(serviceName string, operation string, result string) []executorStubCall {
	return []executorStubCall{
		{[]string{operation, serviceName}, []byte(nil), nil},
		{[]string{inspect, serviceName}, []byte(result), nil},
	}
}

func firstMetricsCallFails(serviceName string) []executorStubCall {
	return []executorStubCall{
		{metricsExecutorCommands(serviceName), []byte(nil), errors.New(errorMessage)},
	}
}

func firstMetricsCallSucceeds(serviceName string, result string) []executorStubCall {
	return []executorStubCall{
		{metricsExecutorCommands(serviceName), []byte(result), nil},
	}
}

func executeArguments(serviceName string, operation string) []string {
	return []string{executableName, serviceName, operation}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name           string
		operation      string
		expectedResult system.Result
		executorCalls  []executorStubCall
	}{
		// start command test cases

		{
			"Start: first executor call fails",
			Start,
			system.Failure(serviceName, Start, executorType, messageExecutorCommandFailed(failedStartPrefix, string([]byte(nil)), errorMessage)),
			firstCommandCallFails(serviceName, Start),
		},
		{
			"Start: second executor call fails",
			Start,
			system.Failure(serviceName, Start, executorType, messageExecutorInspectFailed(failedStartPrefix, errorMessage)),
			secondCommandCallFails(serviceName, Start),
		},
		{
			"Start: container not found in inspect result",
			Start,
			system.Failure(serviceName, Start, executorType, messageExecutorInspectFailed(failedStartPrefix, messageContainerNotFound(serviceName))),
			secondCommandCallSucceeds(serviceName, Start, "[]"),
		},
		{
			"Start: more than one container instance found in inspect result",
			Start,
			system.Failure(serviceName, Start, executorType, messageExecutorInspectFailed(failedStartPrefix, messageMoreThanOneContainerFound(serviceName))),
			secondCommandCallSucceeds(serviceName, Start, "[{\"State\": {\"Running\": false}}, {\"State\": {\"Running\": false}}]"),
		},
		{
			"Start: inspect result says service is not running as expected",
			Start,
			system.Failure(serviceName, Start, executorType, messageServiceIsNotRunningButShouldBe(failedStartPrefix)),
			secondCommandCallSucceeds(serviceName, Start, "[{\"State\": {\"Running\": false}}]"),
		},
		{
			"Start: isContainerRunning json.Decode Failure",
			Start,
			system.Failure(serviceName, Start, executorType, messageExecutorInspectFailed(failedStartPrefix, jsonDecodeFailureErrorMessage)),
			secondCommandCallSucceeds(serviceName, Start, ""),
		},
		{
			"Start: Success",
			Start,
			system.Success(serviceName, Start, executorType),
			secondCommandCallSucceeds(serviceName, Start, "[{\"State\": {\"Running\": true}}]"),
		},

		// Restart command test cases

		{
			"Restart: first executor call fails",
			Restart,
			system.Failure(serviceName, Restart, executorType, messageExecutorCommandFailed(failedRestartPrefix, string([]byte(nil)), errorMessage)),
			firstCommandCallFails(serviceName, Restart),
		},
		{
			"Restart: second executor call fails",
			Restart,
			system.Failure(serviceName, Restart, executorType, messageExecutorInspectFailed(failedRestartPrefix, errorMessage)),
			secondCommandCallFails(serviceName, Restart),
		},
		{
			"Restart: container not found in inspect result",
			Restart,
			system.Failure(serviceName, Restart, executorType, messageExecutorInspectFailed(failedRestartPrefix, messageContainerNotFound(serviceName))),
			secondCommandCallSucceeds(serviceName, Restart, "[]"),
		},
		{
			"Restart: more than one container instance found in inspect result",
			Restart,
			system.Failure(serviceName, Restart, executorType, messageExecutorInspectFailed(failedRestartPrefix, messageMoreThanOneContainerFound(serviceName))),
			secondCommandCallSucceeds(serviceName, Restart, "[{\"State\": {\"Running\": false}}, {\"State\": {\"Running\": false}}]"),
		},
		{
			"Restart: inspect result says service is not running as expected",
			Restart,
			system.Failure(serviceName, Restart, executorType, messageServiceIsNotRunningButShouldBe(failedRestartPrefix)),
			secondCommandCallSucceeds(serviceName, Restart, "[{\"State\": {\"Running\": false}}]"),
		},
		{
			"Restart: isContainerRunning json.Decode Failure",
			Restart,
			system.Failure(serviceName, Restart, executorType, messageExecutorInspectFailed(failedRestartPrefix, jsonDecodeFailureErrorMessage)),
			secondCommandCallSucceeds(serviceName, Restart, ""),
		},
		{
			"Restart: Success",
			Restart,
			system.Success(serviceName, Restart, executorType),
			secondCommandCallSucceeds(serviceName, Restart, "[{\"State\": {\"Running\": true}}]"),
		},

		// stop command test cases

		{
			"Stop: first executor call fails",
			Stop,
			system.Failure(serviceName, Stop, executorType, messageExecutorCommandFailed(failedStopPrefix, string([]byte(nil)), errorMessage)),
			firstCommandCallFails(serviceName, Stop),
		},
		{
			"Stop: second executor call fails",
			Stop,
			system.Failure(serviceName, Stop, executorType, messageExecutorInspectFailed(failedStopPrefix, errorMessage)),
			secondCommandCallFails(serviceName, Stop),
		},
		{
			"Stop: container not found in inspect result",
			Stop,
			system.Failure(serviceName, Stop, executorType, messageExecutorInspectFailed(failedStopPrefix, messageContainerNotFound(serviceName))),
			secondCommandCallSucceeds(serviceName, Stop, "[]"),
		},
		{
			"Stop: more than one container instance found in inspect result",
			Stop,
			system.Failure(serviceName, Stop, executorType, messageExecutorInspectFailed(failedStopPrefix, messageMoreThanOneContainerFound(serviceName))),
			secondCommandCallSucceeds(serviceName, Stop, "[{\"State\": {\"Running\": true}}, {\"State\": {\"Running\": true}}]"),
		},
		{
			"Stop: inspect result says service is not running as expected",
			Stop,
			system.Failure(serviceName, Stop, executorType, messageServiceIsRunningButShouldNotBe(failedStopPrefix)),
			secondCommandCallSucceeds(serviceName, Stop, "[{\"State\": {\"Running\": true}}]"),
		},
		{
			"Stop: isContainerRunning json.Decode Failure",
			Stop,
			system.Failure(serviceName, Stop, executorType, messageExecutorInspectFailed(failedStopPrefix, jsonDecodeFailureErrorMessage)),
			secondCommandCallSucceeds(serviceName, Stop, ""),
		},
		{
			"Stop: Success",
			Stop,
			system.Success(serviceName, Stop, executorType),
			secondCommandCallSucceeds(serviceName, Stop, "[{\"State\": {\"Running\": false}}]"),
		},

		// metrics command test case

		{
			"MetricsViaExecutor: Failure",
			Metrics,
			system.Failure(serviceName, Metrics, executorType, errorMessage),
			firstMetricsCallFails(serviceName),
		},
		{
			"MetricsViaExecutor: Success (missing memory scale)",
			Metrics,
			system.MetricsSuccess(serviceName, executorType, 1.49, -1, []byte(metricsSuccessRawResult)),
			firstMetricsCallSucceeds(serviceName, "1.49%"+separator+"1234 / 7.786GiB"+separator+metricsSuccessRawResult),
		},
		{
			"MetricsViaExecutor: Success (kb)",
			Metrics,
			system.MetricsSuccess(serviceName, executorType, 1.49, 1264, []byte(metricsSuccessRawResult)),
			firstMetricsCallSucceeds(serviceName, "1.49%"+separator+"1.234KiB / 7.786GiB"+separator+metricsSuccessRawResult),
		},
		{
			"MetricsViaExecutor: Success (mb)",
			Metrics,
			system.MetricsSuccess(serviceName, executorType, 1.49, 1293943, []byte(metricsSuccessRawResult)),
			firstMetricsCallSucceeds(serviceName, "1.49%"+separator+"1.234MiB / 7.786GiB"+separator+metricsSuccessRawResult),
		},
		{
			"MetricsViaExecutor: Success (gb)",
			Metrics,
			system.MetricsSuccess(serviceName, executorType, 1.49, 1324997411, []byte(metricsSuccessRawResult)),
			firstMetricsCallSucceeds(serviceName, "1.49%"+separator+"1.234GiB / 7.786GiB"+separator+metricsSuccessRawResult),
		},
		{
			"MetricsViaExecutor: Success (missing cpu float value)",
			Metrics,
			system.MetricsSuccess(serviceName, executorType, -1.0, 1264, []byte(metricsSuccessRawResult)),
			firstMetricsCallSucceeds(serviceName, "badValue"+separator+"1.234KiB / 7.786GiB"+separator+metricsSuccessRawResult),
		},

		// invalid operation test case

		{
			"operation not supported by executor",
			invalidOperation,
			system.Failure(serviceName, invalidOperation, executorType, messageExecutorOperationNotSupported()),
			[]executorStubCall{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := newExecutor(test.executorCalls)

			result := Execute(executeArguments(serviceName, test.operation), executor.commandExecutor)

			if assert.Equal(t, len(test.executorCalls), executor.Called) {
				for key, executorCall := range test.executorCalls {
					assertArgsAreEqual(t, executorCall.expectedArgs, executor.capturedArgs[key])
				}
			}
			assert.Equal(t, test.expectedResult, result)
		})
	}
}

func TestMissingArguments(t *testing.T) {
	missingArguments := []string{executableName}
	executor := newExecutor([]executorStubCall{})

	result := Execute(missingArguments, executor.commandExecutor)

	assert.Equal(t, 0, executor.Called)
	assert.Equal(t, system.Failure("", "", executorType, messageMissingArguments()), result)
}
