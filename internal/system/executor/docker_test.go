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
	"os/exec"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/stretchr/testify/assert"
)

const (
	serviceName  = "serviceName"
	errorMessage = "errorMessage"
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

func TestExecute(t *testing.T) {
	tests := []struct {
		name                  string
		operation             string
		expectedResult        []byte
		expectedErrorMessage  string
		expectedExecutorCalls int
		executorCalls         []executorStubCall
	}{
		{
			"Start: first call to executor fails",
			START,
			[]byte(nil),
			executorCommandFailedMessage(FAILED_START_PREFIX, string([]byte(nil)), errorMessage),
			1,
			[]executorStubCall{
				{[]string{START, serviceName}, []byte(nil), errors.New(errorMessage)},
			},
		},
		{
			"Start: second call to executor fails",
			START,
			[]byte(nil),
			executorInspectFailedMessage(FAILED_START_PREFIX, errorMessage),
			2,
			[]executorStubCall{
				{[]string{START, serviceName}, []byte(nil), nil},
				{[]string{INSPECT, serviceName}, []byte(nil), errors.New(errorMessage)},
			},
		},
		{
			"Start: container not found in inspect result",
			START,
			[]byte(nil),
			executorInspectFailedMessage(FAILED_START_PREFIX, containerNotFoundMessage(serviceName)),
			2,
			[]executorStubCall{
				{[]string{START, serviceName}, []byte(nil), nil},
				{[]string{INSPECT, serviceName}, []byte("[]"), nil},
			},
		},
		{
			"Start: more than one container instance found in inspect result",
			START,
			[]byte(nil),
			executorInspectFailedMessage(FAILED_START_PREFIX, moreThanOneContainerFoundMessage(serviceName)),
			2,
			[]executorStubCall{
				{[]string{START, serviceName}, []byte(nil), nil},
				{[]string{INSPECT, serviceName}, []byte("[{\"State\": {\"Running\": false}}, {\"State\": {\"Running\": false}}]"), nil},
			},
		},
		{
			"Start: inspect result says service is not running as expected",
			START,
			[]byte(nil),
			serviceIsNotRunningButShouldBeMessage(FAILED_START_PREFIX),
			2,
			[]executorStubCall{
				{[]string{START, serviceName}, []byte(nil), nil},
				{[]string{INSPECT, serviceName}, []byte("[{\"State\": {\"Running\": false}}]"), nil},
			},
		},
		{
			"Restart: first call to executor fails",
			RESTART,
			[]byte(nil),
			executorCommandFailedMessage(FAILED_RESTART_PREFIX, string([]byte(nil)), errorMessage),
			1,
			[]executorStubCall{
				{[]string{RESTART, serviceName}, []byte(nil), errors.New(errorMessage)},
			},
		},
		{
			"Stop: first call to executor fails",
			STOP,
			[]byte(nil),
			executorCommandFailedMessage(FAILED_STOP_PREFIX, string([]byte(nil)), errorMessage),
			1,
			[]executorStubCall{
				{[]string{STOP, serviceName}, []byte(nil), errors.New(errorMessage)},
			},
		},
		{
			"Stop: service is running but shouldn't be",
			STOP,
			[]byte(nil),
			serviceIsRunningButShouldNotBeMessage(FAILED_STOP_PREFIX),
			2,
			[]executorStubCall{
				{[]string{STOP, serviceName}, []byte(nil), nil},
				{[]string{INSPECT, serviceName}, []byte("[{\"State\": {\"Running\": true}}]"), nil},
			},
		}, {
			"Start: no errors encountered in checking for container status",
			STOP,
			[]byte(nil),
			serviceIsRunningButShouldNotBeMessage(FAILED_STOP_PREFIX),
			2,
			[]executorStubCall{
				{[]string{STOP, serviceName}, []byte(nil), nil},
				{[]string{INSPECT, serviceName}, []byte("[{\"State\": {\"Running\": true}}]"), nil},
			},
		},
		{
			"Start: no errors encountered in checking for container status",
			START,
			nil,
			executorInspectFailedMessage(FAILED_START_PREFIX, "EOF"),
			2,
			[]executorStubCall{
				{[]string{START, serviceName}, []byte(nil), nil},
				{[]string{INSPECT, serviceName}, []byte(nil), nil},
			},
		},
		{
			"operation not supported with specified executor",
			UNKNOWN,
			[]byte(nil),
			executorCommandNotSupportedMessage(),
			0,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := newExecutor(test.executorCalls)

			result, err := Execute(test.operation, serviceName, executor.commandExecutor)

			assert.Equal(t, test.expectedExecutorCalls, executor.Called)
			for key, executorCall := range test.executorCalls {
				assertArgsAreEqual(t, executorCall.expectedArgs, executor.capturedArgs[key])
			}

			assert.Equal(t, test.expectedResult, result)
			assert.NotNil(t, err)
			assert.Equal(t, test.expectedErrorMessage, err.Error())
		})
	}
}

func TestExecuteMetrics(t *testing.T) {

	result, _ := Execute(METRICS, clients.CoreDataServiceKey, func(arg ...string) ([]byte, error) {
		return exec.Command("docker", "stats", clients.CoreDataServiceKey, "--no-stream", "--format", "{\"cpu_perc\":\"{{ .CPUPerc }}\",\"mem_usage\":\"{{ .MemUsage }}\",\"mem_perc\":\"{{ .MemPerc }}\",\"net_io\":\"{{ .NetIO }}\",\"block_io\":\"{{ .BlockIO }}\",\"pids\":\"{{ .PIDs }}\"}").CombinedOutput()
	})

	var s string
	s = string(result)
	assert.NotNil(t, s)

	//	 TODO: Future direction will be to mock out the call to Docker stats.
	// Validate that all expected keys are present in the "JSON-like" response
	/*
		kvp := agent.ProcessResponse(s)

		assert.NotNil(t, kvp["cpu_perc"])
		assert.NotNil(t, kvp["mem_usage"])
		assert.NotNil(t, kvp["mem_perc"])
		assert.NotNil(t, kvp["net_io"])
		assert.NotNil(t, kvp["block_io"])
		assert.NotNil(t, kvp["pids"])
	*/
}
