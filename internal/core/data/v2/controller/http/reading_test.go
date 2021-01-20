package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	v2 "github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadingTotalCount(t *testing.T) {
	expectedReadingCount := uint32(656672)
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingTotalCount").Return(expectedReadingCount, nil)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	rc := NewReadingController(dic)

	req, err := http.NewRequest(http.MethodGet, v2.ApiReadingCountRoute, http.NoBody)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(rc.ReadingTotalCount)
	handler.ServeHTTP(recorder, req)

	var actualResponse common.CountResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)
	assert.Equal(t, v2.ApiVersion, actualResponse.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, http.StatusOK, int(actualResponse.StatusCode), "Response status code not as expected")
	assert.Empty(t, actualResponse.Message, "Message should be empty when it is successful")
	assert.Equal(t, expectedReadingCount, actualResponse.Count, "Event count in the response body is not expected")
}

func TestAllReadings(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AllReadings", 0, 20).Return([]models.Reading{}, nil)
	dbClientMock.On("AllReadings", 0, 1).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewReadingController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - get readings without offset, and limit", "", "", false, http.StatusOK},
		{"Valid - get readings with offset, and limit", "0", "1", false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiAllReadingRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllReadings)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestReadingsByTimeRange(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByTimeRange", 0, 100, 0, 10).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	rc := NewReadingController(dic)
	assert.NotNil(t, rc)

	tests := []struct {
		name               string
		start              string
		end                string
		offset             string
		limit              string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - with proper start/end/offset/limit", "0", "100", "0", "10", false, 0, http.StatusOK},
		{"Invalid - invalid start format", "aaa", "100", "0", "10", true, 0, http.StatusBadRequest},
		{"Invalid - invalid end format", "0", "bbb", "0", "10", true, 0, http.StatusBadRequest},
		{"Invalid - empty start", "", "100", "0", "10", true, 0, http.StatusBadRequest},
		{"Invalid - empty end", "0", "", "0", "10", true, 0, http.StatusBadRequest},
		{"Invalid - end before start", "10", "0", "0", "10", true, 0, http.StatusBadRequest},
		{"Invalid - invalid offset format", "0", "100", "aaa", "10", true, 0, http.StatusBadRequest},
		{"Invalid - invalid limit format", "0", "100", "0", "aaa", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiReadingByTimeRangeRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(v2.Offset, testCase.offset)
			query.Add(v2.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Start: testCase.start, v2.End: testCase.end})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(rc.ReadingsByTimeRange)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Readings), "Device count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestReadingsByResourceName(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByResourceName", 0, 20, TestDeviceResourceName).Return([]models.Reading{}, nil)
	dbClientMock.On("ReadingsByResourceName", 0, 1, TestDeviceResourceName).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewReadingController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		resourceName       string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - get readings without offset, and limit", "", "", TestDeviceResourceName, false, http.StatusOK},
		{"Valid - get readings with offset, and limit", "0", "1", TestDeviceResourceName, false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", TestDeviceResourceName, true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", TestDeviceResourceName, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiReadingByResourceNameRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.ResourceName: testCase.resourceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ReadingsByResourceName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestReadingsByDeviceName(t *testing.T) {
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByDeviceName", 0, 20, TestDeviceName).Return([]models.Reading{}, nil)
	dbClientMock.On("ReadingsByDeviceName", 0, 1, TestDeviceName).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewReadingController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		deviceName         string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - get readings without offset, and limit", "", "", TestDeviceName, false, http.StatusOK},
		{"Valid - get readings with offset, and limit", "0", "1", TestDeviceName, false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", TestDeviceName, true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", TestDeviceName, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiReadingByDeviceNameRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.deviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ReadingsByDeviceName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestReadingCountByDeviceName(t *testing.T) {
	expectedReadingCount := uint32(656672)
	deviceName := "deviceA"
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByDeviceName", deviceName).Return(expectedReadingCount, nil)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	rc := NewReadingController(dic)

	req, err := http.NewRequest(http.MethodGet, v2.ApiReadingCountByDeviceNameRoute, http.NoBody)
	req = mux.SetURLVars(req, map[string]string{v2.Name: deviceName})
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(rc.ReadingCountByDeviceName)
	handler.ServeHTTP(recorder, req)

	var actualResponse common.CountResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)
	assert.Equal(t, v2.ApiVersion, actualResponse.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, http.StatusOK, int(actualResponse.StatusCode), "Response status code not as expected")
	assert.Empty(t, actualResponse.Message, "Message should be empty when it is successful")
	assert.Equal(t, expectedReadingCount, actualResponse.Count, "Reading count in the response body is not expected")
}
