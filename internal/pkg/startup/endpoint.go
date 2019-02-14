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
	"github.com/edgexfoundry/go-mod-registry"
	"os"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

type Endpoint struct {
	RegistryClient *registry.Client
}

func (e Endpoint) Monitor(params types.EndpointParams, ch chan string) {
	var endpoint registry.ServiceEndpoint
	var err error
	for {

		// TODO: Once consul cleanup complete add error logging if RegistryClient nil or can't be cast.
		if e.RegistryClient != nil {
			(*e.RegistryClient).GetServiceEndpoint(params.ServiceKey)
		} else {
			// TODO: remove the else when doing consul cleanup one all service have been changed to ues Registry abstraction and rename above to "Monitor"
			var ep consulclient.ServiceEndpoint

			ep, err = consulclient.GetServiceEndpoint(params.ServiceKey)
			if err == nil {
				endpoint.Host = ep.Address
				endpoint.Port = ep.Port
				endpoint.ServiceId = ep.Key
			}
		}

		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
		}
		url := fmt.Sprintf("http://%s:%v%s", endpoint.Host, endpoint.Port, params.Path)
		ch <- url
		time.Sleep(time.Millisecond * time.Duration(params.Interval))
	}
}
