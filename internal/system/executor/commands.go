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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const inspect = "inspect"

// messageExecutorInspectFailed returns a text error message and exists to support unit testing.
func messageExecutorInspectFailed(operationPrefix string, errorMessage string) string {
	return fmt.Sprintf("%s: %s", operationPrefix, errorMessage)
}

// messageServiceIsNotRunningButShouldBe returns a text error message and exists to support unit testing.
func messageServiceIsNotRunningButShouldBe(operationPrefix string) string {
	return fmt.Sprintf("%s: service is not running but should be", operationPrefix)
}

// messageServiceIsRunningButShouldNotBe returns a text error message and exists to support unit testing.
func messageServiceIsRunningButShouldNotBe(operationPrefix string) string {
	return fmt.Sprintf("%s: service is running but shouldn't be", operationPrefix)
}

// messageContainerNotFound returns a text error message and exists to support unit testing.
func messageContainerNotFound(serviceName string) string {
	return fmt.Sprintf("container %s not found", serviceName)
}

// messageMoreThanOneContainerFound returns a text error message and exists to support unit testing.
func messageMoreThanOneContainerFound(serviceName string) string {
	return fmt.Sprintf("multiple containers found with name %s", serviceName)
}

// isContainerRunning delegates service status inspection to the executor and interprets the result to determine and
// return whether a specific service has one and only one running instance.
func isContainerRunning(service string, executor CommandExecutor) (bool, string) {
	// check the status of the container using the json format - include all
	// containers as the container we want to check may be Exited
	output, err := executor(inspect, service)
	if err != nil {
		return false, fmt.Sprintf("%s: %s", err.Error(), output)
	}

	var containerStatus []struct {
		State struct {
			Running bool
		}
	}
	jsonOutput := json.NewDecoder(strings.NewReader(string(output)))
	if err = jsonOutput.Decode(&containerStatus); err != nil {
		return false, err.Error()
	}

	switch {
	case len(containerStatus) < 1:
		return false, messageContainerNotFound(service)
	case len(containerStatus) > 1:
		return false, messageMoreThanOneContainerFound(service)
	default:
		return containerStatus[0].State.Running, ""
	}
}

// executeACommand handles start/stop/restart operation requests by delegating the command to the executor,
// subsequently verifying the state of the service's container is as expected, and returning an appropriate Result
// value.
func executeACommand(
	operation string,
	service string,
	executor CommandExecutor,
	operationPrefix string,
	shouldBeRunning bool) (string, error) {

	if output, err := executor(operation, service); err != nil {
		return "", fmt.Errorf("%s: %s", err.Error(), output)
	}

	isRunning, errorMessage := isContainerRunning(service, executor)
	switch {
	case len(errorMessage) > 0:
		return "", errors.New(messageExecutorInspectFailed(operationPrefix, errorMessage))
	case isRunning != shouldBeRunning:
		if isRunning {
			return "", errors.New(messageServiceIsRunningButShouldNotBe(operationPrefix))
		}
		return "", errors.New(messageServiceIsNotRunningButShouldBe(operationPrefix))
	default:
		return "", nil
	}
}
