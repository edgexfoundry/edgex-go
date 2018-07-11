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
	"net/url"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

/*
Addressable client for interacting with the addressable section of metadata
*/
type AddressableClient interface {
	Add(addr *models.Addressable) (string, error)
	AddressableForName(name string) (models.Addressable, error)
}

type AddressableRestClient struct {
	url      string
	endpoint clients.Endpointer
}

/*
Return an instance of AddressableClient
*/
func NewAddressableClient(params types.EndpointParams, m clients.Endpointer) AddressableClient {
	a := AddressableRestClient{endpoint: m}
	a.init(params)
	return &a
}

func (a *AddressableRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go a.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					a.url = url
				}
			}
		}(ch)
	} else {
		a.url = params.Url
	}
}

// Add an addressable - handle error codes
// Returns the ID of the addressable and an error
func (a *AddressableRestClient) Add(addr *models.Addressable) (string, error) {
	// Marshal the addressable to JSON
	jsonStr, err := json.Marshal(addr)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Post(a.url, "application/json", bytes.NewReader(jsonStr))
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
		return "", errors.New(string(bodyString))
	}

	return bodyString, err
}

// Helper method to decode an addressable and return the addressable
func (d *AddressableRestClient) decodeAddressable(resp *http.Response) (models.Addressable, error) {
	dec := json.NewDecoder(resp.Body)
	addr := models.Addressable{}

	err := dec.Decode(&addr)
	if err != nil {
		return models.Addressable{}, err
	}

	return addr, err
}

// TODO: make method signatures consistent wrt to error return value
// ie. use it everywhere, or not at all!

// Get the addressable by name
func (a *AddressableRestClient) AddressableForName(name string) (models.Addressable, error) {
	req, err := http.NewRequest(http.MethodGet, a.url+"/name/"+url.QueryEscape(name), nil)
	if err != nil {
		return models.Addressable{}, err
	}

	resp, err := makeRequest(req)

	// Check response
	if resp == nil {
		return models.Addressable{}, ErrResponseNil
	}
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("AddressableForName makeRequest failed: %v\n", err)
		return models.Addressable{}, err
	}

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.Addressable{}, err
		}
		bodyString := string(bodyBytes)

		return models.Addressable{}, errors.New(bodyString)
	}

	return a.decodeAddressable(resp)
}
