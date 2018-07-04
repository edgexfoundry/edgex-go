//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/core/data/interfaces"
	"github.com/edgexfoundry/edgex-go/core/db"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"gopkg.in/mgo.v2/bson"
)

func populateDbReadings(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("name%d", i)
		r := models.Reading{}
		r.Name = name
		r.Device = name
		r.Value = name
		var err error
		id, err = dbClient.AddReading(r)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func populateDbValues(dbClient interfaces.DBClient, count int) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("name%d", i)
		v := models.ValueDescriptor{}
		v.Name = name
		v.Description = name
		v.Type = name
		v.UomLabel = name
		v.Labels = []string{name, "LABEL"}
		var err error
		id, err = dbClient.AddValueDescriptor(v)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func populateDbEvents(dbClient interfaces.DBClient, count int, pushed int64) (bson.ObjectId, error) {
	var id bson.ObjectId
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("name%d", i)
		e := models.Event{}
		e.Device = name
		e.Event = name
		e.Pushed = pushed
		var err error
		id, err = dbClient.AddEvent(&e)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func testDBReadings(t *testing.T, dbClient interfaces.DBClient) {
	err := dbClient.ScrubAllEvents()
	if err != nil {
		t.Fatalf("Error removing all readings")
	}

	beforeTime := db.MakeTimestamp()
	id, err := populateDbReadings(dbClient, 100)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	// To have two readings with the same name
	id, err = populateDbReadings(dbClient, 10)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}
	afterTime := db.MakeTimestamp()

	count, err := dbClient.ReadingCount()
	if err != nil {
		t.Fatalf("Error getting readings count:  %v", err)
	}
	if count != 110 {
		t.Fatalf("There should be 110 readings instead of %d", count)
	}

	readings, err := dbClient.Readings()
	if err != nil {
		t.Fatalf("Error getting readings %v", err)
	}
	if len(readings) != 110 {
		t.Fatalf("There should be 110 readings instead of %d", len(readings))
	}
	r3, err := dbClient.ReadingById(id.Hex())
	if err != nil {
		t.Fatalf("Error getting reading by id %v", err)
	}
	if r3.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", r3.Id, id)
	}
	_, err = dbClient.ReadingById("INVALID")
	if err == nil {
		t.Fatalf("Reading should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	readings, err = dbClient.ReadingsByDeviceAndValueDescriptor("name1", "name1", 10)
	if err != nil {
		t.Fatalf("Error getting ReadingsByDeviceAndValueDescriptor: %v", err)
	}
	if len(readings) != 2 {
		t.Fatalf("There should be 2 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByDeviceAndValueDescriptor("name1", "name1", 1)
	if err != nil {
		t.Fatalf("Error getting ReadingsByDeviceAndValueDescriptor: %v", err)
	}
	if len(readings) != 1 {
		t.Fatalf("There should be 1 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByDeviceAndValueDescriptor("name20", "name20", 10)
	if err != nil {
		t.Fatalf("Error getting ReadingsByDeviceAndValueDescriptor: %v", err)
	}
	if len(readings) != 1 {
		t.Fatalf("There should be 1 readings, not %d", len(readings))
	}

	readings, err = dbClient.ReadingsByDevice("name1", 10)
	if err != nil {
		t.Fatalf("Error getting ReadingsByDevice: %v", err)
	}
	if len(readings) != 2 {
		t.Fatalf("There should be 2 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByDevice("name1", 1)
	if err != nil {
		t.Fatalf("Error getting ReadingsByDevice: %v", err)
	}
	if len(readings) != 1 {
		t.Fatalf("There should be 1 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByDevice("name20", 10)
	if err != nil {
		t.Fatalf("Error getting ReadingsByDevice: %v", err)
	}
	if len(readings) != 1 {
		t.Fatalf("There should be 1 readings, not %d", len(readings))
	}

	readings, err = dbClient.ReadingsByValueDescriptor("name1", 10)
	if err != nil {
		t.Fatalf("Error getting ReadingsByValueDescriptor: %v", err)
	}
	if len(readings) != 2 {
		t.Fatalf("There should be 2 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByValueDescriptor("name1", 1)
	if err != nil {
		t.Fatalf("Error getting ReadingsByValueDescriptor: %v", err)
	}
	if len(readings) != 1 {
		t.Fatalf("There should be 1 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByValueDescriptor("name20", 10)
	if err != nil {
		t.Fatalf("Error getting ReadingsByValueDescriptor: %v", err)
	}
	if len(readings) != 1 {
		t.Fatalf("There should be 1 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByValueDescriptor("name", 10)
	if err != nil {
		t.Fatalf("Error getting ReadingsByValueDescriptor: %v", err)
	}
	if len(readings) != 0 {
		t.Fatalf("There should be 0 readings, not %d", len(readings))
	}

	readings, err = dbClient.ReadingsByValueDescriptorNames([]string{"name1", "name2"}, 10)
	if err != nil {
		t.Fatalf("Error getting ReadingsByValueDescriptorNames: %v", err)
	}
	if len(readings) != 4 {
		t.Fatalf("There should be 4 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByValueDescriptorNames([]string{"name1", "name2"}, 1)
	if err != nil {
		t.Fatalf("Error getting ReadingsByValueDescriptorNames: %v", err)
	}
	if len(readings) != 1 {
		t.Fatalf("There should be 1 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByValueDescriptorNames([]string{"name", "noname"}, 10)
	if err != nil {
		t.Fatalf("Error getting ReadingsByValueDescriptorNames: %v", err)
	}
	if len(readings) != 0 {
		t.Fatalf("There should be 0 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByValueDescriptorNames([]string{"name20"}, 10)
	if err != nil {
		t.Fatalf("Error getting ReadingsByValueDescriptorNames: %v", err)
	}
	if len(readings) != 1 {
		t.Fatalf("There should be 1 readings, not %d", len(readings))
	}

	readings, err = dbClient.ReadingsByCreationTime(beforeTime, afterTime+10, 200)
	if err != nil {
		t.Fatalf("Error getting ReadingsByCreationTime: %v", err)
	}
	if len(readings) != 110 {
		t.Fatalf("There should be 110 readings, not %d", len(readings))
	}
	readings, err = dbClient.ReadingsByCreationTime(beforeTime, beforeTime+30, 100)
	if err != nil {
		t.Fatalf("Error getting ReadingsByCreationTime: %v", err)
	}
	if len(readings) != 100 {
		t.Fatalf("There should be 100 readings, not %d", len(readings))
	}

	r := models.Reading{}
	r.Id = id
	r.Name = "name"
	err = dbClient.UpdateReading(r)
	if err != nil {
		t.Fatalf("Error updating reading %v", err)
	}
	r2, err := dbClient.ReadingById(r.Id.Hex())
	if err != nil {
		t.Fatalf("Error getting reading by id %v", err)
	}
	if r2.Name != r.Name {
		t.Fatalf("Did not update reading correctly: %s %s", r.Name, r2.Name)
	}

	err = dbClient.DeleteReadingById("INVALID")
	if err == nil {
		t.Fatalf("Reading should not be deleted")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.DeleteReadingById(id.Hex())
	if err != nil {
		t.Fatalf("Reading should be deleted: %v", err)
	}

	err = dbClient.UpdateReading(r)
	if err == nil {
		t.Fatalf("Update should return error")
	}
}

func testDBEvents(t *testing.T, dbClient interfaces.DBClient) {
	err := dbClient.ScrubAllEvents()
	if err != nil {
		t.Fatalf("Error removing all events")
	}

	beforeTime := db.MakeTimestamp()
	id, err := populateDbEvents(dbClient, 100, 0)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	// To have two events with the same name
	id, err = populateDbEvents(dbClient, 10, 1)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}
	afterTime := db.MakeTimestamp()

	count, err := dbClient.EventCount()
	if err != nil {
		t.Fatalf("Error getting events count:  %v", err)
	}
	if count != 110 {
		t.Fatalf("There should be 110 events instead of %d", count)
	}

	count, err = dbClient.EventCountByDeviceId("name1")
	if err != nil {
		t.Fatalf("Error getting events count:  %v", err)
	}
	if count != 2 {
		t.Fatalf("There should be 2 events instead of %d", count)
	}

	count, err = dbClient.EventCountByDeviceId("name20")
	if err != nil {
		t.Fatalf("Error getting events count:  %v", err)
	}
	if count != 1 {
		t.Fatalf("There should be 1 events instead of %d", count)
	}

	count, err = dbClient.EventCountByDeviceId("name")
	if err != nil {
		t.Fatalf("Error getting events count:  %v", err)
	}
	if count != 0 {
		t.Fatalf("There should be 0 events instead of %d", count)
	}

	events, err := dbClient.Events()
	if err != nil {
		t.Fatalf("Error getting events %v", err)
	}
	if len(events) != 110 {
		t.Fatalf("There should be 110 events instead of %d", len(events))
	}
	e3, err := dbClient.EventById(id.Hex())
	if err != nil {
		t.Fatalf("Error getting event by id %v", err)
	}
	if e3.ID.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", e3.ID, id)
	}
	_, err = dbClient.EventById("INVALID")
	if err == nil {
		t.Fatalf("Event should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	events, err = dbClient.EventsForDeviceLimit("name1", 10)
	if err != nil {
		t.Fatalf("Error getting EventsForDeviceLimit: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("There should be 2 events, not %d", len(events))
	}
	events, err = dbClient.EventsForDeviceLimit("name1", 1)
	if err != nil {
		t.Fatalf("Error getting EventsForDeviceLimit: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("There should be 1 events, not %d", len(events))
	}
	events, err = dbClient.EventsForDeviceLimit("name20", 10)
	if err != nil {
		t.Fatalf("Error getting EventsForDeviceLimit: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("There should be 1 events, not %d", len(events))
	}
	events, err = dbClient.EventsForDeviceLimit("name", 10)
	if err != nil {
		t.Fatalf("Error getting EventsForDeviceLimit: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("There should be 0 events, not %d", len(events))
	}

	events, err = dbClient.EventsForDevice("name1")
	if err != nil {
		t.Fatalf("Error getting EventsForDevice: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("There should be 2 events, not %d", len(events))
	}
	events, err = dbClient.EventsForDevice("name20")
	if err != nil {
		t.Fatalf("Error getting EventsForDevice: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("There should be 1 events, not %d", len(events))
	}
	events, err = dbClient.EventsForDevice("name")
	if err != nil {
		t.Fatalf("Error getting EventsForDevice: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("There should be 0 events, not %d", len(events))
	}

	events, err = dbClient.EventsByCreationTime(beforeTime, afterTime+10, 200)
	if err != nil {
		t.Fatalf("Error getting EventsByCreationTime: %v", err)
	}
	if len(events) != 110 {
		t.Fatalf("There should be 110 events, not %d", len(events))
	}
	events, err = dbClient.EventsByCreationTime(beforeTime, afterTime+10, 100)
	if err != nil {
		t.Fatalf("Error getting EventsByCreationTime: %v", err)
	}
	if len(events) != 100 {
		t.Fatalf("There should be 100 events, not %d", len(events))
	}

	events, err = dbClient.EventsOlderThanAge(0)
	if err != nil {
		t.Fatalf("Error getting EventsOlderThanAge: %v", err)
	}
	if len(events) != 110 {
		t.Fatalf("There should be 110 events, not %d", len(events))
	}
	events, err = dbClient.EventsOlderThanAge(1000000)
	if err != nil {
		t.Fatalf("Error getting EventsOlderThanAge: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("There should be 0 events, not %d", len(events))
	}

	events, err = dbClient.EventsPushed()
	if err != nil {
		t.Fatalf("Error getting EventsOlderThanAge: %v", err)
	}
	if len(events) != 10 {
		t.Fatalf("There should be 10 events, not %d", len(events))
	}

	e := models.Event{}
	e.ID = id
	e.Device = "name"
	err = dbClient.UpdateEvent(e)
	if err != nil {
		t.Fatalf("Error updating event %v", err)
	}
	e2, err := dbClient.EventById(e.ID.Hex())
	if err != nil {
		t.Fatalf("Error getting event by id %v", err)
	}
	if e2.Device != e.Device {
		t.Fatalf("Did not update event correctly: %s %s", e.Device, e2.Device)
	}

	err = dbClient.DeleteEventById("INVALID")
	if err == nil {
		t.Fatalf("Event should not be deleted")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	err = dbClient.DeleteEventById(id.Hex())
	if err != nil {
		t.Fatalf("Event should be deleted: %v", err)
	}

	err = dbClient.UpdateEvent(e)
	if err == nil {
		t.Fatalf("Update should return error")
	}

	err = dbClient.ScrubAllEvents()
	if err != nil {
		t.Fatalf("Error removing all events")
	}

	events, err = dbClient.Events()
	if err != nil {
		t.Fatalf("Error getting events %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("There should be 0 events instead of %d", len(events))
	}
}

func testDBValueDescriptors(t *testing.T, dbClient interfaces.DBClient) {
	err := dbClient.ScrubAllValueDescriptors()
	if err != nil {
		t.Fatalf("Error removing all value descriptors")
	}

	id, err := populateDbValues(dbClient, 110)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	_, err = populateDbValues(dbClient, 110)
	if err == nil {
		t.Fatalf("Should be an error adding a new ValueDescriptor with the same name\n")
	}

	values, err := dbClient.ValueDescriptors()
	if err != nil {
		t.Fatalf("Error getting Values %v", err)
	}
	if len(values) != 110 {
		t.Fatalf("There should be 110 Values instead of %d", len(values))
	}

	v3, err := dbClient.ValueDescriptorById(id.Hex())
	if err != nil {
		t.Fatalf("Error getting Value by id %v", err)
	}
	if v3.Id.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", v3.Id, id)
	}
	_, err = dbClient.ValueDescriptorById("INVALID")
	if err == nil {
		t.Fatalf("Value should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	v3, err = dbClient.ValueDescriptorByName("name1")
	if err != nil {
		t.Fatalf("Error getting Value by id %v", err)
	}
	if v3.Name != "name1" {
		t.Fatalf("Name does not match %s - name1", v3.Name)
	}
	_, err = dbClient.ValueDescriptorByName("INVALID")
	if err == nil {
		t.Fatalf("Value should not be found")
	}
	if err != db.ErrNotFound {
		t.Fatalf("Error is not db.ErrNotFound: %v", err)
	}

	values, err = dbClient.ValueDescriptorsByName([]string{"name1", "name2"})
	if err != nil {
		t.Fatalf("Error getting ValuesByValueDescriptorNames: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("There should be 2 Values, not %d", len(values))
	}
	values, err = dbClient.ValueDescriptorsByName([]string{"name1", "name"})
	if err != nil {
		t.Fatalf("Error getting ValuesByValueDescriptorNames: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("There should be 1 Values, not %d", len(values))
	}
	values, err = dbClient.ValueDescriptorsByName([]string{"name", "INVALID"})
	if err != nil {
		t.Fatalf("Error getting ValuesByValueDescriptorNames: %v", err)
	}
	if len(values) != 0 {
		t.Fatalf("There should be 0 Values, not %d", len(values))
	}

	values, err = dbClient.ValueDescriptorsByUomLabel("name1")
	if err != nil {
		t.Fatalf("Error getting ValuesByValueDescriptorNames: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("There should be 1 Values, not %d", len(values))
	}
	values, err = dbClient.ValueDescriptorsByUomLabel("INVALID")
	if err != nil {
		t.Fatalf("Error getting ValuesByValueDescriptorNames: %v", err)
	}
	if len(values) != 0 {
		t.Fatalf("There should be 0 Values, not %d", len(values))
	}

	values, err = dbClient.ValueDescriptorsByLabel("name1")
	if err != nil {
		t.Fatalf("Error getting ValuesByValueDescriptorNames: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("There should be 1 Values, not %d", len(values))
	}
	values, err = dbClient.ValueDescriptorsByLabel("INVALID")
	if err != nil {
		t.Fatalf("Error getting ValuesByValueDescriptorNames: %v", err)
	}
	if len(values) != 0 {
		t.Fatalf("There should be 0 Values, not %d", len(values))
	}

	values, err = dbClient.ValueDescriptorsByType("name1")
	if err != nil {
		t.Fatalf("Error getting ValuesByValueDescriptorNames: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("There should be 1 Values, not %d", len(values))
	}
	values, err = dbClient.ValueDescriptorsByType("INVALID")
	if err != nil {
		t.Fatalf("Error getting ValuesByValueDescriptorNames: %v", err)
	}
	if len(values) != 0 {
		t.Fatalf("There should be 0 Values, not %d", len(values))
	}

	v := models.ValueDescriptor{}
	v.Id = id
	v.Name = "name"
	err = dbClient.UpdateValueDescriptor(v)
	if err != nil {
		t.Fatalf("Error updating Value %v", err)
	}
	v2, err := dbClient.ValueDescriptorById(v.Id.Hex())
	if err != nil {
		t.Fatalf("Error getting Value by id %v", err)
	}
	if v2.Name != v.Name {
		t.Fatalf("Did not update Value correctly: %s %s", v.Name, v2.Name)
	}

	err = dbClient.DeleteValueDescriptorById("INVALID")
	if err == nil {
		t.Fatalf("Value should not be deleted")
	}

	err = dbClient.DeleteValueDescriptorById(id.Hex())
	if err != nil {
		t.Fatalf("Value should be deleted: %v", err)
	}

	err = dbClient.UpdateValueDescriptor(v)
	if err == nil {
		t.Fatalf("Update should return error")
	}

	err = dbClient.ScrubAllValueDescriptors()
	if err != nil {
		t.Fatalf("Error removing all value descriptors")
	}
}

func TestDataDB(t *testing.T, dbClient interfaces.DBClient) {

	err := dbClient.Connect()
	if err != nil {
		t.Fatalf("Could not connect with mongodb: %v", err)
	}

	testDBReadings(t, dbClient)
	testDBEvents(t, dbClient)
	testDBValueDescriptors(t, dbClient)

	dbClient.CloseSession()
	// Calling CloseSession twice to test that there is no panic when closing an
	// already closed db
	dbClient.CloseSession()
}

func BenchmarkDB(b *testing.B, dbClient interfaces.DBClient) {

	benchmarkReadings(b, dbClient)
	benchmarkEvents(b, dbClient)
	dbClient.CloseSession()
}

func benchmarkReadings(b *testing.B, dbClient interfaces.DBClient) {

	// Remove previous events and readings
	dbClient.ScrubAllEvents()

	var readings []string

	b.Run("AddReading", func(b *testing.B) {
		reading := models.Reading{}
		for i := 0; i < b.N; i++ {
			reading.Name = "test" + strconv.Itoa(i)
			reading.Device = "device" + strconv.Itoa(i/100)
			id, err := dbClient.AddReading(reading)
			if err != nil {
				b.Fatalf("Error add reading: %v", err)
			}
			readings = append(readings, id.Hex())
		}
	})

	b.Run("Readings", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := dbClient.Readings()
			if err != nil {
				b.Fatalf("Error readings: %v", err)
			}
		}
	})

	b.Run("ReadingCount", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := dbClient.ReadingCount()
			if err != nil {
				b.Fatalf("Error reading count: %v", err)
			}
		}
	})

	b.Run("ReadingById", func(b *testing.B) {
		if b.N > len(readings) {
			b.N = len(readings)
		}
		for i := 0; i < b.N; i++ {
			_, err := dbClient.ReadingById(readings[i])
			if err != nil {
				b.Fatalf("Error reading by ID: %v", err)
			}
		}
	})

	b.Run("ReadingsByDevice", func(b *testing.B) {
		if b.N > len(readings)/10 {
			b.N = len(readings) / 10
		}
		for i := 0; i < b.N; i++ {
			device := "device" + strconv.Itoa(i)
			_, err := dbClient.ReadingsByDevice(device, 100)
			if err != nil {
				b.Fatalf("Error reading by device: %v", err)
			}
		}
	})
}

func benchmarkEvents(b *testing.B, dbClient interfaces.DBClient) {

	// Remove previous events and readings
	dbClient.ScrubAllEvents()

	var events []string

	b.Run("AddEvent", func(b *testing.B) {
		event := models.Event{}
		reading := models.Reading{}
		event.Readings = append(event.Readings, reading)
		event.Readings = append(event.Readings, reading)
		event.Readings = append(event.Readings, reading)
		event.Readings = append(event.Readings, reading)
		event.Readings = append(event.Readings, reading)
		for i := 0; i < b.N; i++ {
			event.Device = "device" + strconv.Itoa(i/100)
			id, err := dbClient.AddEvent(&event)
			if err != nil {
				b.Fatalf("Error add event: %v", err)
			}
			events = append(events, id.Hex())
		}
	})

	b.Run("Events", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := dbClient.Events()
			if err != nil {
				b.Fatalf("Error events: %v", err)
			}
		}
	})

	b.Run("EventCount", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := dbClient.EventCount()
			if err != nil {
				b.Fatalf("Error event count: %v", err)
			}
		}
	})

	b.Run("EventById", func(b *testing.B) {
		if b.N > len(events) {
			b.N = len(events)
		}
		for i := 0; i < b.N; i++ {
			_, err := dbClient.EventById(events[i])
			if err != nil {
				b.Fatalf("Error event by ID: %v", err)
			}
		}
	})

	b.Run("EventsForDevice", func(b *testing.B) {
		if b.N > len(events)/10 {
			b.N = len(events) / 10
		}
		for i := 0; i < b.N; i++ {
			device := "device" + strconv.Itoa(i)
			_, err := dbClient.EventsForDevice(device)
			if err != nil {
				b.Fatalf("Error events for device: %v", err)
			}
		}
	})
}
