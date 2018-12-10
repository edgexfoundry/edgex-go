package data

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/messaging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/memory"

	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata/mocks"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/globalsign/mgo/bson"
)

var testEvent models.Event
var testRoutes *mux.Router

const (
	testDeviceName string = "Test Device"
	testOrigin     int64  = 123456789
	testBsonString string = "57e59a71e4b0ca8e6d6d4cc2"
	testUUIDString string = "ca93c8fa-9919-4ec5-85d3-f81b2b6a7bc1"
)

// Mock implementation of the event publisher for testing purposes
type mockEventPublisher struct{}

func TestCheckMaxLimit(t *testing.T) {
	reset()

	testedLimit := math.MinInt32

	expectedNil := checkMaxLimit(testedLimit)

	if expectedNil != nil {
		t.Errorf("Should not exceed limit")
	}
}

func TestCheckMaxLimitOverLimit(t *testing.T) {
	reset()

	testedLimit := math.MaxInt32

	expectedErr := checkMaxLimit(testedLimit)

	if expectedErr == nil {
		t.Errorf("Exceeded limit and should throw error")
	}
}

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

// Supporting methods
// Reset() re-initializes dependencies for each test
func reset() {
	Configuration = &ConfigurationStruct{}
	testEvent.Device = testDeviceName
	testEvent.Origin = testOrigin
	testEvent.Readings = buildReadings()
	dbClient = &memory.MemDB{}
	testEvent.ID, _ = dbClient.AddEvent(testEvent)
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
		return id == "valid"
	})).Return(mockDeviceResultFn, nil)
	client.On("Device", mock.MatchedBy(func(id string) bool {
		return id == "404"
	})).Return(mockDeviceResultFn, types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("Device", mock.Anything).Return(mockDeviceResultFn, fmt.Errorf("some error"))

	mockDeviceForNameResultFn := func(name string) models.Device {
		device := models.Device{Id: bson.NewObjectId(), Name: name, Addressable: mockAddressable}

		return device
	}
	client.On("DeviceForName", mock.MatchedBy(func(name string) bool {
		return name == testEvent.Device
	})).Return(mockDeviceForNameResultFn, nil)
	client.On("DeviceForName", mock.MatchedBy(func(name string) bool {
		return name == "404"
	})).Return(mockDeviceForNameResultFn, types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("DeviceForName", mock.Anything).Return(mockDeviceForNameResultFn, fmt.Errorf("some error"))

	return client
}

func newMockDb() interfaces.DBClient {
	DB := &dbMock.DBClient{}

	DB.On("EventsOlderThanAge", mock.MatchedBy(func(age int64) bool {
		return age == -1
	})).Return(nil, fmt.Errorf("expected testing error"))

	DB.On("ValueDescriptorByName", mock.MatchedBy(func(name string) bool {
		return name == "Temperature"
	})).Return(models.ValueDescriptor{Type: "8"}, nil)

	DB.On("ValueDescriptorByName", mock.MatchedBy(func(name string) bool {
		return name == "Pressure"
	})).Return(models.ValueDescriptor{}, fmt.Errorf("some error"))

	DB.On("EventsForDeviceLimit", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "valid"
	}), mock.Anything).Return([]models.Event{}, nil)

	DB.On("EventsForDeviceLimit", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "error"
	}), mock.Anything).Return(nil, fmt.Errorf("some error"))

	DB.On("EventsForDevice", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "valid"
	})).Return([]models.Event{{Readings: append(buildReadings(), buildReadings()...)}}, nil)

	DB.On("EventsForDevice", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "error"
	})).Return(nil, fmt.Errorf("some error"))

	DB.On("EventsByCreationTime", mock.MatchedBy(func(start int64) bool {
		return start == 0xF00D
	}), mock.Anything, mock.Anything).Return([]models.Event{}, nil)

	DB.On("EventsByCreationTime", mock.MatchedBy(func(start int64) bool {
		return start == 0xBADF00D
	}), mock.Anything, mock.Anything).Return(nil, fmt.Errorf("some error"))

	DB.On("EventsForDevice", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == testEvent.ID
	})).Return([]models.Event{testEvent}, nil)

	DB.On("EventsForDevice", mock.MatchedBy(func(deviceId string) bool {
		return deviceId == "error"
	})).Return(nil, fmt.Errorf("some error"))

	DB.On("DeleteEventById", mock.Anything).Return(nil)

	DB.On("ValueDescriptorByName", mock.MatchedBy(func(name string) bool {
		return name == "valid"
	})).Return(models.ValueDescriptor{}, nil)

	DB.On("ValueDescriptorByName", mock.MatchedBy(func(name string) bool {
		return name == "404"
	})).Return(models.ValueDescriptor{}, db.ErrNotFound)

	DB.On("ValueDescriptorByName", mock.Anything).Return(models.ValueDescriptor{}, fmt.Errorf("some error"))

	DB.On("ValueDescriptorById", mock.MatchedBy(func(id string) bool {
		return id == "valid"
	})).Return(models.ValueDescriptor{}, nil)

	DB.On("ValueDescriptorById", mock.MatchedBy(func(id string) bool {
		return id == "404"
	})).Return(models.ValueDescriptor{}, db.ErrNotFound)

	DB.On("ValueDescriptorById", mock.Anything).Return(models.ValueDescriptor{}, fmt.Errorf("some error"))

	DB.On("ValueDescriptorsByUomLabel", mock.MatchedBy(func(name string) bool {
		return name == "valid"
	})).Return([]models.ValueDescriptor{}, nil)

	DB.On("ValueDescriptorsByUomLabel", mock.MatchedBy(func(name string) bool {
		return name == "404"
	})).Return([]models.ValueDescriptor{}, db.ErrNotFound)

	DB.On("ValueDescriptorsByUomLabel", mock.Anything).Return([]models.ValueDescriptor{}, fmt.Errorf("some error"))

	DB.On("ValueDescriptorsByLabel", mock.MatchedBy(func(name string) bool {
		return name == "valid"
	})).Return([]models.ValueDescriptor{}, nil)

	DB.On("ValueDescriptorsByLabel", mock.MatchedBy(func(name string) bool {
		return name == "404"
	})).Return([]models.ValueDescriptor{}, db.ErrNotFound)

	DB.On("ValueDescriptorsByLabel", mock.Anything).Return([]models.ValueDescriptor{}, fmt.Errorf("some error"))

	DB.On("ValueDescriptorsByType", mock.MatchedBy(func(name string) bool {
		return name == "valid"
	})).Return([]models.ValueDescriptor{}, nil)

	DB.On("ValueDescriptorsByType", mock.MatchedBy(func(name string) bool {
		return name == "404"
	})).Return([]models.ValueDescriptor{}, db.ErrNotFound)

	DB.On("ValueDescriptorsByType", mock.Anything).Return([]models.ValueDescriptor{}, fmt.Errorf("some error"))

	DB.On("ValueDescriptors").Return([]models.ValueDescriptor{}, fmt.Errorf("some error"))

	DB.On("AddValueDescriptor", mock.MatchedBy(func(vd models.ValueDescriptor) bool {
		return vd.Name == "valid"
	})).Return("", nil)

	DB.On("AddValueDescriptor", mock.MatchedBy(func(vd models.ValueDescriptor) bool {
		return vd.Name == "409"
	})).Return("", db.ErrNotUnique)

	DB.On("AddValueDescriptor", mock.Anything).Return("", fmt.Errorf("some error"))

	DB.On("DeleteValueDescriptorById", mock.MatchedBy(func(id string) bool {
		return id == testBsonString
	})).Return(nil)

	DB.On("DeleteValueDescriptorById", mock.Anything).Return(fmt.Errorf("some error"))

	DB.On("Readings").Return([]models.Reading{}, fmt.Errorf("some error"))

	DB.On("AddReading", mock.MatchedBy(func(reading models.Reading) bool {
		return reading.Name == "valid"
	})).Return("", nil)

	DB.On("AddReading", mock.Anything).Return("", fmt.Errorf("some error"))

	DB.On("ReadingById", mock.MatchedBy(func(id string) bool {
		return id == "valid"
	})).Return(models.Reading{}, nil)

	DB.On("ReadingById", mock.MatchedBy(func(id string) bool {
		return id == "404"
	})).Return(models.Reading{}, db.ErrNotFound)

	DB.On("ReadingById", mock.Anything).Return(models.Reading{}, fmt.Errorf("some error"))

	// these are reversed from usual because of a call in events.go
	DB.On("DeleteReadingById", mock.MatchedBy(func(id string) bool {
		return id == "invalid"
	})).Return(fmt.Errorf("some error"))

	DB.On("DeleteReadingById", mock.Anything).Return(nil)

	DB.On("ReadingCount").Return(0, fmt.Errorf("some error"))

	DB.On("ReadingsByDevice", mock.MatchedBy(func(name string) bool {
		return name == "valid"
	}), mock.Anything).Return([]models.Reading{}, nil)

	DB.On("ReadingsByDevice", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	// also used by TestDeleteValueDescriptor
	DB.On("ReadingsByValueDescriptor", mock.MatchedBy(func(name string) bool {
		return name == "valid"
	}), mock.Anything).Return([]models.Reading{}, nil)

	DB.On("ReadingsByValueDescriptor", mock.MatchedBy(func(name string) bool {
		return name == "409"
	}), mock.Anything).Return([]models.Reading{{}}, nil)

	DB.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	DB.On("ReadingsByValueDescriptorNames", mock.MatchedBy(func(names []string) bool {
		return names[0] == "valid"
	}), mock.Anything).Return([]models.Reading{}, nil)

	DB.On("ReadingsByValueDescriptorNames", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	DB.On("ReadingsByCreationTime", mock.MatchedBy(func(start int64) bool {
		return start == 0x0BEEF
	}), mock.Anything, mock.Anything).Return([]models.Reading{}, nil)

	DB.On("ReadingsByCreationTime", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	DB.On("ReadingsByDeviceAndValueDescriptor", mock.MatchedBy(func(device string) bool {
		return device == "valid"
	}), mock.Anything, mock.Anything).Return([]models.Reading{}, nil)

	DB.On("ReadingsByDeviceAndValueDescriptor", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	return DB
}

func buildReadings() []models.Reading {
	ticks := db.MakeTimestamp()
	r1 := models.Reading{Id: bson.NewObjectId().Hex(),
		Name:     "Temperature",
		Value:    "45",
		Origin:   testOrigin,
		Created:  ticks,
		Modified: ticks,
		Pushed:   ticks,
		Device:   testDeviceName}

	r2 := models.Reading{Id: bson.NewObjectId().Hex(),
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
