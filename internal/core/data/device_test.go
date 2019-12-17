package data

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/globalsign/mgo/bson"
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
