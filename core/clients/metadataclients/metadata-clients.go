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
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

var (
	ErrResponseNil error = errors.New("Problem connecting to metadata - reponse was nil")
	ErrNotFound    error = errors.New("Item not found")
)

/*
Addressable client for interacting with the addressable section of metadata
*/
type AddressableClient interface {
	Add(addr *models.Addressable) (string, error)
	AddressableForName(name string) (models.Addressable, error)
}

type AddressableRestClient struct {
	url string
}

/*
Device client for interacting with the device section of metadata
*/
type DeviceClient interface {
	Add(dev *models.Device) (string, error)
	Delete(id string) error
	DeleteByName(name string) error
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
	url string
}

/*
Command client for interacting with the command section of metadata
*/
type CommandClient interface {
	Add(com *models.Command) (string, error)
	Command(id string) (models.Command, error)
	Commands() ([]models.Command, error)
	CommandsForName(name string) ([]models.Command, error)
	Delete(id string) error
	Update(com models.Command) error
}

type CommandRestClient struct {
	url string
}

/*
Service client for interacting with the device service section of metadata
*/
type ServiceClient interface {
	Add(ds *models.DeviceService) (string, error)
	DeviceServiceForName(name string) (models.DeviceService, error)
	UpdateLastConnected(id string, time int64) error
	UpdateLastReported(id string, time int64) error
}

type ServiceRestClient struct {
	url string
}

// Device Profile client for interacting with the device profile section of metadata
type DeviceProfileClient interface {
	Add(dp *models.DeviceProfile) (string, error)
}

type DeviceProfileRestClient struct {
	url string
}

/*
Return an instance of AddressableClient
*/
func NewAddressableClient(metaDbAddressableUrl string) AddressableClient {
	a := AddressableRestClient{url: metaDbAddressableUrl}

	return &a
}

/*
Return an instance of DeviceClient
*/
func NewDeviceClient(metaDbDeviceUrl string) DeviceClient {
	d := DeviceRestClient{url: metaDbDeviceUrl}

	return &d
}

/*
Return an instance of CommandClient
*/
func NewCommandClient(metaDbCommandUrl string) CommandClient {
	c := CommandRestClient{url: metaDbCommandUrl}

	return &c
}

/*
Return an instance of ServiceClient
*/
func NewServiceClient(metaDbServiceUrl string) ServiceClient {
	s := ServiceRestClient{url: metaDbServiceUrl}

	return &s
}

// Return an instance of DeviceProfileClient
func NewDeviceProfileClient(metaDbDeviceProfileUrl string) DeviceProfileClient {
	d := DeviceProfileRestClient{url: metaDbDeviceProfileUrl}

	return &d
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
func (a *AddressableRestClient) Add(addr *models.Addressable) (string, error) {
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
func (d *AddressableRestClient) decodeAddressable(resp *http.Response) (models.Addressable, error) {
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
func (a *AddressableRestClient) AddressableForName(name string) (models.Addressable, error) {
	req, err := http.NewRequest(http.MethodGet, a.url+"/name/"+url.QueryEscape(name), nil)
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
		fmt.Printf("AddressableForName makeRequest failed: %v\n", err)
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
func (d *DeviceRestClient) decodeDeviceSlice(resp *http.Response) ([]models.Device, error) {
	dec := json.NewDecoder(resp.Body)
	dSlice := []models.Device{}

	err := dec.Decode(&dSlice)
	if err != nil {
		fmt.Println(err)
	}

	return dSlice, err
}

// Helper method to decode a device and return the device
func (d *DeviceRestClient) decodeDevice(resp *http.Response) (models.Device, error) {
	dec := json.NewDecoder(resp.Body)
	dev := models.Device{}

	err := dec.Decode(&dev)
	if err != nil {
		fmt.Println(err)
	}

	return dev, err
}

// Get the device by id
func (d *DeviceRestClient) Device(id string) (models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/"+id, nil)
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
func (d *DeviceRestClient) Devices() ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url, nil)
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
func (d *DeviceRestClient) DeviceForName(name string) (models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/name/"+url.QueryEscape(name), nil)
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
func (d *DeviceRestClient) DevicesByLabel(label string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/label/"+url.QueryEscape(label), nil)
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
func (d *DeviceRestClient) DevicesForService(serviceId string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/service/"+serviceId, nil)
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
func (d *DeviceRestClient) DevicesForServiceByName(serviceName string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/servicename/"+url.QueryEscape(serviceName), nil)
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
func (d *DeviceRestClient) DevicesForProfile(profileId string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/profile/"+profileId, nil)
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
func (d *DeviceRestClient) DevicesForProfileByName(profileName string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/profilename/"+url.QueryEscape(profileName), nil)
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
func (d *DeviceRestClient) DevicesForAddressable(addressableId string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/addressable/"+addressableId, nil)
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
func (d *DeviceRestClient) DevicesForAddressableByName(addressableName string) ([]models.Device, error) {
	req, err := http.NewRequest(http.MethodGet, d.url+"/addressablename/"+url.QueryEscape(addressableName), nil)
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
func (d *DeviceRestClient) Add(dev *models.Device) (string, error) {
	jsonStr, err := json.Marshal(dev)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, d.url, bytes.NewReader(jsonStr))
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
func (d *DeviceRestClient) Update(dev models.Device) error {
	jsonStr, err := json.Marshal(&dev)
	if err != nil {
		fmt.Println(err)
		return err
	}

	req, err := http.NewRequest(http.MethodPut, d.url, bytes.NewReader(jsonStr))
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
func (d *DeviceRestClient) UpdateLastConnected(id string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/"+id+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
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
func (d *DeviceRestClient) UpdateLastConnectedByName(name string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/name/"+url.QueryEscape(name)+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
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
func (d *DeviceRestClient) UpdateLastReported(id string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/"+id+"/lastreported/"+strconv.FormatInt(time, 10), nil)
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
func (d *DeviceRestClient) UpdateLastReportedByName(name string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/name/"+url.QueryEscape(name)+"/lastreported/"+strconv.FormatInt(time, 10), nil)
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
func (d *DeviceRestClient) UpdateOpState(id string, opState string) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/"+id+"/opstate/"+opState, nil)
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
func (d *DeviceRestClient) UpdateOpStateByName(name string, opState string) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/name/"+url.QueryEscape(name)+"/opstate/"+opState, nil)
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
func (d *DeviceRestClient) UpdateAdminState(id string, adminState string) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/"+id+"/adminstate/"+adminState, nil)
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
func (d *DeviceRestClient) UpdateAdminStateByName(name string, adminState string) error {
	req, err := http.NewRequest(http.MethodPut, d.url+"/name/"+url.QueryEscape(name)+"/adminstate/"+adminState, nil)
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
func (d *DeviceRestClient) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, d.url+"/id/"+id, nil)
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
func (d *DeviceRestClient) DeleteByName(name string) error {
	req, err := http.NewRequest(http.MethodDelete, d.url+"/name/"+url.QueryEscape(name), nil)
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
func (c *CommandRestClient) decodeCommand(resp *http.Response) (models.Command, error) {
	dec := json.NewDecoder(resp.Body)
	com := models.Command{}
	err := dec.Decode(&com)
	if err != nil {
		fmt.Println(err)
	}

	return com, err
}

// Helper method to decode and return a command slice
func (c *CommandRestClient) decodeCommandSlice(resp *http.Response) ([]models.Command, error) {
	dec := json.NewDecoder(resp.Body)
	comSlice := []models.Command{}
	err := dec.Decode(&comSlice)
	if err != nil {
		fmt.Println(err)
	}

	return comSlice, err
}

// Get a command by id
func (c *CommandRestClient) Command(id string) (models.Command, error) {
	req, err := http.NewRequest(http.MethodGet, c.url+"/"+id, nil)
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
func (c *CommandRestClient) Commands() ([]models.Command, error) {
	req, err := http.NewRequest(http.MethodGet, c.url, nil)
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
func (c *CommandRestClient) CommandsForName(name string) ([]models.Command, error) {
	req, err := http.NewRequest(http.MethodGet, c.url+"/name/"+name, nil)
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
func (c *CommandRestClient) Add(com *models.Command) (string, error) {
	jsonStr, err := json.Marshal(com)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(jsonStr))
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
func (c *CommandRestClient) Update(com models.Command) error {
	jsonStr, err := json.Marshal(&com)
	if err != nil {
		fmt.Println(err)
		return err
	}

	req, err := http.NewRequest(http.MethodPut, c.url, bytes.NewReader(jsonStr))
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
func (c *CommandRestClient) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, c.url+"/id/"+id, nil)
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
func (s *ServiceRestClient) decodeDeviceService(resp *http.Response) (models.DeviceService, error) {
	dec := json.NewDecoder(resp.Body)
	ds := models.DeviceService{}
	err := dec.Decode(&ds)
	if err != nil {
		fmt.Println(err)
	}

	return ds, err
}

// Update the last connected time for the device service
func (s *ServiceRestClient) UpdateLastConnected(id string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, s.url+"/"+id+"/lastconnected/"+strconv.FormatInt(time, 10), nil)
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
func (s *ServiceRestClient) UpdateLastReported(id string, time int64) error {
	req, err := http.NewRequest(http.MethodPut, s.url+"/"+id+"/lastreported/"+strconv.FormatInt(time, 10), nil)
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
func (s *ServiceRestClient) Add(ds *models.DeviceService) (string, error) {
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
func (s *ServiceRestClient) DeviceServiceForName(name string) (models.DeviceService, error) {
	req, err := http.NewRequest(http.MethodGet, s.url+"/name/"+name, nil)
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
func (dpc *DeviceProfileRestClient) Add(dp *models.DeviceProfile) (string, error) {
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
