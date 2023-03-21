/*******************************************************************************
 * Copyright 2019 VMware Inc.
 * Copyright (C) 2023 IOTech Ltd
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
	"gopkg.in/yaml.v3"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
)

func EncodeAndWriteResponse(i interface{}, w http.ResponseWriter, LoggingClient logger.LoggingClient) {
	w.Header().Set(common.ContentType, common.ContentTypeJSON)

	enc := json.NewEncoder(w)
	err := enc.Encode(i)
	// Problems encoding
	if err != nil {
		LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func EncodeAndWriteYamlResponse(i interface{}, w http.ResponseWriter, lc logger.LoggingClient) {
	w.Header().Set(common.ContentType, common.ContentTypeYAML)

	enc := yaml.NewEncoder(w)
	err := enc.Encode(i)
	// Problems encoding
	if err != nil {
		lc.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
