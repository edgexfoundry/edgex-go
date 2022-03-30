//
// Copyright (C) 2020-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type DBClient interface {
	CloseSession()

	AddDeviceProfile(e model.DeviceProfile) (model.DeviceProfile, errors.EdgeX)
	UpdateDeviceProfile(e model.DeviceProfile) errors.EdgeX
	DeviceProfileById(id string) (model.DeviceProfile, errors.EdgeX)
	DeviceProfileByName(name string) (model.DeviceProfile, errors.EdgeX)
	DeleteDeviceProfileById(id string) errors.EdgeX
	DeleteDeviceProfileByName(name string) errors.EdgeX
	DeviceProfileNameExists(name string) (bool, errors.EdgeX)
	AllDeviceProfiles(offset int, limit int, labels []string) ([]model.DeviceProfile, errors.EdgeX)
	DeviceProfilesByModel(offset int, limit int, model string) ([]model.DeviceProfile, errors.EdgeX)
	DeviceProfilesByManufacturer(offset int, limit int, manufacturer string) ([]model.DeviceProfile, errors.EdgeX)
	DeviceProfilesByManufacturerAndModel(offset int, limit int, manufacturer string, model string) ([]model.DeviceProfile, uint32, errors.EdgeX)
	DeviceProfileCountByLabels(labels []string) (uint32, errors.EdgeX)
	DeviceProfileCountByManufacturer(manufacturer string) (uint32, errors.EdgeX)
	DeviceProfileCountByModel(model string) (uint32, errors.EdgeX)

	AddDeviceService(ds model.DeviceService) (model.DeviceService, errors.EdgeX)
	DeviceServiceById(id string) (model.DeviceService, errors.EdgeX)
	DeviceServiceByName(name string) (model.DeviceService, errors.EdgeX)
	DeleteDeviceServiceById(id string) errors.EdgeX
	DeleteDeviceServiceByName(name string) errors.EdgeX
	DeviceServiceNameExists(name string) (bool, errors.EdgeX)
	AllDeviceServices(offset int, limit int, labels []string) ([]model.DeviceService, errors.EdgeX)
	UpdateDeviceService(ds model.DeviceService) errors.EdgeX
	DeviceServiceCountByLabels(labels []string) (uint32, errors.EdgeX)

	AddDevice(d model.Device) (model.Device, errors.EdgeX)
	DeleteDeviceById(id string) errors.EdgeX
	DeleteDeviceByName(name string) errors.EdgeX
	DevicesByServiceName(offset int, limit int, name string) ([]model.Device, errors.EdgeX)
	DeviceIdExists(id string) (bool, errors.EdgeX)
	DeviceNameExists(id string) (bool, errors.EdgeX)
	DeviceById(id string) (model.Device, errors.EdgeX)
	DeviceByName(name string) (model.Device, errors.EdgeX)
	AllDevices(offset int, limit int, labels []string) ([]model.Device, errors.EdgeX)
	DevicesByProfileName(offset int, limit int, profileName string) ([]model.Device, errors.EdgeX)
	UpdateDevice(d model.Device) errors.EdgeX
	DeviceCountByLabels(labels []string) (uint32, errors.EdgeX)
	DeviceCountByProfileName(profileName string) (uint32, errors.EdgeX)
	DeviceCountByServiceName(serviceName string) (uint32, errors.EdgeX)

	AddProvisionWatcher(pw model.ProvisionWatcher) (model.ProvisionWatcher, errors.EdgeX)
	ProvisionWatcherById(id string) (model.ProvisionWatcher, errors.EdgeX)
	ProvisionWatcherByName(name string) (model.ProvisionWatcher, errors.EdgeX)
	ProvisionWatchersByServiceName(offset int, limit int, name string) ([]model.ProvisionWatcher, errors.EdgeX)
	ProvisionWatchersByProfileName(offset int, limit int, name string) ([]model.ProvisionWatcher, errors.EdgeX)
	AllProvisionWatchers(offset int, limit int, labels []string) ([]model.ProvisionWatcher, errors.EdgeX)
	DeleteProvisionWatcherByName(name string) errors.EdgeX
	UpdateProvisionWatcher(pw model.ProvisionWatcher) errors.EdgeX
	ProvisionWatcherCountByLabels(labels []string) (uint32, errors.EdgeX)
	ProvisionWatcherCountByServiceName(name string) (uint32, errors.EdgeX)
	ProvisionWatcherCountByProfileName(name string) (uint32, errors.EdgeX)
}
