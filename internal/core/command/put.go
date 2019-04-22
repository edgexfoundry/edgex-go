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
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"net/http"
	"strings"
)

// NewPutCommand creates and Executor which can be used to execute the PUT related command.
func NewPutCommand(device models.Device, command models.Command, body string, context context.Context, httpCaller internal.HttpCaller) (Executor, error) {
	url := device.Service.Addressable.GetBaseURL() + strings.Replace(command.Put.Action.Path, DEVICEIDURLPARAM, device.Id, -1)
	request, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		return serviceCommand{}, err
	}

	correlationID := context.Value(clients.CorrelationHeader)
	if correlationID != nil {
		request.Header.Set(clients.CorrelationHeader, correlationID.(string))
	}

	return serviceCommand{
		Device:     device,
		HttpCaller: httpCaller,
		Request:    request,
	}, nil
}
