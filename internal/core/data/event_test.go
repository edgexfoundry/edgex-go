/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

package data

import (
	"strconv"
	"sync"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/google/uuid"
	"github.com/globalsign/mgo/bson"
)

//Test methods
func TestEventCount(t *testing.T) {
	reset()
	c, err := countEvents()
	if err != nil {
		t.Errorf(err.Error())
	}

	if c != 1 {
		t.Errorf("expected event count 1, received: %d", c)
	}
}

func TestCountByDevice(t *testing.T) {
	reset()
	count, err := countEventsByDevice(testEvent.Device)
	if err != nil {
		t.Errorf(err.Error())
	}

	if count == 0 {
		t.Errorf("no events found")
	}
}

func TestDeleteByAge(t *testing.T) {
	reset()
	count, err := deleteEventsByAge(-1)
	if err != nil {
		t.Errorf(err.Error())
	}

	if count == 0 {
		t.Errorf("no events deleted")
	}
}

func TestDeleteEventByAgeErrorThrownByEventsOlderThanAge(t *testing.T) {
	reset()
	dbClient = newMockDb()
	_, err := deleteEventsByAge(-1)

	if err == nil {
		t.Errorf("Should throw error")
	}
}

func TestGetEvents(t *testing.T) {
	reset()
	events, err := getEvents(0)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(events) == 0 {
		t.Errorf("no events found")
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event")
	}

	for e := range events {
		testEventWithoutReadings(events[e], t)
	}
}

func TestGetEventsWithLimit(t *testing.T) {
	reset()
	//Put an extra dummy event in the DB
	evt := models.Event{Device: testDeviceName, Origin: testOrigin}
	dbClient.AddEvent(evt)

	events, err := getEvents(1)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event")
	}
}

func TestAddEventWithPersistence(t *testing.T) {
	reset()
	Configuration.PersistData = true
	evt := models.Event{Device: testDeviceName, Origin: testOrigin, Readings: buildReadings()}
	//wire up handlers to listen for device events
	bitEvents := make([]bool, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go handleDomainEvents(bitEvents, &wg, t)

	newId, err := addNewEvent(evt)
	Configuration.PersistData = false
	if err != nil {
		t.Errorf(err.Error())
	}

	_, err = uuid.Parse(newId)
	if err != nil {
		t.Errorf("invalid UUID id: %s", newId)
	}

	wg.Wait()
	for i, val := range bitEvents {
		if !val {
			t.Errorf("event not received in timely fashion, index %v, TestAddEventWithPersistence", i)
		}
	}

	//verify we can load the new event from the database
	_, err = getEventById(newId)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestAddEventNoPersistence(t *testing.T) {
	reset()
	Configuration.PersistData = false
	evt := models.Event{Device: testDeviceName, Origin: testOrigin, Readings: buildReadings()}
	//wire up handlers to listen for device events
	bitEvents := make([]bool, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go handleDomainEvents(bitEvents, &wg, t)

	newId, err := addNewEvent(evt)
	if err != nil {
		t.Errorf(err.Error())
	}
	if bson.IsObjectIdHex(newId) {
		t.Errorf("unexpected bson id %s received", newId)
	}

	wg.Wait()
	for i, val := range bitEvents {
		if !val {
			t.Errorf("event not received in timely fashion, index %v, TestAddEventNoPersistence", i)
		}
	}

	//event was not persisted so we should not find it in the database
	_, err = getEventById(newId)
	if err != nil {
		if x, ok := err.(*errors.ErrEventNotFound); !ok {
			t.Errorf(x.Error())
		}
	}
}

func TestAddEventWithValidationValueDescriptorExistsAndIsInvalid(t *testing.T) {
	reset()
	dbClient = newMockDb()
	Configuration.ValidateCheck = true

	Configuration.PersistData = false
	evt := models.Event{Device: testDeviceName, Origin: testOrigin, Readings: buildReadings()[0:1]}
	//wire up handlers to listen for device events
	bitEvents := make([]bool, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go handleDomainEvents(bitEvents, &wg, t)

	_, err := addNewEvent(evt)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestAddEventWithValidationValueDescriptorNotFound(t *testing.T) {
	reset()
	Configuration.ValidateCheck = true

	Configuration.PersistData = false
	evt := models.Event{Device: testDeviceName, Origin: testOrigin, Readings: buildReadings()}
	//wire up handlers to listen for device events
	bitEvents := make([]bool, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go handleDomainEvents(bitEvents, &wg, t)

	_, err := addNewEvent(evt)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestAddEventWithValidationValueDescriptorDBError(t *testing.T) {
	reset()
	dbClient = newMockDb()
	Configuration.ValidateCheck = true

	Configuration.PersistData = false
	evt := models.Event{Device: testDeviceName, Origin: testOrigin, Readings: buildReadings()[1:]}
	//wire up handlers to listen for device events
	bitEvents := make([]bool, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go handleDomainEvents(bitEvents, &wg, t)

	_, err := addNewEvent(evt)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestUpdateEventNotFound(t *testing.T) {
	reset()
	evt := models.Event{ID: bson.NewObjectId().Hex(), Device: "Not Found", Origin: testOrigin}
	err := updateEvent(evt)
	if err != nil {
		if x, ok := err.(*errors.ErrEventNotFound); !ok {
			t.Errorf("unexpected error type: %s", x.Error())
		}
	} else {
		t.Errorf("expected ErrEventNotFound")
	}
}

func TestUpdateEventDeviceNotFound(t *testing.T) {
	Configuration.MetaDataCheck = true
	reset()
	evt := models.Event{ID: bson.NewObjectId().Hex(), Device: "Not Found", Origin: testOrigin}
	err := updateEvent(evt)
	if err == nil {
		t.Errorf("error expected")
	}
	Configuration.MetaDataCheck = false
}

func TestUpdateEvent(t *testing.T) {
	reset()
	evt := models.Event{ID: testEvent.ID, Device: "Some Value", Origin: testOrigin}
	err := updateEvent(evt)
	if err != nil {
		t.Errorf(err.Error())
	}
	chk, err := dbClient.EventById(testEvent.ID)
	if err != nil {
		t.Errorf(err.Error())
	}
	if chk.Device != "Some Value" {
		t.Errorf("unexpected device value %s", chk.Device)
	}
}

func TestDeleteAllEvents(t *testing.T) {
	reset()
	err := deleteAllEvents()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestGetEventById(t *testing.T) {
	reset()
	_, err := getEventById(testEvent.ID)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestGetEventByIdNotFound(t *testing.T) {
	reset()
	_, err := getEventById("abcxyz")
	if err != nil {
		if x, ok := err.(*errors.ErrEventNotFound); !ok {
			t.Errorf(x.Error())
		}
	}
}

func TestUpdateEventPushDate(t *testing.T) {
	reset()
	old := testEvent.Pushed
	err := updateEventPushDate(testEvent.ID)
	if err != nil {
		t.Errorf(err.Error())
	}
	e, err := getEventById(testEvent.ID)
	if err != nil {
		t.Errorf(err.Error())
	}
	if old == e.Pushed {
		t.Errorf("event.pushed was not updated.")
	}
}

func TestDeleteEventById(t *testing.T) {
	reset()
	err := deleteEventById(testEvent.ID)
	if err != nil {
		t.Errorf(err.Error())
	}

	_, err = getEventById(testEvent.ID)
	if err != nil {
		if x, ok := err.(*errors.ErrEventNotFound); !ok {
			t.Errorf(x.Error())
		}
	}
}

func TestDeleteEvent(t *testing.T) {
	reset()
	err := deleteEvent(testEvent)

	if err != nil {
		t.Errorf(err.Error())
	}

	_, err = getEventById(testEvent.ID)
	if err != nil {
		if x, ok := err.(*errors.ErrEventNotFound); !ok {
			t.Errorf(x.Error())
		}
	}
}

func TestDeleteEventEventDoesNotExist(t *testing.T) {
	reset()
	testEvent.ID = "fake"
	err := deleteEvent(testEvent)

	if err == nil {
		t.Errorf("Event does not exist and should throw error")
	}
}

func TestDeleteEventReadingDoesNotExist(t *testing.T) {
	reset()
	testEvent.Readings[0].Id = "fake"
	err := deleteEvent(testEvent)

	if err == nil {
		t.Errorf("Reading does not exist and should throw error")
	}
}

func TestGetEventsByDeviceIdLimit(t *testing.T) {
	reset()
	dbClient = newMockDb()

	expectedList, expectedNil := getEventsByDeviceIdLimit(0, "valid")

	if expectedNil != nil {
		t.Errorf("Should not throw error")
	}

	if expectedList == nil {
		t.Errorf("Should return a list of events")
	}
}

func TestGetEventsByDeviceIdLimitDBThrowsError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	expectedNil, expectedErr := getEventsByDeviceIdLimit(0, "error")

	if expectedNil != nil {
		t.Errorf("Should not return list")
	}

	if expectedErr == nil {
		t.Errorf("Should throw error")
	}
}

func TestGetEventsByCreationTime(t *testing.T) {
	reset()
	dbClient = newMockDb()

	expectedReadings, expectedNil := getEventsByCreationTime(0, 0xF00D, 0)

	if expectedReadings == nil {
		t.Errorf("Should return Events")
	}

	if expectedNil != nil {
		t.Errorf("Should not throw error")
	}
}

func TestGetEventsByCreationTimeDBThrowsError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	expectedNil, expectedErr := getEventsByCreationTime(0, 0xBADF00D, 0)

	if expectedNil != nil {
		t.Errorf("Should not return list")
	}

	if expectedErr == nil {
		t.Errorf("Should throw error")
	}
}

func TestDeleteEvents(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, expectedNil := deleteEvents("valid")

	if expectedNil != nil {
		t.Errorf("Should not throw error")
	}
}

func TestDeleteEventsDBLookupThrowsError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	expectedZero, expectedErr := deleteEvents("error")

	if expectedZero != 0 {
		t.Errorf("Should return zero on error")
	}

	if expectedErr == nil {
		t.Errorf("Should throw error")
	}
}

func TestScrubPushedEvents(t *testing.T) {
	reset()

	pushedEvent := testEvent
	pushedEvent.Pushed = -1
	pushedEvent.ID = "pushed"
	dbClient.AddEvent(pushedEvent)

	expectedCount := 1

	actualCount, expectedNil := scrubPushedEvents()

	if actualCount != expectedCount {
		t.Errorf("Expected %d deletions, was %d", expectedCount, actualCount)
	}

	if expectedNil != nil {
		t.Errorf("Should not throw error")
	}
}

func testEventWithoutReadings(event models.Event, t *testing.T) {
	if event.ID != testEvent.ID {
		t.Error("eventId mismatch. expected " + testEvent.ID + " received " + event.ID)
	}

	if event.Device != testEvent.Device {
		t.Error("device mismatch. expected " + testDeviceName + " received " + event.Device)
	}

	if event.Origin != testEvent.Origin {
		t.Error("origin mismatch. expected " + strconv.FormatInt(testEvent.Origin, 10) + " received " + strconv.FormatInt(event.Origin, 10))
	}
}
