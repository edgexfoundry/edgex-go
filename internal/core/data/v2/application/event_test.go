package application

import (
	"context"
	"net/http"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testDeviceName     string = "Test Device"
	testUUIDString     string = "ca93c8fa-9919-4ec5-85d3-f81b2b6a7bc1"
	testCreatedTime           = 1600666214495
	testOriginTime            = 1600666185705354000
	nonexistentEventID        = "8ad33474-fbc5-11ea-adc1-0242ac120002"
	testEventCount            = uint32(7778)
)

var persistedEvent = models.Event{
	Id:         testUUIDString,
	DeviceName: testDeviceName,
	Created:    testCreatedTime,
	Origin:     testOriginTime,
	Readings:   buildReadings(),
}

func buildReadings() []models.Reading {
	ticks := utils.MakeTimestamp()

	r1 := models.SimpleReading{
		BaseReading: models.BaseReading{
			Id:           uuid.New().String(),
			Created:      ticks,
			Origin:       testOriginTime,
			DeviceName:   testDeviceName,
			ResourceName: "Temperature",
			ProfileName:  "TempProfile",
			ValueType:    dtos.ValueTypeUint16,
		},
		Value: "45",
	}

	r2 := models.BinaryReading{
		BaseReading: models.BaseReading{
			Id:           uuid.New().String(),
			Created:      ticks,
			Origin:       testOriginTime,
			DeviceName:   testDeviceName,
			ResourceName: "FileData",
			ProfileName:  "FileDataProfile",
		},
		BinaryValue: []byte("1010"),
		MediaType:   "file",
	}

	var readings []models.Reading
	readings = append(readings, r1, r2)
	return readings
}

func newMockDB(persist bool) *dbMock.DBClient {
	myMock := &dbMock.DBClient{}

	if persist {
		myMock.On("AddEvent", mock.Anything).Return(persistedEvent, nil)
		myMock.On("EventById", nonexistentEventID).Return(models.Event{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "event doesn't exist in the database", nil))
		myMock.On("EventById", testUUIDString).Return(persistedEvent, nil)
		myMock.On("DeleteEventById", nonexistentEventID).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "event doesn't exist in the database", nil))
		myMock.On("DeleteEventById", testUUIDString).Return(nil)
		myMock.On("EventTotalCount").Return(testEventCount, nil)
		myMock.On("EventCountByDevice", testDeviceName).Return(testEventCount, nil)
		myMock.On("DeleteEventsByDeviceName", testDeviceName).Return(nil)
		myMock.On("DeleteEventsByAge", int64(0)).Return(nil)
	}

	return myMock
}

func TestAddEvent(t *testing.T) {
	evt := models.Event{
		Id:         testUUIDString,
		DeviceName: testDeviceName,
		Origin:     testOriginTime,
		Readings:   buildReadings(),
	}

	tests := []struct {
		Name        string
		Persistence bool
	}{
		{Name: "Add Event with persistence", Persistence: true},
		{Name: "Add Event without persistence", Persistence: false},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			dbClientMock := newMockDB(testCase.Persistence)

			dic := mocks.NewMockDIC()
			dic.Update(di.ServiceConstructorMap{
				dataContainer.ConfigurationName: func(get di.Get) interface{} {
					return &config.ConfigurationStruct{
						Writable: config.WritableInfo{
							PersistData: testCase.Persistence,
						},
					}
				},
				v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
					return dbClientMock
				},
			})
			_, err := AddEvent(evt, context.Background(), dic)

			require.NoError(t, err)

			if !testCase.Persistence {
				// assert there is no db client function called
				dbClientMock.AssertExpectations(t)
			}
		})
	}
}

func TestEventById(t *testing.T) {
	validEventId := testUUIDString
	emptyEventId := ""
	invalidEventId := "bad"
	notFoundEventId := nonexistentEventID

	dbClientMock := newMockDB(true)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		Name               string
		EventId            string
		ErrorExpected      bool
		ExpectedErrKind    errors.ErrKind
		ExpectedStatusCode int
	}{
		{"Valid - Find Event by Id", validEventId, false, errors.KindUnknown, http.StatusOK},
		{"Invalid - Empty EventId", emptyEventId, true, errors.KindInvalidId, http.StatusBadRequest},
		{"Invalid - EventId is not an UUID", invalidEventId, true, errors.KindInvalidId, http.StatusBadRequest},
		{"Invalid - Event doesn't exist", notFoundEventId, true, errors.KindEntityDoesNotExist, http.StatusNotFound},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			evt, err := EventById(testCase.EventId, dic)

			if testCase.ErrorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.ExpectedStatusCode, err.Code(), "Error code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.EventId, evt.Id, "Event Id not as expected")
			}
		})
	}
}

func TestDeleteEventById(t *testing.T) {
	validEventId := testUUIDString
	emptyEventId := ""
	invalidEventId := "bad"
	notFoundEventId := nonexistentEventID

	dbClientMock := newMockDB(true)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		Name               string
		EventId            string
		ErrorExpected      bool
		ExpectedErrKind    errors.ErrKind
		ExpectedStatusCode int
	}{
		{"Valid - Delete Event by Id", validEventId, false, errors.KindUnknown, http.StatusOK},
		{"Invalid - Empty EventId", emptyEventId, true, errors.KindInvalidId, http.StatusBadRequest},
		{"Invalid - EventId is not an UUID", invalidEventId, true, errors.KindInvalidId, http.StatusBadRequest},
		{"Invalid - Event doesn't exist", notFoundEventId, true, errors.KindEntityDoesNotExist, http.StatusNotFound},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			err := DeleteEventById(testCase.EventId, dic)

			if testCase.ErrorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.ExpectedStatusCode, err.Code(), "Error code not as expected")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEventTotalCount(t *testing.T) {
	dbClientMock := newMockDB(true)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	count, err := EventTotalCount(dic)
	require.NoError(t, err)
	assert.Equal(t, testEventCount, count, "Event total count is not expected")
}

func TestEventCountByDevice(t *testing.T) {
	dbClientMock := newMockDB(true)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	count, err := EventCountByDevice(testDeviceName, dic)
	require.NoError(t, err)
	assert.Equal(t, testEventCount, count, "Event total count is not expected")
}

func TestDeleteEventsByDeviceName(t *testing.T) {
	dbClientMock := newMockDB(true)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		Name               string
		deviceName         string
		ErrorExpected      bool
		ExpectedErrKind    errors.ErrKind
		ExpectedStatusCode int
	}{
		{"Valid - Delete Event by Id", testDeviceName, false, errors.KindInvalidId, http.StatusOK},
		{"Invalid - Empty device name", "", true, errors.KindInvalidId, http.StatusBadRequest},
		{"Invalid - Empty device name with spaces", " \n\t\r ", true, errors.KindInvalidId, http.StatusBadRequest},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			err := DeleteEventsByDeviceName(testCase.deviceName, dic)

			if testCase.ErrorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.ExpectedStatusCode, err.Code(), "Error code not as expected")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEventsByTimeRange(t *testing.T) {
	event1 := persistedEvent
	event2 := persistedEvent
	event2.Created = event2.Created + 20
	event3 := persistedEvent
	event3.Created = event3.Created + 30
	event4 := persistedEvent
	event4.Created = event4.Created + 40
	event5 := persistedEvent
	event5.Created = event5.Created + 50

	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventsByTimeRange", int(event1.Created), int(event5.Created), 0, 10).Return([]models.Event{event5, event4, event3, event2, event1}, nil)
	dbClientMock.On("EventsByTimeRange", int(event2.Created), int(event4.Created), 0, 10).Return([]models.Event{event4, event3, event2}, nil)
	dbClientMock.On("EventsByTimeRange", int(event2.Created), int(event4.Created), 1, 2).Return([]models.Event{event3, event2}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		start              int
		end                int
		offset             int
		limit              int
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - all events", int(event1.Created), int(event5.Created), 0, 10, false, 5, http.StatusOK},
		{"Valid - events trimmed by latest and oldest", int(event2.Created), int(event4.Created), 0, 10, false, 3, http.StatusOK},
		{"Valid - events trimmed by latest and oldest and skipped first", int(event2.Created), int(event4.Created), 1, 2, false, 2, http.StatusOK},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			events, err := EventsByTimeRange(testCase.start, testCase.end, testCase.offset, testCase.limit, dic)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedCount, len(events), "Event total count is not expected")
		})
	}
}

func TestDeleteEventsByAge(t *testing.T) {
	dbClientMock := newMockDB(true)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	err := DeleteEventsByAge(0, dic)
	require.NoError(t, err)
}
