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
package coredata

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

var (
	ErrResponseNil       = errors.New("Response was nil")
	ErrNotFound    error = errors.New("Item not found")
)

// Addressable client for interacting with the addressable section of metadata
type ValueDescriptorClient interface {
	ValueDescriptors() ([]models.ValueDescriptor, error)
	ValueDescriptor(id string) (models.ValueDescriptor, error)
	ValueDescriptorForName(name string) (models.ValueDescriptor, error)
	ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error)
	ValueDescriptorsForDevice(deviceId string) ([]models.ValueDescriptor, error)
	ValueDescriptorsForDeviceByName(deviceName string) ([]models.ValueDescriptor, error)
	ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error)
	Add(vdr *models.ValueDescriptor) (string, error)
	Update(vdr *models.ValueDescriptor) error
	Delete(id string) error
	DeleteByName(name string) error
}

type ValueDescriptorRestClient struct {
	url      string
	endpoint clients.Endpointer
}

func NewValueDescriptorClient(params types.EndpointParams, m clients.Endpointer) ValueDescriptorClient {
	v := ValueDescriptorRestClient{endpoint: m}
	v.init(params)
	return &v
}

func (d *ValueDescriptorRestClient) init(params types.EndpointParams) {
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

// Helper method to get the body from the response after making the request
func getBody(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

// Helper method to make the request and return the response
func makeRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)

	return resp, err
}

// Help method to decode a valuedescriptor slice
func (v *ValueDescriptorRestClient) decodeValueDescriptorSlice(resp *http.Response) ([]models.ValueDescriptor, error) {
	dSlice := make([]models.ValueDescriptor, 0)

	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&dSlice)

	return dSlice, err
}

// Helper method to decode a device and return the device
func (v *ValueDescriptorRestClient) decodeValueDescriptor(resp *http.Response) (models.ValueDescriptor, error) {
	dec := json.NewDecoder(resp.Body)
	vdr := models.ValueDescriptor{}

	err := dec.Decode(&vdr)

	return vdr, err
}

// Get a list of all value descriptors
func (v *ValueDescriptorRestClient) ValueDescriptors() ([]models.ValueDescriptor, error) {
	req, err := http.NewRequest(http.MethodGet, v.url, nil)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}
	if resp == nil {
		return []models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	// Response was not OK
	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.ValueDescriptor{}, err
		}
		return []models.ValueDescriptor{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return v.decodeValueDescriptorSlice(resp)
}

// Get the value descriptor by id
func (v *ValueDescriptorRestClient) ValueDescriptor(id string) (models.ValueDescriptor, error) {
	req, err := http.NewRequest(http.MethodGet, v.url+"/"+id, nil)
	if err != nil {
		return models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return models.ValueDescriptor{}, err
	}
	if resp == nil {
		return models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.ValueDescriptor{}, err
		}

		return models.ValueDescriptor{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return v.decodeValueDescriptor(resp)
}

// Get the value descriptor by name
func (v *ValueDescriptorRestClient) ValueDescriptorForName(name string) (models.ValueDescriptor, error) {
	req, err := http.NewRequest(http.MethodGet, v.url+"/name/"+url.QueryEscape(name), nil)
	if err != nil {
		return models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return models.ValueDescriptor{}, err
	}
	if resp == nil {
		return models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.ValueDescriptor{}, err
		}

		return models.ValueDescriptor{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return v.decodeValueDescriptor(resp)
}

// Get the value descriptors by label
func (v *ValueDescriptorRestClient) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	req, err := http.NewRequest(http.MethodGet, v.url+"/label/"+url.QueryEscape(label), nil)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}
	if resp == nil {
		return []models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.ValueDescriptor{}, err
		}

		return []models.ValueDescriptor{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return v.decodeValueDescriptorSlice(resp)
}

// Get the value descriptors for a device (by id)
func (v *ValueDescriptorRestClient) ValueDescriptorsForDevice(deviceId string) ([]models.ValueDescriptor, error) {
	req, err := http.NewRequest(http.MethodGet, v.url+"/deviceid/"+deviceId, nil)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}
	if resp == nil {
		return []models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.ValueDescriptor{}, err
		}

		return []models.ValueDescriptor{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return v.decodeValueDescriptorSlice(resp)
}

// Get the value descriptors for a device (by name)
func (v *ValueDescriptorRestClient) ValueDescriptorsForDeviceByName(deviceName string) ([]models.ValueDescriptor, error) {
	req, err := http.NewRequest(http.MethodGet, v.url+"/devicename/"+deviceName, nil)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}
	if resp == nil {
		return []models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.ValueDescriptor{}, err
		}

		return []models.ValueDescriptor{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return v.decodeValueDescriptorSlice(resp)
}

// Get the value descriptors for a uomLabel
func (v *ValueDescriptorRestClient) ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error) {
	req, err := http.NewRequest(http.MethodGet, v.url+"/uomlabel/"+uomLabel, nil)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}
	if resp == nil {
		return []models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.ValueDescriptor{}, err
		}

		return []models.ValueDescriptor{}, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}
	return v.decodeValueDescriptorSlice(resp)
}

// Add a value descriptor
func (v *ValueDescriptorRestClient) Add(vdr *models.ValueDescriptor) (string, error) {
	jsonStr, err := json.Marshal(vdr)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, v.url, bytes.NewReader(jsonStr))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeRequest(req)
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
		return "", types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return bodyString, nil
}

// update a value descriptor
func (v *ValueDescriptorRestClient) Update(vdr *models.ValueDescriptor) error {
	jsonStr, err := json.Marshal(&vdr)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, v.url, bytes.NewReader(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeRequest(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}


		return types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return nil
}

// Delete a value descriptor (specified by id)
func (v *ValueDescriptorRestClient) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, v.url+"/id/"+id, nil)
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
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}

		return types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return nil
}

// Delete a value descriptor (specified by name)
func (v *ValueDescriptorRestClient) DeleteByName(name string) error {
	req, err := http.NewRequest(http.MethodDelete, v.url+"/name/"+name, nil)
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
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}

		return types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return nil
}
