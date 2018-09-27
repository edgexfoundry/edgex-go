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
package distro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type DistroClient interface {
	NotifyRegistrations(models.NotifyUpdate) error
}

type distroRestClient struct {
	url      string
	endpoint clients.Endpointer
}

func NewDistroClient(params types.EndpointParams, m clients.Endpointer) DistroClient {
	d := distroRestClient{endpoint: m}
	d.init(params)
	return &d
}

func (d *distroRestClient) init(params types.EndpointParams) {
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

func (d *distroRestClient) NotifyRegistrations(update models.NotifyUpdate) error {
	client := &http.Client{}
	url := d.url + clients.ApiNotifyRegistrationRoute

	data, err := json.Marshal(update)
	if err != nil {
		return errors.New(fmt.Sprintf("error marshaling to json: %s", err.Error()))
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		json, _ := getBody(resp)
		return types.NewErrServiceClient(resp.StatusCode, json)
	}
	return nil
}

func getBody(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}
