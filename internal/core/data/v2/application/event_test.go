package application

import (
	"context"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/globalsign/mgo/bson"

	"github.com/google/uuid"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testEvent models.Event

const (
	testDeviceName string = "Test Device"
	testOrigin     int64  = 123456789
	testUUIDString string = "ca93c8fa-9919-4ec5-85d3-f81b2b6a7bc1"
)

func newAddEventMockDB(persist bool) *dbMock.DBClient {
	myMock := &dbMock.DBClient{}

	if persist {
		mockId := "3c5badcb-2008-47f2-ba78-eb2d992f8422"
		myMock.On("AddEvent", mock.Anything).Return(models.Event{
			Id:         mockId,
			DeviceName: testDeviceName,
			Origin:     testOrigin,
			Readings:   buildReadings(),
		}, nil)
	}

	return myMock
}

func TestAddEventWithPersistence(t *testing.T) {
	reset()

	evt := models.Event{DeviceName: testDeviceName, Origin: testOrigin, Readings: buildReadings()}

	dbClientMock := newAddEventMockDB(true)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	_, err := AddEvent(evt, context.Background(), dic)

	require.NoError(t, err)

	dbClientMock.AssertExpectations(t)
}

func TestAddEventNoPersistence(t *testing.T) {
	reset()

	evt := models.Event{DeviceName: testDeviceName, Origin: testOrigin, Readings: buildReadings()}

	dbClientMock := newAddEventMockDB(false)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		dataContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					PersistData: false,
				},
			}
		},
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	newId, err := AddEvent(evt, context.Background(), dic)

	require.NoError(t, err)

	require.False(t, bson.IsObjectIdHex(newId), "unexpected bson id %s received", newId)

	dbClientMock.AssertExpectations(t)
}

// Supporting methods
// Reset() re-initializes dependencies for each test
func reset() {
	testEvent.Id = testUUIDString
	testEvent.DeviceName = testDeviceName
	testEvent.Origin = testOrigin
	testEvent.Readings = buildReadings()
}

func buildReadings() []models.Reading {
	ticks := db.MakeTimestamp()

	r1 := models.SimpleReading{
		BaseReading: models.BaseReading{
			Id:         uuid.New().String(),
			Created:    ticks,
			Origin:     testOrigin,
			DeviceName: testDeviceName,
			Name:       "Temperature",
			Labels:     []string{"Fahrenheit"},
			ValueType:  dtos.ValueTypeFloat32,
		},
		Value:         "45",
		FloatEncoding: "Base64",
	}

	r2 := models.BinaryReading{
		BaseReading: models.BaseReading{
			Id:         uuid.New().String(),
			Created:    ticks,
			Origin:     testOrigin,
			DeviceName: testDeviceName,
			Name:       "FileData",
			Labels:     []string{"text"},
		},
		BinaryValue: []byte("1010"),
		MediaType:   "file",
	}

	var readings []models.Reading
	readings = append(readings, r1, r2)
	return readings
}
