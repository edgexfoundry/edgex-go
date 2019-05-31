/*******************************************************************************
 * Copyright 2017 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package interfaces

import (
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type DBClient interface {
	CloseSession()

	// Device Report
	GetAllDeviceReports() ([]contract.DeviceReport, error)
	GetDeviceReportByDeviceName(n string) ([]contract.DeviceReport, error)
	GetDeviceReportByName(n string) (contract.DeviceReport, error)
	GetDeviceReportById(id string) (contract.DeviceReport, error)
	AddDeviceReport(dr contract.DeviceReport) (string, error)
	UpdateDeviceReport(dr contract.DeviceReport) error
	GetDeviceReportsByAction(n string) ([]contract.DeviceReport, error)
	DeleteDeviceReportById(id string) error

	// Device
	UpdateDevice(d contract.Device) error
	GetDeviceById(id string) (contract.Device, error)
	GetDeviceByName(n string) (contract.Device, error)
	GetAllDevices() ([]contract.Device, error)
	GetDevicesByProfileId(pid string) ([]contract.Device, error)
	GetDevicesByServiceId(sid string) ([]contract.Device, error)
	GetDevicesWithLabel(l string) ([]contract.Device, error)
	AddDevice(d contract.Device) (string, error)
	DeleteDeviceById(id string) error

	// Device Profile
	UpdateDeviceProfile(dp contract.DeviceProfile) error
	AddDeviceProfile(d contract.DeviceProfile) (string, error)
	GetAllDeviceProfiles() ([]contract.DeviceProfile, error)
	GetDeviceProfileById(id string) (contract.DeviceProfile, error)
	DeleteDeviceProfileById(id string) error
	GetDeviceProfilesByModel(m string) ([]contract.DeviceProfile, error)
	GetDeviceProfilesWithLabel(l string) ([]contract.DeviceProfile, error)
	GetDeviceProfilesByManufacturerModel(man string, mod string) ([]contract.DeviceProfile, error)
	GetDeviceProfilesByManufacturer(man string) ([]contract.DeviceProfile, error)
	GetDeviceProfileByName(n string) (contract.DeviceProfile, error)

	// Addressable
	UpdateAddressable(a contract.Addressable) error
	AddAddressable(a contract.Addressable) (string, error)
	GetAddressableById(id string) (contract.Addressable, error)
	GetAddressableByName(n string) (contract.Addressable, error)
	GetAddressablesByTopic(t string) ([]contract.Addressable, error)
	GetAddressablesByPort(p int) ([]contract.Addressable, error)
	GetAddressablesByPublisher(p string) ([]contract.Addressable, error)
	GetAddressablesByAddress(add string) ([]contract.Addressable, error)
	GetAddressables() ([]contract.Addressable, error)
	DeleteAddressableById(id string) error

	// Device service
	UpdateDeviceService(ds contract.DeviceService) error
	GetDeviceServicesByAddressableId(id string) ([]contract.DeviceService, error)
	GetDeviceServicesWithLabel(l string) ([]contract.DeviceService, error)
	GetDeviceServiceById(id string) (contract.DeviceService, error)
	GetDeviceServiceByName(n string) (contract.DeviceService, error)
	GetAllDeviceServices() ([]contract.DeviceService, error)
	AddDeviceService(ds contract.DeviceService) (string, error)
	DeleteDeviceServiceById(id string) error

	// Provision watcher
	GetProvisionWatcherById(id string) (contract.ProvisionWatcher, error)
	GetAllProvisionWatchers() ([]contract.ProvisionWatcher, error)
	GetProvisionWatcherByName(n string) (contract.ProvisionWatcher, error)
	GetProvisionWatchersByProfileId(id string) ([]contract.ProvisionWatcher, error)
	GetProvisionWatchersByServiceId(id string) ([]contract.ProvisionWatcher, error)
	GetProvisionWatchersByIdentifier(k string, v string) ([]contract.ProvisionWatcher, error)
	AddProvisionWatcher(pw contract.ProvisionWatcher) (string, error)
	UpdateProvisionWatcher(pw contract.ProvisionWatcher) error
	DeleteProvisionWatcherById(id string) error

	// Command
	GetAllCommands() ([]contract.Command, error)
	GetCommandById(id string) (contract.Command, error)
	GetCommandByName(id string) ([]contract.Command, error)

	// Scrub all metadata (only used in test)
	ScrubMetadata() error
}
