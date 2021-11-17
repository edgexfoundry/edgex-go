//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
)

// MetadataDeviceClientName contains the name of the Metadata DeviceClient instance in the DIC.
var MetadataDeviceClientName = di.TypeInstanceToName((*interfaces.DeviceClient)(nil))

// MetadataDeviceProfileClientName contains the name of the Metadata DeviceProfileClient instance in the DIC.
var MetadataDeviceProfileClientName = di.TypeInstanceToName((*interfaces.DeviceProfileClient)(nil))

// MetadataDeviceServiceClientName contains the name of the Metadata DeviceServiceClient instance in the DIC.
var MetadataDeviceServiceClientName = di.TypeInstanceToName((*interfaces.DeviceServiceClient)(nil))

// MetadataProvisionWatcherClientName contains the name of the Metadata ProvisionWatcherClient instance in the DIC.
var MetadataProvisionWatcherClientName = di.TypeInstanceToName((*interfaces.ProvisionWatcherClient)(nil))

// MetadataDeviceClientFrom helper function queries the DIC and returns the Metadata DeviceClient instance.
func MetadataDeviceClientFrom(get di.Get) interfaces.DeviceClient {
	client, ok := get(MetadataDeviceClientName).(interfaces.DeviceClient)
	if !ok {
		return nil
	}

	return client
}

// MetadataDeviceProfileClientFrom helper function queries the DIC and returns the Metadata DeviceProfileClient instance.
func MetadataDeviceProfileClientFrom(get di.Get) interfaces.DeviceProfileClient {
	client, ok := get(MetadataDeviceProfileClientName).(interfaces.DeviceProfileClient)
	if !ok {
		return nil
	}

	return client
}

// MetadataDeviceServiceClientFrom helper function queries the DIC and returns the Metadata DeviceServiceClient instance.
func MetadataDeviceServiceClientFrom(get di.Get) interfaces.DeviceServiceClient {
	client, ok := get(MetadataDeviceServiceClientName).(interfaces.DeviceServiceClient)
	if !ok {
		return nil
	}

	return client
}

// MetadataProvisionWatcherClientFrom helper function queries the DIC and returns the Metadata ProvisionWatcherClient instance.
func MetadataProvisionWatcherClientFrom(get di.Get) interfaces.ProvisionWatcherClient {
	client, ok := get(MetadataProvisionWatcherClientName).(interfaces.ProvisionWatcherClient)
	if !ok {
		return nil
	}

	return client
}
