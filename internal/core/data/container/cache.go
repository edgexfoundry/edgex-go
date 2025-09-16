//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/core/data/cache"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

var DeviceInfoCacheInterfaceName = di.TypeInstanceToName((*cache.DeviceInfoCache)(nil))

// DeviceInfoCacheFrom helper function queries the DIC and returns the cache.DeviceInfoCache implementation.
func DeviceInfoCacheFrom(get di.Get) cache.DeviceInfoCache {
	return get(DeviceInfoCacheInterfaceName).(cache.DeviceInfoCache)
}
