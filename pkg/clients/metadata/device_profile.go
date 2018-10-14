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

// Device Profile client for interacting with the device profile section of metadata
type DeviceProfileClient interface {
	Add(dp *models.DeviceProfile) (string, error)
	Delete(id string) error
	DeleteByName(name string) error
	DeviceProfile(id string) (models.DeviceProfile, error)
	DeviceProfiles() ([]models.DeviceProfile, error)
	DeviceProfileForName(name string) (models.DeviceProfile, error)
	Update(dp models.DeviceProfile) error
	Upload(yamlString string) (string, error)
	UploadFile(yamlFilePath string) (string, error)
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

// Helper method to request and decode a device profile
func (dpc *DeviceProfileRestClient) requestDeviceProfile(url string) (models.DeviceProfile, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return models.DeviceProfile{}, err
	}

	dp := models.DeviceProfile{}
	err = json.Unmarshal(data, &dp)
	return dp, err
}

// Helper method to request and decode a device profile slice
func (dpc *DeviceProfileRestClient) requestDeviceProfileSlice(url string) ([]models.DeviceProfile, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return []models.DeviceProfile{}, err
	}

	dpSlice := make([]models.DeviceProfile, 0)
	err = json.Unmarshal(data, &dpSlice)
	return dpSlice, err
}

// Add a new device profile to metadata
func (dpc *DeviceProfileRestClient) Add(dp *models.DeviceProfile) (string, error) {
	return clients.PostJsonRequest(dpc.url, dp)
}

// Delete a device profile (specified by id)
func (dpc *DeviceProfileRestClient) Delete(id string) error {
	return clients.DeleteRequest(dpc.url + "/id/" + id)
}

// Delete a device profile (specified by name)
func (dpc *DeviceProfileRestClient) DeleteByName(name string) error {
	return clients.DeleteRequest(dpc.url + "/name/" + url.QueryEscape(name))
}

// Get the device profile by id
func (dpc *DeviceProfileRestClient) DeviceProfile(id string) (models.DeviceProfile, error) {
	return dpc.requestDeviceProfile(dpc.url + "/" + id)
}

// Get a list of all devices
func (dpc *DeviceProfileRestClient) DeviceProfiles() ([]models.DeviceProfile, error) {
	return dpc.requestDeviceProfileSlice(dpc.url)
}

// Get the device profile by name
func (dpc *DeviceProfileRestClient) DeviceProfileForName(name string) (models.DeviceProfile, error) {
	return dpc.requestDeviceProfile(dpc.url + "/name/" + name)
}

// Update an existing device profile in metadata
func (dpc *DeviceProfileRestClient) Update(dp models.DeviceProfile) error {
	return clients.UpdateRequest(dpc.url, dp)
}

func (dpc *DeviceProfileRestClient) Upload(yamlString string) (string, error) {
	return clients.PostRequest(dpc.url+"/upload", []byte(yamlString), clients.ContentYaml)
}

func (dpc *DeviceProfileRestClient) UploadFile(yamlFilePath string) (string, error) {
	return clients.UploadFileRequest(dpc.url+"/uploadfile", yamlFilePath)
}
