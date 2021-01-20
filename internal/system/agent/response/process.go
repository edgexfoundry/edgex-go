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

package response

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

// Process converts a response string (assumed to contain JSON) to a map.
func Process(response string, lc logger.LoggingClient) map[string]interface{} {
	rsp := make(map[string]interface{})
	err := json.Unmarshal([]byte(response), &rsp)
	if err != nil {
		lc.Error("error unmarshalling response from JSON: %v", err.Error())
	}
	return rsp
}
