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

func commandByDeviceID(
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

	if d.AdminState == contract.Locked {
		return "", errors.NewErrDeviceLocked(d.Name)
	}

	// once command service have its own persistence layer this call will be changed.
	commands, err := dbClient.GetCommandsByDeviceId(d.Id)
	if err != nil {
		return "", err
	}

	var c contract.Command
	for _, one := range commands {
		if commandID == one.Id {
			c = one
			break
		}
	}

	if c.String() == (contract.Command{}).String() {
		return "", errors.NewErrCommandNotAssociatedWithDevice(commandID, deviceID)
	}

	return commandByDevice(d, c, body, queryParams, isPutCommand, ctx, loggingClient)
}

func commandByNames(
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

	if d.AdminState == contract.Locked {
		return "", errors.NewErrDeviceLocked(d.Name)
	}

	command, err := dbClient.GetCommandByNameAndDeviceId(cn, d.Id)
	if err != nil {
		return "", nil
	}

	return commandByDevice(d, command, body, queryParams, isPutCommand, ctx, loggingClient)
}

func commandByDevice(
	device contract.Device,
	command contract.Command,
	body string,
	queryParams string,
	isPutCommand bool,
	ctx context.Context,
	loggingClient logger.LoggingClient) (string, error) {
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

func getCommands(
	ctx context.Context,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) ([]contract.CommandResponse, error) {
	configuration *config.ConfigurationStruct) (int, []contract.CommandResponse, error) {
	devices, err := deviceClient.Devices(ctx)
	if err != nil {
		return nil, err
	}
	cr := []contract.CommandResponse{}
	for _, d := range devices {
		commands, err := dbClient.GetCommandsByDeviceId(d.Id)
		if err != nil {
			return nil, err
		}
		cr = append(cr, contract.CommandResponseFromDevice(d, commands, configuration.Service.Url()))
	}

	return cr, err
}

func getCommandsByDeviceID(
	did string,
	ctx context.Context,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) (contract.CommandResponse, error) {
	d, err := deviceClient.Device(did, ctx)
	if err != nil {
		return contract.CommandResponse{}, nil
	}

	commands, err := dbClient.GetCommandsByDeviceId(d.Id)
	if err != nil {
		return contract.CommandResponse{}, nil
	}

	return contract.CommandResponseFromDevice(d, commands, Configuration.Service.Url()), err
}

func getCommandsByDeviceName(
	dn string,
	ctx context.Context,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	deviceClient metadata.DeviceClient) (contract.CommandResponse, error) {
	d, err := deviceClient.DeviceForName(dn, ctx)
	if err != nil {
		return contract.CommandResponse{}, nil
	}

	commands, err := dbClient.GetCommandsByDeviceId(d.Id)
	if err != nil {
		return contract.CommandResponse{}, nil
	}

	return contract.CommandResponseFromDevice(d, commands, Configuration.Service.Url()), err
}
