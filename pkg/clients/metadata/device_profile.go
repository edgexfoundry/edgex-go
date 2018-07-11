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
	"net/url"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Device Profile client for interacting with the device profile section of metadata
type DeviceProfileClient interface {
	Add(dp *models.DeviceProfile) (string, error)
	Delete(id string) error
	DeleteByName(name string) error
	DeviceProfile(id string) (models.DeviceProfile, error)
	DeviceProfiles() ([]models.DeviceProfile, error)
	DeviceProfileForName(name string) (models.DeviceProfile, error)
}

type DeviceProfileRestClient struct {
	url      string
	endpoint clients.Endpointer
}

// Return an instance of DeviceProfileClient
func NewDeviceProfileClient(params types.EndpointParams, m clients.Endpointer) DeviceProfileClient {
	d := DeviceProfileRestClient{endpoint: m}
	d.init(params)
	return &d
}

func (d *DeviceProfileRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go d.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
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

	req, err := http.NewRequest(http.MethodPost, dpc.url, bytes.NewReader(jsonStr))
	if err != nil {
		return "", err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	// Get the body
	bodyBytes, err := getBody(resp)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)

	// Check the response code
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(bodyString)
	}

	return bodyString, nil
}

// Delete a device profile (specified by id)
func (dpc *DeviceProfileRestClient) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, dpc.url+"/id/"+id, nil)
	if err != nil {
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Delete a device profile (specified by name)
func (dpc *DeviceProfileRestClient) DeleteByName(name string) error {
	req, err := http.NewRequest(http.MethodDelete, dpc.url+"/name/"+url.QueryEscape(name), nil)
	if err != nil {
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Get the device profile by id
func (dpc *DeviceProfileRestClient) DeviceProfile(id string) (models.DeviceProfile, error) {
	req, err := http.NewRequest(http.MethodGet, dpc.url+"/"+id, nil)
	if err != nil {
		return models.DeviceProfile{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return models.DeviceProfile{}, err
	}
	if resp == nil {
		return models.DeviceProfile{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.DeviceProfile{}, err
		}
		bodyString := string(bodyBytes)

		return models.DeviceProfile{}, errors.New(bodyString)
	}

	return dpc.decodeDeviceProfile(resp)
}

// Get a list of all devices
func (dpc *DeviceProfileRestClient) DeviceProfiles() ([]models.DeviceProfile, error) {
	req, err := http.NewRequest(http.MethodGet, dpc.url, nil)
	if err != nil {
		return []models.DeviceProfile{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.DeviceProfile{}, err
	}
	if resp == nil {
		return []models.DeviceProfile{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.DeviceProfile{}, err
		}
		bodyString := string(bodyBytes)

		return []models.DeviceProfile{}, errors.New(bodyString)
	}
	return dpc.decodeDeviceProfileSlice(resp)
}

// Get the device profile by name
func (dpc *DeviceProfileRestClient) DeviceProfileForName(name string) (models.DeviceProfile, error) {
	req, err := http.NewRequest(http.MethodGet, dpc.url+"/name/"+name, nil)
	if err != nil {
		return models.DeviceProfile{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return models.DeviceProfile{}, err
	}
	if resp == nil {
		return models.DeviceProfile{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.DeviceProfile{}, err
		}
		bodyString := string(bodyBytes)

		return models.DeviceProfile{}, errors.New(bodyString)
	}

	return dpc.decodeDeviceProfile(resp)
}

// Help method to decode a device profile slice
func (dpc *DeviceProfileRestClient) decodeDeviceProfileSlice(resp *http.Response) ([]models.DeviceProfile, error) {
	dec := json.NewDecoder(resp.Body)
	dSlice := []models.DeviceProfile{}

	err := dec.Decode(&dSlice)
	if err != nil {
		return []models.DeviceProfile{}, err
	}

	return dSlice, err
}

// Helper method to decode and return a device profile
func (dpc *DeviceProfileRestClient) decodeDeviceProfile(resp *http.Response) (models.DeviceProfile, error) {
	dec := json.NewDecoder(resp.Body)
	ds := models.DeviceProfile{}
	err := dec.Decode(&ds)
	if err != nil {
		return models.DeviceProfile{}, err
	}

	return ds, err
}
