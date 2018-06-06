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
package metadata

import (
	"github.com/edgexfoundry/edgex-go/core/db"
	"github.com/edgexfoundry/edgex-go/core/db/mongo"
	"github.com/edgexfoundry/edgex-go/core/domain/enums"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

type DBClient interface {
	CloseSession()

	Connect() error

	// Schedule event
	GetAllScheduleEvents(se *[]models.ScheduleEvent) error
	AddScheduleEvent(se *models.ScheduleEvent) error
	GetScheduleEventByName(se *models.ScheduleEvent, n string) error
	UpdateScheduleEvent(se models.ScheduleEvent) error
	GetScheduleEventById(se *models.ScheduleEvent, id string) error
	GetScheduleEventsByScheduleName(se *[]models.ScheduleEvent, n string) error
	GetScheduleEventsByAddressableId(se *[]models.ScheduleEvent, id string) error
	GetScheduleEventsByServiceName(se *[]models.ScheduleEvent, n string) error
	DeleteScheduleEvent(se models.ScheduleEvent) error

	// Schedule
	GetAllSchedules(s *[]models.Schedule) error
	AddSchedule(s *models.Schedule) error
	GetScheduleByName(s *models.Schedule, n string) error
	UpdateSchedule(s models.Schedule) error
	GetScheduleById(s *models.Schedule, id string) error
	DeleteSchedule(s models.Schedule) error

	// Device Report
	GetAllDeviceReports(dr *[]models.DeviceReport) error
	GetDeviceReportByDeviceName(dr *[]models.DeviceReport, n string) error
	GetDeviceReportByName(dr *models.DeviceReport, n string) error
	GetDeviceReportById(dr *models.DeviceReport, id string) error
	AddDeviceReport(dr *models.DeviceReport) error
	UpdateDeviceReport(dr *models.DeviceReport) error
	GetDeviceReportsByScheduleEventName(dr *[]models.DeviceReport, n string) error
	DeleteDeviceReport(dr models.DeviceReport) error

	// Device
	UpdateDevice(d models.Device) error
	GetDeviceById(d *models.Device, id string) error
	GetDeviceByName(d *models.Device, n string) error
	GetAllDevices(d *[]models.Device) error
	GetDevicesByProfileId(d *[]models.Device, pid string) error
	GetDevicesByServiceId(d *[]models.Device, sid string) error
	GetDevicesByAddressableId(d *[]models.Device, aid string) error
	GetDevicesWithLabel(d *[]models.Device, l []string) error
	AddDevice(d *models.Device) error
	DeleteDevice(d models.Device) error
	UpdateDeviceProfile(dp *models.DeviceProfile) error
	AddDeviceProfile(d *models.DeviceProfile) error
	GetAllDeviceProfiles(d *[]models.DeviceProfile) error
	GetDeviceProfileById(d *models.DeviceProfile, id string) error
	DeleteDeviceProfile(dp models.DeviceProfile) error
	GetDeviceProfilesByModel(dp *[]models.DeviceProfile, m string) error
	GetDeviceProfilesWithLabel(dp *[]models.DeviceProfile, l []string) error
	GetDeviceProfilesByManufacturerModel(dp *[]models.DeviceProfile, man string, mod string) error
	GetDeviceProfilesByManufacturer(dp *[]models.DeviceProfile, man string) error
	GetDeviceProfileByName(dp *models.DeviceProfile, n string) error

	UpdateAddressable(ra *models.Addressable, r *models.Addressable) error
	AddAddressable(a *models.Addressable) error
	GetAddressableById(a *models.Addressable, id string) error
	GetAddressableByName(a *models.Addressable, n string) error
	GetAddressablesByTopic(a *[]models.Addressable, t string) error
	GetAddressablesByPort(a *[]models.Addressable, p int) error
	GetAddressablesByPublisher(a *[]models.Addressable, p string) error
	GetAddressablesByAddress(a *[]models.Addressable, add string) error
	GetAddressables(d *[]models.Addressable) error
	DeleteAddressable(a models.Addressable) error

	// Device service
	UpdateDeviceService(ds models.DeviceService) error
	GetDeviceServicesByAddressableId(d *[]models.DeviceService, id string) error
	GetDeviceServicesWithLabel(d *[]models.DeviceService, l []string) error
	GetDeviceServiceById(d *models.DeviceService, id string) error
	GetDeviceServiceByName(d *models.DeviceService, n string) error
	GetAllDeviceServices(d *[]models.DeviceService) error
	AddDeviceService(ds *models.DeviceService) error
	DeleteDeviceService(ds models.DeviceService) error

	// Provision watcher
	GetProvisionWatcherById(pw *models.ProvisionWatcher, id string) error
	GetAllProvisionWatchers(pw *[]models.ProvisionWatcher) error
	GetProvisionWatcherByName(pw *models.ProvisionWatcher, n string) error
	GetProvisionWatcherByProfileId(pw *[]models.ProvisionWatcher, id string) error
	GetProvisionWatchersByServiceId(pw *[]models.ProvisionWatcher, id string) error
	GetProvisionWatchersByIdentifier(pw *[]models.ProvisionWatcher, k string, v string) error
	AddProvisionWatcher(pw *models.ProvisionWatcher) error
	UpdateProvisionWatcher(pw models.ProvisionWatcher) error
	DeleteProvisionWatcher(pw models.ProvisionWatcher) error

	// Command
	GetCommandById(c *models.Command, id string) error
	GetCommandByName(d *[]models.Command, id string) error
	AddCommand(c *models.Command) error
	GetAllCommands(d *[]models.Command) error
	UpdateCommand(c *models.Command, r *models.Command) error
	DeleteCommandById(id string) error

	GetDeviceProfilesUsingCommand(dp *[]models.DeviceProfile, c models.Command) error
}

func getDatabase(dbType string, config db.Configuration) (DBClient, error) {
	switch dbType {
	case enums.MongoStr:
		return mongo.NewClient(config), nil
	case enums.MemoryStr:
	}
	return nil, db.ErrNotFound
}
