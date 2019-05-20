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
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"net/http"
)

func commandByDeviceID(deviceID string, commandID string, body string, isPutCommand bool, ctx context.Context) (string, int) {
	device, err := mdc.Device(deviceID, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return err.Error(), chk.StatusCode
		} else {
			return err.Error(), http.StatusInternalServerError
		}
	}

	if device.AdminState == contract.Locked {
		LoggingClient.Error(device.Name + " is in admin locked state")

		return "", http.StatusLocked
	}

	command, err := cc.Command(commandID, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return err.Error(), chk.StatusCode
		} else {
			return err.Error(), http.StatusInternalServerError
		}
	}

	return commandByDevice(device, command, body, isPutCommand, ctx)
}

func commandByNames(dn string, cn string, body string, isPutCommand bool, ctx context.Context) (string, int) {
	d, err := mdc.DeviceForName(dn, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())
		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return err.Error(), chk.StatusCode
		} else {
			return err.Error(), http.StatusInternalServerError
		}
	}

	//Match command name with commands associated with Device Profile.
	var c contract.Command
	for _, one := range d.Profile.CoreCommands {
		if cn == one.Name {
			c = one
			break
		}
	}

	if c.String() == (contract.Command{}).String() {
		errMsg := fmt.Sprintf("Command with name '%v' not found.", cn)
		LoggingClient.Error(errMsg)
		return errMsg, http.StatusNotFound
	}

	return commandByDevice(d, c, body, isPutCommand, ctx)
}

func commandByDevice(device contract.Device, command contract.Command, body string, isPutCommand bool, ctx context.Context) (string, int) {
	var ex Executor
	var err error
	if isPutCommand {
		ex, err = NewPutCommand(device, command, body, ctx, &http.Client{})
	} else {
		ex, err = NewGetCommand(device, command, ctx, &http.Client{})
	}

	if err != nil {
		return "", http.StatusInternalServerError
	}

	responseBody, responseCode, err := ex.Execute()
	if err != nil {
		return "", http.StatusInternalServerError
	}

	return responseBody, responseCode
}

func getCommands(ctx context.Context) (int, []contract.CommandResponse, error) {
	devices, err := mdc.Devices(ctx)
	if err != nil {
		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, nil, chk
		} else {
			return http.StatusInternalServerError, nil, err
		}
	}
	var cr []contract.CommandResponse
	for _, d := range devices {
		cr = append(cr, contract.CommandResponseFromDevice(d, d.Profile.CoreCommands, Configuration.Service.Url()))
	}
	return http.StatusOK, cr, err

}

func getCommandsByDeviceID(did string, ctx context.Context) (int, contract.CommandResponse, error) {
	d, err := mdc.Device(did, ctx)
	if err != nil {
		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, contract.CommandResponse{}, chk
		} else {
			return http.StatusInternalServerError, contract.CommandResponse{}, err
		}
	}
	return http.StatusOK, contract.CommandResponseFromDevice(d, d.Profile.CoreCommands, Configuration.Service.Url()), err
}

func getCommandsByDeviceName(dn string, ctx context.Context) (int, contract.CommandResponse, error) {
	d, err := mdc.DeviceForName(dn, ctx)
	if err != nil {
		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, contract.CommandResponse{}, err
		} else {
			return http.StatusInternalServerError, contract.CommandResponse{}, err
		}
	}
	return http.StatusOK, contract.CommandResponseFromDevice(d, d.Profile.CoreCommands, Configuration.Service.Url()), err
}
