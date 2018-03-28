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
 * @original author: Spencer Bull, Dell
 * @version: 0.5.0
 * @updated by:  Jim White, Dell Technologies, Feb 27, 2108
 * Added func makeTimestamp and import of time to support it (Fede C. initiated during mono repo work)
 *******************************************************************************/
package command

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/core/clients/metadataclients"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

func issueCommand(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	return resp, err
}

func commandByDeviceID(did string, cid string, b string, p bool) (string, int) {
	var dc = metadataclients.NewDeviceClient(configuration.MetaDeviceURL)
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

	var cc = metadataclients.NewCommandClient(configuration.MetaCommandURL)
	c, err := cc.Command(cid)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		// send 403 no command exists
		return "", http.StatusForbidden
	}
	if p {
		url := d.Service.Addressable.GetBaseURL() + strings.Replace(c.Put.Action.Path, DEVICEIDURLPARAM, d.Id.Hex(), -1)
		loggingClient.Info("Issuing PUT command to: " + url)
		req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(b))
		if err != nil {
			return "", http.StatusInternalServerError
		}
		resp, err := issueCommand(req)
		if err != nil {
			return "", http.StatusBadGateway
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return buf.String(), resp.StatusCode
	} else {
		url := d.Service.Addressable.GetBaseURL() + strings.Replace(c.Get.Action.Path, DEVICEIDURLPARAM, d.Id.Hex(), -1)
		loggingClient.Info("Issuing GET command to: " + url)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return "", http.StatusInternalServerError
		}
		resp, err := issueCommand(req)
		if err != nil {
			return "", http.StatusBadGateway
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return buf.String(), resp.StatusCode
	}
}

func putDeviceAdminState(did string, as string) (int, error) {
	var dc = metadataclients.NewDeviceClient(configuration.MetaDeviceURL)
	err := dc.UpdateAdminState(did, as)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		return http.StatusServiceUnavailable, err
	}
	return http.StatusOK, err
}

func putDeviceAdminStateByName(dn string, as string) (int, error) {
	var dc = metadataclients.NewDeviceClient(configuration.MetaDeviceURL)
	err := dc.UpdateAdminStateByName(dn, as)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		return http.StatusServiceUnavailable, err
	}
	return http.StatusOK, err
}

func putDeviceOpState(did string, as string) (int, error) {
	var dc = metadataclients.NewDeviceClient(configuration.MetaDeviceURL)
	err := dc.UpdateOpState(did, as)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		return http.StatusServiceUnavailable, err
	}
	return http.StatusOK, err
}

func putDeviceOpStateByName(dn string, as string) (int, error) {
	var dc = metadataclients.NewDeviceClient(configuration.MetaDeviceURL)
	err := dc.UpdateOpStateByName(dn, as)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		return http.StatusServiceUnavailable, err
	}
	return http.StatusOK, err
}

func getCommands() (int, []models.CommandResponse, error) {
	var dc = metadataclients.NewDeviceClient(configuration.MetaDeviceURL)
	devices, err := dc.Devices()
	if err != nil {
		return http.StatusServiceUnavailable, nil, err
	}
	var cr []models.CommandResponse
	for _, d := range devices {
		cr = append(cr, models.CommandResponseFromDevice(d, constructCommandURL()))
	}
	return http.StatusOK, cr, err

}

func getCommandsByDeviceID(did string) (int, models.CommandResponse, error) {
	var dc = metadataclients.NewDeviceClient(configuration.MetaDeviceURL)
	d, err := dc.Device(did)
	if err != nil {
		return http.StatusServiceUnavailable, models.CommandResponse{}, err
	}
	return http.StatusOK, models.CommandResponseFromDevice(d, constructCommandURL()), err
}

func getCommandsByDeviceName(dn string) (int, models.CommandResponse, error) {
	var dc = metadataclients.NewDeviceClient(configuration.MetaDeviceURL)
	d, err := dc.DeviceForName(dn)
	if err != nil {
		return http.StatusServiceUnavailable, models.CommandResponse{}, err
	}
	return http.StatusOK, models.CommandResponseFromDevice(d, constructCommandURL()), err
}

func constructCommandURL() string {
	return configuration.URLProtocol + configuration.ServiceAddress + ":" + strconv.Itoa(configuration.ServicePort)
}
