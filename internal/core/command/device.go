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
	"context"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces"
)

func executeCommandByDeviceID(
	ctx context.Context,
	deviceID string,
	commandID string,
	body string,
	queryParams string,
	isPutCommand bool,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) (string, error) {

	d, err := deviceClient.Device(ctx, deviceID)
	if err != nil {
		return "", err
	}

	if d.AdminState == contract.Locked {
		return "", errors.NewErrDeviceLocked(d.Name)
	}

	// once command service have its own persistence layer this call will be changed.
	commands, err := dbClient.GetCommandsByDeviceId(d.Id)
	if err != nil {
		return "", err
	}

	var c contract.Command
	for _, command := range commands {
		if commandID == command.Id {
			c = command
			break
		}
	}

	if c.String() == (contract.Command{}).String() {
		return "", errors.NewErrCommandNotAssociatedWithDevice(commandID, deviceID)
	}

	return executeCommandByDevice(ctx, d, c, body, queryParams, isPutCommand, lc)
}

func executeCommandByName(
	ctx context.Context,
	dn string,
	cn string,
	body string,
	queryParams string,
	isPutCommand bool,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) (string, error) {

	d, err := deviceClient.DeviceForName(ctx, dn)
	if err != nil {
		return "", err
	}

	if d.AdminState == contract.Locked {
		return "", errors.NewErrDeviceLocked(d.Name)
	}

	command, err := dbClient.GetCommandByNameAndDeviceId(cn, d.Id)
	if err != nil {
		return "", err
	}

	return executeCommandByDevice(ctx, d, command, body, queryParams, isPutCommand, lc)
}

func executeCommandByDevice(
	ctx context.Context,
	device contract.Device,
	command contract.Command,
	body string,
	queryParams string,
	isPutCommand bool,
	lc logger.LoggingClient) (string, error) {

	var ex Executor
	var err error
	if isPutCommand {
		ex, err = NewPutCommand(device, command, body, ctx, &http.Client{}, lc)
	} else {
		ex, err = NewGetCommand(device, command, queryParams, ctx, &http.Client{}, lc)
	}

	if err != nil {
		return "", err
	}

	responseBody, responseCode, err := ex.Execute()
	if err != nil {
		return "", err
	}
	if responseCode != http.StatusOK {
		return "", types.NewErrServiceClient(responseCode, []byte(responseBody))
	}

	return responseBody, nil
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
