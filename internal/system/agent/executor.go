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

package agent

import (
	"errors"
	"fmt"
	"os/exec"
)

// runExec provides the common callout to the configuration-defined executor.
func runExec(serviceName string, operation string) (string, error) {
	bytes, err := exec.Command(Configuration.ExecutorPath, serviceName, operation).CombinedOutput()
	return string(bytes), err
}

// operationViaExecutor delegates a start/stop/restart operation request to the configuration-defined executor.
func operationViaExecutor(serviceName string, operation string) (string, error) {
	switch operation {
	case start:
		fallthrough
	case stop:
		fallthrough
	case restart:
		return runExec(serviceName, operation)
	default:
		return "", errors.New(fmt.Sprintf("Unsupported operation: %s", operation))
	}
}

// metricsViaExecutor delegates a metrics request to the configuration-defined executor.
func metricsViaExecutor(serviceName string) (string, error) {
	result, err := runExec(serviceName, metrics)
	if err != nil {
		return "", err
	}
	return result, nil
}
