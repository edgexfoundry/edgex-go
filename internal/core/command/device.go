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
	deviceID string,
	commandID string,
	body string,
	queryParams string,
	isPutCommand bool,
	ctx context.Context,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) (string, error) {
	d, err := deviceClient.Device(deviceID, ctx)
	if err != nil {
		return "", err
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

	return executeCommandByDevice(d, c, body, queryParams, isPutCommand, ctx, loggingClient)
}

func executeCommandByName(
	dn string,
	cn string,
	body string,
	queryParams string,
	isPutCommand bool,
	ctx context.Context,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) (string, error) {
	d, err := deviceClient.DeviceForName(dn, ctx)
	if err != nil {
		return "", err
	}

	command, err := dbClient.GetCommandByNameAndDeviceId(cn, d.Id)
	if err != nil {
		return "", nil
	}

	return executeCommandByDevice(d, command, body, queryParams, isPutCommand, ctx, loggingClient)
}

func executeCommandByDevice(
	device contract.Device,
	command contract.Command,
	body string,
	queryParams string,
	isPutCommand bool,
	ctx context.Context,
	loggingClient logger.LoggingClient) (string, error) {
	if device.AdminState == contract.Locked {
		return "", errors.NewErrDeviceLocked(device.Name)
	}

	var ex Executor
	var err error
	if isPutCommand {
		ex, err = NewPutCommand(device, command, body, ctx, &http.Client{}, loggingClient)
	} else {
		ex, err = NewGetCommand(device, command, queryParams, ctx, &http.Client{}, loggingClient)
	}

	if err != nil {
		return "", err
	}

	responseBody, responseCode, err := ex.Execute()
	if err != nil {
		return "", err
	}
	if responseCode != 200 {
		return "", types.NewErrServiceClient(responseCode, []byte(responseBody))
	}

	return responseBody, nil
}

func getAllCommands(
	ctx context.Context,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) ([]contract.CommandResponse, error) {
	configuration *config.ConfigurationStruct) (int, []contract.CommandResponse, error) {
	devices, err := deviceClient.Devices(ctx)
	if err != nil {
		return nil, err
	}

	var responses []contract.CommandResponse
	for _, d := range devices {
		cr, err := newCommandResponse(d, dbClient)
		if err != nil {
			return nil, err
		}

		responses = append(responses, cr)
	}

	return responses, err
}

func getCommandsByDeviceID(
	did string,
	ctx context.Context,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) (contract.CommandResponse, error) {
	d, err := deviceClient.Device(did, ctx)
	if err != nil {
		return contract.CommandResponse{}, nil
	}

	return newCommandResponse(d, dbClient)
}

func getCommandsByDeviceName(
	dn string,
	ctx context.Context,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) (contract.CommandResponse, error) {
	d, err := deviceClient.DeviceForName(dn, ctx)
	if err != nil {
		return contract.CommandResponse{}, nil
	}

	return newCommandResponse(d, dbClient)
}

func newCommandResponse(
	d contract.Device,
	dbClient interfaces.DBClient) (contract.CommandResponse, error) {
	commands, err := dbClient.GetCommandsByDeviceId(d.Id)
	if err != nil {
		return contract.CommandResponse{}, nil
	}

	return contract.CommandResponseFromDevice(d, commands, Configuration.Service.Url()), err
}
