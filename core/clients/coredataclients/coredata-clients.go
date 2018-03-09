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
 * @microservice: core-clients-go library
 * @original author: Ryan Comer, Dell
 * @version: 0.5.0
 * updated for Kei Ohmura.  Kei made the original contribution in the core-clients-go repo
 * fix - moves `decodeValueDescriptorSlice()` method to the ValueDescriptorClient.
 * added - methods to the ValueDescriptorClient
 *******************************************************************************/
package coredataclients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

var (
	ErrResponseNil       = errors.New("Response was nil")
	ErrNotFound    error = errors.New("Item not found")
)

// Addressable client for interacting with the addressable section of metadata
type ValueDescriptorClient struct {
	url string
}

func NewValueDescriptorClient(valueDescriptorUrl string) ValueDescriptorClient {
	return ValueDescriptorClient{url: valueDescriptorUrl}
}

// Helper method to get the body from the response after making the request
func getBody(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return body, err
}

// Helper method to make the request and return the response
func makeRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	return resp, err
}

// Help method to decode a valuedescriptor slice
func (v *ValueDescriptorClient) decodeValueDescriptorSlice(resp *http.Response) ([]models.ValueDescriptor, error) {
	dSlice := make([]models.ValueDescriptor, 0)

	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&dSlice)
	if err != nil {
		fmt.Println(err)
	}

	return dSlice, err
}

// Helper method to decode a device and return the device
func (v *ValueDescriptorClient) decodeValueDescriptor(resp *http.Response) (models.ValueDescriptor, error) {
	dec := json.NewDecoder(resp.Body)
	vdr := models.ValueDescriptor{}

	err := dec.Decode(&vdr)
	if err != nil {
		fmt.Println(err)
	}

	return vdr, err
}

// Get a list of all value descriptors
func (v *ValueDescriptorClient) ValueDescriptors() ([]models.ValueDescriptor, error) {
	req, err := http.NewRequest("GET", v.url, nil)
	if err != nil {
		fmt.Println(err.Error())
		return []models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.ValueDescriptor{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	// Reponse was not OK
	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.ValueDescriptor{}, err
		}
		bodyString := string(bodyBytes)
		return []models.ValueDescriptor{}, errors.New(string(bodyString))
	}

	return v.decodeValueDescriptorSlice(resp)
}

// Get the value descriptor by id
func (v *ValueDescriptorClient) ValueDescriptor(id string) (models.ValueDescriptor, error) {
	req, err := http.NewRequest("GET", v.url+"/"+id, nil)
	if err != nil {
		fmt.Println(err)
		return models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return models.ValueDescriptor{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return models.ValueDescriptor{}, err
		}
		bodyString := string(bodyBytes)

		return models.ValueDescriptor{}, errors.New(bodyString)
	}

	return v.decodeValueDescriptor(resp)
}

// Get the value descriptor by name
func (v *ValueDescriptorClient) ValueDescriptorForName(name string) (models.ValueDescriptor, error) {
	req, err := http.NewRequest("GET", v.url+"/name/"+url.QueryEscape(name), nil)
	if err != nil {
		fmt.Println(err)
		return models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return models.ValueDescriptor{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return models.ValueDescriptor{}, err
		}
		bodyString := string(bodyBytes)

		return models.ValueDescriptor{}, errors.New(bodyString)
	}
	return v.decodeValueDescriptor(resp)
}

// Get the value descriptors by label
func (v *ValueDescriptorClient) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	req, err := http.NewRequest("GET", v.url+"/label/"+url.QueryEscape(label), nil)
	if err != nil {
		fmt.Println(err)
		return []models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.ValueDescriptor{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.ValueDescriptor{}, err
		}
		bodyString := string(bodyBytes)

		return []models.ValueDescriptor{}, errors.New(bodyString)
	}
	return v.decodeValueDescriptorSlice(resp)
}

// Get the value descriptors for a device (by id)
func (v *ValueDescriptorClient) ValueDescriptorsForDevice(deviceId string) ([]models.ValueDescriptor, error) {
	req, err := http.NewRequest("GET", v.url+"/deviceid/"+deviceId, nil)
	if err != nil {
		fmt.Println(err)
		return []models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.ValueDescriptor{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.ValueDescriptor{}, err
		}
		bodyString := string(bodyBytes)

		return []models.ValueDescriptor{}, errors.New(bodyString)
	}
	return v.decodeValueDescriptorSlice(resp)
}

// Get the value descriptors for a device (by name)
func (v *ValueDescriptorClient) ValueDescriptorsForDeviceByName(deviceName string) ([]models.ValueDescriptor, error) {
	req, err := http.NewRequest("GET", v.url+"/devicename/"+deviceName, nil)
	if err != nil {
		fmt.Println(err)
		return []models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.ValueDescriptor{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.ValueDescriptor{}, err
		}
		bodyString := string(bodyBytes)

		return []models.ValueDescriptor{}, errors.New(bodyString)
	}
	return v.decodeValueDescriptorSlice(resp)
}

// Get the value descriptors for a uomLabel
func (v *ValueDescriptorClient) ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error) {
	req, err := http.NewRequest("GET", v.url+"/uomlabel/"+uomLabel, nil)
	if err != nil {
		fmt.Println(err)
		return []models.ValueDescriptor{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.ValueDescriptor{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.ValueDescriptor{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.ValueDescriptor{}, err
		}
		bodyString := string(bodyBytes)

		return []models.ValueDescriptor{}, errors.New(bodyString)
	}
	return v.decodeValueDescriptorSlice(resp)
}

// Add a value descriptor
func (v *ValueDescriptorClient) Add(vdr *models.ValueDescriptor) (string, error) {
	jsonStr, err := json.Marshal(vdr)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req, err := http.NewRequest("POST", v.url, bytes.NewReader(jsonStr))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	// Get the response body
	bodyBytes, err := getBody(resp)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	bodyString := string(bodyBytes)

	if resp.StatusCode != 200 {
		return "", errors.New(bodyString)
	}

	return bodyString, nil
}

// update a value descriptor
func (v *ValueDescriptorClient) Update(vdr *models.ValueDescriptor) error {
	jsonStr, err := json.Marshal(&vdr)
	if err != nil {
		fmt.Println(err)
		return err
	}

	req, err := http.NewRequest("PUT", v.url, bytes.NewReader(jsonStr))
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Delete a value descriptor (specified by id)
func (v *ValueDescriptorClient) Delete(id string) error {
	req, err := http.NewRequest("DELETE", v.url+"/id/"+id, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Delete a value descriptor (specified by name)
func (v *ValueDescriptorClient) DeleteByName(name string) error {
	req, err := http.NewRequest("DELETE", v.url+"/name/"+name, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}
