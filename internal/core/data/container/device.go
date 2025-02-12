//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/cache"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

// DeviceStoreInterfaceName contains the name of the cache.ActiveDeviceStore implementation in the DIC.
var DeviceStoreInterfaceName = di.TypeInstanceToName((*cache.ActiveDeviceStore)(nil))

// DeviceStoreFrom helper function queries the DIC and returns the cache.ActiveDeviceStore implementation.
func DeviceStoreFrom(get di.Get) cache.ActiveDeviceStore {
	return get(DeviceStoreInterfaceName).(cache.ActiveDeviceStore)
}
