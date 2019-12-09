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
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/messaging"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/interfaces/mocks"
	correlation "github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/mock"
)

// Test methods
func TestEventCount(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventCount").Return(1, nil)

	c, err := countEvents(dbClientMock)
	if err != nil {
		t.Errorf(err.Error())
	}

	if c != 1 {
		t.Errorf("expected event count 1, received: %d", c)
	}
}

func TestCountByDevice(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventCountByDeviceId", mock.Anything).Return(2, nil)

	count, err := countEventsByDevice(testEvent.Device, context.Background(), dbClientMock)
	if err != nil {
		t.Errorf(err.Error())
	}

	if count == 0 {
		t.Errorf("no events found")
	}
}

func buildEvents() []models.Event {
	events := []models.Event{}
	events = append(events, models.Event{
		ID:     "1",
		Device: testDeviceName,
		Readings: []models.Reading{
			{Id: "1"},
			{Id: "2"},
		},
	})
	return events
}

func newDeleteEventsOlderThanAgeMockDB() *dbMock.DBClient {
	myMock := &dbMock.DBClient{}

	myMock.On("EventsOlderThanAge", mock.MatchedBy(func(age int64) bool {
		return age == -1
	})).Return(buildEvents(), nil).Maybe()

	myMock.On("DeleteReadingById", mock.MatchedBy(func(id string) bool {
		return id == "1"
	})).Return(nil)

	myMock.On("DeleteReadingById", mock.MatchedBy(func(id string) bool {
		return id == "2"
	})).Return(nil)

	myMock.On("DeleteEventById", mock.MatchedBy(func(id string) bool {
		return id == "1"
	})).Return(nil)

	return myMock
}

func TestDeleteByAge(t *testing.T) {
	reset()
	dbClientMock := newDeleteEventsOlderThanAgeMockDB()
	count, err := deleteEventsByAge(-1, logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf(err.Error())
	}

	if count == 0 {
		t.Errorf("deleteEventsByAge returned 0; expected non-zero")
	}

	dbClientMock.AssertExpectations(t)
}

func TestDeleteEventByAgeErrorThrownByEventsOlderThanAge(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventsOlderThanAge", mock.MatchedBy(func(age int64) bool {
		return age == -1
	})).Return([]models.Event{}, fmt.Errorf("some error"))

	_, err := deleteEventsByAge(-1, logger.NewMockClient(), dbClientMock)

	if err == nil {
		t.Errorf("Should throw error")
	}
}

func TestGetEvents(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("Events").Return([]models.Event{testEvent}, nil)

	events, err := getEvents(0, dbClientMock)
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

func newGetEventsWithLimitMockDB(expectedLimit int) *dbMock.DBClient {
	myMock := &dbMock.DBClient{}

	myMock.On("EventsWithLimit", mock.MatchedBy(func(limit int) bool {
		return limit == expectedLimit
	})).Return(func(limit int) []models.Event {
		events := make([]models.Event, 0)
		for i := 0; i < limit; i++ {
			events = append(events, testEvent)
		}
		return events
	}, nil)

	return myMock
}

func TestGetEventsWithLimit(t *testing.T) {
	reset()

	limit := 1
	dbClientMock := newGetEventsWithLimitMockDB(limit)

	events, err := getEvents(limit, dbClientMock)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(events) != limit {
		t.Errorf("expected %d event", limit)
	}

	dbClientMock.AssertExpectations(t)
}

func newAddEventMockDB(persist bool) *dbMock.DBClient {
	myMock := &dbMock.DBClient{}

	if persist {
		myMock.On("AddEvent", mock.Anything).Return("3c5badcb-2008-47f2-ba78-eb2d992f8422", nil)
	}

	return myMock
}

func TestAddEventWithPersistence(t *testing.T) {
	reset()

	// no need to mock this since it's all in process
	msgClient, _ := messaging.NewMessageClient(msgTypes.MessageBusConfig{
		PublishHost: msgTypes.HostInfo{
			Host:     "*",
			Protocol: "tcp",
			Port:     5563,
		},
		Type: "zero",
	})

	dbClientMock := newAddEventMockDB(true)
	chEvents := make(chan interface{}, 10)
	Configuration.Writable.PersistData = true
	evt := models.Event{Device: testDeviceName, Origin: testOrigin, Readings: buildReadings()}
	// wire up handlers to listen for device events
	bitEvents := make([]bool, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go handleDomainEvents(bitEvents, chEvents, &wg, t)

	_, err := addNewEvent(
		correlation.Event{Event: evt},
		context.Background(),
		logger.NewMockClient(),
		dbClientMock,
		chEvents,
		msgClient)

	Configuration.Writable.PersistData = false
	if err != nil {
		t.Errorf(err.Error())
	}

	wg.Wait()
	for i, val := range bitEvents {
		if !val {
			t.Errorf("event not received in timely fashion, index %v, TestAddEventWithPersistence", i)
		}
	}

	dbClientMock.AssertExpectations(t)
}

func TestAddEventNoPersistence(t *testing.T) {
	reset()
	msgClient, _ := messaging.NewMessageClient(msgTypes.MessageBusConfig{
		PublishHost: msgTypes.HostInfo{
			Host:     "*",
			Protocol: "tcp",
			Port:     5563,
		},
		Type: "zero",
	})

	dbClientMock := newAddEventMockDB(false)
	Configuration.Writable.PersistData = false
	evt := models.Event{Device: testDeviceName, Origin: testOrigin, Readings: buildReadings()}
	// wire up handlers to listen for device events
	bitEvents := make([]bool, 2)
	chEvents := make(chan interface{})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go handleDomainEvents(bitEvents, chEvents, &wg, t)

	newId, err := addNewEvent(
		correlation.Event{Event: evt},
		context.Background(),
		logger.NewMockClient(),
		dbClientMock,
		chEvents,
		msgClient)

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

	dbClientMock.AssertExpectations(t)
}

func TestUpdateEventNotFound(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventById", mock.Anything).Return(models.Event{}, fmt.Errorf("Event not found"))

	evt := models.Event{ID: bson.NewObjectId().Hex(), Device: "Not Found", Origin: testOrigin}
	err := updateEvent(correlation.Event{Event: evt}, context.Background(), dbClientMock)
	if err != nil {
		if x, ok := err.(errors.ErrEventNotFound); !ok {
			t.Errorf("unexpected error type: %s", x.Error())
		}
	} else {
		t.Errorf("expected ErrEventNotFound")
	}
}

func newUpdateEventMockDB(expectedDevice string) *dbMock.DBClient {
	myMock := &dbMock.DBClient{}

	myMock.On("EventById", mock.MatchedBy(func(id string) bool {
		return id == testEvent.ID
	})).Return(testEvent, nil).Maybe()

	myMock.On("UpdateEvent", mock.MatchedBy(func(event correlation.Event) bool {
		return event.ID == testEvent.ID && event.Device == expectedDevice
	})).Return(nil)

	return myMock
}

func TestUpdateEvent(t *testing.T) {
	reset()
	expectedDevice := "Some Value"
	dbClientMock := newUpdateEventMockDB(expectedDevice)

	evt := models.Event{ID: testEvent.ID, Device: expectedDevice, Origin: testOrigin}
	err := updateEvent(correlation.Event{Event: evt}, context.Background(), dbClientMock)
	if err != nil {
		t.Errorf(err.Error())
	}

	dbClientMock.AssertExpectations(t)
}

func TestDeleteAllEvents(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ScrubAllEvents").Return(nil)

	err := deleteAllEvents(dbClientMock)
	if err != nil {
		t.Errorf(err.Error())
	}
	dbClientMock.AssertExpectations(t)
}

func TestGetEventById(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventById", mock.MatchedBy(func(id string) bool {
		return id == testEvent.ID
	})).Return(testEvent, nil)

	_, err := getEventById(testEvent.ID, dbClientMock)
	if err != nil {
		t.Errorf(err.Error())
	}

	dbClientMock.AssertExpectations(t)
}

func TestGetEventByIdNotFound(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventById", mock.Anything).Return(testEvent, db.ErrNotFound)

	_, err := getEventById("abcxyz", dbClientMock)
	if err != nil {
		if x, ok := err.(errors.ErrEventNotFound); !ok {
			t.Errorf(x.Error())
		}
	}

	dbClientMock.AssertExpectations(t)
}

func TestUpdateEventPushDate(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventById", mock.MatchedBy(func(id string) bool {
		return id == testEvent.ID
	})).Return(testEvent, nil)
	dbClientMock.On("UpdateEvent", mock.MatchedBy(func(event correlation.Event) bool {
		return event.ID == testEvent.ID
	})).Return(nil)

	err := updateEventPushDate(testEvent.ID, context.Background(), dbClientMock)
	if err != nil {
		t.Errorf(err.Error())
	}

	dbClientMock.AssertExpectations(t)
}

func newDeleteEventMockDB() *dbMock.DBClient {
	myMock := &dbMock.DBClient{}
	myMock.On("EventById", mock.MatchedBy(func(id string) bool {
		return id == testEvent.ID
	})).Return(testEvent, nil)
	myMock.On("DeleteReadingById", mock.MatchedBy(func(id string) bool {
		return id == testEvent.Readings[0].Id
	})).Return(nil)
	myMock.On("DeleteReadingById", mock.MatchedBy(func(id string) bool {
		return id == testEvent.Readings[1].Id
	})).Return(nil)
	myMock.On("DeleteEventById", mock.MatchedBy(func(id string) bool {
		return id == testEvent.ID
	})).Return(nil)
	return myMock
}

func TestDeleteEventById(t *testing.T) {
	reset()
	dbClientMock := newDeleteEventMockDB()

	err := deleteEventById(testEvent.ID, logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf(err.Error())
	}

	dbClientMock.AssertExpectations(t)
}

func TestDeleteEvent(t *testing.T) {
	reset()
	dbClientMock := newDeleteEventMockDB()

	err := deleteEvent(testEvent, logger.NewMockClient(), dbClientMock)

	if err != nil {
		t.Errorf(err.Error())
	}

	_, err = getEventById(testEvent.ID, dbClientMock)
	if err != nil {
		if x, ok := err.(errors.ErrEventNotFound); !ok {
			t.Errorf(x.Error())
		}
	}

	dbClientMock.AssertExpectations(t)
}

func TestDeleteEventEventDoesNotExist(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteEventById", mock.MatchedBy(func(id string) bool {
		return id == testEvent.ID
	})).Return(db.ErrNotFound)
	dbClientMock.On("DeleteReadingsByDevice", mock.MatchedBy(func(device string) bool {
		return device == testDeviceName
	})).Return(nil)

	testEvent.Readings = nil
	err := deleteEvent(testEvent, logger.NewMockClient(), dbClientMock)

	if err == nil {
		t.Errorf("Event does not exist and should throw error")
	}
}

func TestDeleteEventReadingDoesNotExist(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteEventById", mock.MatchedBy(func(id string) bool {
		return id == testEvent.ID
	})).Return(db.ErrNotFound)
	dbClientMock.On("DeleteReadingById", mock.MatchedBy(func(id string) bool {
		return id == testEvent.Readings[0].Id || id == testEvent.Readings[1].Id
	})).Return(db.ErrNotFound)
	dbClientMock.On("DeleteReadingsByDevice", mock.MatchedBy(func(device string) bool {
		return device == testDeviceName
	})).Return(nil)

	err := deleteEvent(testEvent, logger.NewMockClient(), dbClientMock)

	if err == nil {
		t.Errorf("Reading does not exist and should throw error")
	}
}

func TestGetEventsByDeviceIdLimit(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}

	dbClientMock.On("EventsForDeviceLimit", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "valid"
	}), mock.Anything).Return([]models.Event{testEvent}, nil)

	expectedList, expectedNil := getEventsByDeviceIdLimit(0, "valid", logger.NewMockClient(), dbClientMock)

	if expectedNil != nil {
		t.Errorf("Should not throw error")
	}

	if expectedList == nil {
		t.Errorf("Should return a list of events")
	}

	if expectedList[0].ID != testEvent.ID {
		t.Errorf("Didn't successfully return testEvent")
	}
}

func TestGetEventsByDeviceIdLimitDBThrowsError(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}

	dbClientMock.On("EventsForDeviceLimit", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "error"
	}), mock.Anything).Return(nil, fmt.Errorf("some error"))

	expectedNil, expectedErr := getEventsByDeviceIdLimit(0, "error", logger.NewMockClient(), dbClientMock)

	if expectedNil != nil {
		t.Errorf("Should not return list")
	}

	if expectedErr == nil {
		t.Errorf("Should throw error")
	}
}

func TestGetEventsByCreationTime(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventsByCreationTime", mock.MatchedBy(func(start int64) bool {
		return start == 0xF00D
	}), mock.Anything, mock.Anything).Return([]models.Event{}, nil)

	expectedReadings, expectedNil := getEventsByCreationTime(0, 0xF00D, 0, logger.NewMockClient(), dbClientMock)

	if expectedReadings == nil {
		t.Errorf("Should return Events")
	}

	if expectedNil != nil {
		t.Errorf("Should not throw error")
	}
}

func TestGetEventsByCreationTimeDBThrowsError(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventsByCreationTime", mock.MatchedBy(func(start int64) bool {
		return start == 0xBADF00D
	}), mock.Anything, mock.Anything).Return(nil, fmt.Errorf("some error"))

	expectedNil, expectedErr := getEventsByCreationTime(0, 0xBADF00D, 0, logger.NewMockClient(), dbClientMock)

	if expectedNil != nil {
		t.Errorf("Should not return list")
	}

	if expectedErr == nil {
		t.Errorf("Should throw error")
	}
}

func TestDeleteEvents(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteEventsByDevice", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == testUUIDString
	})).Return(0, nil)

	_, expectedNil := deleteEvents(testUUIDString, dbClientMock)

	if expectedNil != nil {
		t.Errorf("Should not throw error")
	}

	dbClientMock.AssertExpectations(t)
}

func TestScrubPushedEvents(t *testing.T) {
	reset()

	pushedEvents := []models.Event{testEvent, testEvent}
	pushedEvents[1].ID = testUUIDString

	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventsPushed").Return(pushedEvents, nil)
	dbClientMock.On("DeleteReadingById", mock.MatchedBy(func(id string) bool {
		return id == pushedEvents[0].Readings[0].Id || id == pushedEvents[0].Readings[1].Id
	})).Return(nil).Times(4)

	dbClientMock.On("DeleteEventById", mock.MatchedBy(func(id string) bool {
		return id == pushedEvents[0].ID || id == pushedEvents[1].ID
	})).Return(nil).Twice()

	expectedCount := 2
	actualCount, expectedNil := scrubPushedEvents(logger.NewMockClient(), dbClientMock)

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
