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
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/edgexfoundry/core-clients-go/metadataclients"
	"github.com/edgexfoundry/core-domain-go/models"
)

func issueCommand(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

func commandByDeviceID(did string, cid string, b string, p bool) (io.ReadCloser, int) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	d, err := dc.Device(did)

	if err != nil {
		loggingClient.Error(err.Error(), "")
		// send 403: no device exists by the id provided
		return nil, http.StatusForbidden
	}

	if p && (d.AdminState == models.Locked) {
		loggingClient.Error(d.Name + " is in admin locked state")
		// send 422: device is locked
		return nil, http.StatusUnprocessableEntity
	}

	var cc = metadataclients.NewCommandClient(configuration.Metadbcommandurl)
	c, err := cc.Command(cid)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		// send 403 no command exists
		return nil, http.StatusForbidden
	}
	if p {
		loggingClient.Info("Issuing PUT command to: " + string(c.Put.Action.Path))
		req, err := http.NewRequest(PUT, c.Put.Action.Path, strings.NewReader(b))
		resp, err := issueCommand(req)
		if err != nil {
			return nil, http.StatusBadGateway
		}
		return resp.Body, resp.StatusCode
	} else {
		loggingClient.Info("Issuing GET command to: " + c.Get.Action.Path)
		req, err := http.NewRequest(PUT, c.Get.Action.Path, nil)
		resp, err := issueCommand(req)
		if err != nil {
			return nil, http.StatusBadGateway
		}
		return nil, resp.StatusCode
	}
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

func getCommands() (int, []models.CommandResponse, error) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	fmt.Print(configuration.Metadbdeviceurl)
	devices, err := dc.Devices()
	if err != nil {
		return http.StatusServiceUnavailable, nil, err
	}
	var cr []models.CommandResponse
	for _, d := range devices {
		cr = append(cr, d.CommandResponse())
	}
	return http.StatusOK, cr, err

}

func getCommandsByDeviceID(did string) (int, models.CommandResponse, error) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	d, err := dc.Device(did)
	if err != nil {
		return http.StatusServiceUnavailable, d.CommandResponse(), err
	}
	return http.StatusOK, d.CommandResponse(), err
}

func getCommandsByDeviceName(dn string) (int, models.CommandResponse, error) {
	var dc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	d, err := dc.DeviceForName(dn)
	if err != nil {
		return http.StatusServiceUnavailable, d.CommandResponse(), err
	}
	return http.StatusOK, d.CommandResponse(), err
}
