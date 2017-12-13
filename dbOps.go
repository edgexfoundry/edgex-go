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
 *
 * @microservice: core-metadata-go service
 * @author: Spencer Bull & Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package main

import (
	"strings"
	"time"

	"github.com/edgexfoundry/core-domain-go/models"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type DataStore struct {
	s *mgo.Session
}

func (ds DataStore) dataStore() *DataStore {
	return &DataStore{ds.s.Copy()}
}

// Connect to the database
func dbConnect() bool {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {

		mongoDBDialInfo := &mgo.DialInfo{
			Addrs:    []string{DOCKERMONGO},
			Timeout:  time.Duration(configuration.MongoDBConnectTimeout) * time.Millisecond,
			Database: DATABASE,
			Username: DBUSER,
			Password: DBPASS,
		}
		s, err := mgo.DialWithInfo(mongoDBDialInfo)
		// Set timeout based on configuration
		s.SetSocketTimeout(time.Duration(configuration.MongoDBConnectTimeout) * time.Millisecond)
		if err != nil {
			return false
		}
		DS.s = s
		return true
	}
	return false
}

/* ----------------------- Schedule Event ------------------------------*/
func getAllScheduleEvents(se *[]models.ScheduleEvent) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAllScheduleEvents(se)
	}
	return nil
}
func addScheduleEvent(se *models.ScheduleEvent) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoAddScheduleEvent(se)
	}
	return nil
}
func getScheduleEventByName(se *models.ScheduleEvent, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetScheduleEventByName(se, n)
	}
	return nil
}
func updateScheduleEvent(se models.ScheduleEvent) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateScheduleEvent(se)
	}
	return nil
}
func getScheduleEventById(se *models.ScheduleEvent, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetScheduleEventById(se, id)
	}
	return nil
}
func getScheduleEventsByScheduleName(se *[]models.ScheduleEvent, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetScheduleEventsByScheduleName(se, n)
	}
	return nil
}
func getScheduleEventsByAddressableId(se *[]models.ScheduleEvent, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetScheduleEventsByAddressableId(se, id)
	}
	return nil
}
func getScheduleEventsByServiceName(se *[]models.ScheduleEvent, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetScheduleEventsByServiceName(se, n)
	}
	return nil
}

/* -------------------------- Schedule ---------------------------------*/
func getAllSchedules(s *[]models.Schedule) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAllSchedules(s)
	}
	return nil
}
func addSchedule(s *models.Schedule) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoAddSchedule(s)
	}
	return nil
}
func getScheduleByName(s *models.Schedule, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetScheduleByName(s, n)
	}
	return nil
}
func updateSchedule(s models.Schedule) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateSchedule(s)
	}
	return nil
}
func getScheduleById(s *models.Schedule, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetScheduleById(s, id)
	}
	return nil
}

/* ------------------------Device Report -------------------------------*/
func getAllDeviceReports(dr *[]models.DeviceReport) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAllDeviceReports(dr)
	}
	return nil
}
func getDeviceReportByDeviceName(dr *[]models.DeviceReport, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceReportByDeviceName(dr, n)
	}
	return nil
}
func getDeviceReportByName(dr *models.DeviceReport, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceReportByName(dr, n)
	}
	return nil
}
func getDeviceReportById(dr *models.DeviceReport, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceReportById(dr, id)
	}
	return nil
}
func addDeviceReport(dr *models.DeviceReport) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoAddDeviceReport(dr)
	}
	return nil
}
func updateDeviceReport(dr *models.DeviceReport) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateDeviceReport(dr)
	}
	return nil
}
func getDeviceReportsByScheduleEventName(dr *[]models.DeviceReport, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceReportsByScheduleEventName(dr, n)
	}
	return nil
}

// ------------------------------------- DEVICE --------------------------------------------

func UpdateDevice(d models.Device) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateDevice(d)
	}
	return nil
}
func getDeviceById(d *models.Device, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceById(d, id)
	}
	return nil
}
func getDeviceByName(d *models.Device, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceByName(d, n)
	}
	return nil
}
func getAllDevices(d *[]models.Device) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAllDevices(d)
	}
	return nil
}
func getDevicesByProfileId(d *[]models.Device, pid string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDevicesByProfileId(d, pid)
	}
	return nil
}
func getDevicesByProfileName(d *[]models.Device, pn string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDevicesByProfileName(d, pn)
	}
	return nil
}
func getDevicesByServiceId(d *[]models.Device, sid string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDevicesByServiceId(d, sid)
	}
	return nil
}
func getDevicesByServiceName(d *[]models.Device, sn string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDevicesByServiceName(d, sn)
	}
	return nil
}
func getDevicesByAddressableId(d *[]models.Device, aid string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDevicesByAddressableId(d, aid)
	}
	return nil
}
func getDevicesByAddressableName(d *[]models.Device, an string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDevicesByAddressableName(d, an)
	}
	return nil
}
func getDevicesWithLabel(d *[]models.Device, l []string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDevicesWithLabel(d, l)
	}
	return nil
}
func addDevice(d *models.Device) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoAddNewDevice(d)
	}
	return nil
}
func updateDeviceProfile(dp *models.DeviceProfile) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateDeviceProfile(dp)
	}
	return nil
}
func addDeviceProfile(d *models.DeviceProfile) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoAddNewDeviceProfile(d)
	}
	return nil
}
func getAllDeviceProfiles(d *[]models.DeviceProfile) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceProfiles(d, bson.M{})
	}
	return nil
}
func getDeviceProfileById(d *models.DeviceProfile, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceProfileById(d, id)
	}
	return nil
}
func deleteDeviceProfileById(dpid string) error {
	if err := deleteById(DPCOL, dpid); err != nil {
		return err
	}
	return nil
}
func deleteDeviceProfileByName(n string) error {
	var dp models.DeviceProfile
	getDeviceProfileByName(&dp, n)
	// Delete all of the commands for the device profile
	for i := 0; i < len(dp.Commands); i++ {
		// TODO Figure out how to store MONGO ID
		if err := deleteById(COMCOL, dp.Commands[i].Id.Hex()); err != nil {
			return err
		}
	}
	// Delete the device profile
	if err := deleteByName(DPCOL, n); err != nil {
		return err
	}
	return nil
}

//func getDeviceProfilesByCommandName(d *[]models.DeviceProfile, cn string) error {
//	if strings.Compare(DATABASE, MONGOSTR) == 0 {
//		return mgoGetDeviceProfilesByCommandName(d, cn)
//	}
//	return nil
//}
func getDeviceProfilesByModel(dp *[]models.DeviceProfile, m string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceProfilesByModel(dp, m)
	}
	return nil
}
func getDeviceProfilesWithLabel(dp *[]models.DeviceProfile, l []string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceProfilesWithLabel(dp, l)
	}
	return nil
}
func getDeviceProfilesByManufacturerModel(dp *[]models.DeviceProfile, man string, mod string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceProfilesByManufacturerModel(dp, man, mod)
	}
	return nil
}
func getDeviceProfilesByManufacturer(dp *[]models.DeviceProfile, man string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceProfilesByManufacturer(dp, man)
	}
	return nil
}
func getDeviceProfileByName(dp *models.DeviceProfile, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceProfileByName(dp, n)
	}
	return nil
}
func updateAddressable(ra *models.Addressable, r *models.Addressable) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateAddressable(ra, r)
	}
	return nil
}
func addAddressable(a *models.Addressable) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoAddNewAddressable(a)
	}
	return nil
}
func getAddressableById(a *models.Addressable, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAddressableById(a, id)
	}
	return nil
}
func getAddressableByName(a *models.Addressable, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAddressableByName(a, n)
	}
	return nil
}
func getAddressablesByTopic(a *[]models.Addressable, t string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAddressablesByTopic(a, t)
	}
	return nil
}
func getAddressablesByPort(a *[]models.Addressable, p int) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAddressablesByPort(a, p)
	}
	return nil
}
func getAddressablesByPublisher(a *[]models.Addressable, p string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAddressablesByPublisher(a, p)
	}
	return nil
}
func getAddressablesByAddress(a *[]models.Addressable, add string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAddressablesByAddress(a, add)
	}
	return nil
}
func getAddressable(d *models.Addressable, q bson.M) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAddressable(d, q)
	}
	return nil
}
func getAddressables(d *[]models.Addressable, q bson.M) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAddressables(d, q)
	}
	return nil
}
func isAddressableAssociatedToDevice(a models.Addressable) (bool, error) {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoIsAddressableAssociatedToDevice(a)
	}
	return false, nil
}
func isAddressableAssociatedToDeviceService(a models.Addressable) (bool, error) {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoIsAddressableAssociatedToDeviceService(a)
	}
	return false, nil
}

// ------------------------ DEVICE SERVICE -----------------------

func updateDeviceService(ds models.DeviceService) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateDeviceService(ds)
	}
	return nil
}
func getDeviceServicesByAddressableName(d *[]models.DeviceService, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceServicesByAddressableName(d, n)
	}
	return nil
}
func getDeviceServicesByAddressableId(d *[]models.DeviceService, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceServicesByAddressableId(d, id)
	}
	return nil
}
func getDeviceServicesWithLabel(d *[]models.DeviceService, l []string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceServicesWithLabel(d, l)
	}
	return nil
}
func getDeviceServiceById(d *models.DeviceService, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceServiceById(d, id)
	}
	return nil
}
func getDeviceServiceByName(d *models.DeviceService, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceServiceByName(d, n)
	}
	return nil
}
func getAllDeviceServices(d *[]models.DeviceService) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAllDeviceServices(d)
	}
	return nil
}
func addDeviceService(ds *models.DeviceService) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoAddNewDeviceService(ds)
	}
	return nil
}

/* -----------------------Provision Watcher ----------------------*/
func getProvisionWatcherById(pw *models.ProvisionWatcher, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetProvisionWatcherById(pw, id)
	}
	return nil
}
func getAllProvisionWatchers(pw *[]models.ProvisionWatcher) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAllProvisionWatchers(pw)
	}
	return nil
}
func getProvisionWatcherByName(pw *models.ProvisionWatcher, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetProvisionWatcherByName(pw, n)
	}
	return nil
}
func getProvisionWatcherByProfileId(pw *[]models.ProvisionWatcher, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetProvisionWatcherByProfileId(pw, id)
	}
	return nil
}
func getProvisionWatchersByProfileName(pw *[]models.ProvisionWatcher, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetProvisionWatchersByProfileName(pw, n)
	}
	return nil
}
func getProvisionWatchersByServiceId(pw *[]models.ProvisionWatcher, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetProvisionWatchersByServiceId(pw, id)
	}
	return nil
}
func getProvisionWatchersByServiceName(pw *[]models.ProvisionWatcher, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetProvisionWatchersByServiceName(pw, n)
	}
	return nil
}
func getProvisionWatchersByIdentifier(pw *[]models.ProvisionWatcher, k string, v string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetProvisionWatchersByIdentifier(pw, k, v)
	}
	return nil
}
func addProvisionWatcher(pw *models.ProvisionWatcher) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoAddProvisionWatcher(pw)
	}
	return nil
}
func updateProvisionWatcher(pw models.ProvisionWatcher) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateProvisionWatcher(pw)
	}
	return nil
}

/* -----------------------COMMAND ----------------------*/
func getCommandById(c *models.Command, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		err := mgoGetCommandById(c, id)
		if err == mgo.ErrNotFound {
			return ErrNotFound
		} else {
			return err
		}
	}
	return nil
}
func getCommandByName(d *[]models.Command, id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetCommandByName(d, id)
	}
	return nil
}
func addCommand(c *models.Command) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoAddCommand(c)
	}
	return nil
}
func getAllCommands(d *[]models.Command) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetAllCommands(d)
	}
	return nil
}
func updateCommand(c *models.Command, r *models.Command) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateCommand(c, r)
	}
	return nil
}
func deleteByName(c string, n string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoDeleteByName(c, n)
	}
	return nil
}
func deleteCommandById(id string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoDeleteCommandById(id)
	}

	return nil
}

// Get the device profiles that are using the command
func getDeviceProfilesUsingCommand(dp *[]models.DeviceProfile, c models.Command) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoGetDeviceProfilesUsingCommand(dp, c)
	}

	return nil
}
func deleteById(c string, did string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoDeleteById(c, did)
	}
	return nil

}
func setByName(c string, n string, pv2 string, p2 string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateByName(c, n, pv2, p2)
	}
	return nil
}
func setByNameInt(c string, n string, pv2 string, p2 int64) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateByNameInt(c, n, pv2, p2)
	}
	return nil
}
func setById(c string, did string, pv2 string, p2 string) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateById(c, did, pv2, p2)
	}
	return nil
}
func setByIdInt(c string, did string, pv2 string, p2 int64) error {
	if strings.Compare(DATABASE, MONGOSTR) == 0 {
		return mgoUpdateByIdInt(c, did, pv2, p2)
	}
	return nil
}
