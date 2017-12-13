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
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package coredataclients

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/edgexfoundry/core-domain-go/models"
	"io/ioutil"
	"net/http"
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
func decodeValueDescriptorSlice(resp *http.Response) ([]models.ValueDescriptor, error) {
	dSlice := make([]models.ValueDescriptor, 0)

	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&dSlice)
	if err != nil {
		fmt.Println(err)
	}

	return dSlice, err
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

	return decodeValueDescriptorSlice(resp)
}
