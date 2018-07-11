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
	"net/http"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

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
	url      string
	endpoint clients.Endpointer
}

/*
Return an instance of CommandClient
*/
func NewCommandClient(params types.EndpointParams, m clients.Endpointer) CommandClient {
	c := CommandRestClient{endpoint: m}
	c.init(params)
	return &c
}

func (c *CommandRestClient) init(params types.EndpointParams) {
	if params.UseRegistry {
		ch := make(chan string, 1)
		go c.endpoint.Monitor(params, ch)
		go func(ch chan string) {
			for {
				select {
				case url := <-ch:
					c.url = url
				}
			}
		}(ch)
	} else {
		c.url = params.Url
	}
}

// Helper method to decode and return a command
func (c *CommandRestClient) decodeCommand(resp *http.Response) (models.Command, error) {
	dec := json.NewDecoder(resp.Body)
	com := models.Command{}
	err := dec.Decode(&com)
	if err != nil {
		return models.Command{}, err
	}

	return com, err
}

// Helper method to decode and return a command slice
func (c *CommandRestClient) decodeCommandSlice(resp *http.Response) ([]models.Command, error) {
	dec := json.NewDecoder(resp.Body)
	comSlice := []models.Command{}
	err := dec.Decode(&comSlice)
	if err != nil {
		return []models.Command{}, err
	}

	return comSlice, err
}

// Get a command by id
func (c *CommandRestClient) Command(id string) (models.Command, error) {
	req, err := http.NewRequest(http.MethodGet, c.url+"/"+id, nil)
	if err != nil {
		return models.Command{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return models.Command{}, err
	}
	if resp == nil {
		return models.Command{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
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
		return []models.Command{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return []models.Command{}, err
	}
	if resp == nil {
		return []models.Command{}, ErrResponseNil
	}

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
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
		return []models.Command{}, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return []models.Command{}, err
	}
	if resp == nil {
		return []models.Command{}, ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Get the response body
		bodyBytes, err := getBody(resp)
		if err != nil {
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
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(jsonStr))
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

	// Get the response body
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

// Update a command
func (c *CommandRestClient) Update(com models.Command) error {
	jsonStr, err := json.Marshal(&com)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, c.url, bytes.NewReader(jsonStr))
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

// Delete a command
func (c *CommandRestClient) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, c.url+"/id/"+id, nil)
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
