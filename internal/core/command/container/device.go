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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/metadata"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// MetadataDeviceClientName contains the name of the client implementation in the DIC.
var MetadataDeviceClientName = di.TypeInstanceToName((*metadata.DeviceClient)(nil))

// MetadataDeviceClientFrom helper function queries the DIC and returns the client implementation.
func MetadataDeviceClientFrom(get di.Get) metadata.DeviceClient {
	return get(MetadataDeviceClientName).(metadata.DeviceClient)
}
