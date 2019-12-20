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
	"bytes"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// DefaultErrorCode the default value used when an error has occurred and no response code was received.
const DefaultErrorCode = 0

// serviceCommand type which encapsulates command information to be sent to the command service.
type serviceCommand struct {
	contract.Device
	internal.HttpCaller
	*http.Request
	logger.LoggingClient
}

// Execute sends the command to the command service
func (sc serviceCommand) Execute() (string, int, error) {
	sc.LoggingClient.Debug("Issuing" + sc.Request.Method + " command to: " + sc.Request.URL.String())
	resp, reqErr := sc.HttpCaller.Do(sc.Request)
	if reqErr != nil {
		sc.LoggingClient.Error(reqErr.Error())
		return "", http.StatusInternalServerError, reqErr

	}

	buf := new(bytes.Buffer)
	_, readErr := buf.ReadFrom(resp.Body)

	if readErr != nil {
		return "", DefaultErrorCode, readErr
	}
	return buf.String(), resp.StatusCode, nil
}

func newServiceCommand(
	device contract.Device,
	caller internal.HttpCaller,
	req *http.Request,
	lc logger.LoggingClient) serviceCommand {
	return serviceCommand{
		Device:        device,
		HttpCaller:    caller,
		Request:       req,
		LoggingClient: lc,
	}
}
