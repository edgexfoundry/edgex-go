//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/metadata"
)

// MetadataDeviceClientName contains the name of the Metadata device client instance in the DIC.
var MetadataDeviceClientName = "V2MetadataDeviceClient"

// MetadataDeviceClientFrom helper function queries the DIC and returns the Metadata device client instance.
func MetadataDeviceClientFrom(get di.Get) metadata.DeviceClient {
	return get(MetadataDeviceClientName).(metadata.DeviceClient)
}
