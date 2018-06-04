/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
package metadata

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/edgexfoundry/edgex-go/core/clients"
	"github.com/edgexfoundry/edgex-go/core/clients/types"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

// Device Profile client for interacting with the device profile section of metadata
type DeviceProfileClient interface {
	Add(dp *models.DeviceProfile) (string, error)
}

type DeviceProfileRestClient struct {
	url string
	endpoint clients.Endpointer
}

// Return an instance of DeviceProfileClient
func NewDeviceProfileClient(params types.EndpointParams, m clients.Endpointer) DeviceProfileClient {
	d := DeviceProfileRestClient{endpoint:m}
	d.init(params)
	return &d
}

func(d *DeviceProfileRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go d.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for true {
				select {
				case url := <- ch:
					d.url = url
				}
			}
		}(ch)
	} else {
		d.url = params.Url
	}
}

// Add a new device profile to metadata
func (dpc *DeviceProfileRestClient) Add(dp *models.DeviceProfile) (string, error) {
	jsonStr, err := json.Marshal(dp)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Post(dpc.url, "application/json", bytes.NewReader(jsonStr))
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	// Get the response
	bodyBytes, err := getBody(resp)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)

	// Check the response code
	if resp.StatusCode != 200 {
		return "", errors.New(bodyString)
	}

	return bodyString, nil
}