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

package command

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/command/errors"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// NewGetCommand creates and Executor which can be used to execute the GET related command.
func NewGetCommand(
	device contract.Device,
	command contract.Command,
	context context.Context,
	httpCaller internal.HttpCaller,
	lc logger.LoggingClient,
	originalRequest *http.Request) (Executor, error) {

	queryParams := originalRequest.URL.RawQuery
	urlResult := device.Service.Addressable.GetBaseURL() + strings.Replace(
		command.Get.Action.Path,
		DEVICEIDURLPARAM,
		device.Id,
		-1) + "?" + queryParams
	validURL, err := url.ParseRequestURI(urlResult)
	if err != nil {
		return serviceCommand{}, err
	}
	deviceServiceProxiedRequest, err := http.NewRequest(http.MethodGet, validURL.String(), nil)
	if err != nil {
		return serviceCommand{}, err
	}

	err = addHeadersToRequest(originalRequest, deviceServiceProxiedRequest, context)
	if err != nil {
		return serviceCommand{}, errors.NewErrParsingOriginalRequest("header")
	}

	return newServiceCommand(device, httpCaller, deviceServiceProxiedRequest, lc), nil
}
