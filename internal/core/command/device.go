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
			return "", chk.StatusCode
		} else {
			return "", http.StatusInternalServerError
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
			return "", chk.StatusCode
		} else {
			return "", http.StatusInternalServerError
		}
	}

	var ex Executor
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

func putDeviceAdminState(did string, as string, ctx context.Context) (int, error) {
	err := mdc.UpdateAdminState(did, as, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, chk
		} else {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusOK, err
}

func putDeviceAdminStateByName(dn string, as string, ctx context.Context) (int, error) {
	err := mdc.UpdateAdminStateByName(dn, as, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, chk
		} else {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusOK, err
}

func putDeviceOpState(did string, as string, ctx context.Context) (int, error) {
	err := mdc.UpdateOpState(did, as, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, chk
		} else {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusOK, err
}

func putDeviceOpStateByName(dn string, as string, ctx context.Context) (int, error) {
	err := mdc.UpdateOpStateByName(dn, as, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, chk
		} else {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusOK, err
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
