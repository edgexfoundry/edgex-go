/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package handle

import (
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

// httpJSONResult marshals an arbitrary struct to JSON, sets the appropriate HTTP header and status code, and writes the
// converted JSON content.  If there is an marshaling error, it returns an http.StatusBadRequest result.
// Modified version of /internal/pkg/encoding.go to provide implementation capable of setting both content-type and
// non-200 status code.
func httpJSONResult(w http.ResponseWriter, statusCode int, i interface{}) {
	encoded, err := json.Marshal(i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(statusCode)
	_, _ = w.Write(encoded)
}
