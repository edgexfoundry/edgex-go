/*******************************************************************************
 * Copyright (C) 2019 IOTech Ltd
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
package pkg

import (
	"encoding/json"
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"net/http"
)

// Report EdgeX version
func VersionHandler(w http.ResponseWriter, _ *http.Request) {
	res := struct {
		Version string `json:"version"`
	}{edgex.Version}
	w.Header().Add(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&res)
}
