//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package test

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
)

func TestMetadataDB(t *testing.T, db interfaces.DBClient) {
	// Remove previous metadata
	db.ScrubMetadata()

	testDBAddressables(t, db)
	testDBCommand(t, db)
	testDBDeviceService(t, db)
	testDBSchedule(t, db)
	testDBDeviceReport(t, db)
	testDBScheduleEvent(t, db)
	testDBDeviceProfile(t, db)
	testDBDevice(t, db)
	testDBProvisionWatcher(t, db)

	db.CloseSession()
	// Calling CloseSession twice to test that there is no panic when closing an
	// already closed db
	db.CloseSession()
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

func getDeviceService(db interfaces.DBClient, i int) (models.DeviceService, error) {
	name := fmt.Sprintf("name%d", i)
	ds := models.DeviceService{}
	ds.Name = name
	ds.AdminState = "UNLOCKED"
	ds.Addressable = getAddressable(i, "ds_")
	ds.Labels = append(ds.Labels, name)
	ds.OperatingState = "ENABLED"
	ds.LastConnected = 5
	ds.LastReported = 5
	ds.Description = name

	_, err := db.AddAddressable(ds.Addressable)
	if err != nil {
		return ds, fmt.Errorf("Error creating addressable: %v", err)
	}

	return ds, nil
}

func getCommand(db interfaces.DBClient, i int) models.Command {
	name := fmt.Sprintf("name%d", i)
	c := models.Command{}
	c.Name = name
	c.Put = &models.Put{}
	c.Get = &models.Get{}
	return c
}

func getDeviceProfile(db interfaces.DBClient, i int) (models.DeviceProfile, error) {
	name := fmt.Sprintf("name%d", i)
	dp := models.DeviceProfile{}
	dp.Name = name
	dp.Manufacturer = name
	dp.Model = name
	dp.Labels = append(dp.Labels, name)
	// TODO
	// dp.DeviceResources = append(dp.DeviceResources, name)
	// dp.Resources = append(dp.Resources, name)
	c := getCommand(db, i)
	newId, err := db.AddCommand(c)
	if err != nil {
		return dp, err
	}
	c.Id = newId
	dp.Commands = append(dp.Commands, c)
	return dp, nil
}

func populateAddressable(db interfaces.DBClient, count int) (string, error) {
	var id string
	for i := 0; i < count; i++ {
		var err error
		a := getAddressable(i, "")
		id, err = db.AddAddressable(a)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func populateCommand(db interfaces.DBClient, count int) (string, error) {
	var id string
	for i := 0; i < count; i++ {
		c := getCommand(db, i)
		newId, err := db.AddCommand(c)
		if err != nil {
			return id, err
		}
		id = newId
	}
	return id, nil
}

func populateDeviceService(db interfaces.DBClient, count int) (string, error) {
	var id string

	for i := 0; i < count; i++ {
		ds, err := getDeviceService(db, i)
		if err != nil {
			return id, nil
		}
		id, err = db.AddDeviceService(ds)
		if err != nil {
			return id, fmt.Errorf("Error creating device service: %v", err)
		}
	}
	return id, nil
}

func populateSchedule(db interfaces.DBClient, count int) (bson.ObjectId, error) {
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
		err = db.AddSchedule(&s)
		if err != nil {
			return id, err
		}
		id = s.Id
	}
	return id, nil
}

func populateDeviceReport(db interfaces.DBClient, count int) (string, error) {
	var id string
	for i := 0; i < count; i++ {
		var err error
		name := fmt.Sprintf("name%d", i)
		dr := models.DeviceReport{}
		dr.Name = name
		dr.Device = name
		dr.Event = name
		dr.Expected = append(dr.Expected, name)
		id, err = db.AddDeviceReport(dr)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func populateScheduleEvent(db interfaces.DBClient, count int) (bson.ObjectId, error) {
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
		_, err = db.AddAddressable(se.Addressable)
		if err != nil {
			return id, fmt.Errorf("Error creating addressable: %v", err)
		}
		err = db.AddScheduleEvent(&se)
		if err != nil {
			return id, err
		}
		id = se.Id
	}
	return id, nil
}

func populateDevice(db interfaces.DBClient, count int) (string, error) {
	var id string
	for i := 0; i < count; i++ {
		var err error
		name := fmt.Sprintf("name%d", i)
		d := models.Device{}
		d.Name = name
		d.AdminState = "UNLOCKED"
		d.OperatingState = "ENABLED"
		d.LastConnected = 4
		d.LastReported = 4
		d.Labels = append(d.Labels, name)

		d.Addressable = getAddressable(i, "device")
		_, err = db.AddAddressable(d.Addressable)
		if err != nil {
			return id, fmt.Errorf("Error creating addressable: %v", err)
		}

		d.Service, err = getDeviceService(db, i)
		if err != nil {
			return id, nil
		}
		_, err = db.AddDeviceService(d.Service)
		if err != nil {
			return id, fmt.Errorf("Error creating DeviceService: %v", err)
		}
		d.Profile, err = getDeviceProfile(db, i)
		if err != nil {
			return id, fmt.Errorf("Error getting DeviceProfile: %v", err)
		}
		id, err = db.AddDeviceProfile(d.Profile)
		if err != nil {
			return id, fmt.Errorf("Error creating DeviceProfile: %v", err)
		}
		id, err = db.AddDevice(d)
		if err != nil {
			return id, err
		}
		id = d.Id
	}
	return id, nil
}

func populateDeviceProfile(db interfaces.DBClient, count int) (string, error) {
	var id string
	for i := 0; i < count; i++ {
		dp, err := getDeviceProfile(db, i)
		if err != nil {
			return id, fmt.Errorf("Error getting DeviceProfile: %v", err)
		}
		id, err = db.AddDeviceProfile(dp)
		if err != nil {
			return id, err
		}
		id = dp.Id
	}
	return id, nil
}

func populateProvisionWatcher(db interfaces.DBClient, count int) (string, error) {
	var id string
	for i := 0; i < count; i++ {
		var err error
		name := fmt.Sprintf("name%d", i)
		d := models.ProvisionWatcher{}
		d.Name = name
		d.OperatingState = "ENABLED"
		d.Identifiers = make(map[string]string)
		d.Identifiers["name"] = name

		d.Service, err = getDeviceService(db, i)
		if err != nil {
			return id, err
		}
		_, err = db.AddDeviceService(d.Service)
		if err != nil {
			return id, fmt.Errorf("Error creating DeviceService: %v", err)
		}
		d.Profile, err = getDeviceProfile(db, i)
		if err != nil {
			return id, fmt.Errorf("Error getting DeviceProfile: %v", err)
		}
		id, err = db.AddDeviceProfile(d.Profile)
		if err != nil {
			return id, fmt.Errorf("Error creating DeviceProfile: %v", err)
		}
		id, err = db.AddProvisionWatcher(d)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func clearAddressables(t *testing.T, db interfaces.DBClient) {
	addrs, err := db.GetAddressables()
	if err != nil {
		t.Fatalf("Error getting addressables %v", err)
	}
	for _, a := range addrs {
		if err = db.DeleteAddressableById(a.Id); err != nil {
			t.Fatalf("Error removing addressable %v: %v", a, err)
		}
	}
}

func clearDevices(t *testing.T, db interfaces.DBClient) {
	ds, err := db.GetAllDevices()
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	for _, d := range ds {
		if err = db.DeleteDeviceById(d.Id); err != nil {
			t.Fatalf("Error removing device %v: %v", d, err)
		}
	}
}

func clearSchedules(t *testing.T, db interfaces.DBClient) {
	var ss []models.Schedule
	err := db.GetAllSchedules(&ss)
	if err != nil {
		t.Fatalf("Error getting schedule %v", err)
	}
	for _, s := range ss {
		if err = db.DeleteScheduleById(s.Id.Hex()); err != nil {
			t.Fatalf("Error removing schedule %v: %v", s, err)
		}
	}
}

func clearScheduleEvents(t *testing.T, db interfaces.DBClient) {
	var ss []models.ScheduleEvent
	err := db.GetAllScheduleEvents(&ss)
	if err != nil {
		t.Fatalf("Error getting schedule %v", err)
	}
	for _, s := range ss {
		if err = db.DeleteScheduleEventById(s.Id.Hex()); err != nil {
			t.Fatalf("Error removing schedule %v: %v", s, err)
		}
	}
}

func clearDeviceServices(t *testing.T, db interfaces.DBClient) {
	dss, err := db.GetAllDeviceServices()
	if err != nil {
		t.Fatalf("Error getting deviceServices %v", err)
	}
	for _, ds := range dss {
		if err = db.DeleteDeviceServiceById(ds.Id); err != nil {
			t.Fatalf("Error removing deviceService %v: %v", ds, err)
		}
	}
}

func clearDeviceReports(t *testing.T, db interfaces.DBClient) {
	drs, err := db.GetAllDeviceReports()
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	for _, ds := range drs {
		if err = db.DeleteDeviceReportById(ds.Id); err != nil {
			t.Fatalf("Error removing deviceReport %v: %v", ds, err)
		}
	}
}

func clearDeviceProfiles(t *testing.T, db interfaces.DBClient) {
	var dps []models.DeviceProfile
	dps, err := db.GetAllDeviceProfiles()
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}

	for _, ds := range dps {
		if err = db.DeleteDeviceProfileById(ds.Id); err != nil {
			t.Fatalf("Error removing deviceProfile %v: %v", ds, err)
		}
	}
}

func testDBAddressables(t *testing.T, db interfaces.DBClient) {
	var addrs []models.Addressable

	clearAddressables(t, db)

	id, err := populateAddressable(db, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	// Error to have an Addressable with the same name
	_, err = populateAddressable(db, 1)
	if err == nil {
		t.Fatalf("Should not be able to add a duplicated addressable\n")
	}

	addrs, err = db.GetAddressables()
	if err != nil {
		t.Fatalf("Error getting addressables %v", err)
	}
	if len(addrs) != 100 {
		t.Fatalf("There should be 100 addressables instead of %d", len(addrs))
	}
	a := models.Addressable{}
	a, err = db.GetAddressableById(id)
	if err != nil {
		t.Fatalf("Error getting addressable by id %v", err)
	}
	if a.Id != id {
		t.Fatalf("Id does not match %s - %s", a.Id, id)
	}
	a, err = db.GetAddressableById("INVALID")
	if err == nil {
		t.Fatalf("Addressable should not be found")
	}
	a, err = db.GetAddressableByName("name1")
	if err != nil {
		t.Fatalf("Error getting addressable by name %v", err)
	}
	if a.Name != "name1" {
		t.Fatalf("name does not match %s - %s", a.Name, "name1")
	}
	a, err = db.GetAddressableByName("INVALID")
	if err == nil {
		t.Fatalf("Addressable should not be found")
	}

	addrs, err = db.GetAddressablesByTopic("name1")
	if err != nil {
		t.Fatalf("Error getting addressables by topic: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("There should be 1 addressable, not %d", len(addrs))
	}

	addrs, err = db.GetAddressablesByTopic("INVALID")
	if err != nil {
		t.Fatalf("Error getting addressables by topic: %v", err)
	}
	if len(addrs) != 0 {
		t.Fatalf("There should be no addressables, not %d", len(addrs))
	}

	addrs, err = db.GetAddressablesByPort(2)
	if err != nil {
		t.Fatalf("Error getting addressables by port: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("There should be 1 addressable, not %d", len(addrs))
	}

	addrs, err = db.GetAddressablesByPort(-1)
	if err != nil {
		t.Fatalf("Error getting addressables by port: %v", err)
	}
	if len(addrs) != 0 {
		t.Fatalf("There should be no addressables, not %d", len(addrs))
	}

	addrs, err = db.GetAddressablesByPublisher("name1")
	if err != nil {
		t.Fatalf("Error getting addressables by publisher: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("There should be 1 addressable, not %d", len(addrs))
	}

	addrs, err = db.GetAddressablesByPublisher("INVALID")
	if err != nil {
		t.Fatalf("Error getting addressables by publisher: %v", err)
	}
	if len(addrs) != 0 {
		t.Fatalf("There should be no addressables, not %d", len(addrs))
	}

	addrs, err = db.GetAddressablesByAddress("name1")
	if err != nil {
		t.Fatalf("Error getting addressables by Address: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("There should be 1 addressable, not %d", len(addrs))
	}

	addrs, err = db.GetAddressablesByAddress("INVALID")
	if err != nil {
		t.Fatalf("Error getting addressables by Address: %v", err)
	}
	if len(addrs) != 0 {
		t.Fatalf("There should be no addressables, not %d", len(addrs))
	}

	a, err = db.GetAddressableById(id)
	if err != nil {
		t.Fatalf("Error getting addressable %v", err)
	}
	a, err = db.GetAddressableByName("name1")
	if err != nil {
		t.Fatalf("Error getting addressable %v", err)
	}

	a.Name = "name"
	err = db.UpdateAddressable(a)
	if err != nil {
		t.Fatalf("Error updating Addressable %v", err)
	}
	a, err = db.GetAddressableByName("name1")
	if err == nil {
		t.Fatalf("Addressable name1 should be renamed")
	}
	a, err = db.GetAddressableByName("name")
	if err != nil {
		t.Fatalf("Addressable name should be renamed")
	}

	// aa.Name = "name2"
	// err = db.UpdateAddressable(&aa, &a)
	// if err == nil {
	// 	t.Fatalf("Error updating Addressable %v", err)
	// }

	a.Id = "INVALID"
	a.Name = "INVALID"
	err = db.DeleteAddressableById(a.Id)
	if err == nil {
		t.Fatalf("Addressable should not be deleted")
	}

	a, err = db.GetAddressableByName("name")
	if err != nil {
		t.Fatalf("Error getting addressable")
	}
	err = db.DeleteAddressableById(a.Id)
	if err != nil {
		t.Fatalf("Addressable should be deleted: %v", err)
	}

	clearAddressables(t, db)
}

func testDBCommand(t *testing.T, db interfaces.DBClient) {
	commands, err := db.GetAllCommands()
	if err != nil {
		t.Fatalf("Error getting commands %v", err)
	}
	for _, c := range commands {
		if err = db.DeleteCommandById(c.Id); err != nil {
			t.Fatalf("Error removing command %v", err)
		}
	}

	id, err := populateCommand(db, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	commands, err = db.GetAllCommands()
	if err != nil {
		t.Fatalf("Error getting commands %v", err)
	}
	if len(commands) != 100 {
		t.Fatalf("There should be 100 commands instead of %d", len(commands))
	}
	c, err := db.GetCommandById(id)
	if err != nil {
		t.Fatalf("Error getting command by id %v", err)
	}
	if c.Id != id {
		t.Fatalf("Id does not match %s - %s", c.Id, id)
	}
	c, err = db.GetCommandById("INVALID")
	if err == nil {
		t.Fatalf("Command should not be found")
	}

	commands, err = db.GetCommandByName("name1")
	if err != nil {
		t.Fatalf("Error getting commands by name %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("There should be 1 commands instead of %d", len(commands))
	}

	commands, err = db.GetCommandByName("INVALID")
	if err != nil {
		t.Fatalf("Error getting commands by name %v", err)
	}
	if len(commands) != 0 {
		t.Fatalf("There should be 1 commands instead of %d", len(commands))
	}

	c.Id = id
	c.Get = &models.Get{}
	c.Put = &models.Put{}
	c.Name = "name"
	err = db.UpdateCommand(c)
	if err != nil {
		t.Fatalf("Error updating Command %v", err)
	}

	c.Id = "INVALID"
	err = db.UpdateCommand(c)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = db.DeleteCommandById("INVALID")
	if err == nil {
		t.Fatalf("Command should not be deleted")
	}

	err = db.DeleteCommandById(id)
	if err != nil {
		t.Fatalf("Command should be deleted: %v", err)
	}
}

func testDBDeviceService(t *testing.T, db interfaces.DBClient) {
	var deviceServices []models.DeviceService

	clearDeviceServices(t, db)
	clearAddressables(t, db)

	id, err := populateDeviceService(db, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	ds2 := models.DeviceService{}
	ds2.Name = "name1"
	_, err = db.AddDeviceService(ds2)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	deviceServices, err = db.GetAllDeviceServices()
	if err != nil {
		t.Fatalf("Error getting deviceServices %v", err)
	}
	if len(deviceServices) != 100 {
		t.Fatalf("There should be 100 deviceServices instead of %d", len(deviceServices))
	}
	ds, err := db.GetDeviceServiceById(id)
	if err != nil {
		t.Fatalf("Error getting deviceService by id %v", err)
	}
	if ds.Id != id {
		t.Fatalf("Id does not match %s - %s", ds.Id, id)
	}
	ds, err = db.GetDeviceServiceById("INVALID")
	if err == nil {
		t.Fatalf("DeviceService should not be found")
	}

	ds, err = db.GetDeviceServiceByName("name1")
	if err != nil {
		t.Fatalf("Error getting deviceServices by name %v", err)
	}
	if ds.Name != "name1" {
		t.Fatalf("The ds should be named name1 instead of %s", ds.Name)
	}

	ds, err = db.GetDeviceServiceByName("INVALID")
	if err == nil {
		t.Fatalf("There should be a not found error")
	}

	deviceServices, err = db.GetDeviceServicesByAddressableId(ds.Addressable.Id)
	if err != nil {
		t.Fatalf("Error getting deviceServices by addressable id %v", err)
	}
	if len(deviceServices) != 1 {
		t.Fatalf("There should be 1 deviceServices instead of %d", len(deviceServices))
	}
	deviceServices, err = db.GetDeviceServicesByAddressableId(bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting deviceServices by addressable id")
	}

	deviceServices, err = db.GetDeviceServicesWithLabel("name3")
	if err != nil {
		t.Fatalf("Error getting deviceServices by addressable id %v", err)
	}
	if len(deviceServices) != 1 {
		t.Fatalf("There should be 1 deviceServices instead of %d", len(deviceServices))
	}
	deviceServices, err = db.GetDeviceServicesWithLabel("INVALID")
	if err != nil {
		t.Fatalf("Error getting deviceServices by addressable id %v", err)
	}
	if len(deviceServices) != 0 {
		t.Fatalf("There should be 0 deviceServices instead of %d", len(deviceServices))
	}

	ds.Id = id
	ds.Name = "name"
	err = db.UpdateDeviceService(ds)
	if err != nil {
		t.Fatalf("Error updating DeviceService %v", err)
	}

	ds.Id = "INVALID"
	err = db.UpdateDeviceService(ds)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = db.DeleteDeviceServiceById(ds.Id)
	if err == nil {
		t.Fatalf("DeviceService should not be deleted")
	}

	ds.Id = id
	err = db.DeleteDeviceServiceById(ds.Id)
	if err != nil {
		t.Fatalf("DeviceService should be deleted: %v", err)
	}

	clearDeviceServices(t, db)
}

func testDBSchedule(t *testing.T, db interfaces.DBClient) {
	var schedules []models.Schedule

	clearSchedules(t, db)
	id, err := populateSchedule(db, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	e := models.Schedule{}
	e.Name = "name1"
	err = db.AddSchedule(&e)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	err = db.GetAllSchedules(&schedules)
	if err != nil {
		t.Fatalf("Error getting schedules %v", err)
	}
	if len(schedules) != 100 {
		t.Fatalf("There should be 100 schedules instead of %d", len(schedules))
	}

	err = db.GetScheduleById(&e, id.Hex())
	if err != nil {
		t.Fatalf("Error getting schedule by id %v", err)
	}
	if e.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	err = db.GetScheduleById(&e, "INVALID")
	if err == nil {
		t.Fatalf("Schedule should not be found")
	}

	err = db.GetScheduleByName(&e, "name1")
	if err != nil {
		t.Fatalf("Error getting schedule by id %v", err)
	}
	if e.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	err = db.GetScheduleByName(&e, "INVALID")
	if err == nil {
		t.Fatalf("Schedule should not be found")
	}

	e2 := models.Schedule{}
	e2.Id = id
	e2.Name = "name"
	err = db.UpdateSchedule(e2)
	if err != nil {
		t.Fatalf("Error updating Schedule %v", err)
	}

	e2.Id = "INVALID"
	err = db.UpdateSchedule(e2)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = db.DeleteScheduleById(e2.Id.Hex())
	if err == nil {
		t.Fatalf("Schedule should not be deleted")
	}

	e2.Id = id
	err = db.DeleteScheduleById(e2.Id.Hex())
	if err != nil {
		t.Fatalf("Schedule should be deleted: %v", err)
	}
}

func testDBDeviceReport(t *testing.T, db interfaces.DBClient) {
	var deviceReports []models.DeviceReport

	clearDeviceReports(t, db)

	id, err := populateDeviceReport(db, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	e := models.DeviceReport{}
	e.Name = "name1"
	_, err = db.AddDeviceReport(e)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	deviceReports, err = db.GetAllDeviceReports()
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	if len(deviceReports) != 100 {
		t.Fatalf("There should be 100 deviceReports instead of %d", len(deviceReports))
	}

	e, err = db.GetDeviceReportById(id)
	if err != nil {
		t.Fatalf("Error getting deviceReport by id %v", err)
	}
	if e.Id != id {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	e, err = db.GetDeviceReportById("INVALID")
	if err == nil {
		t.Fatalf("DeviceReport should not be found")
	}

	e, err = db.GetDeviceReportByName("name1")
	if err != nil {
		t.Fatalf("Error getting deviceReport by id %v", err)
	}
	if e.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	e, err = db.GetDeviceReportByName("INVALID")
	if err == nil {
		t.Fatalf("DeviceReport should not be found")
	}

	deviceReports, err = db.GetDeviceReportByDeviceName("name1")
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	if len(deviceReports) != 1 {
		t.Fatalf("There should be 1 deviceReports instead of %d", len(deviceReports))
	}

	deviceReports, err = db.GetDeviceReportByDeviceName("name")
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	if len(deviceReports) != 0 {
		t.Fatalf("There should be 0 deviceReports instead of %d", len(deviceReports))
	}

	deviceReports, err = db.GetDeviceReportsByScheduleEventName("name1")
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	if len(deviceReports) != 1 {
		t.Fatalf("There should be 1 deviceReports instead of %d", len(deviceReports))
	}

	deviceReports, err = db.GetDeviceReportsByScheduleEventName("name")
	if err != nil {
		t.Fatalf("Error getting deviceReports %v", err)
	}
	if len(deviceReports) != 0 {
		t.Fatalf("There should be 0 deviceReports instead of %d", len(deviceReports))
	}

	e2 := models.DeviceReport{}
	e2.Id = id
	e2.Name = "name"
	err = db.UpdateDeviceReport(e2)
	if err != nil {
		t.Fatalf("Error updating DeviceReport %v", err)
	}

	e2.Id = "INVALID"
	err = db.UpdateDeviceReport(e2)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = db.DeleteDeviceReportById(e2.Id)
	if err == nil {
		t.Fatalf("DeviceReport should not be deleted")
	}

	e2.Id = id
	err = db.DeleteDeviceReportById(e2.Id)
	if err != nil {
		t.Fatalf("DeviceReport should be deleted: %v", err)
	}
}

func testDBScheduleEvent(t *testing.T, db interfaces.DBClient) {
	var scheduleEvents []models.ScheduleEvent

	clearScheduleEvents(t, db)
	clearAddressables(t, db)
	id, err := populateScheduleEvent(db, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	e := models.ScheduleEvent{}
	e.Name = "name1"
	err = db.AddScheduleEvent(&e)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}
	e.Name = "name_not_used"
	e.Addressable.Name = "unused"
	err = db.AddScheduleEvent(&e)
	if err == nil {
		t.Fatalf("Should be an error adding an event with not existing addressable")
	}

	err = db.GetAllScheduleEvents(&scheduleEvents)
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 100 {
		t.Fatalf("There should be 100 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = db.GetScheduleEventById(&e, id.Hex())
	if err != nil {
		t.Fatalf("Error getting ScheduleEvent by id %v", err)
	}
	if e.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	err = db.GetScheduleEventById(&e, "INVALID")
	if err == nil {
		t.Fatalf("ScheduleEvent should not be found")
	}

	err = db.GetScheduleEventByName(&e, "name1")
	if err != nil {
		t.Fatalf("Error getting ScheduleEvent by id %v", err)
	}
	if e.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", e.Id, id)
	}
	err = db.GetScheduleEventByName(&e, "INVALID")
	if err == nil {
		t.Fatalf("ScheduleEvent should not be found")
	}

	err = db.GetScheduleEventsByScheduleName(&scheduleEvents, "name1")
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 1 {
		t.Fatalf("There should be 1 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = db.GetScheduleEventsByScheduleName(&scheduleEvents, "name")
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 0 {
		t.Fatalf("There should be 0 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = db.GetScheduleEventsByAddressableId(&scheduleEvents, e.Addressable.Id)
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 1 {
		t.Fatalf("There should be 1 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = db.GetScheduleEventsByAddressableId(&scheduleEvents, bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 0 {
		t.Fatalf("There should be 0 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = db.GetScheduleEventsByServiceName(&scheduleEvents, "name1")
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 1 {
		t.Fatalf("There should be 1 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	err = db.GetScheduleEventsByServiceName(&scheduleEvents, "name")
	if err != nil {
		t.Fatalf("Error getting ScheduleEvents %v", err)
	}
	if len(scheduleEvents) != 0 {
		t.Fatalf("There should be 0 ScheduleEvents instead of %d", len(scheduleEvents))
	}

	e.Id = id
	e.Name = "name"
	err = db.UpdateScheduleEvent(e)
	if err != nil {
		t.Fatalf("Error updating ScheduleEvent %v", err)
	}

	e.Id = "INVALID"
	err = db.UpdateScheduleEvent(e)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = db.DeleteScheduleEventById(e.Id.Hex())
	if err == nil {
		t.Fatalf("ScheduleEvent should not be deleted")
	}

	e.Id = id
	err = db.DeleteScheduleEventById(e.Id.Hex())
	if err != nil {
		t.Fatalf("ScheduleEvent should be deleted: %v", err)
	}
}

func testDBDeviceProfile(t *testing.T, db interfaces.DBClient) {
	var deviceProfiles []models.DeviceProfile

	clearAddressables(t, db)
	clearDeviceProfiles(t, db)
	id, err := populateDeviceProfile(db, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	dp := models.DeviceProfile{}
	dp.Name = "name1"
	id, err = db.AddDeviceProfile(dp)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	deviceProfiles, err = db.GetAllDeviceProfiles()
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 100 {
		t.Fatalf("There should be 100 deviceProfiles instead of %d", len(deviceProfiles))
	}

	dp, err = db.GetDeviceProfileById(id)
	if err != nil {
		t.Fatalf("Error getting deviceProfile by id %v", err)
	}
	if dp.Id != id {
		t.Fatalf("Id does not match %s - %s", dp.Id, id)
	}
	dp, err = db.GetDeviceProfileById("INVALID")
	if err == nil {
		t.Fatalf("DeviceProfile should not be found")
	}

	dp, err = db.GetDeviceProfileByName("name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfile by id %v", err)
	}
	if dp.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", dp.Id, id)
	}
	dp, err = db.GetDeviceProfileByName("INVALID")
	if err == nil {
		t.Fatalf("DeviceProfile should not be found")
	}

	deviceProfiles, err = db.GetDeviceProfilesByModel("name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 1 {
		t.Fatalf("There should be 1 deviceProfiles instead of %d", len(deviceProfiles))
	}

	deviceProfiles, err = db.GetDeviceProfilesByModel("name")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 0 {
		t.Fatalf("There should be 0 deviceProfiles instead of %d", len(deviceProfiles))
	}

	deviceProfiles, err = db.GetDeviceProfilesByManufacturer("name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 1 {
		t.Fatalf("There should be 1 deviceProfiles instead of %d", len(deviceProfiles))
	}

	deviceProfiles, err = db.GetDeviceProfilesByManufacturer("name")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 0 {
		t.Fatalf("There should be 0 deviceProfiles instead of %d", len(deviceProfiles))
	}

	deviceProfiles, err = db.GetDeviceProfilesByManufacturerModel("name1", "name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 1 {
		t.Fatalf("There should be 1 deviceProfiles instead of %d", len(deviceProfiles))
	}

	deviceProfiles, err = db.GetDeviceProfilesByManufacturerModel("name", "name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 0 {
		t.Fatalf("There should be 0 deviceProfiles instead of %d", len(deviceProfiles))
	}

	deviceProfiles, err = db.GetDeviceProfilesWithLabel("name1")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 1 {
		t.Fatalf("There should be 1 deviceProfiles instead of %d", len(deviceProfiles))
	}

	deviceProfiles, err = db.GetDeviceProfilesWithLabel("name")
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 0 {
		t.Fatalf("There should be 0 deviceProfiles instead of %d", len(deviceProfiles))
	}

	c := models.Command{}
	c.Id = dp.Commands[0].Id

	deviceProfiles, err = db.GetDeviceProfilesUsingCommand(c)
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 1 {
		t.Fatalf("There should be 1 deviceProfiles instead of %d", len(deviceProfiles))
	}

	c.Id = bson.NewObjectId().Hex()
	deviceProfiles, err = db.GetDeviceProfilesUsingCommand(c)
	if err != nil {
		t.Fatalf("Error getting deviceProfiles %v", err)
	}
	if len(deviceProfiles) != 0 {
		t.Fatalf("There should be 0 deviceProfiles instead of %d", len(deviceProfiles))
	}

	d2 := models.DeviceProfile{}
	d2.Id = id
	d2.Name = "name"
	err = db.UpdateDeviceProfile(d2)
	if err != nil {
		t.Fatalf("Error updating DeviceProfile %v", err)
	}

	d2.Id = "INVALID"
	err = db.UpdateDeviceProfile(d2)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = db.DeleteDeviceProfileById(d2.Id)
	if err == nil {
		t.Fatalf("DeviceProfile should not be deleted")
	}

	d2.Id = id
	err = db.DeleteDeviceProfileById(d2.Id)
	if err != nil {
		t.Fatalf("DeviceProfile should be deleted: %v", err)
	}

	clearDeviceProfiles(t, db)
}

func testDBDevice(t *testing.T, db interfaces.DBClient) {
	var devices []models.Device

	clearDeviceProfiles(t, db)
	clearDeviceServices(t, db)
	clearAddressables(t, db)
	clearDevices(t, db)
	id, err := populateDevice(db, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	d := models.Device{}
	d.Name = "name1"
	d.Id, err = db.AddDevice(d)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	devices, err = db.GetAllDevices()
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 100 {
		t.Fatalf("There should be 100 devices instead of %d", len(devices))
	}

	d, err = db.GetDeviceById(id)
	if err != nil {
		t.Fatalf("Error getting device by id %v", err)
	}
	if d.Id != id {
		t.Fatalf("Id does not match %s - %s", d.Id, id)
	}
	d, err = db.GetDeviceById("INVALID")
	if err == nil {
		t.Fatalf("Device should not be found")
	}

	d, err = db.GetDeviceByName("name1")
	if err != nil {
		t.Fatalf("Error getting device by id %v", err)
	}
	if d.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", d.Id, id)
	}
	d, err = db.GetDeviceByName("INVALID")
	if err == nil {
		t.Fatalf("Device should not be found")
	}

	devices, err = db.GetDevicesByProfileId(d.Profile.Id)
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("There should be 1 devices instead of %d", len(devices))
	}

	devices, err = db.GetDevicesByProfileId(bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("There should be 0 devices instead of %d", len(devices))
	}

	devices, err = db.GetDevicesByServiceId(d.Service.Id)
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("There should be 1 devices instead of %d", len(devices))
	}

	devices, err = db.GetDevicesByServiceId(bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("There should be 0 devices instead of %d", len(devices))
	}

	devices, err = db.GetDevicesByAddressableId(d.Addressable.Id)
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("There should be 1 devices instead of %d", len(devices))
	}

	devices, err = db.GetDevicesByAddressableId(bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("There should be 0 devices instead of %d", len(devices))
	}

	devices, err = db.GetDevicesWithLabel("name1")
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("There should be 1 devices instead of %d", len(devices))
	}

	devices, err = db.GetDevicesWithLabel("name")
	if err != nil {
		t.Fatalf("Error getting devices %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("There should be 0 devices instead of %d", len(devices))
	}

	d.Id = id
	d.Name = "name"
	err = db.UpdateDevice(d)
	if err != nil {
		t.Fatalf("Error updating Device %v", err)
	}

	d.Id = "INVALID"
	err = db.UpdateDevice(d)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = db.DeleteDeviceById(d.Id)
	if err == nil {
		t.Fatalf("Device should not be deleted")
	}

	d.Id = id
	err = db.DeleteDeviceById(d.Id)
	if err != nil {
		t.Fatalf("Device should be deleted: %v", err)
	}
}

func testDBProvisionWatcher(t *testing.T, db interfaces.DBClient) {
	var provisionWatchers []models.ProvisionWatcher

	clearDeviceProfiles(t, db)
	clearDeviceServices(t, db)
	clearAddressables(t, db)
	id, err := populateProvisionWatcher(db, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	pw := models.ProvisionWatcher{}
	pw.Name = "name1"
	_, err = db.AddProvisionWatcher(pw)
	if err == nil {
		t.Fatalf("Should be an error adding an existing name")
	}

	provisionWatchers, err = db.GetAllProvisionWatchers()
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 100 {
		t.Fatalf("There should be 100 provisionWatchers instead of %d", len(provisionWatchers))
	}

	pw, err = db.GetProvisionWatcherById(id)
	if err != nil {
		t.Fatalf("Error getting provisionWatcher by id %v", err)
	}
	if pw.Id != id {
		t.Fatalf("Id does not match %s - %s", pw.Id, id)
	}
	pw, err = db.GetProvisionWatcherById("INVALID")
	if err == nil {
		t.Fatalf("ProvisionWatcher should not be found")
	}

	pw, err = db.GetProvisionWatcherByName("name1")
	if err != nil {
		t.Fatalf("Error getting provisionWatcher by id %v", err)
	}
	if pw.Name != "name1" {
		t.Fatalf("Id does not match %s - %s", pw.Id, id)
	}
	pw, err = db.GetProvisionWatcherByName("INVALID")
	if err == nil {
		t.Fatalf("ProvisionWatcher should not be found")
	}

	provisionWatchers, err = db.GetProvisionWatchersByServiceId(pw.Service.Id)
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 1 {
		t.Fatalf("There should be 1 provisionWatchers instead of %d", len(provisionWatchers))
	}

	provisionWatchers, err = db.GetProvisionWatchersByServiceId(bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 0 {
		t.Fatalf("There should be 0 provisionWatchers instead of %d", len(provisionWatchers))
	}

	provisionWatchers, err = db.GetProvisionWatchersByProfileId(pw.Profile.Id)
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 1 {
		t.Fatalf("There should be 1 provisionWatchers instead of %d", len(provisionWatchers))
	}

	provisionWatchers, err = db.GetProvisionWatchersByProfileId(bson.NewObjectId().Hex())
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 0 {
		t.Fatalf("There should be 0 provisionWatchers instead of %d", len(provisionWatchers))
	}

	provisionWatchers, err = db.GetProvisionWatchersByIdentifier("name", "name1")
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 1 {
		t.Fatalf("There should be 1 provisionWatchers instead of %d", len(provisionWatchers))
	}

	provisionWatchers, err = db.GetProvisionWatchersByIdentifier("name", "invalid")
	if err != nil {
		t.Fatalf("Error getting provisionWatchers %v", err)
	}
	if len(provisionWatchers) != 0 {
		t.Fatalf("There should be 0 provisionWatchers instead of %d", len(provisionWatchers))
	}

	pw.Name = "name"
	err = db.UpdateProvisionWatcher(pw)
	if err != nil {
		t.Fatalf("Error updating ProvisionWatcher %v", err)
	}

	pw.Id = "INVALID"
	err = db.UpdateProvisionWatcher(pw)
	if err == nil {
		t.Fatalf("Should return error")
	}

	err = db.DeleteProvisionWatcherById(pw.Id)
	if err == nil {
		t.Fatalf("ProvisionWatcher should not be deleted")
	}

	pw.Id = id
	err = db.DeleteProvisionWatcherById(pw.Id)
	if err != nil {
		t.Fatalf("ProvisionWatcher should be deleted: %v", err)
	}
}
