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
package command

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/core/clients/metadataclients"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
)

var loggingClient = logger.NewClient(COMMAND, false, "")

// CommandClient : client to interact with core command
type CommandClient interface {
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

// Get : issue GET command
func (cc *CommandRestClient) Get(id string, cID string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, cc.url+"/"+id+"/"+COMMAND+"/"+cID, nil)
	if err != nil {
		loggingClient.Error(err.Error())
		return "", err
	}

	resp, err := doReq(req)
	if err != nil {
		loggingClient.Error(err.Error())
		return "", err
	}
	defer resp.Body.Close()

	json, err := getBody(resp)
	if resp.StatusCode != http.StatusOK {
		if err != nil {
			loggingClient.Error(err.Error())
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
		loggingClient.Error(err.Error())
		return "", err
	}

	resp, err := doReq(req)
	if err != nil {
		loggingClient.Error(err.Error())
		return "", err
	}
	if resp == nil {
		loggingClient.Error(ErrResponseNil.Error())
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	json, err := getBody(resp)
	if err != nil {
		loggingClient.Error(err.Error())
		return "", err
	} else if resp.StatusCode != http.StatusOK {
		loggingClient.Error(string(json))
		return "", errors.New(string(json))
	}

	return string(json), err
}

func doReq(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		loggingClient.Error(err.Error())
	}
	return resp, err
}

func getBody(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		loggingClient.Error(err.Error())
		return []byte{}, err
	}

	return body, nil
}
