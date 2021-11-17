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

package container

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// LoggingClientInterfaceName contains the name of the logger.LoggingClient implementation in the DIC.
var LoggingClientInterfaceName = di.TypeInstanceToName((*logger.LoggingClient)(nil))

// LoggingClientFrom helper function queries the DIC and returns the logger.loggingClient implementation.
func LoggingClientFrom(get di.Get) logger.LoggingClient {
	loggingClient, ok := get(LoggingClientInterfaceName).(logger.LoggingClient)
	if !ok {
		return nil
	}

	return loggingClient
}
