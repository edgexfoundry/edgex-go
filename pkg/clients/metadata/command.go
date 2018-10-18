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

// Helper method to request and decode a command
func (c *CommandRestClient) requestCommand(url string) (models.Command, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return models.Command{}, err
	}

	com := models.Command{}
	err = json.Unmarshal(data, &com)
	return com, err
}

// Helper method to request and decode a command slice
func (c *CommandRestClient) requestCommandSlice(url string) ([]models.Command, error) {
	data, err := clients.GetRequest(url)
	if err != nil {
		return []models.Command{}, err
	}

	comSlice := make([]models.Command, 0)
	err = json.Unmarshal(data, &comSlice)
	return comSlice, err
}

// Get a command by id
func (c *CommandRestClient) Command(id string) (models.Command, error) {
	return c.requestCommand(c.url + "/" + id)
}

// Get a list of all the commands
func (c *CommandRestClient) Commands() ([]models.Command, error) {
	return c.requestCommandSlice(c.url)
}

// Get a list of commands for a certain name
func (c *CommandRestClient) CommandsForName(name string) ([]models.Command, error) {
	return c.requestCommandSlice(c.url + "/name/" + name)
}

// Add a new command
func (c *CommandRestClient) Add(com *models.Command) (string, error) {
	return clients.PostJsonRequest(c.url, com)
}

// Update a command
func (c *CommandRestClient) Update(com models.Command) error {
	return clients.UpdateRequest(c.url, com)
}

// Delete a command
func (c *CommandRestClient) Delete(id string) error {
	return clients.DeleteRequest(c.url + "/id/" + id)
}
