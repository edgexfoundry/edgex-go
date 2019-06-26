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

const (
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
	METRICS = "metrics"
	UNKNOWN = "unknown"
	INSPECT = "inspect"

	FAILED_START_PREFIX   = "Error starting service"
	FAILED_RESTART_PREFIX = "Error restarting service"
	FAILED_STOP_PREFIX    = "Error stopping service"
)

type CommandExecutor func(arg ...string) ([]byte, error)

func Execute(operation string, service string, executor CommandExecutor) ([]byte, error) {
	switch operation {
	case START:
		return nil, executeACommand(operation, service, executor, FAILED_START_PREFIX, true)
	case RESTART:
		return nil, executeACommand(operation, service, executor, FAILED_RESTART_PREFIX, true)
	case STOP:
		return nil, executeACommand(operation, service, executor, FAILED_STOP_PREFIX, false)
	case METRICS:
		return executor(
			"stats",
			service,
			"--no-stream",
			"--format",
			"{\"cpu_perc\":\"{{ .CPUPerc }}\",\"mem_usage\":\"{{ .MemUsage }}\",\"mem_perc\":\"{{ .MemPerc }}\",\"net_io\":\"{{ .NetIO }}\",\"block_io\":\"{{ .BlockIO }}\",\"pids\":\"{{ .PIDs }}\"}")
	default:
		return nil, fmt.Errorf("operation not supported with specified executor")
	}
}

func executorCommandFailedMessage(operationPrefix string, result string, errorMessage string) string {
	return fmt.Sprintf("%s: %s (%s)", operationPrefix, errorMessage, result)
}

func executorCommandNotSupportedMessage() string {
	return fmt.Sprintf("operation not supported with specified executor")
}

func executorInspectFailedMessage(operationPrefix string, errorMessage string) string {
	return fmt.Sprintf("%s: %s", operationPrefix, errorMessage)
}

func serviceIsNotRunningButShouldBeMessage(operationPrefix string) string {
	return fmt.Sprintf("%s: service is not running but should be", operationPrefix)
}

func serviceIsRunningButShouldNotBeMessage(operationPrefix string) string {
	return fmt.Sprintf("%s: service is running but shouldn't be", operationPrefix)
}

func executeACommand(
	operation string,
	service string,
	executor CommandExecutor,
	operationPrefix string,
	shouldBeRunning bool) error {

	if result, err := executor(operation, service); err != nil {
		return errors.New(executorCommandFailedMessage(operationPrefix, string(result), err.Error()))
	}

	isRunning, err := isContainerRunning(service, executor)
	switch {
	case err != nil:
		return errors.New(executorInspectFailedMessage(operationPrefix, err.Error()))
	case isRunning != shouldBeRunning:
		if isRunning {
			return errors.New(serviceIsRunningButShouldNotBeMessage(operationPrefix))
		}
		return errors.New(serviceIsNotRunningButShouldBeMessage(operationPrefix))
	default:
		return nil
	}
}

func containerNotFoundMessage(serviceName string) string {
	return fmt.Sprintf("container %s not found", serviceName)
}

func moreThanOneContainerFoundMessage(serviceName string) string {
	return fmt.Sprintf("multiple containers found with name %s", serviceName)
}

func isContainerRunning(service string, executor CommandExecutor) (bool, error) {
	// check the status of the container using the json format - include all
	// containers as the container we want to check may be Exited
	stringOutput, err := executor(INSPECT, service)
	if err != nil {
		return false, err
	}

	var containerStatus []struct {
		State struct {
			Running bool
		}
	}
	jsonOutput := json.NewDecoder(strings.NewReader(string(stringOutput)))
	if err = jsonOutput.Decode(&containerStatus); err != nil {
		return false, err
	}

	switch {
	case len(containerStatus) < 1:
		return false, errors.New(containerNotFoundMessage(service))
	case len(containerStatus) > 1:
		return false, errors.New(moreThanOneContainerFoundMessage(service))
	default:
		return containerStatus[0].State.Running, nil
	}
}
