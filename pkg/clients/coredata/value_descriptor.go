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
	"context"
	"encoding/json"
	"net/url"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Addressable client for interacting with the addressable section of metadata
type ValueDescriptorClient interface {
	ValueDescriptors(ctx context.Context) ([]models.ValueDescriptor, error)
	ValueDescriptor(id string, ctx context.Context) (models.ValueDescriptor, error)
	ValueDescriptorForName(name string, ctx context.Context) (models.ValueDescriptor, error)
	ValueDescriptorsByLabel(label string, ctx context.Context) ([]models.ValueDescriptor, error)
	ValueDescriptorsForDevice(deviceId string, ctx context.Context) ([]models.ValueDescriptor, error)
	ValueDescriptorsForDeviceByName(deviceName string, ctx context.Context) ([]models.ValueDescriptor, error)
	ValueDescriptorsByUomLabel(uomLabel string, ctx context.Context) ([]models.ValueDescriptor, error)
	Add(vdr *models.ValueDescriptor, ctx context.Context) (string, error)
	Update(vdr *models.ValueDescriptor, ctx context.Context) error
	Delete(id string, ctx context.Context) error
	DeleteByName(name string, ctx context.Context) error
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

// Helper method to request and decode a valuedescriptor slice
func (v *ValueDescriptorRestClient) requestValueDescriptorSlice(url string, ctx context.Context) ([]models.ValueDescriptor, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return []models.ValueDescriptor{}, err
	}

	dSlice := make([]models.ValueDescriptor, 0)
	err = json.Unmarshal(data, &dSlice)
	return dSlice, err
}

// Helper method to request and decode a device
func (v *ValueDescriptorRestClient) requestValueDescriptor(url string, ctx context.Context) (models.ValueDescriptor, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return models.ValueDescriptor{}, err
	}

	vdr := models.ValueDescriptor{}
	err = json.Unmarshal(data, &vdr)
	return vdr, err
}

// Get a list of all value descriptors
func (v *ValueDescriptorRestClient) ValueDescriptors(ctx context.Context) ([]models.ValueDescriptor, error) {
	return v.requestValueDescriptorSlice(v.url, ctx)
}

// Get the value descriptor by id
func (v *ValueDescriptorRestClient) ValueDescriptor(id string, ctx context.Context) (models.ValueDescriptor, error) {
	return v.requestValueDescriptor(v.url+"/"+id, ctx)
}

// Get the value descriptor by name
func (v *ValueDescriptorRestClient) ValueDescriptorForName(name string, ctx context.Context) (models.ValueDescriptor, error) {
	return v.requestValueDescriptor(v.url+"/name/"+url.QueryEscape(name), ctx)
}

// Get the value descriptors by label
func (v *ValueDescriptorRestClient) ValueDescriptorsByLabel(label string, ctx context.Context) ([]models.ValueDescriptor, error) {
	return v.requestValueDescriptorSlice(v.url+"/label/"+url.QueryEscape(label), ctx)
}

// Get the value descriptors for a device (by id)
func (v *ValueDescriptorRestClient) ValueDescriptorsForDevice(deviceId string, ctx context.Context) ([]models.ValueDescriptor, error) {
	return v.requestValueDescriptorSlice(v.url+"/deviceid/"+deviceId, ctx)
}

// Get the value descriptors for a device (by name)
func (v *ValueDescriptorRestClient) ValueDescriptorsForDeviceByName(deviceName string, ctx context.Context) ([]models.ValueDescriptor, error) {
	return v.requestValueDescriptorSlice(v.url+"/devicename/"+deviceName, ctx)
}

// Get the value descriptors for a uomLabel
func (v *ValueDescriptorRestClient) ValueDescriptorsByUomLabel(uomLabel string, ctx context.Context) ([]models.ValueDescriptor, error) {
	return v.requestValueDescriptorSlice(v.url+"/uomlabel/"+uomLabel, ctx)
}

// Add a value descriptor
func (v *ValueDescriptorRestClient) Add(vdr *models.ValueDescriptor, ctx context.Context) (string, error) {
	return clients.PostJsonRequest(v.url, vdr, ctx)
}

// Update a value descriptor
func (v *ValueDescriptorRestClient) Update(vdr *models.ValueDescriptor, ctx context.Context) error {
	return clients.UpdateRequest(v.url, vdr, ctx)
}

// Delete a value descriptor (specified by id)
func (v *ValueDescriptorRestClient) Delete(id string, ctx context.Context) error {
	return clients.DeleteRequest(v.url+"/id/"+id, ctx)
}

// Delete a value descriptor (specified by name)
func (v *ValueDescriptorRestClient) DeleteByName(name string, ctx context.Context) error {
	return clients.DeleteRequest(v.url+"/name/"+name, ctx)
}
