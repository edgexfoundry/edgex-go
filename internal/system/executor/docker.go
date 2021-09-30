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
)

const (
	Start   = "start"
	Stop    = "stop"
	Restart = "restart"
	Metrics = "metrics"

	failedStartPrefix   = "Error starting service"
	failedRestartPrefix = "Error restarting service"
	failedStopPrefix    = "Error stopping service"
)

// CommandExecutor defines the function signature implemented by an underlying executor.  The executor's responsibility
// is to take a series of arguments (i.e. service name, operation, etc.) "execute" the requested operation, and return
// a result.  This abstraction was introduced to support unit testing.
type CommandExecutor func(arg ...string) ([]byte, error)

// messageExecutorOperationNotSupported returns a text error message and exists to support unit testing.
func messageExecutorOperationNotSupported() string {
	return "operation not supported by executor"
}

// messageMissingArguments returns a text error message and exists to support unit testing.
func messageMissingArguments() string {
	return "missing <service> and <operation> command line arguments"
}

// Execute is called from main (which supplies an executor) to process a request.
func Execute(args []string, executor CommandExecutor) (res interface{}, err error) {
	switch {
	case len(args) > 2:
		service := args[1]
		operation := args[2]

		switch operation {
		case Start:
			return executeACommand(operation, service, executor, failedStartPrefix, true)
		case Restart:
			return executeACommand(operation, service, executor, failedRestartPrefix, true)
		case Stop:
			return executeACommand(operation, service, executor, failedStopPrefix, false)
		case Metrics:
			return gatherMetrics(service, executor)
		default:
			return res, errors.New(messageExecutorOperationNotSupported())
		}
	default:
		return res, errors.New(messageMissingArguments())
	}
}
