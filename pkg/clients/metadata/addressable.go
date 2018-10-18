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
	"encoding/json"
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
	Addressable(id string) (models.Addressable, error)
	AddressableForName(name string) (models.Addressable, error)
	Update(addr models.Addressable) error
	Delete(id string) error
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

// Helper method to request and decode an addressable
func (a *AddressableRestClient) requestAddressable(url string) (models.Addressable, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return models.Addressable{}, err
	}

	add := models.Addressable{}
	err = json.Unmarshal(data, &add)
	return add, err
}

// Add an addressable - handle error codes
// Returns the ID of the addressable and an error
func (a *AddressableRestClient) Add(addr *models.Addressable) (string, error) {
	return clients.PostJsonRequest(a.url, addr)
}

// Get an addressable by id
func (a *AddressableRestClient) Addressable(id string) (models.Addressable, error) {
	return a.requestAddressable(a.url + "/" + id)
}

// Get the addressable by name
func (a *AddressableRestClient) AddressableForName(name string) (models.Addressable, error) {
	return a.requestAddressable(a.url + "/name/" + url.QueryEscape(name))
}

// Update a addressable
func (a *AddressableRestClient) Update(addr models.Addressable) error {
	return clients.UpdateRequest(a.url, addr)
}

// Delete a addressable (specified by id)
func (a *AddressableRestClient) Delete(id string) error {
	return clients.DeleteRequest(a.url + "/id/" + id)
}
