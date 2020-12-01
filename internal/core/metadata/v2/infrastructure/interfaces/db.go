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
	DeviceProfileByName(name string) (model.DeviceProfile, errors.EdgeX)
	DeleteDeviceProfileById(id string) errors.EdgeX
	DeleteDeviceProfileByName(name string) errors.EdgeX
	DeviceProfileNameExists(name string) (bool, errors.EdgeX)
	AllDeviceProfiles(offset int, limit int, labels []string) ([]model.DeviceProfile, errors.EdgeX)
	DeviceProfilesByModel(offset int, limit int, model string) ([]model.DeviceProfile, errors.EdgeX)

	AddDeviceService(e model.DeviceService) (model.DeviceService, errors.EdgeX)
	DeviceServiceById(id string) (model.DeviceService, errors.EdgeX)
	DeviceServiceByName(name string) (model.DeviceService, errors.EdgeX)
	DeleteDeviceServiceById(id string) errors.EdgeX
	DeleteDeviceServiceByName(name string) errors.EdgeX
	DeviceServiceNameExists(name string) (bool, errors.EdgeX)
	AllDeviceServices(offset int, limit int, labels []string) ([]model.DeviceService, errors.EdgeX)

	AddDevice(d model.Device) (model.Device, errors.EdgeX)
	DeleteDeviceById(id string) errors.EdgeX
	DeleteDeviceByName(name string) errors.EdgeX
	DevicesByServiceName(offset int, limit int, name string) ([]model.Device, errors.EdgeX)
	DeviceIdExists(id string) (bool, errors.EdgeX)
	DeviceNameExists(id string) (bool, errors.EdgeX)
	DeviceById(id string) (model.Device, errors.EdgeX)
	DeviceByName(name string) (model.Device, errors.EdgeX)
	AllDevices(offset int, limit int, labels []string) ([]model.Device, errors.EdgeX)
}
