package data

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

var testEvent contract.Event

const (
	testDeviceName string = "Test Device"
	testOrigin     int64  = 123456789
	testBsonString string = "57e59a71e4b0ca8e6d6d4cc2"
	testUUIDString string = "ca93c8fa-9919-4ec5-85d3-f81b2b6a7bc1"
)

func TestCheckMaxLimit(t *testing.T) {
	reset()

	testedLimit := math.MinInt32

	expectedNil := checkMaxLimit(testedLimit, logger.NewMockClient())

	if expectedNil != nil {
		t.Errorf("Should not exceed limit")
	}
}

func TestCheckMaxLimitOverLimit(t *testing.T) {
	reset()

	testedLimit := math.MaxInt32

	expectedErr := checkMaxLimit(testedLimit, logger.NewMockClient())

	if expectedErr == nil {
		t.Errorf("Exceeded limit and should throw error")
	}
}

// Supporting methods
// Reset() re-initializes dependencies for each test
func reset() {
	Configuration = &ConfigurationStruct{}
	testEvent.ID = testBsonString
	testEvent.Device = testDeviceName
	testEvent.Origin = testOrigin
	testEvent.Readings = buildReadings()
}

func newMockDeviceClient() *mocks.DeviceClient {
	client := &mocks.DeviceClient{}

	protocols := getProtocols()

	mockDeviceResultFn := func(id string, ctx context.Context) contract.Device {
		if bson.IsObjectIdHex(id) {
			return contract.Device{Id: id, Name: testDeviceName, Protocols: protocols}
		}
		return contract.Device{}
	}
	client.On("Device", "valid", context.Background()).Return(mockDeviceResultFn, nil)
	client.On("Device", "404", context.Background()).Return(mockDeviceResultFn,
		types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("Device", mock.Anything, context.Background()).Return(mockDeviceResultFn, fmt.Errorf("some error"))

	mockDeviceForNameResultFn := func(name string, ctx context.Context) contract.Device {
		device := contract.Device{Id: uuid.New().String(), Name: name, Protocols: protocols}

		return device
	}
	client.On("DeviceForName", testDeviceName, context.Background()).Return(mockDeviceForNameResultFn, nil)
	client.On("DeviceForName", "404", context.Background()).Return(mockDeviceForNameResultFn,
		types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("DeviceForName", mock.Anything, context.Background()).Return(mockDeviceForNameResultFn,
		fmt.Errorf("some error"))

	return client
}

func getProtocols() map[string]contract.ProtocolProperties {
	p1 := make(map[string]string)
	p1["host"] = "localhost"
	p1["port"] = "1234"
	p1["unitID"] = "1"

	p2 := make(map[string]string)
	p2["serialPort"] = "/dev/USB0"
	p2["baudRate"] = "19200"
	p2["dataBits"] = "8"
	p2["stopBits"] = "1"
	p2["parity"] = "0"
	p2["unitID"] = "2"

	wrap := make(map[string]contract.ProtocolProperties)
	wrap["modbus-ip"] = p1
	wrap["modbus-rtu"] = p2

	return wrap
}

func buildReadings() []contract.Reading {
	ticks := db.MakeTimestamp()
	r1 := contract.Reading{Id: bson.NewObjectId().Hex(),
		Name:     "Temperature",
		Value:    "45",
		Origin:   testOrigin,
		Created:  ticks,
		Modified: ticks,
		Pushed:   ticks,
		Device:   testDeviceName}

	r2 := contract.Reading{Id: bson.NewObjectId().Hex(),
		Name:     "Pressure",
		Value:    "1.01325",
		Origin:   testOrigin,
		Created:  ticks,
		Modified: ticks,
		Pushed:   ticks,
		Device:   testDeviceName}
	readings := []contract.Reading{}
	readings = append(readings, r1, r2)
	return readings
}

func handleDomainEvents(bitEvents []bool, chEvents <-chan interface{}, wait *sync.WaitGroup, t *testing.T) {
	until := time.Now().Add(500 * time.Millisecond) // Kill this loop after half second.
	for time.Now().Before(until) {
		select {
		case evt := <-chEvents:
			switch evt.(type) {
			case DeviceLastReported:
				e := evt.(DeviceLastReported)
				if e.DeviceName != testDeviceName {
					t.Errorf("DeviceLastReported name mismatch %s", e.DeviceName)
					return
				}
				bitEvents[0] = true
				break
			case DeviceServiceLastReported:
				e := evt.(DeviceServiceLastReported)
				if e.DeviceName != testDeviceName {
					t.Errorf("DeviceLastReported name mismatch %s", e.DeviceName)
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
