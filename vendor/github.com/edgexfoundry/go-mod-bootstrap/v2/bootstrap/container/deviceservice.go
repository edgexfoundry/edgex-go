//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
)

// DeviceServiceCallbackClientName contains the name of the DeviceServiceCallbackClient instance in the DIC.
var DeviceServiceCallbackClientName = di.TypeInstanceToName((*interfaces.DeviceServiceCallbackClient)(nil))

// DeviceServiceCommandClientName contains the name of the DeviceServiceCommandClient instance in the DIC.
var DeviceServiceCommandClientName = di.TypeInstanceToName((*interfaces.DeviceServiceCommandClient)(nil))

// DeviceServiceCallbackClientFrom helper function queries the DIC and returns the DeviceServiceCallbackClient instance.
func DeviceServiceCallbackClientFrom(get di.Get) interfaces.DeviceServiceCallbackClient {
	client, ok := get(DeviceServiceCallbackClientName).(interfaces.DeviceServiceCallbackClient)
	if !ok {
		return nil
	}

	return client
}

// DeviceServiceCommandClientFrom helper function queries the DIC and returns the DeviceServiceCommandClient instance.
func DeviceServiceCommandClientFrom(get di.Get) interfaces.DeviceServiceCommandClient {
	client, ok := get(DeviceServiceCommandClientName).(interfaces.DeviceServiceCommandClient)
	if !ok {
		return nil
	}

	return client
}
