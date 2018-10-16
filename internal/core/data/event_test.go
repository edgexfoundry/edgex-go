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
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	"math"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/interfaces/mocks"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/data/messaging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/memory"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata/mocks"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"gopkg.in/mgo.v2/bson"
)

var testEvent models.Event
var testRoutes *mux.Router

const (
	testDeviceName string = "Test Device"
	testOrigin     int64  = 123456789
)

// Mock implementation of the event publisher for testing purposes
type mockEventPublisher struct{}

func newMockEventPublisher(config messaging.PubSubConfiguration) messaging.EventPublisher {
	return &mockEventPublisher{}
}

func (zep *mockEventPublisher) SendEventMessage(e models.Event) error {
	return nil
}

func TestMain(m *testing.M) {
	testRoutes = LoadRestRoutes()
	LoggingClient = logger.NewMockClient()
	mdc = newMockDeviceClient()
	ep = newMockEventPublisher(messaging.PubSubConfiguration{})
	chEvents = make(chan interface{}, 10)
	os.Exit(m.Run())
}

//Test methods
func TestCount(t *testing.T) {
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
	dbClient.AddEvent(&evt)

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
	if !bson.IsObjectIdHex(newId) {
		t.Errorf("invalid bson id: %s", newId)
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
	Configuration.ValidateCheck= true

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
	Configuration.ValidateCheck= true

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
	Configuration.ValidateCheck= true

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
	evt := models.Event{ID: bson.NewObjectId(), Device: "Not Found", Origin: testOrigin}
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
	evt := models.Event{ID: bson.NewObjectId(), Device: "Not Found", Origin: testOrigin}
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
	chk, err := dbClient.EventById(testEvent.ID.Hex())
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
	_, err := getEventById(testEvent.ID.Hex())
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
	err := updateEventPushDate(testEvent.ID.Hex())
	if err != nil {
		t.Errorf(err.Error())
	}
	e, err := getEventById(testEvent.ID.Hex())
	if err != nil {
		t.Errorf(err.Error())
	}
	if old == e.Pushed {
		t.Errorf("event.pushed was not updated.")
	}
}

func TestDeleteEventById(t *testing.T) {
	reset()
	err := deleteEventById(testEvent.ID.Hex())
	if err != nil {
		t.Errorf(err.Error())
	}

	_, err = getEventById(testEvent.ID.Hex())
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

	_, err = getEventById(testEvent.ID.Hex())
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

func TestTestLimit(t *testing.T) {
	testedLimit := math.MinInt32

	expectedNil := checkMaxLimit(testedLimit)

	if expectedNil != nil {
		t.Errorf("Should not exceed limit")
	}
}

func TestTestLimitOverLimit(t *testing.T) {
	testedLimit := math.MaxInt32

	expectedErr := checkMaxLimit(testedLimit)

	if expectedErr == nil {
		t.Errorf("Exceeded limit and should throw error")
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
	dbClient.AddEvent(&pushedEvent)

	expectedCount:= 1

	actualCount, expectedNil := scrubPushedEvents()

	if actualCount != expectedCount {
		t.Errorf("Expected %d deletions, was %d", expectedCount, actualCount)
	}

	if expectedNil != nil {
		t.Errorf("Should not throw error")
	}
}

// Supporting methods
// Reset() re-initializes dependencies for each test
func reset() {
	Configuration = &ConfigurationStruct{}
	testEvent.Device = testDeviceName
	testEvent.Origin = testOrigin
	testEvent.Readings = buildReadings()
	dbClient = &memory.MemDB{}
	testEvent.ID, _ = dbClient.AddEvent(&testEvent)
}

func newMockDeviceClient() *mocks.DeviceClient {
	client := &mocks.DeviceClient{}

	mockAddressable := models.Addressable{
		Address:  "localhost",
		Name:     "Test Addressable",
		Port:     3000,
		Protocol: "http"}

	mockDeviceResultFn := func(id string) models.Device {
		if bson.IsObjectIdHex(id) {
			return models.Device{Id: bson.ObjectIdHex(id), Name: testEvent.Device, Addressable: mockAddressable}
		}
		return models.Device{}
	}
	client.On("Device", mock.MatchedBy(func(id string) bool {
		return bson.IsObjectIdHex(id)
	})).Return(mockDeviceResultFn, nil)
	client.On("Device", mock.MatchedBy(func(id string) bool {
		return !bson.IsObjectIdHex(id)
	})).Return(mockDeviceResultFn, fmt.Errorf("id is not bson ObjectIdHex"))

	mockDeviceForNameResultFn := func(name string) models.Device {
		device := models.Device{Id: bson.NewObjectId(), Name: name, Addressable: mockAddressable}

		return device
	}
	client.On("DeviceForName", mock.MatchedBy(func(name string) bool {
		return name == testEvent.Device
	})).Return(mockDeviceForNameResultFn, nil)
	client.On("DeviceForName", mock.MatchedBy(func(name string) bool {
		return name != testEvent.Device
	})).Return(mockDeviceForNameResultFn, fmt.Errorf("no device found for name"))

	return client
}

func newMockDb() interfaces.DBClient {
	db := &dbMock.DBClient{}

	db.On("EventsOlderThanAge", mock.MatchedBy(func(age int64) bool {
		return age == -1
	})).Return(nil, fmt.Errorf("expected testing error"))

	db.On("ValueDescriptorByName", mock.MatchedBy(func(name string) bool {
		return name == "Temperature"
	})).Return(models.ValueDescriptor{Type: "8"}, nil)

	db.On("ValueDescriptorByName", mock.MatchedBy(func(name string) bool {
		return name == "Pressure"
	})).Return(models.ValueDescriptor{}, fmt.Errorf("some error"))

	db.On("EventsForDeviceLimit", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "valid"
	}), mock.Anything).Return([]models.Event{}, nil)

	db.On("EventsForDeviceLimit", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "error"
	}), mock.Anything).Return(nil, fmt.Errorf("some error"))

	db.On("EventsForDevice", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "valid"
	})).Return([]models.Event{{Readings: append(buildReadings(), buildReadings()...)}}, nil)

	db.On("EventsForDevice", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "error"
	})).Return(nil, fmt.Errorf("some error"))

	db.On("EventsByCreationTime", mock.MatchedBy(func(start int64) bool {
		return start == 0xF00D
	}), mock.Anything, mock.Anything).Return([]models.Event{}, nil)

	db.On("EventsByCreationTime", mock.MatchedBy(func(start int64) bool {
		return start == 0xBADF00D
	}), mock.Anything, mock.Anything).Return(nil, fmt.Errorf("some error"))

	db.On("EventsForDevice", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == testEvent.ID.Hex()
	})).Return([]models.Event{testEvent}, nil)

	db.On("EventsForDevice", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "error"
	})).Return(nil, fmt.Errorf("some error"))

	db.On("DeleteEventById", mock.Anything).Return(nil)

	db.On("DeleteReadingById", mock.Anything).Return(nil)

	return db
}

func buildReadings() []models.Reading {
	ticks := db.MakeTimestamp()
	r1 := models.Reading{Id: bson.NewObjectId(),
		Name:     "Temperature",
		Value:    "45",
		Origin:   testOrigin,
		Created:  ticks,
		Modified: ticks,
		Pushed:   ticks,
		Device:   testDeviceName}

	r2 := models.Reading{Id: bson.NewObjectId(),
		Name:     "Pressure",
		Value:    "1.01325",
		Origin:   testOrigin,
		Created:  ticks,
		Modified: ticks,
		Pushed:   ticks,
		Device:   testDeviceName}
	readings := []models.Reading{}
	readings = append(readings, r1, r2)
	return readings
}

func testEventWithoutReadings(event models.Event, t *testing.T) {
	if event.ID.Hex() != testEvent.ID.Hex() {
		t.Error("eventId mismatch. expected " + testEvent.ID.Hex() + " received " + event.ID.Hex())
	}

	if event.Device != testEvent.Device {
		t.Error("device mismatch. expected " + testDeviceName + " received " + event.Device)
	}

	if event.Origin != testEvent.Origin {
		t.Error("origin mismatch. expected " + strconv.FormatInt(testEvent.Origin, 10) + " received " + strconv.FormatInt(event.Origin, 10))
	}
}

func handleDomainEvents(bitEvents []bool, wait *sync.WaitGroup, t *testing.T) {
	until := time.Now().Add(250 * time.Millisecond) //Kill this loop after quarter second.
	for time.Now().Before(until) {
		select {
		case evt := <-chEvents:
			switch evt.(type) {
			case DeviceLastReported:
				e := evt.(DeviceLastReported)
				if e.DeviceName != testDeviceName {
					t.Errorf("DeviceLastReported name mistmatch %s", e.DeviceName)
					return
				}
				bitEvents[0] = true
				break
			case DeviceServiceLastReported:
				e := evt.(DeviceServiceLastReported)
				if e.DeviceName != testDeviceName {
					t.Errorf("DeviceLastReported name mistmatch %s", e.DeviceName)
					return
				}
				bitEvents[1] = true
				break
			}
		default:
			//	Without a default case in here, the select block will hang.
		}
	}
	wait.Done()
}
