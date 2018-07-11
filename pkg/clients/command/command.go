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
package command

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

var (
	COMMAND    = "command"
	JSONHEADER = "application/json"

	ErrResponseNil       = errors.New("Response was nil")
	ErrorCommandNotFound = errors.New("Command not found")
)

// CommandClient : client to interact with core command
type CommandClient interface {
	Get(id string, cID string) (string, error)
	Put(id string, cID string, body string) (string, error)
}

type CommandRestClient struct {
	url      string
	endpoint clients.Endpointer
}

// NewCommandClient : Create an instance of CommandClient
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

// Get : issue GET command
func (cc *CommandRestClient) Get(id string, cID string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, cc.url+"/"+id+"/"+COMMAND+"/"+cID, nil)
	if err != nil {
		return "", err
	}

	resp, err := doReq(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	json, err := getBody(resp)
	if resp.StatusCode != http.StatusOK {
		if err != nil {
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
		return "", err
	}

	resp, err := doReq(req)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	json, err := getBody(resp)
	if err != nil {
		return "", err
	} else if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(json))
	}

	return string(json), err
}

func doReq(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)

	return resp, err
}

func getBody(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}
