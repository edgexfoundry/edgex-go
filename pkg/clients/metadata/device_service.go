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
	"strconv"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

/*
Service client for interacting with the device service section of metadata
*/
type DeviceServiceClient interface {
	Add(ds *models.DeviceService, ctx context.Context) (string, error)
	DeviceServiceForName(name string, ctx context.Context) (models.DeviceService, error)
	UpdateLastConnected(id string, time int64, ctx context.Context) error
	UpdateLastReported(id string, time int64, ctx context.Context) error
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

// Helper method to request and decode a device service
func (s *DeviceServiceRestClient) requestDeviceService(url string, ctx context.Context) (models.DeviceService, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return models.DeviceService{}, err
	}

	ds := models.DeviceService{}
	err = json.Unmarshal(data, &ds)
	return ds, err
}

// Update the last connected time for the device service
func (s *DeviceServiceRestClient) UpdateLastConnected(id string, time int64, ctx context.Context) error {
	_, err := clients.PutRequest(s.url+"/"+id+"/lastconnected/"+strconv.FormatInt(time, 10), nil, ctx)
	return err
}

// Update the last reported time for the device service
func (s *DeviceServiceRestClient) UpdateLastReported(id string, time int64, ctx context.Context) error {
	_, err := clients.PutRequest(s.url+"/"+id+"/lastreported/"+strconv.FormatInt(time, 10), nil, ctx)
	return err
}

// Add a new deviceservice
func (s *DeviceServiceRestClient) Add(ds *models.DeviceService, ctx context.Context) (string, error) {
	return clients.PostJsonRequest(s.url, ds, ctx)
}

// Request deviceservice for specified name
func (s *DeviceServiceRestClient) DeviceServiceForName(name string, ctx context.Context) (models.DeviceService, error) {
	return s.requestDeviceService(s.url+"/name/"+name, ctx)
}
