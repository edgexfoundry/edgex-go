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
package metadataclients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/edgexfoundry/core-domain-go/models"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var (
	ErrResponseNil error = errors.New("Problem connecting to metadata - reponse was nil")
	ErrNotFound    error = errors.New("Item not found")
)

/*
Addressable client for interacting with the addressable section of metadata
*/
type AddressableClient struct {
	url string
}

/*
Device client for interacting with the device section of metadata
*/
type DeviceClient struct {
	url string
}

/*
Command client for interacting with the command section of metadata
*/
type CommandClient struct {
	url string
}

/*
Service client for interacting with the device service section of metadata
*/
type ServiceClient struct {
	url string
}

// Device Profile client for interacting with the device profile section of metadata
type DeviceProfileClient struct {
	url string
}

/*
Return an instance of AddressableClient
*/
func NewAddressableClient(metaDbAddressableUrl string) AddressableClient {
	return AddressableClient{url: metaDbAddressableUrl}
}

/*
Return an instance of DeviceClient
*/
func NewDeviceClient(metaDbDeviceUrl string) DeviceClient {
	return DeviceClient{url: metaDbDeviceUrl}
}

/*
Return an instance of CommandClient
*/
func NewCommandClient(metaDbCommandUrl string) CommandClient {
	return CommandClient{url: metaDbCommandUrl}
}

/*
Return an instance of ServiceClient
*/
func NewServiceClient(metaDbServiceUrl string) ServiceClient {
	return ServiceClient{url: metaDbServiceUrl}
}

// Return an instance of DeviceProfileClient
func NewDeviceProfileClient(metaDbDeviceProfileUrl string) DeviceProfileClient {
	return DeviceProfileClient{url: metaDbDeviceProfileUrl}
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

// Helper method to get the body from the response after making the request
func getBody(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}

	return body, nil
}

// ***************** ADDRESSABLE CLIENT METHODS ***********************

// Add an addressable - handle error codes
// Returns the ID of the addressable and an error
func (a *AddressableClient) Add(addr *models.Addressable) (string, error) {
	// Marshal the addressable to JSON
	jsonStr, err := json.Marshal(addr)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Post(a.url, "application/json", bytes.NewReader(jsonStr))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil.Error())
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
		return "", errors.New(string(bodyString))
	}

	return bodyString, err
}

// Helper method to decode an addressable and return the addressable
func (d *AddressableClient) decodeAddressable(resp *http.Response) (models.Addressable, error) {
	dec := json.NewDecoder(resp.Body)
	addr := models.Addressable{}

	err := dec.Decode(&addr)
	if err != nil {
		fmt.Println(err)
	}

	return addr, err
}

// TODO: make method signatures consistent wrt to error return value
// ie. use it everywhere, or not at all!

// Get the addressable by name
func (a *AddressableClient) AddressableForName(name string) (models.Addressable, error) {
	req, err := http.NewRequest("GET", a.url+"/name/"+url.QueryEscape(name), nil)
	if err != nil {
		fmt.Println(err)
		return models.Addressable{}, err
	}

	resp, err := makeRequest(req)

	// Check response
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return models.Addressable{}, ErrResponseNil
	}
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("AddressableForName makeRequest failed: %s", err)
		return models.Addressable{}, err
	}

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return models.Addressable{}, err
		}
		bodyString := string(bodyBytes)

		fmt.Println(bodyString)
		return models.Addressable{}, errors.New(bodyString)
	}

	return a.decodeAddressable(resp)
}

// ***************** DEVICE CLIENT METHODS ***********************
// Help method to decode a device slice
func (d *DeviceClient) decodeDeviceSlice(resp *http.Response) ([]models.Device, error) {
	dec := json.NewDecoder(resp.Body)
	dSlice := []models.Device{}

	err := dec.Decode(&dSlice)
	if err != nil {
		fmt.Println(err)
	}

	return dSlice, err
}

// Helper method to decode a device and return the device
func (d *DeviceClient) decodeDevice(resp *http.Response) (models.Device, error) {
	dec := json.NewDecoder(resp.Body)
	dev := models.Device{}

	err := dec.Decode(&dev)
	if err != nil {
		fmt.Println(err)
	}

	return dev, err
}

// Get the device by id
func (d *DeviceClient) Device(id string) (models.Device, error) {
	req, err := http.NewRequest("GET", d.url+"/"+id, nil)
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

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return models.Device{}, errors.New(bodyString)
	}

	return d.decodeDevice(resp)
}

// Get a list of all devices
func (d *DeviceClient) Devices() ([]models.Device, error) {
	req, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		fmt.Println(err)
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Device{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the device by name
func (d *DeviceClient) DeviceForName(name string) (models.Device, error) {
	req, err := http.NewRequest("GET", d.url+"/name/"+url.QueryEscape(name), nil)
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

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return models.Device{}, errors.New(bodyString)
	}
	return d.decodeDevice(resp)
}

// Get the device by label
func (d *DeviceClient) DevicesByLabel(label string) ([]models.Device, error) {
	req, err := http.NewRequest("GET", d.url+"/label/"+url.QueryEscape(label), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Device{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the devices that are on a service
func (d *DeviceClient) DevicesForService(serviceId string) ([]models.Device, error) {
	req, err := http.NewRequest("GET", d.url+"/service/"+serviceId, nil)
	if err != nil {
		fmt.Println(err)
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Device{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the devices that are on a service(by name)
func (d *DeviceClient) DevicesForServiceByName(serviceName string) ([]models.Device, error) {
	req, err := http.NewRequest("GET", d.url+"/servicename/"+url.QueryEscape(serviceName), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Device{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the devices for a profile
func (d *DeviceClient) DevicesForProfile(profileId string) ([]models.Device, error) {
	req, err := http.NewRequest("GET", d.url+"/profile/"+profileId, nil)
	if err != nil {
		fmt.Println(err)
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Device{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the devices for a profile (by name)
func (d *DeviceClient) DevicesForProfileByName(profileName string) ([]models.Device, error) {
	req, err := http.NewRequest("GET", d.url+"/profilename/"+url.QueryEscape(profileName), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Device{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}
	return d.decodeDeviceSlice(resp)
}

// Get the devices for an addressable
func (d *DeviceClient) DevicesForAddressable(addressableId string) ([]models.Device, error) {
	req, err := http.NewRequest("GET", d.url+"/addressable/"+addressableId, nil)
	if err != nil {
		fmt.Println(err)
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Device{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}

	return d.decodeDeviceSlice(resp)
}

// Get the devices for an addressable (by name)
func (d *DeviceClient) DevicesForAddressableByName(addressableName string) ([]models.Device, error) {
	req, err := http.NewRequest("GET", d.url+"/addressablename/"+url.QueryEscape(addressableName), nil)
	if err != nil {
		fmt.Println(err)
		return []models.Device{}, err
	}

	// Make the request and get response
	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return []models.Device{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Device{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Device{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Device{}, errors.New(bodyString)
	}

	return d.decodeDeviceSlice(resp)
}

// Add a device - handle error codes
func (d *DeviceClient) Add(dev *models.Device) (string, error) {
	jsonStr, err := json.Marshal(dev)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req, err := http.NewRequest("POST", d.url, bytes.NewReader(jsonStr))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	// Get the body
	bodyBytes, err := getBody(resp)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	bodyString := string(bodyBytes)

	if resp.StatusCode != 200 {
		fmt.Println(bodyString)
		return "", errors.New(bodyString)
	}

	return bodyString, nil
}

// Update a device - handle error codes
func (d *DeviceClient) Update(dev models.Device) error {
	jsonStr, err := json.Marshal(&dev)
	if err != nil {
		fmt.Println(err)
		return err
	}

	req, err := http.NewRequest("PUT", d.url, bytes.NewReader(jsonStr))
	if err != nil {
		fmt.Println(err)
		return err
	}

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
		// Get the response body
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

// Update the lastConnected value for a device (specified by id)
func (d *DeviceClient) UpdateLastConnected(id string, time int64) error {
	req, err := http.NewRequest("PUT", d.url+"/"+id+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

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
		// Get the response body
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

// Update the lastConnected value for a device (specified by name)
func (d *DeviceClient) UpdateLastConnectedByName(name string, time int64) error {
	req, err := http.NewRequest("PUT", d.url+"/name/"+url.QueryEscape(name)+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

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
		// Get the response body
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

// Update the lastReported value for a device (specified by id)
func (d *DeviceClient) UpdateLastReported(id string, time int64) error {
	req, err := http.NewRequest("PUT", d.url+"/"+id+"/lastreported/"+strconv.FormatInt(time, 10), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

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
		// Get the response body
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

// Update the lastReported value for a device (specified by name)
func (d *DeviceClient) UpdateLastReportedByName(name string, time int64) error {
	req, err := http.NewRequest("PUT", d.url+"/name/"+url.QueryEscape(name)+"/lastreported/"+strconv.FormatInt(time, 10), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return err
}

// Update the opState value for a device (specified by id)
func (d *DeviceClient) UpdateOpState(id string, opState string) error {
	req, err := http.NewRequest("PUT", d.url+"/"+id+"/opstate/"+opState, nil)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return err
}

// Update the opState value for a device (specified by name)
func (d *DeviceClient) UpdateOpStateByName(name string, opState string) error {
	req, err := http.NewRequest("PUT", d.url+"/name/"+url.QueryEscape(name)+"/opstate/"+opState, nil)
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
		// Get the response body
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

// Update the adminState value for a device (specified by id)
func (d *DeviceClient) UpdateAdminState(id string, adminState string) error {
	req, err := http.NewRequest("PUT", d.url+"/"+id+"/adminstate/"+adminState, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

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
		// Get the response body
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

// Update the adminState value for a device (specified by name)
func (d *DeviceClient) UpdateAdminStateByName(name string, adminState string) error {
	req, err := http.NewRequest("PUT", d.url+"/name/"+url.QueryEscape(name)+"/adminstate/"+adminState, nil)
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
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		bodyString := string(bodyBytes)

		return errors.New(bodyString)
	}

	return err
}

// Delete a device (specified by id)
func (d *DeviceClient) Delete(id string) error {
	req, err := http.NewRequest("DELETE", d.url+"/id/"+id, nil)
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
		// Get the response body
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

// Delete a device (specified by name)
func (d *DeviceClient) DeleteByName(name string) error {
	req, err := http.NewRequest("DELETE", d.url+"/name/"+url.QueryEscape(name), nil)
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
		// Get the response body
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

// ************************** COMMAND CLIENT METHODS ****************************

// Helper method to decode and return a command
func (c *CommandClient) decodeCommand(resp *http.Response) (models.Command, error) {
	dec := json.NewDecoder(resp.Body)
	com := models.Command{}
	err := dec.Decode(&com)
	if err != nil {
		fmt.Println(err)
	}

	return com, err
}

// Helper method to decode and return a command slice
func (c *CommandClient) decodeCommandSlice(resp *http.Response) ([]models.Command, error) {
	dec := json.NewDecoder(resp.Body)
	comSlice := []models.Command{}
	err := dec.Decode(&comSlice)
	if err != nil {
		fmt.Println(err)
	}

	return comSlice, err
}

// Get a command by id
func (c *CommandClient) Command(id string) (models.Command, error) {
	req, err := http.NewRequest("GET", c.url+"/"+id, nil)
	if err != nil {
		fmt.Println(err)
		return models.Command{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return models.Command{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return models.Command{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return models.Command{}, err
		}
		bodyString := string(bodyBytes)

		return models.Command{}, errors.New(bodyString)
	}

	return c.decodeCommand(resp)
}

// Get a list of all the commands
func (c *CommandClient) Commands() ([]models.Command, error) {
	req, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		fmt.Println(err)
		return []models.Command{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return []models.Command{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Command{}, ErrResponseNil
	}

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Command{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Command{}, errors.New(bodyString)
	}

	return c.decodeCommandSlice(resp)
}

// Get a list of commands for a certain name
func (c *CommandClient) CommandsForName(name string) ([]models.Command, error) {
	req, err := http.NewRequest("GET", c.url+"/name/"+name, nil)
	if err != nil {
		fmt.Println(err)
		return []models.Command{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		fmt.Println(err)
		return []models.Command{}, err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return []models.Command{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return []models.Command{}, err
		}
		bodyString := string(bodyBytes)

		return []models.Command{}, errors.New(bodyString)
	}

	return c.decodeCommandSlice(resp)
}

// Add a new command
func (c *CommandClient) Add(com *models.Command) (string, error) {
	jsonStr, err := json.Marshal(com)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(jsonStr))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

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

// Update a command
func (c *CommandClient) Update(com models.Command) error {
	jsonStr, err := json.Marshal(&com)
	if err != nil {
		fmt.Println(err)
		return err
	}

	req, err := http.NewRequest("PUT", c.url, bytes.NewReader(jsonStr))
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
		// Get the response body
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

// Delete a command
func (c *CommandClient) Delete(id string) error {
	req, err := http.NewRequest("DELETE", c.url+"/id/"+id, nil)
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
		// Get the response body
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

// ********************** SERVICE CLIENT METHODS **************************
// Helper method to decode and return a deviceservice
func (s *ServiceClient) decodeDeviceService(resp *http.Response) (models.DeviceService, error) {
	dec := json.NewDecoder(resp.Body)
	ds := models.DeviceService{}
	err := dec.Decode(&ds)
	if err != nil {
		fmt.Println(err)
	}

	return ds, err
}

// Update the last connected time for the device service
func (s *ServiceClient) UpdateLastConnected(id string, time int64) error {
	req, err := http.NewRequest("PUT", s.url+"/"+id+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
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
		// Get the response body
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

// Update the last reported time for the device service
func (s *ServiceClient) UpdateLastReported(id string, time int64) error {
	req, err := http.NewRequest("PUT", s.url+"/"+id+"/lastreported/"+strconv.FormatInt(time, 10), nil)
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
		// Get the response body
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

// Add a new deviceservice
func (s *ServiceClient) Add(ds *models.DeviceService) (string, error) {
	jsonStr, err := json.Marshal(ds)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Post(s.url, "application/json", bytes.NewReader(jsonStr))
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
		fmt.Println(bodyString)
		return "", errors.New(bodyString)
	}

	return bodyString, nil
}

// Request deviceservice for specified name
func (s *ServiceClient) DeviceServiceForName(name string) (models.DeviceService, error) {
	req, err := http.NewRequest("GET", s.url+"/name/"+name, nil)
	if err != nil {
		fmt.Printf("DeviceServiceForName NewRequest failed: %v\n", err)
		return models.DeviceService{}, err
	}

	resp, err := makeRequest(req)
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return models.DeviceService{}, ErrResponseNil
	}
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("DeviceServiceForName makeRequest failed: %v\n", err)
		return models.DeviceService{}, err
	}

	if resp.StatusCode != 200 {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
			fmt.Println(err.Error())
			return models.DeviceService{}, err
		}
		bodyString := string(bodyBytes)

		return models.DeviceService{}, errors.New(bodyString)
	}

	return s.decodeDeviceService(resp)
}

// ***************** DEVICE PROFILE METHODS *************************

// Add a new device profile to metadata
func (dpc *DeviceProfileClient) Add(dp *models.DeviceProfile) (string, error) {
	jsonStr, err := json.Marshal(dp)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Post(dpc.url, "application/json", bytes.NewReader(jsonStr))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	// Get the response
	bodyBytes, err := getBody(resp)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	bodyString := string(bodyBytes)

	// Check the response code
	if resp.StatusCode != 200 {
		fmt.Println(bodyString)
		return "", errors.New(bodyString)
	}

	return bodyString, nil
}
