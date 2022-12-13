//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

const (
	testDeviceResourceName = "TestDeviceResource"
	testDeviceName         = "TestDevice"
	testProfileName        = "TestProfile"
	testSourceName         = "testSourceName"
	testUUIDString         = "ca93c8fa-9919-4ec5-85d3-f81b2b6a7bc1"
	testOriginTime         = 1600666185705354000
	nonexistentEventID     = "8ad33474-fbc5-11ea-adc1-0242ac120002"
	testEventCount         = uint32(7778)
)

var persistedEvent = models.Event{
	Id:         testUUIDString,
	DeviceName: testDeviceName,
	SourceName: testSourceName,
	Origin:     testOriginTime,
	Readings:   buildReadings(),
}

func buildReadings() []models.Reading {
	ticks := utils.MakeTimestamp()

	r1 := models.SimpleReading{
		BaseReading: models.BaseReading{
			Id:           uuid.New().String(),
			Origin:       ticks,
			DeviceName:   testDeviceName,
			ResourceName: testDeviceResourceName,
			ProfileName:  "TempProfile",
			ValueType:    common.ValueTypeUint16,
		},
		Value: "45",
	}

	r2 := models.BinaryReading{
		BaseReading: models.BaseReading{
			Id:           uuid.New().String(),
			Origin:       ticks + 20,
			DeviceName:   testDeviceName,
			ResourceName: testDeviceResourceName,
			ProfileName:  "FileDataProfile",
		},
		BinaryValue: []byte("1010"),
		MediaType:   "file",
	}

	r3 := models.SimpleReading{
		BaseReading: models.BaseReading{
			Id:           uuid.New().String(),
			Origin:       ticks + 30,
			DeviceName:   testDeviceName,
			ResourceName: testDeviceResourceName,
			ProfileName:  "TempProfile",
			ValueType:    common.ValueTypeUint16,
		},
		Value: "33",
	}

	r4 := models.SimpleReading{
		BaseReading: models.BaseReading{
			Id:           uuid.New().String(),
			Origin:       ticks + 40,
			DeviceName:   testDeviceName,
			ResourceName: testDeviceResourceName,
			ProfileName:  "TempProfile",
			ValueType:    common.ValueTypeUint16,
		},
		Value: "44",
	}

	r5 := models.SimpleReading{
		BaseReading: models.BaseReading{
			Id:           uuid.New().String(),
			Origin:       ticks + 50,
			DeviceName:   testDeviceName,
			ResourceName: testDeviceResourceName,
			ProfileName:  "TempProfile",
			ValueType:    common.ValueTypeUint16,
		},
		Value: "55",
	}

	var readings []models.Reading
	readings = append(readings, r1, r2, r3, r4, r5)
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
		myMock.On("EventCountByDeviceName", testDeviceName).Return(testEventCount, nil)
		myMock.On("DeleteEventsByDeviceName", testDeviceName).Return(nil)
		myMock.On("DeleteEventsByAge", int64(0)).Return(nil)
	}

	return myMock
}

func TestValidateEvent(t *testing.T) {
	evt := models.Event{
		Id:          testUUIDString,
		DeviceName:  testDeviceName,
		ProfileName: testProfileName,
		SourceName:  testSourceName,
		Origin:      testOriginTime,
		Readings:    buildReadings(),
	}

	tests := []struct {
		Name          string
		event         models.Event
		profileName   string
		deviceName    string
		sourceName    string
		errorExpected bool
	}{
		{"Valid - profileName/deviceName matches", persistedEvent, testProfileName, testDeviceName, testSourceName, false},
		{"Invalid - empty profile name", persistedEvent, "", testDeviceName, testSourceName, true},
		{"Invalid - inconsistent profile name", persistedEvent, "inconsistent", testDeviceName, testSourceName, true},
		{"Invalid - empty device name", persistedEvent, testProfileName, "", testSourceName, true},
		{"Invalid - inconsistent profile name", persistedEvent, testProfileName, "inconsistent", testSourceName, true},
		{"Invalid - empty source name", persistedEvent, "", testDeviceName, "", true},
		{"Invalid - inconsistent source name", persistedEvent, testProfileName, testDeviceName, "inconsistent", true},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			dbClientMock := newMockDB(true)

			dic := mocks.NewMockDIC()
			dic.Update(di.ServiceConstructorMap{
				container.DBClientInterfaceName: func(get di.Get) interface{} {
					return dbClientMock
				},
			})

			app := NewCoreDataApp(dic)
			err := app.ValidateEvent(evt, testCase.profileName, testCase.deviceName, testCase.sourceName, context.Background(), dic)

			if testCase.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddEvent(t *testing.T) {
	evt := models.Event{
		Id:          testUUIDString,
		DeviceName:  testDeviceName,
		ProfileName: testProfileName,
		Origin:      testOriginTime,
		Readings:    buildReadings(),
	}

	tests := []struct {
		Name          string
		Persistence   bool
		errorExpected bool
	}{
		{"Valid - Add Event with persistence", true, false},
		{"Valid - Add Event without persistence", false, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			dbClientMock := newMockDB(testCase.Persistence)

			dic := mocks.NewMockDIC()
			dic.Update(di.ServiceConstructorMap{
				container.ConfigurationName: func(get di.Get) interface{} {
					return &config.ConfigurationStruct{
						Writable: config.WritableInfo{
							PersistData: testCase.Persistence,
						},
					}
				},
				container.DBClientInterfaceName: func(get di.Get) interface{} {
					return dbClientMock
				},
			})

			// TODO: Add Metric dependencies to DIC??
			app := NewCoreDataApp(dic)
			err := app.AddEvent(evt, context.Background(), dic)

			if testCase.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			app := NewCoreDataApp(dic)
			evt, err := app.EventById(testCase.EventId, dic)

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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			app := NewCoreDataApp(dic)
			err := app.DeleteEventById(testCase.EventId, dic)

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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	app := NewCoreDataApp(dic)
	count, err := app.EventTotalCount(dic)
	require.NoError(t, err)
	assert.Equal(t, testEventCount, count, "Event total count is not expected")
}

func TestEventCountByDeviceName(t *testing.T) {
	dbClientMock := newMockDB(true)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	app := NewCoreDataApp(dic)
	count, err := app.EventCountByDeviceName(testDeviceName, dic)
	require.NoError(t, err)
	assert.Equal(t, testEventCount, count, "Event total count is not expected")
}

func TestDeleteEventsByDeviceName(t *testing.T) {
	dbClientMock := newMockDB(true)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			app := NewCoreDataApp(dic)
			err := app.DeleteEventsByDeviceName(testCase.deviceName, dic)

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
	event2.Origin = event2.Origin + 20
	event3 := persistedEvent
	event3.Origin = event3.Origin + 30
	event4 := persistedEvent
	event4.Origin = event4.Origin + 40
	event5 := persistedEvent
	event5.Origin = event5.Origin + 50

	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("EventCountByTimeRange", int(event1.Origin), int(event5.Origin)).Return(uint32(5), nil)
	dbClientMock.On("EventsByTimeRange", int(event1.Origin), int(event5.Origin), 0, 10).Return([]models.Event{event5, event4, event3, event2, event1}, nil)
	dbClientMock.On("EventCountByTimeRange", int(event2.Origin), int(event4.Origin)).Return(uint32(3), nil)
	dbClientMock.On("EventsByTimeRange", int(event2.Origin), int(event4.Origin), 0, 10).Return([]models.Event{event4, event3, event2}, nil)
	dbClientMock.On("EventsByTimeRange", int(event2.Origin), int(event4.Origin), 1, 2).Return([]models.Event{event3, event2}, nil)
	dbClientMock.On("EventsByTimeRange", int(event2.Origin), int(event4.Origin), 4, 2).Return(nil, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		ExpectedErrKind    errors.ErrKind
		expectedCount      int
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - all events", int(event1.Origin), int(event5.Origin), 0, 10, false, "", 5, uint32(5), http.StatusOK},
		{"Valid - events trimmed by latest and oldest", int(event2.Origin), int(event4.Origin), 0, 10, false, "", 3, uint32(3), http.StatusOK},
		{"Valid - events trimmed by latest and oldest and skipped first", int(event2.Origin), int(event4.Origin), 1, 2, false, "", 2, uint32(3), http.StatusOK},
		{"Invalid - bounds out of range", int(event2.Origin), int(event4.Origin), 4, 2, true, errors.KindRangeNotSatisfiable, 0, uint32(0), http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			app := NewCoreDataApp(dic)
			events, totalCount, err := app.EventsByTimeRange(testCase.start, testCase.end, testCase.offset, testCase.limit, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(events), "Event count is not expected")
				assert.Equal(t, testCase.expectedTotalCount, totalCount, "Total count is not expected")
			}
		})
	}
}

func TestDeleteEventsByAge(t *testing.T) {
	dbClientMock := newMockDB(true)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	app := NewCoreDataApp(dic)
	err := app.DeleteEventsByAge(0, dic)
	require.NoError(t, err)
}
