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
	"strconv"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

/*
Device client for interacting with the device section of metadata
*/
type DeviceClient interface {
	Add(dev *models.Device) (string, error)
	Delete(id string) error
	DeleteByName(name string) error
	CheckForDevice(token string) (models.Device, error)
	Device(id string) (models.Device, error)
	DeviceForName(name string) (models.Device, error)
	Devices() ([]models.Device, error)
	DevicesByLabel(label string) ([]models.Device, error)
	DevicesForAddressable(addressableid string) ([]models.Device, error)
	DevicesForAddressableByName(addressableName string) ([]models.Device, error)
	DevicesForProfile(profileid string) ([]models.Device, error)
	DevicesForProfileByName(profileName string) ([]models.Device, error)
	DevicesForService(serviceid string) ([]models.Device, error)
	DevicesForServiceByName(serviceName string) ([]models.Device, error)
	Update(dev models.Device) error
	UpdateAdminState(id string, adminState string) error
	UpdateAdminStateByName(name string, adminState string) error
	UpdateLastConnected(id string, time int64) error
	UpdateLastConnectedByName(name string, time int64) error
	UpdateLastReported(id string, time int64) error
	UpdateLastReportedByName(name string, time int64) error
	UpdateOpState(id string, opState string) error
	UpdateOpStateByName(name string, opState string) error
}

type DeviceRestClient struct {
	url      string
	endpoint clients.Endpointer
}

/*
Return an instance of DeviceClient
*/
func NewDeviceClient(params types.EndpointParams, m clients.Endpointer) DeviceClient {
	d := DeviceRestClient{endpoint: m}
	d.init(params)
	return &d
}

func (d *DeviceRestClient) init(params types.EndpointParams) {
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

// Helper method to request and decode a device
func (d *DeviceRestClient) requestDevice(url string) (models.Device, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return models.Device{}, err
	}

	dev := models.Device{}
	err = json.Unmarshal(data, &dev)
	return dev, err
}

// Helper method to request and decode a device slice
func (d *DeviceRestClient) requestDeviceSlice(url string) ([]models.Device, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return []models.Device{}, err
	}

	dSlice := make([]models.Device, 0)
	err = json.Unmarshal(data, &dSlice)
	return dSlice, err
}

//Use the models.Event.Device property for the supplied token parameter.
//The above property is currently double-purposed and needs to be refactored.
//This call replaces the previous two calls necessary to lookup a device by id followed by name.
func (d *DeviceRestClient) CheckForDevice(token string) (models.Device, error) {
	return d.requestDevice(d.url + "/check/" + token)
}

// Get the device by id
func (d *DeviceRestClient) Device(id string) (models.Device, error) {
	return d.requestDevice(d.url + "/" + id)
}

// Get a list of all devices
func (d *DeviceRestClient) Devices() ([]models.Device, error) {
	return d.requestDeviceSlice(d.url)
}

// Get the device by name
func (d *DeviceRestClient) DeviceForName(name string) (models.Device, error) {
	return d.requestDevice(d.url + "/name/" + url.QueryEscape(name))
}

// Get the device by label
func (d *DeviceRestClient) DevicesByLabel(label string) ([]models.Device, error) {
	return d.requestDeviceSlice(d.url + "/label/" + url.QueryEscape(label))
}

// Get the devices that are on a service
func (d *DeviceRestClient) DevicesForService(serviceId string) ([]models.Device, error) {
	return d.requestDeviceSlice(d.url + "/service/" + serviceId)
}

// Get the devices that are on a service(by name)
func (d *DeviceRestClient) DevicesForServiceByName(serviceName string) ([]models.Device, error) {
	return d.requestDeviceSlice(d.url + "/servicename/" + url.QueryEscape(serviceName))
}

// Get the devices for a profile
func (d *DeviceRestClient) DevicesForProfile(profileId string) ([]models.Device, error) {
	return d.requestDeviceSlice(d.url + "/profile/" + profileId)
}

// Get the devices for a profile (by name)
func (d *DeviceRestClient) DevicesForProfileByName(profileName string) ([]models.Device, error) {
	return d.requestDeviceSlice(d.url + "/profilename/" + url.QueryEscape(profileName))
}

// Get the devices for an addressable
func (d *DeviceRestClient) DevicesForAddressable(addressableId string) ([]models.Device, error) {
	return d.requestDeviceSlice(d.url + "/addressable/" + addressableId)
}

// Get the devices for an addressable (by name)
func (d *DeviceRestClient) DevicesForAddressableByName(addressableName string) ([]models.Device, error) {
	return d.requestDeviceSlice(d.url + "/addressablename/" + url.QueryEscape(addressableName))
}

// Add a device - handle error codes
func (d *DeviceRestClient) Add(dev *models.Device) (string, error) {
	return clients.PostJsonRequest(d.url, dev)
}

// Update a device - handle error codes
func (d *DeviceRestClient) Update(dev models.Device) error {
	return clients.UpdateRequest(d.url, dev)
}

// Update the lastConnected value for a device (specified by id)
func (d *DeviceRestClient) UpdateLastConnected(id string, time int64) error {
	_, err := clients.PutRequest(d.url+"/"+id+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
	return err
}

// Update the lastConnected value for a device (specified by name)
func (d *DeviceRestClient) UpdateLastConnectedByName(name string, time int64) error {
	_, err := clients.PutRequest(d.url+"/name/"+url.QueryEscape(name)+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
	return err
}

// Update the lastReported value for a device (specified by id)
func (d *DeviceRestClient) UpdateLastReported(id string, time int64) error {
	_, err := clients.PutRequest(d.url+"/"+id+"/lastreported/"+strconv.FormatInt(time, 10), nil)
	return err
}

// Update the lastReported value for a device (specified by name)
func (d *DeviceRestClient) UpdateLastReportedByName(name string, time int64) error {
	_, err := clients.PutRequest(d.url+"/name/"+url.QueryEscape(name)+"/lastreported/"+strconv.FormatInt(time, 10), nil)
	return err
}

// Update the opState value for a device (specified by id)
func (d *DeviceRestClient) UpdateOpState(id string, opState string) error {
	_, err := clients.PutRequest(d.url+"/"+id+"/opstate/"+opState, nil)
	return err
}

// Update the opState value for a device (specified by name)
func (d *DeviceRestClient) UpdateOpStateByName(name string, opState string) error {
	_, err := clients.PutRequest(d.url+"/name/"+url.QueryEscape(name)+"/opstate/"+opState, nil)
	return err
}

// Update the adminState value for a device (specified by id)
func (d *DeviceRestClient) UpdateAdminState(id string, adminState string) error {
	_, err := clients.PutRequest(d.url+"/"+id+"/adminstate/"+adminState, nil)
	return err
}

// Update the adminState value for a device (specified by name)
func (d *DeviceRestClient) UpdateAdminStateByName(name string, adminState string) error {
	_, err := clients.PutRequest(d.url+"/name/"+url.QueryEscape(name)+"/adminstate/"+adminState, nil)
	return err
}

// Delete a device (specified by id)
func (d *DeviceRestClient) Delete(id string) error {
	return clients.DeleteRequest(d.url + "/id/" + id)
}

// Delete a device (specified by name)
func (d *DeviceRestClient) DeleteByName(name string) error {
	return clients.DeleteRequest(d.url + "/name/" + url.QueryEscape(name))
}
