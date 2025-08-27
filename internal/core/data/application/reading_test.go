//
// Copyright (C) 2023 IOTech Ltd
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/query"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

func TestAllReadings(t *testing.T) {
	readings := buildReadings()
	totalCount := uint32(len(readings))

	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AllReadings", 0, 20).Return(readings, nil)
	dbClientMock.On("ReadingTotalCount").Return(totalCount, nil)
	dbClientMock.On("AllReadings", 3, 10).Return([]models.Reading{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		offset             int
		limit              int
		errorExpected      bool
		ExpectedErrKind    errors.ErrKind
		expectedCount      int
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - all readings", 0, 20, false, "", len(readings), totalCount, http.StatusOK},
		{"Invalid - bounds out of range", 3, 10, true, errors.KindRangeNotSatisfiable, 0, 0, http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			readings, total, err := AllReadings(query.Parameters{Offset: testCase.offset, Limit: testCase.limit}, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(readings), "Reading count is not expected")
				assert.Equal(t, testCase.expectedTotalCount, total, "Total count is not expected")
			}
		})
	}
}

func TestReadingsByTimeRange(t *testing.T) {
	readings := buildReadings()
	totalCount5 := uint32(5)
	totalCount3 := uint32(3)

	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByTimeRange", readings[0].GetBaseReading().Origin, readings[4].GetBaseReading().Origin).Return(totalCount5, nil)
	dbClientMock.On("ReadingsByTimeRange", readings[0].GetBaseReading().Origin, readings[4].GetBaseReading().Origin, 0, 10).Return(readings, nil)
	dbClientMock.On("ReadingCountByTimeRange", readings[1].GetBaseReading().Origin, readings[3].GetBaseReading().Origin).Return(totalCount3, nil)
	dbClientMock.On("ReadingsByTimeRange", readings[1].GetBaseReading().Origin, readings[3].GetBaseReading().Origin, 0, 10).Return([]models.Reading{readings[3], readings[2], readings[1]}, nil)
	dbClientMock.On("ReadingsByTimeRange", readings[1].GetBaseReading().Origin, readings[3].GetBaseReading().Origin, 1, 2).Return([]models.Reading{readings[2], readings[1]}, nil)
	dbClientMock.On("ReadingsByTimeRange", readings[1].GetBaseReading().Origin, readings[3].GetBaseReading().Origin, 4, 2).Return(nil, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		start              int64
		end                int64
		offset             int
		limit              int
		errorExpected      bool
		ExpectedErrKind    errors.ErrKind
		expectedCount      int
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - all readings", readings[0].GetBaseReading().Origin, readings[4].GetBaseReading().Origin, 0, 10, false, "", 5, totalCount5, http.StatusOK},
		{"Valid - readings trimmed by latest and oldest", readings[1].GetBaseReading().Origin, readings[3].GetBaseReading().Origin, 0, 10, false, "", 3, totalCount3, http.StatusOK},
		{"Valid - readings trimmed by latest and oldest and skipped first", readings[1].GetBaseReading().Origin, readings[3].GetBaseReading().Origin, 1, 2, false, "", 2, totalCount3, http.StatusOK},
		{"Invalid - bounds out of range", readings[1].GetBaseReading().Origin, readings[3].GetBaseReading().Origin, 4, 2, true, errors.KindRangeNotSatisfiable, 0, uint32(0), http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			readings, totalCount, err := ReadingsByTimeRange(query.Parameters{Start: testCase.start, End: testCase.end, Offset: testCase.offset, Limit: testCase.limit}, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(readings), "Reading count is not expected")
				assert.Equal(t, testCase.expectedTotalCount, totalCount, "Total count is not expected")
			}
		})
	}
}

func TestReadingsByResourceName(t *testing.T) {
	readings := buildReadings()
	totalCount := uint32(len(readings))

	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByResourceName", testDeviceResourceName).Return(totalCount, nil)
	dbClientMock.On("ReadingsByResourceName", 0, 20, testDeviceResourceName).Return(readings, nil)
	dbClientMock.On("ReadingsByResourceName", len(readings)+1, 10, testDeviceResourceName).Return([]models.Reading{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		offset             int
		limit              int
		resourceName       string
		errorExpected      bool
		ExpectedErrKind    errors.ErrKind
		expectedCount      int
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - all readings", 0, 20, testDeviceResourceName, false, "", len(readings), totalCount, http.StatusOK},
		{"Invalid - bounds out of range", len(readings) + 1, 10, testDeviceResourceName, true, errors.KindRangeNotSatisfiable, 0, 0, http.StatusRequestedRangeNotSatisfiable},
		{"Invalid - empty resource name", len(readings) + 1, 10, "", true, errors.KindContractInvalid, 0, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			readings, total, err := ReadingsByResourceName(query.Parameters{Offset: testCase.offset, Limit: testCase.limit}, testCase.resourceName, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(readings), "Reading count is not expected")
				assert.Equal(t, testCase.expectedTotalCount, total, "Total count is not expected")
			}
		})
	}
}

func TestReadingsByDeviceName(t *testing.T) {
	readings := buildReadings()
	totalCount := uint32(len(readings))

	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByDeviceName", testDeviceName).Return(totalCount, nil)
	dbClientMock.On("ReadingsByDeviceName", 0, 20, testDeviceName).Return(readings, nil)
	dbClientMock.On("ReadingsByDeviceName", 3, 10, testDeviceName).Return([]models.Reading{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		offset             int
		limit              int
		deviceName         string
		errorExpected      bool
		ExpectedErrKind    errors.ErrKind
		expectedCount      int
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - all readings", 0, 20, testDeviceName, false, "", len(readings), totalCount, http.StatusOK},
		{"Invalid - bounds out of range", 3, 10, testDeviceName, true, errors.KindRangeNotSatisfiable, 0, 0, http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			readings, total, err := ReadingsByDeviceName(query.Parameters{Offset: testCase.offset, Limit: testCase.limit}, testCase.deviceName, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(readings), "Reading count is not expected")
				assert.Equal(t, testCase.expectedTotalCount, total, "Total count is not expected")
			}
		})
	}
}

func TestReadingCountByDeviceName(t *testing.T) {
	expectedReadingCount := uint32(656672)
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByDeviceName", testDeviceName).Return(expectedReadingCount, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		deviceName         string
		errorExpected      bool
		ExpectedErrKind    errors.ErrKind
		expectedCount      uint32
		expectedStatusCode int
	}{
		{"Valid - all readings", testDeviceName, false, "", expectedReadingCount, http.StatusOK},
		{"Invalid - empty device name", "", true, errors.KindContractInvalid, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			count, err := ReadingCountByDeviceName(testCase.deviceName, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, expectedReadingCount, count, "Reading total count is not expected")
			}
		})
	}
}

func TestPurgeEventsByAutoEvent(t *testing.T) {
	dic := mocks.NewMockDIC()
	deviceName := "testDevice"
	sourceName := "testSource"
	coreDataConfig := container.ConfigurationFrom(dic.Get)
	coreDataConfig.Retention = config.EventRetention{
		Interval:        "10m",
		DefaultMaxCap:   -1,
		DefaultMinCap:   1,
		DefaultDuration: "30m",
	}
	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return coreDataConfig
		},
	})

	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("LatestEventByDeviceNameAndSourceNameAndAgeAndOffset", deviceName, sourceName, mock.Anything, mock.Anything).Return(models.Event{}, nil)
	dbClientMock.On("DeleteEventsByAgeAndDeviceNameAndSourceName", mock.Anything, deviceName, sourceName).Return(nil)
	dbClientMock.On("DeleteEventsByDeviceNameAndSourceName", deviceName, sourceName).Return(nil)
	dbClientMock.On("LatestEventByDeviceNameAndSourceNameAndOffset", deviceName, sourceName, mock.Anything).Return(models.Event{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name      string
		autoEvent models.AutoEvent
	}{
		{"time-based event retention",
			models.AutoEvent{
				Interval:          "",
				OnChange:          false,
				OnChangeThreshold: 0,
				SourceName:        sourceName,
				Retention: models.Retention{
					MaxCap:   -1,
					MinCap:   -1,
					Duration: "10m",
				},
			},
		},
		{"time-based event retention with miniCap",
			models.AutoEvent{
				Interval:          "",
				OnChange:          false,
				OnChangeThreshold: 0,
				SourceName:        sourceName,
				Retention: models.Retention{
					MaxCap:   -1,
					MinCap:   1,
					Duration: "10m",
				},
			},
		},
		{"count-based event retention",
			models.AutoEvent{
				Interval:          "",
				OnChange:          false,
				OnChangeThreshold: 0,
				SourceName:        sourceName,
				Retention: models.Retention{
					MaxCap:   -1,
					MinCap:   -1,
					Duration: "0s",
				},
			},
		},
		{"count-based event retention with miniCap",
			models.AutoEvent{
				Interval:          "",
				OnChange:          false,
				OnChangeThreshold: 0,
				SourceName:        sourceName,
				Retention: models.Retention{
					MaxCap:   -1,
					MinCap:   1,
					Duration: "0s",
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := purgeEventsByAutoEvent(deviceName, testCase.autoEvent, dic)
			require.NoError(t, err)
			duration, parseErr := time.ParseDuration(testCase.autoEvent.Retention.Duration)
			require.NoError(t, parseErr)
			if duration > 0 && testCase.autoEvent.Retention.MinCap <= 0 {
				// time-based retention
				dbClientMock.AssertCalled(t, "DeleteEventsByAgeAndDeviceNameAndSourceName", mock.Anything, deviceName, testCase.autoEvent.SourceName)
			} else if duration > 0 && testCase.autoEvent.Retention.MinCap > 0 {
				// time-based retention with miniCap
				dbClientMock.AssertCalled(t, "LatestEventByDeviceNameAndSourceNameAndAgeAndOffset", deviceName, sourceName, duration.Nanoseconds(), uint32(testCase.autoEvent.Retention.MinCap))
				dbClientMock.AssertCalled(t, "DeleteEventsByAgeAndDeviceNameAndSourceName", mock.Anything, deviceName, testCase.autoEvent.SourceName)
			} else if testCase.autoEvent.Retention.MinCap <= 0 {
				// count-based retention
				dbClientMock.AssertCalled(t, "DeleteEventsByDeviceNameAndSourceName", deviceName, testCase.autoEvent.SourceName)
			} else {
				// count-based retention with miniCap
				dbClientMock.AssertCalled(t, "LatestEventByDeviceNameAndSourceNameAndOffset", deviceName, sourceName, uint32(testCase.autoEvent.Retention.MinCap))
				dbClientMock.AssertCalled(t, "DeleteEventsByAgeAndDeviceNameAndSourceName", mock.Anything, deviceName, testCase.autoEvent.SourceName)
			}

		})
	}
}

func TestNoneAutoEventSourcesByDevice(t *testing.T) {
	deviceName := "test-device"
	profileName := "test-profile"
	dic := mocks.NewMockDIC()
	clientMock := &clientMocks.DeviceProfileClient{}
	clientMock.On("DeviceProfileByName", context.Background(), profileName).Return(responses.DeviceProfileResponse{Profile: dtos.DeviceProfile{
		DeviceResources: []dtos.DeviceResource{{Name: "resource1", Properties: dtos.ResourceProperties{ReadWrite: common.ReadWrite_R}}, {Name: "resource2", Properties: dtos.ResourceProperties{ReadWrite: common.ReadWrite_W}}},
		DeviceCommands:  []dtos.DeviceCommand{{Name: "command1", ReadWrite: common.ReadWrite_R}, {Name: "command2", ReadWrite: common.ReadWrite_W}},
	}}, nil)
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.DeviceProfileClientName: func(get di.Get) interface{} {
			return clientMock
		},
	})

	test := []struct {
		name            string
		device          models.Device
		expectedSources []string
	}{
		{"no auto events in device", models.Device{Name: deviceName, ProfileName: profileName, AutoEvents: nil}, []string{"resource1", "command1"}},
		{"one auto events in device", models.Device{Name: deviceName, ProfileName: profileName, AutoEvents: []models.AutoEvent{{SourceName: "resource1"}}}, []string{"command1"}},
		{"two auto events in device", models.Device{Name: deviceName, ProfileName: profileName, AutoEvents: []models.AutoEvent{{SourceName: "resource1"}, {SourceName: "command1"}}}, nil},
	}
	for _, testCase := range test {
		t.Run(testCase.name, func(t *testing.T) {
			sources, err := noneAutoEventSourcesByDevice(testCase.device, dic)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedSources, sources)
		})
	}
}

func TestProcessNumericReadings_SimpleReadingToNumeric(t *testing.T) {
	test := []struct {
		name      string
		valueType string
		value     any
		expected  any
	}{
		{"int8 string to numeric", common.ValueTypeInt8, int8(123), int64(123)},
		{"int16 string to numeric", common.ValueTypeInt16, int16(1234), int64(1234)},
		{"int32 string to numeric", common.ValueTypeInt32, int32(12345), int64(12345)},
		{"int64 string to numeric", common.ValueTypeInt64, int64(123456), int64(123456)},
		{"uint8 string to numeric", common.ValueTypeUint8, uint8(123), uint64(123)},
		{"uint16 string to numeric", common.ValueTypeUint16, uint16(1234), uint64(1234)},
		{"uint32 string to numeric", common.ValueTypeUint32, uint32(12345), uint64(12345)},
		{"uint64 string to numeric", common.ValueTypeUint64, uint64(123456), uint64(123456)},
		{"float32 string to numeric", common.ValueTypeFloat32, float32(11.123), 11.123},
		{"float64 string to numeric", common.ValueTypeFloat64, 11.123456, 11.123456},
		{"int8Array string to numeric", common.ValueTypeInt8Array, []int8{12, 123}, []int64{12, 123}},
		{"int16Array string to numeric", common.ValueTypeInt16Array, []int16{123, 1234}, []int64{123, 1234}},
		{"int32Array string to numeric", common.ValueTypeInt32Array, []int32{1234, 12345}, []int64{1234, 12345}},
		{"int64Array string to numeric", common.ValueTypeInt64Array, []int64{12345, 123456}, []int64{12345, 123456}},
		{"uint8Array string to numeric", common.ValueTypeUint8Array, []uint8{12, 123}, []uint64{12, 123}},
		{"uint16Array string to numeric", common.ValueTypeUint16Array, []uint16{123, 1234}, []uint64{123, 1234}},
		{"uint32Array string to numeric", common.ValueTypeUint32Array, []uint32{1234, 12345}, []uint64{1234, 12345}},
		{"uint64Array string to numeric", common.ValueTypeUint64Array, []uint64{12345, 123456}, []uint64{12345, 123456}},
		{"float32Array string to numeric", common.ValueTypeFloat32Array, []float32{1.12, 11.123}, []float64{1.12, 11.123}},
		{"float32Array string to numeric", common.ValueTypeFloat64Array, []float64{1.12, 11.123456}, []float64{1.12, 11.123456}},
	}
	for _, testCase := range test {
		t.Run(testCase.name, func(t *testing.T) {
			r, err := dtos.NewSimpleReading(testProfileName, testDeviceName, testDeviceResourceName, testCase.valueType, testCase.value)
			require.NoError(t, err)
			readings := []dtos.BaseReading{r}
			processNumericReadings(true, readings)
			assert.Equal(t, testCase.expected, readings[0].NumericValue)
		})
	}
}

func TestProcessNumericReadings_NumericReadingToString(t *testing.T) {
	test := []struct {
		name      string
		valueType string
		value     any
		expected  any
	}{
		{"int8 numeric to string", common.ValueTypeInt8, int8(123), "123"},
		{"int16 numeric to string", common.ValueTypeInt16, int16(1234), "1234"},
		{"int32 numeric to string", common.ValueTypeInt32, int32(12345), "12345"},
		{"int64 numeric to string", common.ValueTypeInt64, int64(123456), "123456"},
		{"uint8 numeric to string", common.ValueTypeUint8, uint8(123), "123"},
		{"uint16 numeric to string", common.ValueTypeUint16, uint16(1234), "1234"},
		{"uint32 numeric to string", common.ValueTypeUint32, uint32(12345), "12345"},
		{"uint64 numeric to string", common.ValueTypeUint64, uint64(123456), "123456"},
		{"float32 numeric to string", common.ValueTypeFloat32, float32(11.123), "1.1123e+01"},
		{"float64 numeric to string", common.ValueTypeFloat64, 11.123456, "1.1123456e+01"},
	}
	for _, testCase := range test {
		t.Run(testCase.name, func(t *testing.T) {
			r := dtos.NewNumericReading(testProfileName, testDeviceName, testDeviceResourceName, testCase.valueType, testCase.value)
			readings := []dtos.BaseReading{r}
			processNumericReadings(false, readings)

			assert.Equal(t, testCase.expected, readings[0].Value)
		})
	}
}
