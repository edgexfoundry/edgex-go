package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

func TestReadingTotalCount(t *testing.T) {
	expectedReadingCount := uint32(656672)
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingTotalCount").Return(expectedReadingCount, nil)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	rc := NewReadingController(dic)

	req, err := http.NewRequest(http.MethodGet, common.ApiReadingCountRoute, http.NoBody)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(rc.ReadingTotalCount)
	handler.ServeHTTP(recorder, req)

	var actualResponse commonDTO.CountResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)
	assert.Equal(t, common.ApiVersion, actualResponse.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, http.StatusOK, int(actualResponse.StatusCode), "Response status code not as expected")
	assert.Empty(t, actualResponse.Message, "Message should be empty when it is successful")
	assert.Equal(t, expectedReadingCount, actualResponse.Count, "Event count in the response body is not expected")
}

func TestAllReadings(t *testing.T) {
	totalCount := uint32(0)
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingTotalCount").Return(totalCount, nil)
	dbClientMock.On("AllReadings", 0, 20).Return([]models.Reading{}, nil)
	dbClientMock.On("AllReadings", 0, 1).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get readings without offset, and limit", "", "", false, totalCount, http.StatusOK},
		{"Valid - get readings with offset, and limit", "0", "1", false, totalCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", true, totalCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiAllReadingRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllReadings)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
			}
		})
	}
}

func TestReadingsByTimeRange(t *testing.T) {
	totalCount := uint32(0)
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByTimeRange", 0, 100).Return(totalCount, nil)
	dbClientMock.On("ReadingsByTimeRange", 0, 100, 0, 10).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - with proper start/end/offset/limit", "0", "100", "0", "10", false, 0, totalCount, http.StatusOK},
		{"Invalid - invalid start format", "aaa", "100", "0", "10", true, 0, totalCount, http.StatusBadRequest},
		{"Invalid - invalid end format", "0", "bbb", "0", "10", true, 0, totalCount, http.StatusBadRequest},
		{"Invalid - empty start", "", "100", "0", "10", true, 0, totalCount, http.StatusBadRequest},
		{"Invalid - empty end", "0", "", "0", "10", true, 0, totalCount, http.StatusBadRequest},
		{"Invalid - end before start", "10", "0", "0", "10", true, 0, totalCount, http.StatusBadRequest},
		{"Invalid - invalid offset format", "0", "100", "aaa", "10", true, 0, totalCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "0", "100", "0", "aaa", true, 0, totalCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiReadingByTimeRangeRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Start: testCase.start, common.End: testCase.end})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(rc.ReadingsByTimeRange)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Readings), "Device count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
			}
		})
	}
}

func TestReadingsByResourceName(t *testing.T) {
	totalCount := uint32(0)
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByResourceName", TestDeviceResourceName).Return(totalCount, nil)
	dbClientMock.On("ReadingsByResourceName", 0, 20, TestDeviceResourceName).Return([]models.Reading{}, nil)
	dbClientMock.On("ReadingsByResourceName", 0, 1, TestDeviceResourceName).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get readings without offset, and limit", "", "", TestDeviceResourceName, false, totalCount, http.StatusOK},
		{"Valid - get readings with offset, and limit", "0", "1", TestDeviceResourceName, false, totalCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", TestDeviceResourceName, true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", TestDeviceResourceName, true, totalCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiReadingByResourceNameRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.ResourceName: testCase.resourceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ReadingsByResourceName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
			}
		})
	}
}

func TestReadingsByDeviceName(t *testing.T) {
	totalCount := uint32(0)
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByDeviceName", TestDeviceName).Return(totalCount, nil)
	dbClientMock.On("ReadingsByDeviceName", 0, 20, TestDeviceName).Return([]models.Reading{}, nil)
	dbClientMock.On("ReadingsByDeviceName", 0, 1, TestDeviceName).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get readings without offset, and limit", "", "", TestDeviceName, false, totalCount, http.StatusOK},
		{"Valid - get readings with offset, and limit", "0", "1", TestDeviceName, false, totalCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", TestDeviceName, true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", TestDeviceName, true, totalCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiReadingByDeviceNameRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ReadingsByDeviceName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	rc := NewReadingController(dic)

	req, err := http.NewRequest(http.MethodGet, common.ApiReadingCountByDeviceNameRoute, http.NoBody)
	req = mux.SetURLVars(req, map[string]string{common.Name: deviceName})
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(rc.ReadingCountByDeviceName)
	handler.ServeHTTP(recorder, req)

	var actualResponse commonDTO.CountResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)
	assert.Equal(t, common.ApiVersion, actualResponse.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, http.StatusOK, int(actualResponse.StatusCode), "Response status code not as expected")
	assert.Empty(t, actualResponse.Message, "Message should be empty when it is successful")
	assert.Equal(t, expectedReadingCount, actualResponse.Count, "Reading count in the response body is not expected")
}

func TestReadingsByResourceNameAndTimeRange(t *testing.T) {
	totalCount := uint32(0)
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByResourceNameAndTimeRange", TestDeviceResourceName, 0, 100).Return(totalCount, nil)
	dbClientMock.On("ReadingsByResourceNameAndTimeRange", TestDeviceResourceName, 0, 100, 0, 10).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	rc := NewReadingController(dic)
	assert.NotNil(t, rc)

	tests := []struct {
		name               string
		resourceName       string
		start              string
		end                string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid ", TestDeviceResourceName, "0", "100", "0", "10", false, totalCount, http.StatusOK},
		{"Invalid - empty resourceName", "", "0", "100", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid start format", TestDeviceResourceName, "aaa", "100", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid end format", TestDeviceResourceName, "0", "bbb", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - empty start", TestDeviceResourceName, "", "100", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - empty end", TestDeviceResourceName, "0", "", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - end before start", TestDeviceResourceName, "10", "0", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid offset format", TestDeviceResourceName, "0", "100", "aaa", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", TestDeviceResourceName, "0", "100", "0", "aaa", true, totalCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiReadingByResourceNameAndTimeRangeRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.ResourceName: testCase.resourceName, common.Start: testCase.start, common.End: testCase.end})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(rc.ReadingsByResourceNameAndTimeRange)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
			}
		})
	}
}

func TestReadingsByDeviceNameAndResourceName(t *testing.T) {
	totalCount := uint32(0)
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByDeviceNameAndResourceName", TestDeviceName, TestDeviceResourceName).Return(totalCount, nil)
	dbClientMock.On("ReadingsByDeviceNameAndResourceName", TestDeviceName, TestDeviceResourceName, 0, 20).Return([]models.Reading{}, nil)
	dbClientMock.On("ReadingsByDeviceNameAndResourceName", TestDeviceName, TestDeviceResourceName, 0, 1).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewReadingController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceName         string
		resourceName       string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"valid - get readings without offset, and limit", TestDeviceName, TestDeviceResourceName, "", "", false, totalCount, http.StatusOK},
		{"valid - get readings with offset, and limit", TestDeviceName, TestDeviceResourceName, "0", "1", false, totalCount, http.StatusOK},
		{"invalid - empty deviceName", "", TestDeviceResourceName, "0", "1", true, totalCount, http.StatusBadRequest},
		{"invalid - empty resourceName", TestDeviceName, "", "0", "1", true, totalCount, http.StatusBadRequest},
		{"invalid - invalid offset format", TestDeviceName, TestDeviceResourceName, "aaa", "1", true, totalCount, http.StatusBadRequest},
		{"invalid - invalid limit format", TestDeviceName, TestDeviceResourceName, "1", "aaa", true, totalCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiReadingByDeviceNameAndResourceNameRoute, http.NoBody)
			require.NoError(t, err)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceName, common.ResourceName: testCase.resourceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ReadingsByDeviceNameAndResourceName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
			}
		})
	}
}

func TestReadingsByDeviceNameAndResourceNameAndTimeRange(t *testing.T) {
	totalCount := uint32(0)
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByDeviceNameAndResourceNameAndTimeRange", TestDeviceName, TestDeviceResourceName, 0, 100).Return(totalCount, nil)
	dbClientMock.On("ReadingsByDeviceNameAndResourceNameAndTimeRange", TestDeviceName, TestDeviceResourceName, 0, 100, 0, 10).Return([]models.Reading{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	rc := NewReadingController(dic)
	assert.NotNil(t, rc)

	tests := []struct {
		name               string
		deviceName         string
		resourceName       string
		start              string
		end                string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid ", TestDeviceName, TestDeviceResourceName, "0", "100", "0", "10", false, totalCount, http.StatusOK},
		{"Invalid - empty deviceName", "", TestDeviceResourceName, "0", "100", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - empty resourceName", TestDeviceName, "", "0", "100", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid start format", TestDeviceName, TestDeviceResourceName, "aaa", "100", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid end format", TestDeviceName, TestDeviceResourceName, "0", "bbb", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - empty start", TestDeviceName, TestDeviceResourceName, "", "100", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - empty end", TestDeviceName, TestDeviceResourceName, "0", "", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - end before start", TestDeviceName, TestDeviceResourceName, "10", "0", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid offset format", TestDeviceName, TestDeviceResourceName, "0", "100", "aaa", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", TestDeviceName, TestDeviceResourceName, "0", "100", "0", "aaa", true, totalCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiReadingByDeviceNameAndResourceNameAndTimeRangeRoute, http.NoBody)
			require.NoError(t, err)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceName, common.ResourceName: testCase.resourceName, common.Start: testCase.start, common.End: testCase.end})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(rc.ReadingsByDeviceNameAndResourceNameAndTimeRange)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
			}
		})
	}
}

func TestReadingsByDeviceNameAndResourceNamesAndTimeRange(t *testing.T) {
	totalCount := uint32(0)
	testResourceNames := []string{"resource01", "resource02"}
	emptyPayload := make(map[string]interface{})
	testResourceNamesPayload := emptyPayload
	testResourceNamesPayload[common.ResourceNames] = testResourceNames
	dic := mocks.NewMockDIC()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCountByDeviceNameAndTimeRange", TestDeviceName, 0, 100).Return(totalCount, nil)
	dbClientMock.On("ReadingsByDeviceNameAndTimeRange", TestDeviceName, 0, 100, 0, 10).Return([]models.Reading{}, nil)
	dbClientMock.On("ReadingsByDeviceNameAndResourceNamesAndTimeRange", TestDeviceName, testResourceNames, 0, 100, 0, 10).Return([]models.Reading{}, totalCount, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	rc := NewReadingController(dic)
	assert.NotNil(t, rc)

	tests := []struct {
		name               string
		deviceName         string
		payload            map[string]interface{}
		start              string
		end                string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - provide deviceName and nil resourceNames", TestDeviceName, nil, "0", "100", "0", "10", false, totalCount, http.StatusOK},
		{"Valid - provide deviceName and empty resourceNames", TestDeviceName, emptyPayload, "0", "100", "0", "10", false, totalCount, http.StatusOK},
		{"Valid - provide deviceName and resourceNames", TestDeviceName, testResourceNamesPayload, "0", "100", "0", "10", false, totalCount, http.StatusOK},
		{"Invalid - empty deviceName", "", testResourceNamesPayload, "0", "100", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid start format", TestDeviceName, testResourceNamesPayload, "aaa", "100", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid end format", TestDeviceName, testResourceNamesPayload, "0", "bbb", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - empty start", TestDeviceName, testResourceNamesPayload, "", "100", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - empty end", TestDeviceName, testResourceNamesPayload, "0", "", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - end before start", TestDeviceName, testResourceNamesPayload, "10", "0", "0", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid offset format", TestDeviceName, testResourceNamesPayload, "0", "100", "aaa", "10", true, totalCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", TestDeviceName, testResourceNamesPayload, "0", "100", "0", "aaa", true, totalCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var reader io.Reader
			if testCase.payload != nil {
				byteData, err := toByteArray(common.ContentTypeJSON, testCase.payload)
				require.NoError(t, err)
				reader = strings.NewReader(string(byteData))
			} else {
				reader = http.NoBody
			}
			req, err := http.NewRequest(http.MethodGet, common.ApiReadingByDeviceNameAndTimeRangeRoute, reader)
			req.Header.Set(common.ContentType, common.ContentTypeJSON)
			require.NoError(t, err)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceName, common.Start: testCase.start, common.End: testCase.end})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(rc.ReadingsByDeviceNameAndResourceNamesAndTimeRange)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiReadingsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
			}
		})
	}
}
