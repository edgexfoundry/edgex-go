/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright (c) 2019 Intel Corporation
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
package endpoint

import (
	"fmt"
	"os"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

type Endpoint struct {
	RegistryClient *registry.Client
}

func (e Endpoint) Monitor(params types.EndpointParams) chan string {
	ch := make(chan string, 1)
	go func() {
		for {
			ch <- e.fetch(params)
			time.Sleep(time.Millisecond * time.Duration(params.Interval))
		}
	}()
	return ch
}

func (e Endpoint) fetch(params types.EndpointParams) string {
	url, err := e.buildURL(params)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout, err.Error())
	}
	return url
}

func (e Endpoint) buildURL(params types.EndpointParams) (string, error) {
	if e.RegistryClient != nil {
		endpoint, err := (*e.RegistryClient).GetServiceEndpoint(params.ServiceKey)
		if err != nil {
			return "", fmt.Errorf("unable to get Service endpoint for %s: %s", params.ServiceKey, err.Error())
		}
		return fmt.Sprintf("http://%s:%v%s", endpoint.Host, endpoint.Port, params.Path), nil
	} else {
		return "", fmt.Errorf("unable to get Service endpoint for %s: Registry client is nil", params.ServiceKey)
	}
}
