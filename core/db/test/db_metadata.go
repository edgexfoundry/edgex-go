//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package test

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/edgex-go/core/db"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/core/metadata/interfaces"
	"gopkg.in/mgo.v2/bson"
)

func TestMetadataDB(t *testing.T, dbClient interfaces.DBClient) {

	testDBAddressables(t, dbClient)
	testDBCommand(t, dbClient)
	testDBDeviceService(t, dbClient)
	testDBSchedule(t, dbClient)
	testDBDeviceReport(t, dbClient)
	testDBScheduleEvent(t, dbClient)
	testDBDeviceProfile(t, dbClient)
	testDBDevice(t, dbClient)
	testDBProvisionWatcher(t, dbClient)

	dbClient.CloseSession()
	// Calling CloseSession twice to test that there is no panic when closing an
	// already closed db
	dbClient.CloseSession()
}

func getAddressable(i int, prefix string) models.Addressable {
	name := fmt.Sprintf("%sname%d", prefix, i)
	a := models.Addressable{}
	a.Name = name
	a.Protocol = name
	a.HTTPMethod = name
	a.Address = name
	a.Port = i
	a.Path = name
	a.Publisher = name
	a.User = name
	a.Password = name
	a.Topic = name
	return a
}

func getDeviceService(dbClient interfaces.DBClient, i int) (models.DeviceService, error) {
	name := fmt.Sprintf("name%d", i)
	ds := models.DeviceService{}
	ds.Name = name
	ds.AdminState = "ENABLED"
	ds.Addressable = getAddressable(i, "ds_")
	ds.Labels = append(ds.Labels, name)
	ds.OperatingState = "ENABLED"
	ds.LastConnected = 5
	ds.LastReported = 5
	ds.Description = name
	var err error
	_, err = dbClient.AddAddressable(&ds.Addressable)
	if err != nil {
		return ds, fmt.Errorf("Error creating addressable: %v", err)
	}
	return ds, nil
}

func getCommand(dbClient interfaces.DBClient, i int) models.Command {
	name := fmt.Sprintf("name%d", i)
	c := models.Command{}
	c.Name = name
	c.Put = &models.Put{}
	c.Get = &models.Get{}
	return c
}

func getDeviceProfile(dbClient interfaces.DBClient, i int) (models.DeviceProfile, error) {
	name := fmt.Sprintf("name%d", i)
	dp := models.DeviceProfile{}
	dp.Name = name
	dp.Manufacturer = name
	dp.Model = name
	dp.Labels = append(dp.Labels, name)
	dp.Objects = dp.Labels
	// TODO
	// dp.DeviceResources = append(dp.DeviceResources, name)
	// dp.Resources = append(dp.Resources, name)
	c := getCommand(dbClient, i)
	err := dbClient.AddCommand(&c)
	if err != nil {
		return dp, err
	}
	dp.Commands = append(dp.Commands, c)
	return dp, nil
}

func populateAddressable(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		var err error
		a := getAddressable(i, "")
		id, err = dbClient.AddAddressable(&a)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func populateCommand(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		c := getCommand(dbClient, i)
		err := dbClient.AddCommand(&c)
		if err != nil {
			return id, err
		}
		id = c.Id
	}
	return id, nil
}

func populateDeviceService(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId

	for i := 0; i < count; i++ {
		ds, err := getDeviceService(dbClient, i)
		if err != nil {
			return id, nil
		}
		err = dbClient.AddDeviceService(&ds)
		id = ds.Id
		if err != nil {
			return id, fmt.Errorf("Error creating device service: %v", err)
		}
	}
	return id, nil
}

func populateSchedule(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		var err error
		name := fmt.Sprintf("name%d", i)
		s := models.Schedule{}
		s.Name = name
		s.Start = name
		s.End = name
		s.Frequency = name
		s.Cron = name
		s.RunOnce = false
		err = dbClient.AddSchedule(&s)
		if err != nil {
			return id, err
		}
		id = s.Id
	}
	return id, nil
}

func populateDeviceReport(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		var err error
		name := fmt.Sprintf("name%d", i)
		dr := models.DeviceReport{}
		dr.Name = name
		dr.Device = name
		dr.Event = name
		dr.Expected = append(dr.Expected, name)
		err = dbClient.AddDeviceReport(&dr)
		if err != nil {
			return id, err
		}
		id = dr.Id
	}
	return id, nil
}

func populateScheduleEvent(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		var err error
		name := fmt.Sprintf("name%d", i)
		se := models.ScheduleEvent{}
		se.Name = name
		se.Schedule = name
		se.Parameters = name
		se.Service = name
		se.Addressable = getAddressable(i, "se_")
		_, err = dbClient.AddAddressable(&se.Addressable)
		if err != nil {
			return id, fmt.Errorf("Error creating addressable: %v", err)
		}
		err = dbClient.AddScheduleEvent(&se)
		if err != nil {
			return id, err
		}
		id = se.Id
	}
	return id, nil
}

func populateDevice(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		var err error
		name := fmt.Sprintf("name%d", i)
		d := models.Device{}
		d.Name = name
		d.AdminState = "ENABLED"
		d.OperatingState = "ENABLED"
		d.LastConnected = 4
		d.LastReported = 4
		d.Labels = append(d.Labels, name)

		d.Addressable = getAddressable(i, "device")
		_, err = dbClient.AddAddressable(&d.Addressable)
		if err != nil {
			return id, fmt.Errorf("Error creating addressable: %v", err)
		}

		d.Service, err = getDeviceService(dbClient, i)
		if err != nil {
			return id, nil
		}
		err = dbClient.AddDeviceService(&d.Service)
		if err != nil {
			return id, fmt.Errorf("Error creating DeviceService: %v", err)
		}
		d.Profile, err = getDeviceProfile(dbClient, i)
		if err != nil {
			return id, fmt.Errorf("Error getting DeviceProfile: %v", err)
		}
		err = dbClient.AddDeviceProfile(&d.Profile)
		if err != nil {
			return id, fmt.Errorf("Error creating DeviceProfile: %v", err)
		}
		err = dbClient.AddDevice(&d)
		if err != nil {
			return id, err
		}
		id = d.Id
	}
	return id, nil
}

func populateDeviceProfile(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		dp, err := getDeviceProfile(dbClient, i)
		if err != nil {
			return id, fmt.Errorf("Error getting DeviceProfile: %v", err)
		}
		err = dbClient.AddDeviceProfile(&dp)
		if err != nil {
			return id, err
		}
		id = dp.Id
	}
	return id, nil
}

func populateProvisionWatcher(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		var err error
		name := fmt.Sprintf("name%d", i)
		d := models.ProvisionWatcher{}
		d.Name = name
		d.OperatingState = "ENABLED"
		d.Identifiers = make(map[string]string)
		d.Identifiers["name"] = name

		d.Service, err = getDeviceService(dbClient, i)
		if err != nil {
			return id, err
		}
		err = dbClient.AddDeviceService(&d.Service)
		if err != nil {
			return id, fmt.Errorf("Error creating DeviceService: %v", err)
		}
		d.Profile, err = getDeviceProfile(dbClient, i)
		if err != nil {
			return id, fmt.Errorf("Error getting DeviceProfile: %v", err)
		}
		err = dbClient.AddDeviceProfile(&d.Profile)
		if err != nil {
			return id, fmt.Errorf("Error creating DeviceProfile: %v", err)
		}
		err = dbClient.AddProvisionWatcher(&d)
		if err != nil {
			return id, err
		}
		id = d.Id
	}
	return id, nil
}

func clearAddressables(t *testing.T, dbClient interfaces.DBClient) {
	var addrs []models.Addressable
	err := dbClient.GetAddressables(&addrs)
	if err != nil {
		t.Fatalf("Error getting addressables %v", err)
	}
	for _, a := range addrs {
		if err = dbClient.DeleteAddressable(a); err != nil {
			t.Fatalf("Error removing addressable %v: %v", a, err)
		}
	}
}

func clearDevices(t *testing.T, dbClient interfaces.DBClient) {
	var ds []models.Device
	err := dbClient.GetAllDevices(&ds)
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	for _, d := range ds {
		if err = dbClient.DeleteDevice(d); err != nil {
			t.Fatalf("Error removing device %v: %v", d, err)
		}
	}
}

func clearSchedules(t *testing.T, dbClient interfaces.DBClient) {
	var ss []models.Schedule
	err := dbClient.GetAllSchedules(&ss)
	if err != nil {
		t.Fatalf("Error getting schedule %v", err)
	}
	for _, s := range ss {
		if err = dbClient.DeleteSchedule(s); err != nil {
			t.Fatalf("Error removing schedule %v: %v", s, err)
		}
	}
}

func clearScheduleEvents(t *testing.T, dbClient interfaces.DBClient) {
	var ss []models.ScheduleEvent
	err := dbClient.GetAllScheduleEvents(&ss)
	if err != nil {
		t.Fatalf("Error getting schedule %v", err)
	}
	for _, s := range ss {
		if err = dbClient.DeleteScheduleEvent(s); err != nil {
			t.Fatalf("Error removing schedule %v: %v", s, err)
		}
	}
}

func clearDeviceServices(t *testing.T, dbClient interfaces.DBClient) {
	var dss []models.DeviceService
	err := dbClient.GetAllDeviceServices(&dss)
	if err != nil {
		t.Fatalf("Error getting deviceServices %v", err)
	}
	for _, ds := range dss {
		if err = dbClient.DeleteDeviceService(ds); err != nil {
			t.Fatalf("Error removing deviceService %v: %v", ds, err)
		}
	}
}

func clearDeviceReports(t *testing.T, dbClient interfaces.DBClient) {
	var drs []models.DeviceReport
	err := dbClient.GetAllDeviceReports(&drs)
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	for _, ds := range drs {
		if err = dbClient.DeleteDeviceReport(ds); err != nil {
			t.Fatalf("Error removing deviceReport %v: %v", ds, err)
		}
	}
}

func clearDeviceProfiles(t *testing.T, dbClient interfaces.DBClient) {
	var dps []models.DeviceProfile
	err := dbClient.GetAllDeviceProfiles(&dps)
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}

	for _, ds := range dps {
		if err = dbClient.DeleteDeviceProfile(ds); err != nil {
			t.Fatalf("Error removing deviceProfile %v: %v", ds, err)
		}
	}
}

func testDBAddressables(t *testing.T, dbClient interfaces.DBClient) {
	var addrs []models.Addressable

	clearAddressables(t, dbClient)

	id, err := populateAddressable(dbClient, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	// Error to have an Addressable with the same name
	_, err = populateAddressable(dbClient, 1)
	if err == nil {
		t.Fatalf("Should not be able to add a duplicated addressable\n")
	}

	err = dbClient.GetAddressables(&addrs)
	if err != nil {
		t.Fatalf("Error getting addressables %v", err)
	}
	if len(addrs) != 100 {
		t.Fatalf("There should be 100 addressables instead of %d", len(addrs))
	}
	a := models.Addressable{}
	err = dbClient.GetAddressableById(&a, id.Hex())
	if err != nil {
		t.Fatalf("Error getting addressable by id %v", err)
	}
	if a.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", a.Id, id)
	}
	err = dbClient.GetAddressableById(&a, "INVALID")
	if err == nil {
		t.Fatalf("Addressable should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}
	err = dbClient.GetAddressableByName(&a, "name1")
	if err != nil {
		t.Fatalf("Error getting addressable by name %v", err)
	}
	if a.Name != "name1" {
		t.Fatalf("name does not match %s - %s", a.Name, "name1")
	}
	err = dbClient.GetAddressableByName(&a, "INVALID")
	if err == nil {
		t.Fatalf("Addressable should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetAddressablesByTopic(&addrs, "name1")
	if err != nil {
		t.Fatalf("Error getting addressables by topic: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("There should be 1 addressable, not %d", len(addrs))
	}

	err = dbClient.GetAddressablesByTopic(&addrs, "INVALID")
	if err != nil {
		t.Fatalf("Error getting addressables by topic: %v", err)
	}
	if len(addrs) != 0 {
		t.Fatalf("There should be no addressables, not %d", len(addrs))
	}

	err = dbClient.GetAddressablesByPort(&addrs, 2)
	if err != nil {
		t.Fatalf("Error getting addressables by port: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("There should be 1 addressable, not %d", len(addrs))
	}

	err = dbClient.GetAddressablesByPort(&addrs, -1)
	if err != nil {
		t.Fatalf("Error getting addressables by port: %v", err)
	}
	if len(addrs) != 0 {
		t.Fatalf("There should be no addressables, not %d", len(addrs))
	}

	err = dbClient.GetAddressablesByPublisher(&addrs, "name1")
	if err != nil {
		t.Fatalf("Error getting addressables by publisher: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("There should be 1 addressable, not %d", len(addrs))
	}

	err = dbClient.GetAddressablesByPublisher(&addrs, "INVALID")
	if err != nil {
		t.Fatalf("Error getting addressables by publisher: %v", err)
	}
	if len(addrs) != 0 {
		t.Fatalf("There should be no addressables, not %d", len(addrs))
	}

	err = dbClient.GetAddressablesByAddress(&addrs, "name1")
	if err != nil {
		t.Fatalf("Error getting addressables by Address: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("There should be 1 addressable, not %d", len(addrs))
	}

	err = dbClient.GetAddressablesByAddress(&addrs, "INVALID")
	if err != nil {
		t.Fatalf("Error getting addressables by Address: %v", err)
	}
	if len(addrs) != 0 {
		t.Fatalf("There should be no addressables, not %d", len(addrs))
	}

	err = dbClient.GetAddressableById(&a, id.Hex())
	if err != nil {
		t.Fatalf("Error getting addressable %v", err)
	}
	err = dbClient.GetAddressableByName(&a, "name1")
	if err != nil {
		t.Fatalf("Error getting addressable %v", err)
	}

	aa := models.Addressable{}
	aa.Id = id
	aa.Name = "name"
	err = dbClient.UpdateAddressable(&aa, &a)
	if err != nil {
		t.Fatalf("Error updating Addressable %v", err)
	}
	err = dbClient.GetAddressableByName(&a, "name1")
	if err == nil {
		t.Fatalf("Addresable name1 should be renamed")
	}
	err = dbClient.GetAddressableByName(&a, "name")
	if err != nil {
		t.Fatalf("Addresable name should be renamed")
	}

	// aa.Name = "name2"
	// err = dbClient.UpdateAddressable(&aa, &a)
	// if err == nil {
	// 	t.Fatalf("Error updating Addressable %v", err)
	// }

	a.Id = "INVALID"
	a.Name = "INVALID"
	err = dbClient.DeleteAddressable(a)
	if err == nil {
		t.Fatalf("Addressable should not be deleted")
	}

	err = dbClient.GetAddressableByName(&a, "name")
	if err != nil {
		t.Fatalf("Error getting addressable")
	}
	err = dbClient.DeleteAddressable(a)
	if err != nil {
		t.Fatalf("Addressable should be deleted: %v", err)
	}

	clearAddressables(t, dbClient)
}

func testDBCommand(t *testing.T, dbClient interfaces.DBClient) {
	var commands []models.Command
	err := dbClient.GetAllCommands(&commands)
	if err != nil {
		t.Fatalf("Error getting commands %v", err)
	}
	for _, c := range commands {
		if err = dbClient.DeleteCommandById(c.Id.Hex()); err != nil {
			t.Fatalf("Error removing command %v", err)
		}
	}

	id, err := populateCommand(dbClient, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	err = dbClient.GetAllCommands(&commands)
	if err != nil {
		t.Fatalf("Error getting commands %v", err)
	}
	if len(commands) != 100 {
		t.Fatalf("There should be 100 commands instead of %d", len(commands))
	}
	c := models.Command{}
	err = dbClient.GetCommandById(&c, id.Hex())
	if err != nil {
		t.Fatalf("Error getting command by id %v", err)
	}
	if c.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", c.Id, id)
	}
	err = dbClient.GetCommandById(&c, "INVALID")
	if err == nil {
		t.Fatalf("Command should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetCommandByName(&commands, "name1")
	if err != nil {
		t.Fatalf("Error getting commands by name %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("There should be 1 commands instead of %d", len(commands))
	}

	err = dbClient.GetCommandByName(&commands, "INVALID")
	if err != nil {
		t.Fatalf("Error getting commands by name %v", err)
	}
	if len(commands) != 0 {
		t.Fatalf("There should be 1 commands instead of %d", len(commands))
	}

	cc := models.Command{}
	cc.Id = id
	cc.Get = &models.Get{}
	cc.Put = &models.Put{}
	cc.Name = "name"
	err = dbClient.UpdateCommand(&cc, &c)
	if err != nil {
		t.Fatalf("Error updating Command %v", err)
	}

	c.Id = "INVALID"
	err = dbClient.UpdateCommand(&cc, &c)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = dbClient.DeleteCommandById("INVALID")
	if err == nil {
		t.Fatalf("Command should not be deleted")
	}

	err = dbClient.DeleteCommandById(id.Hex())
	if err != nil {
		t.Fatalf("Command should be deleted: %v", err)
	}
}

func testDBDeviceService(t *testing.T, dbClient interfaces.DBClient) {
	var deviceServices []models.DeviceService

	clearDeviceServices(t, dbClient)
	id, err := populateDeviceService(dbClient, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	ds2 := models.DeviceService{}
	ds2.Name = "name1"
	err = dbClient.AddDeviceService(&ds2)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	err = dbClient.GetAllDeviceServices(&deviceServices)
	if err != nil {
		t.Fatalf("Error getting deviceServices %v", err)
	}
	if len(deviceServices) != 100 {
		t.Fatalf("There should be 100 deviceServices instead of %d", len(deviceServices))
	}
	ds := models.DeviceService{}
	err = dbClient.GetDeviceServiceById(&ds, id.Hex())
	if err != nil {
		t.Fatalf("Error getting deviceService by id %v", err)
	}
	if ds.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", ds.Id, id)
	}
	err = dbClient.GetDeviceServiceById(&ds, "INVALID")
	if err == nil {
		t.Fatalf("DeviceService should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetDeviceServiceByName(&ds, "name1")
	if err != nil {
		t.Fatalf("Error getting deviceServices by name %v", err)
	}
	if ds.Name != "name1" {
		t.Fatalf("The ds should be named name1 instead of %s", ds.Name)
	}

	err = dbClient.GetDeviceServiceByName(&ds, "INVALID")
	if err == nil {
		t.Fatalf("There should be a not found error")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetDeviceServicesByAddressableId(&deviceServices, ds.Addressable.Id.Hex())
	if err != nil {
		t.Fatalf("Error getting deviceServices by addressable id %v", err)
	}
	if len(deviceServices) != 1 {
		t.Fatalf("There should be 1 deviceServices instead of %d", len(deviceServices))
	}
	err = dbClient.GetDeviceServicesByAddressableId(&deviceServices, bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting deviceServices by addressable id")
	}

	err = dbClient.GetDeviceServicesWithLabel(&deviceServices, "name3")
	if err != nil {
		t.Fatalf("Error getting deviceServices by addressable id %v", err)
	}
	if len(deviceServices) != 1 {
		t.Fatalf("There should be 1 deviceServices instead of %d", len(deviceServices))
	}
	err = dbClient.GetDeviceServicesWithLabel(&deviceServices, "INVALID")
	if err != nil {
		t.Fatalf("Error getting deviceServices by addressable id %v", err)
	}
	if len(deviceServices) != 0 {
		t.Fatalf("There should be 0 deviceServices instead of %d", len(deviceServices))
	}

	ds.Id = id
	ds.Name = "name"
	err = dbClient.UpdateDeviceService(ds)
	if err != nil {
		t.Fatalf("Error updating DeviceService %v", err)
	}

	ds.Id = "INVALID"
	err = dbClient.UpdateDeviceService(ds)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = dbClient.DeleteDeviceService(ds)
	if err == nil {
		t.Fatalf("DeviceService should not be deleted")
	}

	ds.Id = id
	err = dbClient.DeleteDeviceService(ds)
	if err != nil {
		t.Fatalf("DeviceService should be deleted: %v", err)
	}

	clearDeviceServices(t, dbClient)
}

func testDBSchedule(t *testing.T, dbClient interfaces.DBClient) {
	var schedules []models.Schedule

	clearSchedules(t, dbClient)
	id, err := populateSchedule(dbClient, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	e := models.Schedule{}
	e.Name = "name1"
	err = dbClient.AddSchedule(&e)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	err = dbClient.GetAllSchedules(&schedules)
	if err != nil {
		t.Fatalf("Error getting schedules %v", err)
	}
	if len(schedules) != 100 {
		t.Fatalf("There should be 100 schedules instead of %d", len(schedules))
	}

	err = dbClient.GetScheduleById(&e, id.Hex())
	if err != nil {
		t.Fatalf("Error getting schedule by id %v", err)
	}
	if e.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	err = dbClient.GetScheduleById(&e, "INVALID")
	if err == nil {
		t.Fatalf("Schedule should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetScheduleByName(&e, "name1")
	if err != nil {
		t.Fatalf("Error getting schedule by id %v", err)
	}
	if e.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	err = dbClient.GetScheduleByName(&e, "INVALID")
	if err == nil {
		t.Fatalf("Schedule should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	e2 := models.Schedule{}
	e2.Id = id
	e2.Name = "name"
	err = dbClient.UpdateSchedule(e2)
	if err != nil {
		t.Fatalf("Error updating Schedule %v", err)
	}

	e2.Id = "INVALID"
	err = dbClient.UpdateSchedule(e2)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = dbClient.DeleteSchedule(e2)
	if err == nil {
		t.Fatalf("Schedule should not be deleted")
	}

	e2.Id = id
	err = dbClient.DeleteSchedule(e2)
	if err != nil {
		t.Fatalf("Schedule should be deleted: %v", err)
	}
}

func testDBDeviceReport(t *testing.T, dbClient interfaces.DBClient) {
	var deviceReports []models.DeviceReport

	clearDeviceReports(t, dbClient)

	id, err := populateDeviceReport(dbClient, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	e := models.DeviceReport{}
	e.Name = "name1"
	err = dbClient.AddDeviceReport(&e)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	err = dbClient.GetAllDeviceReports(&deviceReports)
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	if len(deviceReports) != 100 {
		t.Fatalf("There should be 100 deviceReports instead of %d", len(deviceReports))
	}

	err = dbClient.GetDeviceReportById(&e, id.Hex())
	if err != nil {
		t.Fatalf("Error getting deviceReport by id %v", err)
	}
	if e.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	err = dbClient.GetDeviceReportById(&e, "INVALID")
	if err == nil {
		t.Fatalf("DeviceReport should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetDeviceReportByName(&e, "name1")
	if err != nil {
		t.Fatalf("Error getting deviceReport by id %v", err)
	}
	if e.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	err = dbClient.GetDeviceReportByName(&e, "INVALID")
	if err == nil {
		t.Fatalf("DeviceReport should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetDeviceReportByDeviceName(&deviceReports, "name1")
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	if len(deviceReports) != 1 {
		t.Fatalf("There should be 1 deviceReports instead of %d", len(deviceReports))
	}

	err = dbClient.GetDeviceReportByDeviceName(&deviceReports, "name")
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	if len(deviceReports) != 0 {
		t.Fatalf("There should be 0 deviceReports instead of %d", len(deviceReports))
	}

	err = dbClient.GetDeviceReportsByScheduleEventName(&deviceReports, "name1")
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	if len(deviceReports) != 1 {
		t.Fatalf("There should be 1 deviceReports instead of %d", len(deviceReports))
	}

	err = dbClient.GetDeviceReportsByScheduleEventName(&deviceReports, "name")
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	if len(deviceReports) != 0 {
		t.Fatalf("There should be 0 deviceReports instead of %d", len(deviceReports))
	}

	e2 := models.DeviceReport{}
	e2.Id = id
	e2.Name = "name"
	err = dbClient.UpdateDeviceReport(&e2)
	if err != nil {
		t.Fatalf("Error updating DeviceReport %v", err)
	}

	e2.Id = "INVALID"
	err = dbClient.UpdateDeviceReport(&e2)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = dbClient.DeleteDeviceReport(e2)
	if err == nil {
		t.Fatalf("DeviceReport should not be deleted")
	}

	e2.Id = id
	err = dbClient.DeleteDeviceReport(e2)
	if err != nil {
		t.Fatalf("DeviceReport should be deleted: %v", err)
	}
}

func testDBScheduleEvent(t *testing.T, dbClient interfaces.DBClient) {
	var scheduleEvents []models.ScheduleEvent

	clearScheduleEvents(t, dbClient)
	clearAddressables(t, dbClient)
	id, err := populateScheduleEvent(dbClient, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	e := models.ScheduleEvent{}
	e.Name = "name1"
	err = dbClient.AddScheduleEvent(&e)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}
	e.Name = "name_not_used"
	e.Addressable.Name = "unused"
	err = dbClient.AddScheduleEvent(&e)
	if err == nil {
		t.Fatalf("Should be an error adding an event with not existing addressable")
	}

	err = dbClient.GetAllScheduleEvents(&scheduleEvents)
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 100 {
		t.Fatalf("There should be 100 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = dbClient.GetScheduleEventById(&e, id.Hex())
	if err != nil {
		t.Fatalf("Error getting ScheduleEvent by id %v", err)
	}
	if e.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	err = dbClient.GetScheduleEventById(&e, "INVALID")
	if err == nil {
		t.Fatalf("ScheduleEvent should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetScheduleEventByName(&e, "name1")
	if err != nil {
		t.Fatalf("Error getting ScheduleEvent by id %v", err)
	}
	if e.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	err = dbClient.GetScheduleEventByName(&e, "INVALID")
	if err == nil {
		t.Fatalf("ScheduleEvent should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetScheduleEventsByScheduleName(&scheduleEvents, "name1")
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 1 {
		t.Fatalf("There should be 1 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = dbClient.GetScheduleEventsByScheduleName(&scheduleEvents, "name")
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 0 {
		t.Fatalf("There should be 0 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = dbClient.GetScheduleEventsByAddressableId(&scheduleEvents, e.Addressable.Id.Hex())
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 1 {
		t.Fatalf("There should be 1 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = dbClient.GetScheduleEventsByAddressableId(&scheduleEvents, bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 0 {
		t.Fatalf("There should be 0 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = dbClient.GetScheduleEventsByServiceName(&scheduleEvents, "name1")
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 1 {
		t.Fatalf("There should be 1 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = dbClient.GetScheduleEventsByServiceName(&scheduleEvents, "name")
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 0 {
		t.Fatalf("There should be 0 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	e.Id = id
	e.Name = "name"
	err = dbClient.UpdateScheduleEvent(e)
	if err != nil {
		t.Fatalf("Error updating ScheduleEvent %v", err)
	}

	e.Id = "INVALID"
	err = dbClient.UpdateScheduleEvent(e)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = dbClient.DeleteScheduleEvent(e)
	if err == nil {
		t.Fatalf("ScheduleEvent should not be deleted")
	}

	e.Id = id
	err = dbClient.DeleteScheduleEvent(e)
	if err != nil {
		t.Fatalf("ScheduleEvent should be deleted: %v", err)
	}
}

func testDBDeviceProfile(t *testing.T, dbClient interfaces.DBClient) {
	var deviceProfiles []models.DeviceProfile

	clearAddressables(t, dbClient)
	clearDeviceProfiles(t, dbClient)
	id, err := populateDeviceProfile(dbClient, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	dp := models.DeviceProfile{}
	dp.Name = "name1"
	err = dbClient.AddDeviceProfile(&dp)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	err = dbClient.GetAllDeviceProfiles(&deviceProfiles)
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 100 {
		t.Fatalf("There should be 100 deviceProfiles instead of %d", len(deviceProfiles))
	}

	err = dbClient.GetDeviceProfileById(&dp, id.Hex())
	if err != nil {
		t.Fatalf("Error getting deviceProfile by id %v", err)
	}
	if dp.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", dp.Id, id)
	}
	err = dbClient.GetDeviceProfileById(&dp, "INVALID")
	if err == nil {
		t.Fatalf("DeviceProfile should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetDeviceProfileByName(&dp, "name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfile by id %v", err)
	}
	if dp.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", dp.Id, id)
	}
	err = dbClient.GetDeviceProfileByName(&dp, "INVALID")
	if err == nil {
		t.Fatalf("DeviceProfile should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetDeviceProfilesByModel(&deviceProfiles, "name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 1 {
		t.Fatalf("There should be 1 deviceProfiles instead of %d", len(deviceProfiles))
	}

	err = dbClient.GetDeviceProfilesByModel(&deviceProfiles, "name")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 0 {
		t.Fatalf("There should be 0 deviceProfiles instead of %d", len(deviceProfiles))
	}

	err = dbClient.GetDeviceProfilesByManufacturer(&deviceProfiles, "name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 1 {
		t.Fatalf("There should be 1 deviceProfiles instead of %d", len(deviceProfiles))
	}

	err = dbClient.GetDeviceProfilesByManufacturer(&deviceProfiles, "name")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 0 {
		t.Fatalf("There should be 0 deviceProfiles instead of %d", len(deviceProfiles))
	}

	err = dbClient.GetDeviceProfilesByManufacturerModel(&deviceProfiles, "name1", "name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 1 {
		t.Fatalf("There should be 1 deviceProfiles instead of %d", len(deviceProfiles))
	}

	err = dbClient.GetDeviceProfilesByManufacturerModel(&deviceProfiles, "name", "name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 0 {
		t.Fatalf("There should be 0 deviceProfiles instead of %d", len(deviceProfiles))
	}

	err = dbClient.GetDeviceProfilesWithLabel(&deviceProfiles, "name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 1 {
		t.Fatalf("There should be 1 deviceProfiles instead of %d", len(deviceProfiles))
	}

	err = dbClient.GetDeviceProfilesWithLabel(&deviceProfiles, "name")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 0 {
		t.Fatalf("There should be 0 deviceProfiles instead of %d", len(deviceProfiles))
	}

	c := models.Command{}
	c.Id = dp.Commands[0].Id

	err = dbClient.GetDeviceProfilesUsingCommand(&deviceProfiles, c)
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 1 {
		t.Fatalf("There should be 1 deviceProfiles instead of %d", len(deviceProfiles))
	}

	c.Id = bson.NewObjectId()
	err = dbClient.GetDeviceProfilesUsingCommand(&deviceProfiles, c)
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 0 {
		t.Fatalf("There should be 0 deviceProfiles instead of %d", len(deviceProfiles))
	}

	d2 := models.DeviceProfile{}
	d2.Id = id
	d2.Name = "name"
	err = dbClient.UpdateDeviceProfile(&d2)
	if err != nil {
		t.Fatalf("Error updating DeviceProfile %v", err)
	}

	d2.Id = "INVALID"
	err = dbClient.UpdateDeviceProfile(&d2)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = dbClient.DeleteDeviceProfile(d2)
	if err == nil {
		t.Fatalf("DeviceProfile should not be deleted")
	}

	d2.Id = id
	err = dbClient.DeleteDeviceProfile(d2)
	if err != nil {
		t.Fatalf("DeviceProfile should be deleted: %v", err)
	}

	clearDeviceProfiles(t, dbClient)
}

func testDBDevice(t *testing.T, dbClient interfaces.DBClient) {
	var devices []models.Device

	clearDeviceProfiles(t, dbClient)
	clearDeviceServices(t, dbClient)
	clearAddressables(t, dbClient)
	clearDevices(t, dbClient)
	id, err := populateDevice(dbClient, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	d := models.Device{}
	d.Name = "name1"
	err = dbClient.AddDevice(&d)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	err = dbClient.GetAllDevices(&devices)
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 100 {
		t.Fatalf("There should be 100 devices instead of %d", len(devices))
	}

	err = dbClient.GetDeviceById(&d, id.Hex())
	if err != nil {
		t.Fatalf("Error getting device by id %v", err)
	}
	if d.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", d.Id, id)
	}
	err = dbClient.GetDeviceById(&d, "INVALID")
	if err == nil {
		t.Fatalf("Device should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetDeviceByName(&d, "name1")
	if err != nil {
		t.Fatalf("Error getting device by id %v", err)
	}
	if d.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", d.Id, id)
	}
	err = dbClient.GetDeviceByName(&d, "INVALID")
	if err == nil {
		t.Fatalf("Device should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetDevicesByProfileId(&devices, d.Profile.Id.Hex())
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("There should be 1 devices instead of %d", len(devices))
	}

	err = dbClient.GetDevicesByProfileId(&devices, bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("There should be 0 devices instead of %d", len(devices))
	}

	err = dbClient.GetDevicesByServiceId(&devices, d.Service.Id.Hex())
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("There should be 1 devices instead of %d", len(devices))
	}

	err = dbClient.GetDevicesByServiceId(&devices, bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("There should be 0 devices instead of %d", len(devices))
	}

	err = dbClient.GetDevicesByAddressableId(&devices, d.Addressable.Id.Hex())
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("There should be 1 devices instead of %d", len(devices))
	}

	err = dbClient.GetDevicesByAddressableId(&devices, bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("There should be 0 devices instead of %d", len(devices))
	}

	err = dbClient.GetDevicesWithLabel(&devices, "name1")
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("There should be 1 devices instead of %d", len(devices))
	}

	err = dbClient.GetDevicesWithLabel(&devices, "name")
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("There should be 0 devices instead of %d", len(devices))
	}

	d.Id = id
	d.Name = "name"
	err = dbClient.UpdateDevice(d)
	if err != nil {
		t.Fatalf("Error updating Device %v", err)
	}

	d.Id = "INVALID"
	err = dbClient.UpdateDevice(d)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = dbClient.DeleteDevice(d)
	if err == nil {
		t.Fatalf("Device should not be deleted")
	}

	d.Id = id
	err = dbClient.DeleteDevice(d)
	if err != nil {
		t.Fatalf("Device should be deleted: %v", err)
	}
}

func testDBProvisionWatcher(t *testing.T, dbClient interfaces.DBClient) {
	var provisionWatchers []models.ProvisionWatcher

	clearDeviceProfiles(t, dbClient)
	clearDeviceServices(t, dbClient)
	clearAddressables(t, dbClient)
	id, err := populateProvisionWatcher(dbClient, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	pw := models.ProvisionWatcher{}
	pw.Name = "name1"
	err = dbClient.AddProvisionWatcher(&pw)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	err = dbClient.GetAllProvisionWatchers(&provisionWatchers)
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 100 {
		t.Fatalf("There should be 100 provisionWatchers instead of %d", len(provisionWatchers))
	}

	err = dbClient.GetProvisionWatcherById(&pw, id.Hex())
	if err != nil {
		t.Fatalf("Error getting provisionWatcher by id %v", err)
	}
	if pw.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", pw.Id, id)
	}
	err = dbClient.GetProvisionWatcherById(&pw, "INVALID")
	if err == nil {
		t.Fatalf("ProvisionWatcher should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetProvisionWatcherByName(&pw, "name1")
	if err != nil {
		t.Fatalf("Error getting provisionWatcher by id %v", err)
	}
	if pw.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", pw.Id, id)
	}
	err = dbClient.GetProvisionWatcherByName(&pw, "INVALID")
	if err == nil {
		t.Fatalf("ProvisionWatcher should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.GetProvisionWatchersByServiceId(&provisionWatchers, pw.Service.Id.Hex())
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 1 {
		t.Fatalf("There should be 1 provisionWatchers instead of %d", len(provisionWatchers))
	}

	err = dbClient.GetProvisionWatchersByServiceId(&provisionWatchers, bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 0 {
		t.Fatalf("There should be 0 provisionWatchers instead of %d", len(provisionWatchers))
	}

	err = dbClient.GetProvisionWatchersByProfileId(&provisionWatchers, pw.Profile.Id.Hex())
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 1 {
		t.Fatalf("There should be 1 provisionWatchers instead of %d", len(provisionWatchers))
	}

	err = dbClient.GetProvisionWatchersByProfileId(&provisionWatchers, bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 0 {
		t.Fatalf("There should be 0 provisionWatchers instead of %d", len(provisionWatchers))
	}

	err = dbClient.GetProvisionWatchersByIdentifier(&provisionWatchers, "name", "name1")
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 1 {
		t.Fatalf("There should be 1 provisionWatchers instead of %d", len(provisionWatchers))
	}

	err = dbClient.GetProvisionWatchersByIdentifier(&provisionWatchers, "name", "invalid")
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 0 {
		t.Fatalf("There should be 0 provisionWatchers instead of %d", len(provisionWatchers))
	}

	pw.Name = "name"
	err = dbClient.UpdateProvisionWatcher(pw)
	if err != nil {
		t.Fatalf("Error updating ProvisionWatcher %v", err)
	}

	pw.Id = "INVALID"
	err = dbClient.UpdateProvisionWatcher(pw)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = dbClient.DeleteProvisionWatcher(pw)
	if err == nil {
		t.Fatalf("ProvisionWatcher should not be deleted")
	}

	pw.Id = id
	err = dbClient.DeleteProvisionWatcher(pw)
	if err != nil {
		t.Fatalf("ProvisionWatcher should be deleted: %v", err)
	}
}
