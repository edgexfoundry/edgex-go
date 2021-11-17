/********************************************************************************
 *  Copyright 2020 Dell Inc.
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

package mqtt

import (
	"fmt"
)

const (
	// Different Client operations.
	PublishOperation   = "Publish"
	SubscribeOperation = "Subscribe"
	ConnectOperation   = "Connect"
)

// TimeoutErr defines an error representing operations which have not completed and surpassed the allowed wait time.
type TimeoutErr struct {
	operation string
	message   string
}

func (te TimeoutErr) Error() string {
	return fmt.Sprintf("Timeout occured while performing a '%s' operation: %s", te.operation, te.message)
}

// NewTimeoutError creates a new TimeoutErr.
func NewTimeoutError(operation string, message string) TimeoutErr {
	return TimeoutErr{
		operation: operation,
		message:   message,
	}
}

// OperationErr defines an error representing operations which have failed.
type OperationErr struct {
	operation string
	message   string
}

func (oe OperationErr) Error() string {
	return fmt.Sprintf("An error occured while performing a '%s' operation: %s", oe.operation, oe.message)
}

// NewOperationErr creates a new OperationErr
func NewOperationErr(operation string, message string) OperationErr {
	return OperationErr{
		operation: operation,
		message:   message,
	}
}
