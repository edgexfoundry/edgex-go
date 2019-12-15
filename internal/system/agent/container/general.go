/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/system/agent/clients"

	"github.com/edgexfoundry/go-mod-bootstrap/di"
)

// GeneralClientsName contains the name of the clients.Clients implementation in the DIC.
var GeneralClientsName = di.TypeInstanceToName((*clients.General)(nil))

// GeneralClientsFrom helper function queries the DIC and returns the clients.General implementation.
func GeneralClientsFrom(get di.Get) *clients.General {
	return get(GeneralClientsName).(*clients.General)
}
