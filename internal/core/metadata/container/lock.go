//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/utils"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

// CapacityCheckLockName contains the name of the metadata's utils.CapacityCheckLock implementation in the DIC.
var CapacityCheckLockName = di.TypeInstanceToName((*utils.CapacityCheckLock)(nil))

// CapacityCheckLockFrom helper function queries the DIC and returns metadata's utils.CapacityCheckLock implementation.
func CapacityCheckLockFrom(get di.Get) *utils.CapacityCheckLock {
	return get(CapacityCheckLockName).(*utils.CapacityCheckLock)
}
