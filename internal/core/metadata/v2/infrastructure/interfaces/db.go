//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type DBClient interface {
	CloseSession()

	AddDeviceProfile(e model.DeviceProfile) (model.DeviceProfile, errors.EdgeX)
}
