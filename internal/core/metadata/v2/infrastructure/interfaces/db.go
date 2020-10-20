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
	UpdateDeviceProfile(e model.DeviceProfile) errors.EdgeX
	GetDeviceProfileByName(name string) (model.DeviceProfile, errors.EdgeX)
	DeleteDeviceProfileById(id string) errors.EdgeX
	DeleteDeviceProfileByName(name string) errors.EdgeX
	DeviceProfileNameExists(name string) (bool, errors.EdgeX)
	GetDeviceProfiles(offset int, limit int, labels []string) ([]model.DeviceProfile, errors.EdgeX)

	AddDeviceService(e model.DeviceService) (model.DeviceService, errors.EdgeX)
	GetDeviceServiceById(id string) (model.DeviceService, errors.EdgeX)
	GetDeviceServiceByName(name string) (model.DeviceService, errors.EdgeX)
	DeleteDeviceServiceById(id string) errors.EdgeX
	DeleteDeviceServiceByName(name string) errors.EdgeX
	DeviceServiceNameExists(name string) (bool, errors.EdgeX)
	GetDeviceServices(offset int, limit int, labels []string) ([]model.DeviceService, errors.EdgeX)

	AddDevice(d model.Device) (model.Device, errors.EdgeX)
}
