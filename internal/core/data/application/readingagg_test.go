//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"net/http"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/core/data/infrastructure/interfaces"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/query"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/stretchr/testify/assert"
)

var (
	device1     = "device1"
	device2     = "device2"
	resource1   = "resource1"
	resource2   = "resource2"
	invalidFunc = "invalid"
	validStart  = int64(1755244649710610498)
	validEnd    = int64(1765244649710610498)
	aggReading  = models.SimpleReading{
		BaseReading: models.BaseReading{
			DeviceName:   device1,
			ResourceName: resource1,
			ProfileName:  "TempProfile",
			ValueType:    common.ValueTypeUint16,
		},
		Value: "9",
	}
	aggReading2 = models.SimpleReading{
		BaseReading: models.BaseReading{
			DeviceName:   device2,
			ResourceName: resource2,
			ProfileName:  "TempProfile",
			ValueType:    common.ValueTypeUint16,
		},
		Value: "5",
	}
	aggReading3 = models.SimpleReading{
		BaseReading: models.BaseReading{
			DeviceName:   device2,
			ResourceName: resource1,
			ProfileName:  "TempProfile",
			ValueType:    common.ValueTypeInt8,
		},
		Value: "2",
	}
	aggReading4 = models.SimpleReading{
		BaseReading: models.BaseReading{
			DeviceName:   device1,
			ResourceName: resource2,
			ProfileName:  "TempProfile",
			ValueType:    common.ValueTypeUint16,
		},
		Value: "7",
	}
)

func TestAllReadingsAggregation(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AllReadingsAggregation", "MAX").Return([]models.Reading{aggReading, aggReading2}, nil)
	dbClientMock.On("AllReadingsAggregation", invalidFunc).Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unknown sql func", nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		aggFunc            string
		errorExpected      bool
		expectedCount      int
		ExpectedErrKind    errors.ErrKind
		expectedStatusCode int
	}{
		{"Valid - all aggregated readings", common.MaxFunc, false, 2, "", http.StatusOK},
		{"InValid - invalid aggregate function", invalidFunc, true, 0, errors.KindServerError, http.StatusInternalServerError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := AllAggregateReadings(testCase.aggFunc, dic)
			if testCase.errorExpected {
				assert.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(result), "Result count not as expected")
			}
		})
	}
}

func TestAllAggregateReadingsByTimeRange(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AllReadingsAggregationByTimeRange", "MAX", validStart, validEnd).Return([]models.Reading{aggReading, aggReading2}, nil)
	dbClientMock.On("AllReadingsAggregationByTimeRange", invalidFunc, int64(0), int64(1)).Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unknown sql func", nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		aggFunc            string
		start              int64
		end                int64
		errorExpected      bool
		expectedCount      int
		ExpectedErrKind    errors.ErrKind
		expectedStatusCode int
	}{
		{"Valid - all aggregated readings by time range", common.MaxFunc, validStart, validEnd, false, 2, "", http.StatusOK},
		{"InValid - invalid aggregate function", invalidFunc, 0, 1, true, 0, errors.KindServerError, http.StatusInternalServerError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := AllAggregateReadingsByTimeRange(testCase.aggFunc, query.Parameters{Start: testCase.start, End: testCase.end}, dic)
			if testCase.errorExpected {
				assert.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(result), "Result count not as expected")
			}
		})
	}
}

func TestAggregateReadingsByResourceName(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsAggregationByResourceName", resource1, "MIN").Return([]models.Reading{aggReading, aggReading3}, nil)
	dbClientMock.On("ReadingsAggregationByResourceName", resource1, invalidFunc).Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unknown sql func", nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		resourceName       string
		aggFunc            string
		errorExpected      bool
		expectedCount      int
		ExpectedErrKind    errors.ErrKind
		expectedStatusCode int
	}{
		{"Valid - all aggregated readings by resource", resource1, common.MinFunc, false, 2, "", http.StatusOK},
		{"InValid - invalid aggregate function", resource1, invalidFunc, true, 0, errors.KindServerError, http.StatusInternalServerError},
		{"InValid - empty resource name", "", "", true, 0, errors.KindContractInvalid, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := AggregateReadingsByResourceName(testCase.resourceName, testCase.aggFunc, dic)
			if testCase.errorExpected {
				assert.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(result), "Result count not as expected")
			}
		})
	}
}

func TestAggregateReadingsByResourceNameAndTimeRange(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsAggregationByResourceNameAndTimeRange", resource1, "MIN", validStart, validEnd).Return([]models.Reading{aggReading, aggReading3}, nil)
	dbClientMock.On("ReadingsAggregationByResourceNameAndTimeRange", resource1, invalidFunc, int64(0), int64(1)).Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unknown sql func", nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		resourceName       string
		aggFunc            string
		start              int64
		end                int64
		errorExpected      bool
		expectedCount      int
		ExpectedErrKind    errors.ErrKind
		expectedStatusCode int
	}{
		{"Valid - all aggregated readings by resource and time range", resource1, common.MinFunc, validStart, validEnd, false, 2, "", http.StatusOK},
		{"InValid - invalid aggregate function", resource1, invalidFunc, int64(0), int64(1), true, 0, errors.KindServerError, http.StatusInternalServerError},
		{"InValid - empty resource name", "", "", int64(0), int64(1), true, 0, errors.KindContractInvalid, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := AggregateReadingsByResourceNameAndTimeRange(testCase.resourceName, testCase.aggFunc, query.Parameters{Start: testCase.start, End: testCase.end}, dic)
			if testCase.errorExpected {
				assert.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(result), "Result count not as expected")
			}
		})
	}
}

func TestAggregateReadingsByDeviceName(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsAggregationByDeviceName", device1, "AVG").Return([]models.Reading{aggReading, aggReading4}, nil)
	dbClientMock.On("ReadingsAggregationByDeviceName", device1, invalidFunc).Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unknown sql func", nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		deviceName         string
		aggFunc            string
		errorExpected      bool
		expectedCount      int
		ExpectedErrKind    errors.ErrKind
		expectedStatusCode int
	}{
		{"Valid - all aggregated readings by device", device1, common.AvgFunc, false, 2, "", http.StatusOK},
		{"InValid - invalid aggregate function", device1, invalidFunc, true, 0, errors.KindServerError, http.StatusInternalServerError},
		{"InValid - empty device name", "", "", true, 0, errors.KindContractInvalid, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := AggregateReadingsByDeviceName(testCase.deviceName, testCase.aggFunc, dic)
			if testCase.errorExpected {
				assert.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(result), "Result count not as expected")
			}
		})
	}
}

func TestAggregateReadingsByDeviceNameAndTimeRange(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsAggregationByDeviceNameAndTimeRange", device1, "MIN", validStart, validEnd).Return([]models.Reading{aggReading, aggReading4}, nil)
	dbClientMock.On("ReadingsAggregationByDeviceNameAndTimeRange", device1, invalidFunc, int64(0), int64(1)).Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unknown sql func", nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		deviceName         string
		aggFunc            string
		start              int64
		end                int64
		errorExpected      bool
		expectedCount      int
		ExpectedErrKind    errors.ErrKind
		expectedStatusCode int
	}{
		{"Valid - all aggregated readings by device and time range", device1, common.MinFunc, validStart, validEnd, false, 2, "", http.StatusOK},
		{"InValid - invalid aggregate function", device1, invalidFunc, int64(0), int64(1), true, 0, errors.KindServerError, http.StatusInternalServerError},
		{"InValid - empty device name", "", "", int64(0), int64(1), true, 0, errors.KindContractInvalid, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := AggregateReadingsByDeviceNameAndTimeRange(testCase.deviceName, testCase.aggFunc, query.Parameters{Start: testCase.start, End: testCase.end}, dic)
			if testCase.errorExpected {
				assert.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(result), "Result count not as expected")
			}
		})
	}
}

func TestAggregateReadingsByDeviceNameAndResourceName(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsAggregationByDeviceNameAndResourceName", device1, resource1, "COUNT").Return([]models.Reading{aggReading}, nil)
	dbClientMock.On("ReadingsAggregationByDeviceNameAndResourceName", device1, resource1, invalidFunc).Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unknown sql func", nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		deviceName         string
		resourceName       string
		aggFunc            string
		errorExpected      bool
		expectedCount      int
		ExpectedErrKind    errors.ErrKind
		expectedStatusCode int
	}{
		{"Valid - all aggregated readings by device and resource", device1, resource1, common.CountFunc, false, 1, "", http.StatusOK},
		{"InValid - invalid aggregate function", device1, resource1, invalidFunc, true, 0, errors.KindServerError, http.StatusInternalServerError},
		{"InValid - empty device name", "", resource1, "", true, 0, errors.KindContractInvalid, http.StatusBadRequest},
		{"InValid - empty resource name", device1, "", "", true, 0, errors.KindContractInvalid, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := AggregateReadingsByDeviceNameAndResourceName(testCase.deviceName, testCase.resourceName, testCase.aggFunc, dic)
			if testCase.errorExpected {
				assert.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(result), "Result count not as expected")
			}
		})
	}
}

func TestAggregateReadingsByDeviceNameAndResourceNameAndTimeRange(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange", device1, resource1, "MIN", validStart, validEnd).Return([]models.Reading{aggReading, aggReading4}, nil)
	dbClientMock.On("ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange", device1, resource1, invalidFunc, int64(0), int64(1)).Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unknown sql func", nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name               string
		deviceName         string
		resourceName       string
		aggFunc            string
		start              int64
		end                int64
		errorExpected      bool
		expectedCount      int
		ExpectedErrKind    errors.ErrKind
		expectedStatusCode int
	}{
		{"Valid - all aggregated readings by device and time range", device1, resource1, common.MinFunc, validStart, validEnd, false, 2, "", http.StatusOK},
		{"InValid - invalid aggregate function", device1, resource1, invalidFunc, int64(0), int64(1), true, 0, errors.KindServerError, http.StatusInternalServerError},
		{"InValid - empty device name", "", resource1, "", int64(0), int64(1), true, 0, errors.KindContractInvalid, http.StatusBadRequest},
		{"InValid - empty resource name", device1, "", "", int64(0), int64(1), true, 0, errors.KindContractInvalid, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := AggregateReadingsByDeviceNameAndResourceNameAndTimeRange(testCase.deviceName, testCase.resourceName, testCase.aggFunc, query.Parameters{Start: testCase.start, End: testCase.end}, dic)
			if testCase.errorExpected {
				assert.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(result), "Result count not as expected")
			}
		})
	}
}

func TestGetReadingAggregation(t *testing.T) {
	dic := mocks.NewMockDIC()

	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange", device1, resource1, "MIN", validStart, validEnd).Return([]models.Reading{aggReading}, nil)
	expectedDTO := dtos.FromReadingModelToDTO(aggReading)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	aggDBFunc := func(dbClient interfaces.DBClient) ([]models.Reading, errors.EdgeX) {
		return dbClientMock.ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange(device1, resource1, "MIN", validStart, validEnd)
	}
	result, err := getReadingAggregation(dic, aggDBFunc)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result), "Result count not as expected")
	assert.Equal(t, expectedDTO, result[0], "Reading aggregated readings not as expected")
}
