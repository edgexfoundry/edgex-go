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
	"github.com/edgexfoundry/edgex-go/core/domain/enums"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

type DBClient interface {
	CloseSession()

	Connect() error

	// Schedule event
	getAllScheduleEvents(se *[]models.ScheduleEvent) error
	addScheduleEvent(se *models.ScheduleEvent) error
	getScheduleEventByName(se *models.ScheduleEvent, n string) error
	updateScheduleEvent(se models.ScheduleEvent) error
	getScheduleEventById(se *models.ScheduleEvent, id string) error
	getScheduleEventsByScheduleName(se *[]models.ScheduleEvent, n string) error
	getScheduleEventsByAddressableId(se *[]models.ScheduleEvent, id string) error
	getScheduleEventsByServiceName(se *[]models.ScheduleEvent, n string) error
	deleteScheduleEvent(se models.ScheduleEvent) error

	// Schedule
	getAllSchedules(s *[]models.Schedule) error
	addSchedule(s *models.Schedule) error
	getScheduleByName(s *models.Schedule, n string) error
	updateSchedule(s models.Schedule) error
	getScheduleById(s *models.Schedule, id string) error
	deleteSchedule(s models.Schedule) error

	// Device Report
	getAllDeviceReports(dr *[]models.DeviceReport) error
	getDeviceReportByDeviceName(dr *[]models.DeviceReport, n string) error
	getDeviceReportByName(dr *models.DeviceReport, n string) error
	getDeviceReportById(dr *models.DeviceReport, id string) error
	addDeviceReport(dr *models.DeviceReport) error
	updateDeviceReport(dr *models.DeviceReport) error
	getDeviceReportsByScheduleEventName(dr *[]models.DeviceReport, n string) error
	deleteDeviceReport(dr models.DeviceReport) error

	// Device
	updateDevice(d models.Device) error
	getDeviceById(d *models.Device, id string) error
	getDeviceByName(d *models.Device, n string) error
	getAllDevices(d *[]models.Device) error
	getDevicesByProfileId(d *[]models.Device, pid string) error
	getDevicesByServiceId(d *[]models.Device, sid string) error
	getDevicesByAddressableId(d *[]models.Device, aid string) error
	getDevicesWithLabel(d *[]models.Device, l []string) error
	addDevice(d *models.Device) error
	deleteDevice(d models.Device) error
	updateDeviceProfile(dp *models.DeviceProfile) error
	addDeviceProfile(d *models.DeviceProfile) error
	getAllDeviceProfiles(d *[]models.DeviceProfile) error
	getDeviceProfileById(d *models.DeviceProfile, id string) error
	deleteDeviceProfile(dp models.DeviceProfile) error
	getDeviceProfilesByModel(dp *[]models.DeviceProfile, m string) error
	getDeviceProfilesWithLabel(dp *[]models.DeviceProfile, l []string) error
	getDeviceProfilesByManufacturerModel(dp *[]models.DeviceProfile, man string, mod string) error
	getDeviceProfilesByManufacturer(dp *[]models.DeviceProfile, man string) error
	getDeviceProfileByName(dp *models.DeviceProfile, n string) error

	updateAddressable(ra *models.Addressable, r *models.Addressable) error
	addAddressable(a *models.Addressable) error
	getAddressableById(a *models.Addressable, id string) error
	getAddressableByName(a *models.Addressable, n string) error
	getAddressablesByTopic(a *[]models.Addressable, t string) error
	getAddressablesByPort(a *[]models.Addressable, p int) error
	getAddressablesByPublisher(a *[]models.Addressable, p string) error
	getAddressablesByAddress(a *[]models.Addressable, add string) error
	getAddressables(d *[]models.Addressable) error
	deleteAddressable(a models.Addressable) error

	// Device service
	updateDeviceService(ds models.DeviceService) error
	getDeviceServicesByAddressableId(d *[]models.DeviceService, id string) error
	getDeviceServicesWithLabel(d *[]models.DeviceService, l []string) error
	getDeviceServiceById(d *models.DeviceService, id string) error
	getDeviceServiceByName(d *models.DeviceService, n string) error
	getAllDeviceServices(d *[]models.DeviceService) error
	addDeviceService(ds *models.DeviceService) error
	deleteDeviceService(ds models.DeviceService) error

	// Provision watcher
	getProvisionWatcherById(pw *models.ProvisionWatcher, id string) error
	getAllProvisionWatchers(pw *[]models.ProvisionWatcher) error
	getProvisionWatcherByName(pw *models.ProvisionWatcher, n string) error
	getProvisionWatcherByProfileId(pw *[]models.ProvisionWatcher, id string) error
	getProvisionWatchersByServiceId(pw *[]models.ProvisionWatcher, id string) error
	getProvisionWatchersByIdentifier(pw *[]models.ProvisionWatcher, k string, v string) error
	addProvisionWatcher(pw *models.ProvisionWatcher) error
	updateProvisionWatcher(pw models.ProvisionWatcher) error
	deleteProvisionWatcher(pw models.ProvisionWatcher) error

	// Command
	getCommandById(c *models.Command, id string) error
	getCommandByName(d *[]models.Command, id string) error
	addCommand(c *models.Command) error
	getAllCommands(d *[]models.Command) error
	updateCommand(c *models.Command, r *models.Command) error
	deleteCommandById(id string) error

	getDeviceProfilesUsingCommand(dp *[]models.DeviceProfile, c models.Command) error
}

func getDatabase(dbType string) (DBClient, error) {
	switch dbType {
	case enums.MongoStr:
		return &mongoDB{}, nil
	case enums.MemoryStr:
	}
	return nil, ErrNotFound
}
