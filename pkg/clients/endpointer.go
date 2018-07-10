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
package clients

import "github.com/edgexfoundry/edgex-go/pkg/clients/types"

//Endpointer is the interface for types that need to implement or simulate integration
//with a service discovery provider.
type Endpointer interface {
	//Monitor is responsible for looking up information about the service endpoint corresponding
	//to the params.ServiceKey property. The name "Monitor" implies that this lookup will be done
	//at a regular interval. Information about the service from the discovery provider should be
	//used to construct a URL which will then be pushed to the supplied channel.
	Monitor(params types.EndpointParams, ch chan string)
}
