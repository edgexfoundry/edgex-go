//
// Copyright (C) 2023 IOTech Ltd
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/query"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
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

	app := NewCoreDataApp(dic)

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
			readings, total, err := app.AllReadings(query.Parameters{Offset: testCase.offset, Limit: testCase.limit}, dic)
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

	app := NewCoreDataApp(dic)

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
			readings, totalCount, err := app.ReadingsByTimeRange(query.Parameters{Start: testCase.start, End: testCase.end, Offset: testCase.offset, Limit: testCase.limit}, dic)
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

	app := NewCoreDataApp(dic)

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
			readings, total, err := app.ReadingsByResourceName(query.Parameters{Offset: testCase.offset, Limit: testCase.limit}, testCase.resourceName, dic)
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

	app := NewCoreDataApp(dic)

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
			readings, total, err := app.ReadingsByDeviceName(query.Parameters{Offset: testCase.offset, Limit: testCase.limit}, testCase.deviceName, dic)
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

	app := NewCoreDataApp(dic)

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
			count, err := app.ReadingCountByDeviceName(testCase.deviceName, dic)
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
