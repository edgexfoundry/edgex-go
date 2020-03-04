/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
	"context"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

func executeCommandByDeviceID(
	originalRequest *http.Request,
	body string,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpCaller internal.HttpCaller) (deviceServiceResponse *http.Response, theResponseBody string, failure error) {

	if originalRequest == nil {
		return nil, "", errors.NewErrExtractingInfoFromRequest()
	}

	ctx := originalRequest.Context()
	deviceID, commandID, err := extractDeviceIdAndCommandIdFromRequest(originalRequest)
	if err != nil {
		return nil, "", err
	}

	d, err := deviceClient.Device(ctx, deviceID)
	if err != nil {
		return nil, "", err
	}

	if d.AdminState == contract.Locked {
		return nil, "", errors.NewErrDeviceLocked(d.Name)
	}

	// once command service have its own persistence layer this call will be changed.
	commands, err := dbClient.GetCommandsByDeviceId(d.Id)
	if err != nil {
		return nil, "", err
	}

	var c contract.Command
	for _, command := range commands {
		if commandID == command.Id {
			c = command
			break
		}
	}

	if c.String() == (contract.Command{}).String() {
		return nil, "", errors.NewErrCommandNotAssociatedWithDevice(commandID, deviceID)
	}

	return executeCommandByDevice(ctx, d, c, body, lc, originalRequest, httpCaller)
}

// extractDeviceIdAndCommandIdFromRequest extracts deviceID and commandID from r, which
// is the HTTP request parameter, and returns the deviceID, commandID to caller, or, if not
// successfully extracted, the associated error is returned.
func extractDeviceIdAndCommandIdFromRequest(r *http.Request) (string, string, error) {
	vars := mux.Vars(r)
	deviceID := vars[ID]
	commandID := vars[COMMANDID]

	if deviceID == "" || commandID == "" {
		return deviceID, commandID, errors.NewErrExtractingInfoFromRequest()
	}

	return deviceID, commandID, nil
}

func executeCommandByName(
	originalRequest *http.Request,
	ctx context.Context,
	dn string,
	cn string,
	body string,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	httpCaller internal.HttpCaller) (deviceServiceResponse *http.Response, theResponseBody string, failure error) {

	d, err := deviceClient.DeviceForName(ctx, dn)
	if err != nil {
		return nil, "", err
	}

	if d.AdminState == contract.Locked {
		return nil, "", errors.NewErrDeviceLocked(d.Name)
	}

	command, err := dbClient.GetCommandByNameAndDeviceId(cn, d.Id)
	if err != nil {
		return nil, "", err
	}

	return executeCommandByDevice(ctx, d, command, body, lc, originalRequest, httpCaller)
}

func executeCommandByDevice(
	ctx context.Context,
	device contract.Device,
	command contract.Command,
	body string,
	lc logger.LoggingClient,
	originalRequest *http.Request,
	httpCaller internal.HttpCaller) (deviceServiceResponse *http.Response, theResponseBody string, failure error) {

	var method string
	var ex Executor
	var err error

	if originalRequest == nil {
		return nil, "", errors.NewErrParsingOriginalRequest("method")
	}

	switch originalRequest.Method {
	case http.MethodPut:
		ex, err = NewPutCommand(device, command, body, ctx, httpCaller, lc, originalRequest)
	case http.MethodGet:
		ex, err = NewGetCommand(device, command, ctx, httpCaller, lc, originalRequest)
	default:
		lc.Error(fmt.Sprintf("unknown method: %s", method))
	}

	if err != nil {
		return nil, "", err
	}

	deviceServiceResponse, err = ex.Execute()
	if err != nil {
		return nil, "", err
	}

	responseBody := new(bytes.Buffer)
	_, readErr := responseBody.ReadFrom(deviceServiceResponse.Body)
	if readErr != nil {
		return nil, "", readErr
	}

	return deviceServiceResponse, responseBody.String(), nil
}

func getAllCommands(
	ctx context.Context,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	configuration *config.ConfigurationStruct) ([]contract.CommandResponse, error) {

	devices, err := deviceClient.Devices(ctx)
	if err != nil {
		return nil, err
	}

	var responses []contract.CommandResponse
	for _, d := range devices {
		cr, err := newCommandResponse(d, dbClient, configuration)
		if err != nil {
			return nil, err
		}

		responses = append(responses, cr)
	}

	return responses, nil
}

func getCommandsByDeviceID(
	ctx context.Context,
	did string,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	configuration *config.ConfigurationStruct) (contract.CommandResponse, error) {

	d, err := deviceClient.Device(ctx, did)
	if err != nil {
		return contract.CommandResponse{}, err
	}

	return newCommandResponse(d, dbClient, configuration)
}

func getCommandsByDeviceName(
	ctx context.Context,
	dn string,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient,
	configuration *config.ConfigurationStruct) (contract.CommandResponse, error) {

	d, err := deviceClient.DeviceForName(ctx, dn)
	if err != nil {
		return contract.CommandResponse{}, err
	}

	return newCommandResponse(d, dbClient, configuration)
}

func newCommandResponse(
	d contract.Device,
	dbClient interfaces.DBClient,
	configuration *config.ConfigurationStruct) (contract.CommandResponse, error) {

	commands, err := dbClient.GetCommandsByDeviceId(d.Id)
	if err != nil {
		return contract.CommandResponse{}, err
	}

	return contract.CommandResponseFromDevice(d, commands, configuration.Service.Url()), nil
}
