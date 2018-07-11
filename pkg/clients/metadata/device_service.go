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
	"fmt"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

/*
Service client for interacting with the device service section of metadata
*/
type DeviceServiceClient interface {
	Add(ds *models.DeviceService) (string, error)
	DeviceServiceForName(name string) (models.DeviceService, error)
	UpdateLastConnected(id string, time int64) error
	UpdateLastReported(id string, time int64) error
}

type DeviceServiceRestClient struct {
	url      string
	endpoint clients.Endpointer
}

/*
Return an instance of DeviceServiceClient
*/
func NewDeviceServiceClient(params types.EndpointParams, m clients.Endpointer) DeviceServiceClient {
	s := DeviceServiceRestClient{endpoint: m}
	s.init(params)
	return &s
}

func (d *DeviceServiceRestClient) init(params types.EndpointParams) {
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

// Helper method to decode and return a deviceservice
func (s *DeviceServiceRestClient) decodeDeviceService(resp *http.Response) (models.DeviceService, error) {
	dec := json.NewDecoder(resp.Body)
	ds := models.DeviceService{}
	err := dec.Decode(&ds)
	if err != nil {
		return models.DeviceService{}, err
	}

	return ds, err
}

// Update the last connected time for the device service
func (s *DeviceServiceRestClient) UpdateLastConnected(id string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, s.url+"/"+id+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
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

// Update the last reported time for the device service
func (s *DeviceServiceRestClient) UpdateLastReported(id string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, s.url+"/"+id+"/lastreported/"+strconv.FormatInt(time, 10), nil)
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

// Add a new deviceservice
func (s *DeviceServiceRestClient) Add(ds *models.DeviceService) (string, error) {
	jsonStr, err := json.Marshal(ds)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Post(s.url, "application/json", bytes.NewReader(jsonStr))
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	// Get the response body
	bodyBytes, err := getBody(resp)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(bodyString)
	}

	return bodyString, nil
}

// Request deviceservice for specified name
func (s *DeviceServiceRestClient) DeviceServiceForName(name string) (models.DeviceService, error) {
	req, err := http.NewRequest(http.MethodGet, s.url+"/name/"+name, nil)
	if err != nil {
		fmt.Printf("DeviceServiceForName NewRequest failed: %v\n", err)
		return models.DeviceService{}, err
	}

	resp, err := makeRequest(req)
	if resp == nil {
		return models.DeviceService{}, ErrResponseNil
	}
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("DeviceServiceForName makeRequest failed: %v\n", err)
		return models.DeviceService{}, err
	}

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.DeviceService{}, err
		}
		bodyString := string(bodyBytes)

		return models.DeviceService{}, errors.New(bodyString)
	}

	return s.decodeDeviceService(resp)
}
