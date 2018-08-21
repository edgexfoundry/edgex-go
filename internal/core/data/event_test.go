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
	"os"
	"strconv"
	"testing"

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
	"sync"
	"time"
)

var testEvent models.Event
var testRoutes *mux.Router

const (
	TEST_DEVICE_NAME string = "Test Device"
	TEST_ORIGIN      int64  = 123456789
)

func TestMain(m *testing.M) {
	Configuration = &ConfigurationStruct{}
	testRoutes = LoadRestRoutes()
	LoggingClient = logger.NewMockClient()
	mdc = newMockDeviceClient()
	ep, _ = messaging.NewEventPublisher(messaging.MOCK, messaging.PubSubConfiguration{})
	chEvents = make(chan interface{}, 10)
	os.Exit(m.Run())
}

//Test methods
func TestCount(t *testing.T) {
	prepareDB()
	c, err := count()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if c != 1 {
		t.Errorf("expected event count 1, received: %s", strconv.Itoa(c))
		return
	}
}

func TestCountByDevice(t *testing.T) {
	prepareDB()
	count, err := countByDevice(testEvent.Device)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if count == 0 {
		t.Errorf("no events found")
		return
	}
}

func TestDeleteByAge(t *testing.T) {
	prepareDB()
	time.Sleep(time.Millisecond * time.Duration(5))
	age := db.MakeTimestamp()
	count, err := deleteByAge(age)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if count == 0 {
		t.Errorf("no events found")
		return
	}
}

func TestGetEvents(t *testing.T) {
	prepareDB()
	events, err := getEvents(0)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if len(events) == 0 {
		t.Errorf("no events found")
		return
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event")
		return
	}

	for e := range events {
		testEventWithoutReadings(events[e], t)
	}
}

func TestGetEventsWithLimit(t *testing.T) {
	prepareDB()
	//Put an extra dummy event in the DB
	evt := models.Event{Device: TEST_DEVICE_NAME, Origin: TEST_ORIGIN}
	dbClient.AddEvent(&evt)

	events, err := getEvents(1)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event")
		return
	}
}

func TestAddEventWithPersistence(t *testing.T) {
	prepareDB()
	Configuration.PersistData = true
	evt := models.Event{Device: TEST_DEVICE_NAME, Origin: TEST_ORIGIN, Readings: buildListOfMockReadings()}
	//wire up handlers to listen for device events
	bitEvents := make([]bool, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go handleDomainEvents(bitEvents, &wg, t)

	newId, err := addNew(evt)
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
			return
		}
	}
}

func TestAddEventNoPersistence(t *testing.T) {
	prepareDB()
	Configuration.PersistData = false
	evt := models.Event{Device: TEST_DEVICE_NAME, Origin: TEST_ORIGIN, Readings: buildListOfMockReadings()}
	//wire up handlers to listen for device events
	bitEvents := make([]bool, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go handleDomainEvents(bitEvents, &wg, t)

	newId, err := addNew(evt)
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
			return
		}
	}
}

func TestUpdateEventNotFound(t *testing.T) {
	prepareDB()
	evt := models.Event{ID: bson.NewObjectId(), Device: "Not Found", Origin: TEST_ORIGIN}
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
	prepareDB()
	evt := models.Event{ID: bson.NewObjectId(), Device: "Not Found", Origin: TEST_ORIGIN}
	err := updateEvent(evt)
	if err == nil {
		t.Errorf("error expected")
	}
	Configuration.MetaDataCheck = false
}

func TestUpdateEvent(t *testing.T) {
	prepareDB()
	evt := models.Event{ID: testEvent.ID, Device: "Some Value", Origin: TEST_ORIGIN}
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

func TestDeleteAll(t *testing.T) {
	prepareDB()
	err := deleteAll()
	if err != nil {
		t.Errorf(err.Error())
	}
}

/*
func TestGetEventHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/event", nil)
	w := httptest.NewRecorder()
	configuration.ReadMaxLimit = 1
	testRoutes.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Error("value expected, status code " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
		return
	}

	if len(w.Body.String()) == 0 {
		t.Error("response was empty " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
		return
	}

	events := []models.Event{}
	json.Unmarshal([]byte(w.Body.String()), &events)
	for e := range events {
		testEventWithoutReadings(events[e], t)
	}
	configuration.ReadMaxLimit = 0
}

func TestGetEventHandlerMaxExceeded(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/event", nil)
	w := httptest.NewRecorder()
	configuration.ReadMaxLimit = 0
	testRoutes.ServeHTTP(w, req)

	if w.Code != 413 {
		t.Error("413 exceeded, status code " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
		return
	}
}


func TestGetEventByIdHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/event/"+testEvent.ID.Hex(), nil)
	w := httptest.NewRecorder()

	testRoutes.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Error("value expected, status code " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
		return
	}

	if len(w.Body.String()) == 0 {
		t.Error("response was empty " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
		return
	}

	event := models.Event{}
	json.Unmarshal([]byte(w.Body.String()), &event)
	testEventWithoutReadings(event, t)
}
*/
// Supporting methods
func prepareDB() {
	testEvent.Device = TEST_DEVICE_NAME
	testEvent.Origin = TEST_ORIGIN
	testEvent.Readings = buildListOfMockReadings()
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

func buildListOfMockReadings() []models.Reading {
	ticks := db.MakeTimestamp()
	r1 := models.Reading{Id: bson.NewObjectId(),
		Name:     "Temperature",
		Value:    "45",
		Origin:   TEST_ORIGIN,
		Created:  ticks,
		Modified: ticks,
		Pushed:   ticks,
		Device:   TEST_DEVICE_NAME}

	r2 := models.Reading{Id: bson.NewObjectId(),
		Name:     "Pressure",
		Value:    "1.01325",
		Origin:   TEST_ORIGIN,
		Created:  ticks,
		Modified: ticks,
		Pushed:   ticks,
		Device:   TEST_DEVICE_NAME}
	readings := []models.Reading{}
	readings = append(readings, r1, r2)
	return readings
}

func testEventWithoutReadings(event models.Event, t *testing.T) {
	if event.ID.Hex() != testEvent.ID.Hex() {
		t.Error("eventId mismatch. expected " + testEvent.ID.Hex() + " received " + event.ID.Hex())
	}

	if event.Device != testEvent.Device {
		t.Error("device mismatch. expected " + TEST_DEVICE_NAME + " received " + event.Device)
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
				if e.DeviceName != TEST_DEVICE_NAME {
					t.Errorf("DeviceLastReported name mistmatch %s", e.DeviceName)
					return
				}
				setEventBit(0, true, bitEvents)
				break
			case DeviceServiceLastReported:
				e := evt.(DeviceServiceLastReported)
				if e.DeviceName != TEST_DEVICE_NAME {
					t.Errorf("DeviceLastReported name mistmatch %s", e.DeviceName)
					return
				}
				setEventBit(1, true, bitEvents)
				break
			}
		default:
			//	Without a default case in here, the select block will hang.
		}
	}
	wait.Done()
}

func setEventBit(index int, value bool, source []bool) {
	for i, oldVal := range source {
		if i == index {
			source[i] = value
		} else {
			source[i] = oldVal
		}
	}
}
