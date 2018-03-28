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
 * @author: Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/
package commandclients

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/core/clients/metadataclients"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

// CommandClient : client to interact with core command
type CommandClient interface {
	Devices() ([]models.Device, error)
	Device(id string) (models.Device, error)
	DeviceByName(n string) (models.Device, error)
	Get(id string, cID string) (string, error)
	Put(id string, cID string, body string) (string, error)
}

type CommandRestClient struct {
	url string
}

// NewCommandClient : Create an instance of CommandClient
func NewCommandClient(commandURL string) CommandClient {
	c := CommandRestClient{url: commandURL}
	return &c
}

// Devices : return all Devices
func (cc *CommandRestClient) Devices() ([]models.Device, error) {
	dc := metadataclients.NewDeviceClient(cc.url)

	return dc.Devices()
}

// Device : return device by id
func (cc *CommandRestClient) Device(id string) (models.Device, error) {
	dc := metadataclients.NewDeviceClient(cc.url)
	return dc.Device(id)
}

// DeviceByName : return device by name
func (cc *CommandRestClient) DeviceByName(n string) (models.Device, error) {
	dc := metadataclients.NewDeviceClient(cc.url)
	return dc.DeviceForName(n)
}

// Get : issue GET command
func (cc *CommandRestClient) Get(id string, cID string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, cc.url+"/"+id+"/"+COMMAND+"/"+cID, nil)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	resp, err := doReq(req)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	defer resp.Body.Close()

	json, err := getBody(resp)
	if resp.StatusCode != http.StatusOK {
		if err != nil {
			fmt.Println(err.Error())
			return "", err
		}
		return "", errors.New(string(json))
	}
	return string(json), err
}

// Put : Issue PUT command
func (cc *CommandRestClient) Put(id string, cID string, body string) (string, error) {
	req, err := http.NewRequest(http.MethodPut, cc.url+"/"+id+"/"+COMMAND+"/"+cID, strings.NewReader(body))
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	resp, err := doReq(req)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	if resp == nil {
		fmt.Println(ErrResponseNil)
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	json, err := getBody(resp)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	} else if resp.StatusCode != http.StatusOK {
		fmt.Println(errors.New(string(json)))
		return "", errors.New(string(json))
	}

	return string(json), err
}

func doReq(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	return resp, err
}

func getBody(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}

	return body, nil
}
