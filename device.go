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
 *
 * @microservice: core-command-go service
 * @author: Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/
package main


import (
	"net/http"

	"github.com/edgexfoundry/core-clients-go/metadataclients"
	"github.com/edgexfoundry/core-domain-go/models"
)

func commandByDeviceID(did string, cid string, b string, p bool) (string, int) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	d, err := dc.Device(did)

	if err != nil {
		loggingClient.Error(err.Error(), "")
		// send 403: no device exists by the id provided
		return "", http.StatusForbidden
	}

	if p && (d.AdminState == models.Locked) {
		loggingClient.Error(d.Name + " is in admin locked state")
		// send 422: device is locked
		return "", http.StatusUnprocessableEntity
	}

	var cc = metadataclients.NewCommandClient(configuration.Metadbcommandurl)
	c, err := cc.Command(cid)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		// send 403 no command exists
		return "", http.StatusForbidden
	}
	if p {
		loggingClient.Info("Issuing PUT command to: " + string(c.Put.Action.Path))
		body, status, err := issueCommand(c.Put.Action.Path, b, true)
		if err != nil {
			return "", http.StatusBadGateway
		}
		return body, status
	}
	loggingClient.Info("Issuing GET command to: " + c.Get.Action.Path)
	gbody, gstatus, gerr := issueCommand(c.Get.Action.Path, "", false)
	if gerr != nil {
		return "", http.StatusBadGateway
	}
	return gbody, gstatus
}

func putDeviceAdminState(did string, as string) (int, error) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	err := dc.UpdateAdminState(did, as)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		return http.StatusServiceUnavailable, err
	}
	return http.StatusOK, err
}

func putDeviceAdminStateByName(dn string, as string) (int, error) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	err := dc.UpdateAdminStateByName(dn, as)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		return http.StatusServiceUnavailable, err
	}
	return http.StatusOK, err
}

func putDeviceOpState(did string, as string) (int, error) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	err := dc.UpdateOpState(did, as)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		return http.StatusServiceUnavailable, err
	}
	return http.StatusOK, err
}

func putDeviceOpStateByName(dn string, as string) (int, error) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	err := dc.UpdateOpStateByName(dn, as)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		return http.StatusServiceUnavailable, err
	}
	return http.StatusOK, err
}

func getCommands() (int, []models.Device, error) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	devices, err := dc.Devices()
	if err != nil {
		// Need to Check if Device Exist TODO
		return http.StatusServiceUnavailable, nil, err
	}
	return http.StatusOK, devices, err

}

func getCommandsByDeviceID(did string) (int, models.Device, error) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	d, err := dc.Device(did)
	if err != nil {
		return http.StatusServiceUnavailable, d, err
	}
	return http.StatusOK, d, err
}

func getCommandsByDeviceName(dn string) (int, models.Device, error) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	d, err := dc.DeviceForName(dn)
	if err != nil {
		return http.StatusServiceUnavailable, d, err
	}
	return http.StatusOK, d, err
}
