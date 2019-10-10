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
	goErrors "errors"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/command/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func commandByDeviceID(deviceID string, commandID string, body string, queryParams string, isPutCommand bool,
	ctx context.Context) (string, error) {
	d, err := mdc.Device(deviceID, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		return "", err
	}

	if d.AdminState == contract.Locked {
		LoggingClient.Error(d.Name + " is in admin locked state")
		return "", errors.NewErrDeviceLocked(d.Name)
	}

	// once command service have its own persistence layer this call will be changed.
	commands, err := dbClient.GetCommandsByDeviceId(d.Id)
	if err != nil {
		LoggingClient.Error(err.Error())
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
		errMsg := fmt.Sprintf("Command with id '%v' does not belong to device with id '%v'.", commandID, deviceID)
		LoggingClient.Error(errMsg)
		err = goErrors.New(errMsg)
		return "", err
	}

	return commandByDevice(d, c, body, queryParams, isPutCommand, ctx)
}

func commandByNames(dn string, cn string, body string, queryParams string, isPutCommand bool,
	ctx context.Context) (string, error) {
	d, err := mdc.DeviceForName(dn, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		return "", err
	}

	if d.AdminState == contract.Locked {
		LoggingClient.Error(d.Name + " is in admin locked state")
		return "", errors.NewErrDeviceLocked(d.Name)
	}

	command, err := dbClient.GetCommandByNameAndDeviceId(cn, d.Id)
	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			return "", err
		} else {
			return "", err
		}
	}

	return commandByDevice(d, command, body, queryParams, isPutCommand, ctx)
}

func commandByDevice(device contract.Device, command contract.Command, body string, queryParams string,
	isPutCommand bool, ctx context.Context) (string, error) {
	var ex Executor
	var err error
	if isPutCommand {
		ex, err = NewPutCommand(device, command, body, ctx, &http.Client{})
	} else {
		ex, err = NewGetCommand(device, command, queryParams, ctx, &http.Client{})
	}

	if err != nil {
		LoggingClient.Error(err.Error())
		return "", err
	}

	responseBody, responseCode, err := ex.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		return "", err
	}
	if responseCode != 200 {
		err = goErrors.New("change me")
		return "", err
	}

	return responseBody, nil
}

func getCommands(ctx context.Context) ([]contract.CommandResponse, error) {
	devices, err := mdc.Devices(ctx)
	if err != nil {
		return nil, err
	}
	cr := []contract.CommandResponse{}
	for _, d := range devices {
		commands, err := dbClient.GetCommandsByDeviceId(d.Id)
		if err != nil {
			LoggingClient.Error(err.Error())
			return nil, err
		}
		cr = append(cr, contract.CommandResponseFromDevice(d, commands, Configuration.Service.Url()))
	}
	return cr, err

}

func getCommandsByDeviceID(did string, ctx context.Context) (contract.CommandResponse, error) {
	d, err := mdc.Device(did, ctx)
	if err != nil {
		return nil, err
	}

	commands, err := dbClient.GetCommandsByDeviceId(d.Id)
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}

	return contract.CommandResponseFromDevice(d, commands, Configuration.Service.Url()), err
}

func getCommandsByDeviceName(dn string, ctx context.Context) (contract.CommandResponse, error) {
	d, err := mdc.DeviceForName(dn, ctx)
	if err != nil {
		return contract.CommandResponse{}, err
	}

	commands, err := dbClient.GetCommandsByDeviceId(d.Id)
	if err != nil {
		LoggingClient.Error(err.Error())
		return contract.CommandResponse{}, err
	}

	return contract.CommandResponseFromDevice(d, commands, Configuration.Service.Url()), err
}
