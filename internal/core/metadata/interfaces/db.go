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
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

type DBClient interface {
	CloseSession()

	// Schedule event
	GetAllScheduleEvents(se *[]contract.ScheduleEvent) error
	AddScheduleEvent(se *contract.ScheduleEvent) error
	GetScheduleEventByName(se *contract.ScheduleEvent, n string) error
	UpdateScheduleEvent(se contract.ScheduleEvent) error
	GetScheduleEventById(se *contract.ScheduleEvent, id string) error
	GetScheduleEventsByScheduleName(se *[]contract.ScheduleEvent, n string) error
	GetScheduleEventsByAddressableId(se *[]contract.ScheduleEvent, id string) error
	GetScheduleEventsByServiceName(se *[]contract.ScheduleEvent, n string) error
	DeleteScheduleEventById(id string) error

	// Schedule
	GetAllSchedules(s *[]contract.Schedule) error
	AddSchedule(s *contract.Schedule) error
	GetScheduleByName(s *contract.Schedule, n string) error
	UpdateSchedule(s contract.Schedule) error
	GetScheduleById(s *contract.Schedule, id string) error
	DeleteScheduleById(id string) error

	// Device Report
	GetAllDeviceReports() ([]contract.DeviceReport, error)
	GetDeviceReportByDeviceName(n string) ([]contract.DeviceReport, error)
	GetDeviceReportByName(n string) (contract.DeviceReport, error)
	GetDeviceReportById(id string) (contract.DeviceReport, error)
	AddDeviceReport(dr contract.DeviceReport) (string, error)
	UpdateDeviceReport(dr contract.DeviceReport) error
	GetDeviceReportsByScheduleEventName(n string) ([]contract.DeviceReport, error)
	DeleteDeviceReportById(id string) error

	// Device
	UpdateDevice(d contract.Device) error
	GetDeviceById(d *contract.Device, id string) error
	GetDeviceByName(d *contract.Device, n string) error
	GetAllDevices(d *[]contract.Device) error
	GetDevicesByProfileId(d *[]contract.Device, pid string) error
	GetDevicesByServiceId(d *[]contract.Device, sid string) error
	GetDevicesByAddressableId(d *[]contract.Device, aid string) error
	GetDevicesWithLabel(d *[]contract.Device, l string) error
	AddDevice(d *contract.Device) error
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
	GetDeviceProfilesUsingCommand(c contract.Command) ([]contract.DeviceProfile, error)

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
	GetProvisionWatcherById(pw *contract.ProvisionWatcher, id string) error
	GetAllProvisionWatchers(pw *[]contract.ProvisionWatcher) error
	GetProvisionWatcherByName(pw *contract.ProvisionWatcher, n string) error
	GetProvisionWatchersByProfileId(pw *[]contract.ProvisionWatcher, id string) error
	GetProvisionWatchersByServiceId(pw *[]contract.ProvisionWatcher, id string) error
	GetProvisionWatchersByIdentifier(pw *[]contract.ProvisionWatcher, k string, v string) error
	AddProvisionWatcher(pw *contract.ProvisionWatcher) error
	UpdateProvisionWatcher(pw contract.ProvisionWatcher) error
	DeleteProvisionWatcherById(id string) error

	// Command
	GetCommandById(id string) (contract.Command, error)
	GetCommandByName(id string) ([]contract.Command, error)
	AddCommand(c contract.Command) (string, error)
	GetAllCommands() ([]contract.Command, error)
	UpdateCommand(c *contract.Command) error
	DeleteCommandById(id string) error

	// Scrub all metadata (only used in test)
	ScrubMetadata() error
}
