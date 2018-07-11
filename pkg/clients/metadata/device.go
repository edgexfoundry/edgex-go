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

// Help method to decode a device slice
func (d *DeviceRestClient) decodeDeviceSlice(resp *http.Response) ([]models.Device, error) {
	dec := json.NewDecoder(resp.Body)
	dSlice := []models.Device{}

	err := dec.Decode(&dSlice)
	if err != nil {
		return []models.Device{}, err
	}

	return dSlice, err
}

// Helper method to decode a device and return the device
func (d *DeviceRestClient) decodeDevice(resp *http.Response) (models.Device, error) {
	dec := json.NewDecoder(resp.Body)
	dev := models.Device{}

	err := dec.Decode(&dev)
	if err != nil {
		return models.Device{}, err
	}

	return dev, err
}

//Use the models.Event.Device property for the supplied token parameter.
//The above property is currently double-purposed and needs to be refactored.
//This call replaces the previous two calls necessary to lookup a device by id followed by name.
func (d *DeviceRestClient) CheckForDevice(token string) (models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/check/"+token, nil)
	if err != nil {
		fmt.Println(err)
		return models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return models.Device{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return models.Device{}, types.ErrNotFound{}
	case http.StatusOK:
		return d.decodeDevice(resp)
	}

	// Unexpected http status. Get the response body
	bodyBytes, err := getBody(resp)
	if err != nil {
		fmt.Println(err.Error())
		return models.Device{}, err
	}
	bodyString := string(bodyBytes)

	return models.Device{}, errors.New(bodyString)
}

// Get the device by id
func (d *DeviceRestClient) Device(id string) (models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/"+id, nil)
	if err != nil {
		return models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return models.Device{}, err
	}
	if resp == nil {
		return models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return models.Device{}, errors.New(bodyString)
	}

	return d.decodeDevice(resp)
}

// Get a list of all devices
func (d *DeviceRestClient) Devices() ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url, nil)
	if err != nil {
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.Device{}, err
	}
	if resp == nil {
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the device by name
func (d *DeviceRestClient) DeviceForName(name string) (models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/name/"+url.QueryEscape(name), nil)
	if err != nil {
		return models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return models.Device{}, err
	}
	if resp == nil {
		return models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return models.Device{}, errors.New(bodyString)
	}
	return d.decodeDevice(resp)
}

// Get the device by label
func (d *DeviceRestClient) DevicesByLabel(label string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/label/"+url.QueryEscape(label), nil)
	if err != nil {
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.Device{}, err
	}
	if resp == nil {
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the devices that are on a service
func (d *DeviceRestClient) DevicesForService(serviceId string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/service/"+serviceId, nil)
	if err != nil {
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.Device{}, err
	}
	if resp == nil {
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the devices that are on a service(by name)
func (d *DeviceRestClient) DevicesForServiceByName(serviceName string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/servicename/"+url.QueryEscape(serviceName), nil)
	if err != nil {
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.Device{}, err
	}
	if resp == nil {
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the devices for a profile
func (d *DeviceRestClient) DevicesForProfile(profileId string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/profile/"+profileId, nil)
	if err != nil {
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.Device{}, err
	}
	if resp == nil {
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the devices for a profile (by name)
func (d *DeviceRestClient) DevicesForProfileByName(profileName string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/profilename/"+url.QueryEscape(profileName), nil)
	if err != nil {
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.Device{}, err
	}
	if resp == nil {
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the devices for an addressable
func (d *DeviceRestClient) DevicesForAddressable(addressableId string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/addressable/"+addressableId, nil)
	if err != nil {
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.Device{}, err
	}
	if resp == nil {
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}

	return d.decodeDeviceSlice(resp)
}

// Get the devices for an addressable (by name)
func (d *DeviceRestClient) DevicesForAddressableByName(addressableName string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/addressablename/"+url.QueryEscape(addressableName), nil)
	if err != nil {
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		return []models.Device{}, err
	}
	if resp == nil {
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}

	return d.decodeDeviceSlice(resp)
}

// Add a device - handle error codes
func (d *DeviceRestClient) Add(dev *models.Device) (string, error) {
	jsonStr, err := json.Marshal(dev)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, d.url, bytes.NewReader(jsonStr))
	if err != nil {
		return "", err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	// Get the body
	bodyBytes, err := getBody(resp)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(bodyString)
	}

	return bodyString, nil
}

// Update a device - handle error codes
func (d *DeviceRestClient) Update(dev models.Device) error {
	jsonStr, err := json.Marshal(&dev)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, d.url, bytes.NewReader(jsonStr))
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Update the lastConnected value for a device (specified by id)
func (d *DeviceRestClient) UpdateLastConnected(id string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/"+id+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Update the lastConnected value for a device (specified by name)
func (d *DeviceRestClient) UpdateLastConnectedByName(name string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/name/"+url.QueryEscape(name)+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Update the lastReported value for a device (specified by id)
func (d *DeviceRestClient) UpdateLastReported(id string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/"+id+"/lastreported/"+strconv.FormatInt(time, 10), nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Update the lastReported value for a device (specified by name)
func (d *DeviceRestClient) UpdateLastReportedByName(name string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/name/"+url.QueryEscape(name)+"/lastreported/"+strconv.FormatInt(time, 10), nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return err
}

// Update the opState value for a device (specified by id)
func (d *DeviceRestClient) UpdateOpState(id string, opState string) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/"+id+"/opstate/"+opState, nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return err
}

// Update the opState value for a device (specified by name)
func (d *DeviceRestClient) UpdateOpStateByName(name string, opState string) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/name/"+url.QueryEscape(name)+"/opstate/"+opState, nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Update the adminState value for a device (specified by id)
func (d *DeviceRestClient) UpdateAdminState(id string, adminState string) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/"+id+"/adminstate/"+adminState, nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Update the adminState value for a device (specified by name)
func (d *DeviceRestClient) UpdateAdminStateByName(name string, adminState string) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/name/"+url.QueryEscape(name)+"/adminstate/"+adminState, nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return err
}

// Delete a device (specified by id)
func (d *DeviceRestClient) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, d.url+"/id/"+id, nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}

// Delete a device (specified by name)
func (d *DeviceRestClient) DeleteByName(name string) error {
	req, err := http.NewRequest(http.MethodDelete, d.url+"/name/"+url.QueryEscape(name), nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return nil
}
