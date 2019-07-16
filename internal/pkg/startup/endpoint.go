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
package startup

import (
	"fmt"
	"os"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	registryTypes "github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

type Endpoint struct {
	RegistryClient *registry.Client
}

func (e Endpoint) Monitor(params types.EndpointParams, ch chan string) {
	var endpoint registryTypes.ServiceEndpoint
	var err error
	for {

		if e.RegistryClient != nil {
			endpoint, err = (*e.RegistryClient).GetServiceEndpoint(params.ServiceKey)
			if err != nil {
				err = fmt.Errorf("unable to get Service endpoint for %s: %s", params.ServiceKey, err.Error())
			}
		} else {
			err = fmt.Errorf("unable to get Service endpoint for %s: Registry client is nil", params.ServiceKey)
		}

		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
		}

		var url string
		if endpoint.Host != "" {
			url = fmt.Sprintf("http://%s:%v%s", endpoint.Host, endpoint.Port, params.Path)
		} else {
			url = params.Url
		}

		ch <- url
		time.Sleep(time.Millisecond * time.Duration(params.Interval))
	}
}
