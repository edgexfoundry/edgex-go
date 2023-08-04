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

	"github.com/labstack/echo/v4"
)

func EncodeAndWriteResponse(i interface{}, w *echo.Response, LoggingClient logger.LoggingClient) error {
	w.Header().Set(common.ContentType, common.ContentTypeJSON)

	enc := json.NewEncoder(w)
	err := enc.Encode(i)

	// Problems encoding
	if err != nil {
		LoggingClient.Error("Error encoding the data: " + err.Error())
		// set Response.Committed to false in order to rewrite the status code
		w.Committed = false
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func EncodeAndWriteYamlResponse(i interface{}, w *echo.Response, lc logger.LoggingClient) error {
	w.Header().Set(common.ContentType, common.ContentTypeYAML)

	enc := yaml.NewEncoder(w)
	err := enc.Encode(i)
	// Problems encoding
	if err != nil {
		lc.Error("Error encoding the data: " + err.Error())
		// set Response.Committed to false in order to rewrite the status code
		w.Committed = false
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return nil
}
