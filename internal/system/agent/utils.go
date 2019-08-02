/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package agent

import (
	"encoding/json"
	"net/http"
)

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func ProcessResponse(response string) map[string]interface{} {
	rsp := make(map[string]interface{})
	err := json.Unmarshal([]byte(response), &rsp)
	if err != nil {
		LoggingClient.Error("error unmarshalling response from JSON: %v", err.Error())
	}
	return rsp
}

func ProcessFetchedResponse(response []byte) map[string]interface{} {
	rsp := make(map[string]interface{})
	err := json.Unmarshal(response, &rsp)
	if err != nil {
		LoggingClient.Error("error unmarshalling response from JSON: %v", err.Error())
	}
	return rsp
}
