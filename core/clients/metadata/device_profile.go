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

	"github.com/edgexfoundry/edgex-go/core/domain/models"

)

// Device Profile client for interacting with the device profile section of metadata
type DeviceProfileClient interface {
	Add(dp *models.DeviceProfile) (string, error)
}

type DeviceProfileRestClient struct {
	url string
}

// Return an instance of DeviceProfileClient
func NewDeviceProfileClient(metaDbDeviceProfileUrl string) DeviceProfileClient {
	d := DeviceProfileRestClient{url: metaDbDeviceProfileUrl}

	return &d
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
