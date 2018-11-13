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
	"context"
	"encoding/json"
	"net/url"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Device Profile client for interacting with the device profile section of metadata
type DeviceProfileClient interface {
	Add(dp *models.DeviceProfile, ctx context.Context) (string, error)
	Delete(id string, ctx context.Context) error
	DeleteByName(name string, ctx context.Context) error
	DeviceProfile(id string, ctx context.Context) (models.DeviceProfile, error)
	DeviceProfiles(ctx context.Context) ([]models.DeviceProfile, error)
	DeviceProfileForName(name string, ctx context.Context) (models.DeviceProfile, error)
	Update(dp models.DeviceProfile, ctx context.Context) error
	Upload(yamlString string, ctx context.Context) (string, error)
	UploadFile(yamlFilePath string, ctx context.Context) (string, error)
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
func (dpc *DeviceProfileRestClient) requestDeviceProfile(url string, ctx context.Context) (models.DeviceProfile, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return models.DeviceProfile{}, err
	}

	dp := models.DeviceProfile{}
	err = json.Unmarshal(data, &dp)
	return dp, err
}

// Helper method to request and decode a device profile slice
func (dpc *DeviceProfileRestClient) requestDeviceProfileSlice(url string, ctx context.Context) ([]models.DeviceProfile, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return []models.DeviceProfile{}, err
	}

	dpSlice := make([]models.DeviceProfile, 0)
	err = json.Unmarshal(data, &dpSlice)
	return dpSlice, err
}

// Add a new device profile to metadata
func (dpc *DeviceProfileRestClient) Add(dp *models.DeviceProfile, ctx context.Context) (string, error) {
	return clients.PostJsonRequest(dpc.url, dp, ctx)
}

// Delete a device profile (specified by id)
func (dpc *DeviceProfileRestClient) Delete(id string, ctx context.Context) error {
	return clients.DeleteRequest(dpc.url+"/id/"+id, ctx)
}

// Delete a device profile (specified by name)
func (dpc *DeviceProfileRestClient) DeleteByName(name string, ctx context.Context) error {
	return clients.DeleteRequest(dpc.url+"/name/"+url.QueryEscape(name), ctx)
}

// Get the device profile by id
func (dpc *DeviceProfileRestClient) DeviceProfile(id string, ctx context.Context) (models.DeviceProfile, error) {
	return dpc.requestDeviceProfile(dpc.url+"/"+id, ctx)
}

// Get a list of all devices
func (dpc *DeviceProfileRestClient) DeviceProfiles(ctx context.Context) ([]models.DeviceProfile, error) {
	return dpc.requestDeviceProfileSlice(dpc.url, ctx)
}

// Get the device profile by name
func (dpc *DeviceProfileRestClient) DeviceProfileForName(name string, ctx context.Context) (models.DeviceProfile, error) {
	return dpc.requestDeviceProfile(dpc.url+"/name/"+name, ctx)
}

// Update an existing device profile in metadata
func (dpc *DeviceProfileRestClient) Update(dp models.DeviceProfile, ctx context.Context) error {
	return clients.UpdateRequest(dpc.url, dp, ctx)
}

func (dpc *DeviceProfileRestClient) Upload(yamlString string, ctx context.Context) (string, error) {
	return clients.PostRequest(dpc.url+"/upload", []byte(yamlString), clients.ContentYaml, ctx)
}

func (dpc *DeviceProfileRestClient) UploadFile(yamlFilePath string, ctx context.Context) (string, error) {
	return clients.UploadFileRequest(dpc.url+"/uploadfile", yamlFilePath, ctx)
}
