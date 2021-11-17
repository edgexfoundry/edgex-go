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

package redis

import (
	"fmt"
	"strings"
)

// DisconnectErr represents errors which occur when attempting to disconnect from a Redis server.
type DisconnectErr struct {
	// disconnectErrors contains the descriptive error messages that occur while attempting to disconnect one or more
	// underlying clients.
	disconnectErrors []string
}

// Error constructs an appropriate error message based on the error descriptions provided.
func (d DisconnectErr) Error() string {
	return fmt.Sprintf("Unable to disconnect client(s): %s", strings.Join(d.disconnectErrors, ","))
}

// NewDisconnectErr created a new DisconnectErr
func NewDisconnectErr(disconnectErrors []string) DisconnectErr {
	return DisconnectErr{
		disconnectErrors: disconnectErrors,
	}
}
