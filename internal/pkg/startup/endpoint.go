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
package startup

import (
	"fmt"
	"os"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/internal/pkg/registry"
	registryTypes "github.com/edgexfoundry/edgex-go/internal/pkg/registry/types"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
)

type Endpoint struct{}

func (e Endpoint) Monitor(params types.EndpointParams, ch chan string) {
	var endpoint registryTypes.ServiceEndpoint
	var err error
	for {
		if registry.Client != nil {
			endpoint, err = registry.Client.GetServiceEndpoint(params.ServiceKey)
		} else {
			// TODO: remove the else when doing consul cleanup one all service have been changed to ues Registry abstraction.
			endpoint, err = consulclient.GetServiceEndpoint(params.ServiceKey)
		}
		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
		}
		url := fmt.Sprintf("http://%s:%v%s", endpoint.Address, endpoint.Port, params.Path)
		ch <- url
		time.Sleep(time.Millisecond * time.Duration(params.Interval))
	}
}
