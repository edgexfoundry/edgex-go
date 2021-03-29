package application

import (
	"net/http"
	"testing"

	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllReadings(t *testing.T) {
	readings := buildReadings()

	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AllReadings", 0, 20).Return(readings, nil)
	dbClientMock.On("AllReadings", 3, 10).Return([]models.Reading{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedStatusCode int
	}{
		{"Valid - all readings", 0, 20, false, "", len(readings), http.StatusOK},
		{"Invalid - bounds out of range", 3, 10, true, errors.KindRangeNotSatisfiable, 0, http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			readings, err := AllReadings(testCase.offset, testCase.limit, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(readings), "Reading total count is not expected")
			}
		})
	}
}

func TestReadingsByTimeRange(t *testing.T) {
	readings := buildReadings()

	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByTimeRange", int(readings[0].GetBaseReading().Origin), int(readings[4].GetBaseReading().Origin), 0, 10).Return(readings, nil)
	dbClientMock.On("ReadingsByTimeRange", int(readings[1].GetBaseReading().Origin), int(readings[3].GetBaseReading().Origin), 0, 10).Return([]models.Reading{readings[3], readings[2], readings[1]}, nil)
	dbClientMock.On("ReadingsByTimeRange", int(readings[1].GetBaseReading().Origin), int(readings[3].GetBaseReading().Origin), 1, 2).Return([]models.Reading{readings[2], readings[1]}, nil)
	dbClientMock.On("ReadingsByTimeRange", int(readings[1].GetBaseReading().Origin), int(readings[3].GetBaseReading().Origin), 4, 2).Return(nil, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range", nil))
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
		ExpectedErrKind    errors.ErrKind
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - all readings", int(readings[0].GetBaseReading().Origin), int(readings[4].GetBaseReading().Origin), 0, 10, false, "", 5, http.StatusOK},
		{"Valid - readings trimmed by latest and oldest", int(readings[1].GetBaseReading().Origin), int(readings[3].GetBaseReading().Origin), 0, 10, false, "", 3, http.StatusOK},
		{"Valid - readings trimmed by latest and oldest and skipped first", int(readings[1].GetBaseReading().Origin), int(readings[3].GetBaseReading().Origin), 1, 2, false, "", 2, http.StatusOK},
		{"Invalid - bounds out of range", int(readings[1].GetBaseReading().Origin), int(readings[3].GetBaseReading().Origin), 4, 2, true, errors.KindRangeNotSatisfiable, 0, http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			readings, err := ReadingsByTimeRange(testCase.start, testCase.end, testCase.offset, testCase.limit, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(readings), "Reading total count is not expected")
			}
		})
	}
}

func TestReadingsByResourceName(t *testing.T) {
	readings := buildReadings()

	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByResourceName", 0, 20, testDeviceResourceName).Return(readings, nil)
	dbClientMock.On("ReadingsByResourceName", len(readings)+1, 10, testDeviceResourceName).Return([]models.Reading{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedStatusCode int
	}{
		{"Valid - all readings", 0, 20, testDeviceResourceName, false, "", len(readings), http.StatusOK},
		{"Invalid - bounds out of range", len(readings) + 1, 10, testDeviceResourceName, true, errors.KindRangeNotSatisfiable, 0, http.StatusRequestedRangeNotSatisfiable},
		{"Invalid - empty resource name", len(readings) + 1, 10, "", true, errors.KindContractInvalid, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			readings, err := ReadingsByResourceName(testCase.offset, testCase.limit, testCase.resourceName, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(readings), "Reading total count is not expected")
			}
		})
	}
}

func TestReadingsByDeviceName(t *testing.T) {
	readings := buildReadings()

	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByDeviceName", 0, 20, testDeviceName).Return(readings, nil)
	dbClientMock.On("ReadingsByDeviceName", 3, 10, testDeviceName).Return([]models.Reading{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedStatusCode int
	}{
		{"Valid - all readings", 0, 20, testDeviceName, false, "", len(readings), http.StatusOK},
		{"Invalid - bounds out of range", 3, 10, testDeviceName, true, errors.KindRangeNotSatisfiable, 0, http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			readings, err := ReadingsByDeviceName(testCase.offset, testCase.limit, testCase.deviceName, dic)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.ExpectedErrKind, errors.Kind(err), "Error kind not as expected")
				assert.Equal(t, testCase.expectedStatusCode, err.Code(), "Status code not as expected")
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedCount, len(readings), "Reading total count is not expected")
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
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
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
