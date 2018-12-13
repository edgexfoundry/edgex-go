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
	"github.com/globalsign/mgo/bson"
)

type DBClient interface {
	CloseSession()

	Connect() error

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
	GetAllDeviceReports(dr *[]contract.DeviceReport) error
	GetDeviceReportByDeviceName(dr *[]contract.DeviceReport, n string) error
	GetDeviceReportByName(dr *contract.DeviceReport, n string) error
	GetDeviceReportById(dr *contract.DeviceReport, id string) error
	AddDeviceReport(dr *contract.DeviceReport) error
	UpdateDeviceReport(dr *contract.DeviceReport) error
	GetDeviceReportsByScheduleEventName(dr *[]contract.DeviceReport, n string) error
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
	UpdateDeviceProfile(dp *contract.DeviceProfile) error
	AddDeviceProfile(d *contract.DeviceProfile) error
	GetAllDeviceProfiles(d *[]contract.DeviceProfile) error
	GetDeviceProfileById(d *contract.DeviceProfile, id string) error
	DeleteDeviceProfileById(id string) error
	GetDeviceProfilesByModel(dp *[]contract.DeviceProfile, m string) error
	GetDeviceProfilesWithLabel(dp *[]contract.DeviceProfile, l string) error
	GetDeviceProfilesByManufacturerModel(dp *[]contract.DeviceProfile, man string, mod string) error
	GetDeviceProfilesByManufacturer(dp *[]contract.DeviceProfile, man string) error
	GetDeviceProfileByName(dp *contract.DeviceProfile, n string) error
	GetDeviceProfilesUsingCommand(dp *[]contract.DeviceProfile, c contract.Command) error

	// Addressable
	UpdateAddressable(ra *contract.Addressable, r *contract.Addressable) error
	AddAddressable(a *contract.Addressable) (bson.ObjectId, error)
	GetAddressableById(a *contract.Addressable, id string) error
	GetAddressableByName(a *contract.Addressable, n string) error
	GetAddressablesByTopic(a *[]contract.Addressable, t string) error
	GetAddressablesByPort(a *[]contract.Addressable, p int) error
	GetAddressablesByPublisher(a *[]contract.Addressable, p string) error
	GetAddressablesByAddress(a *[]contract.Addressable, add string) error
	GetAddressables(d *[]contract.Addressable) error
	DeleteAddressableById(id string) error

	// Device service
	UpdateDeviceService(ds contract.DeviceService) error
	GetDeviceServicesByAddressableId(d *[]contract.DeviceService, id string) error
	GetDeviceServicesWithLabel(d *[]contract.DeviceService, l string) error
	GetDeviceServiceById(d *contract.DeviceService, id string) error
	GetDeviceServiceByName(d *contract.DeviceService, n string) error
	GetAllDeviceServices(d *[]contract.DeviceService) error
	AddDeviceService(ds *contract.DeviceService) error
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
	GetCommandById(c *contract.Command, id string) error
	GetCommandByName(c *[]contract.Command, id string) error
	AddCommand(c *contract.Command) error
	GetAllCommands(d *[]contract.Command) error
	UpdateCommand(c *contract.Command, r *contract.Command) error
	DeleteCommandById(id string) error

	// Scrub all metadata (only used in test)
	ScrubMetadata() error
}
