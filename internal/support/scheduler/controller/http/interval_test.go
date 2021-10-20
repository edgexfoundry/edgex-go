//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces/mocks"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
				Service: bootstrapConfig.ServiceInfo{
					MaxResultCount: 20,
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
}

func addIntervalRequestData() requests.AddIntervalRequest {
	var testAddIntervalReq = requests.AddIntervalRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		Interval: dtos.Interval{
			Id:       ExampleUUID,
			Name:     TestIntervalName,
			Start:    TestIntervalStart,
			End:      TestIntervalEnd,
			Interval: TestIntervalFrequency,
		},
	}

	return testAddIntervalReq
}

func updateIntervalRequestData() requests.UpdateIntervalRequest {
	testUUID := ExampleUUID
	testIntervalName := TestIntervalName
	testFrequency := TestIntervalFrequency
	var req = requests.UpdateIntervalRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		Interval: dtos.UpdateInterval{
			Id:       &testUUID,
			Name:     &testIntervalName,
			Interval: &testFrequency,
		},
	}

	return req
}

func TestAddInterval(t *testing.T) {
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	schedulerManagerMock := &dbMock.SchedulerManager{}

	valid := addIntervalRequestData()
	model := dtos.ToIntervalModel(valid.Interval)
	dbClientMock.On("AddInterval", model).Return(model, nil)
	schedulerManagerMock.On("AddInterval", model).Return(nil)

	noName := addIntervalRequestData()
	noName.Interval.Name = ""
	noRequestId := addIntervalRequestData()
	noRequestId.RequestId = ""

	duplicatedName := addIntervalRequestData()
	duplicatedName.Interval.Name = "duplicatedName"
	model = dtos.ToIntervalModel(duplicatedName.Interval)
	dbClientMock.On("AddInterval", model).Return(model, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("interval name %s already exists", model.Name), nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.SchedulerManagerName: func(get di.Get) interface{} {
			return schedulerManagerMock
		},
	})
	controller := NewIntervalController(dic)
	assert.NotNil(t, controller)
	tests := []struct {
		name               string
		request            []requests.AddIntervalRequest
		expectedStatusCode int
	}{
		{"Valid", []requests.AddIntervalRequest{valid}, http.StatusCreated},
		{"Valid - no request Id", []requests.AddIntervalRequest{noRequestId}, http.StatusCreated},
		{"Invalid - no name", []requests.AddIntervalRequest{noName}, http.StatusBadRequest},
		{"Invalid - duplicated name", []requests.AddIntervalRequest{duplicatedName}, http.StatusConflict},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiIntervalRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddInterval)
			handler.ServeHTTP(recorder, req)
			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Message is empty")
			} else {
				var res []commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}

func TestIntervalByName(t *testing.T) {
	interval := dtos.ToIntervalModel(addIntervalRequestData().Interval)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("IntervalByName", interval.Name).Return(interval, nil)
	dbClientMock.On("IntervalByName", notFoundName).Return(models.Interval{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "interval doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewIntervalController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		intervalName       string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - find interval by name", interval.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - interval not found by name", notFoundName, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiIntervalByNameRoute, testCase.intervalName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.intervalName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.IntervalByName)
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
				var res responseDTO.IntervalResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.intervalName, res.Interval.Name, "Name not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestAllIntervals(t *testing.T) {
	expectedTotalIntervalCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("IntervalTotalCount").Return(expectedTotalIntervalCount, nil)
	dbClientMock.On("AllIntervals", 0, 20).Return([]models.Interval{}, nil)
	dbClientMock.On("AllIntervals", 0, 1).Return([]models.Interval{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewIntervalController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get intervals without offset and limit", "", "", false, expectedTotalIntervalCount, http.StatusOK},
		{"Valid - get intervals with offset and limit", "0", "1", false, expectedTotalIntervalCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", true, expectedTotalIntervalCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", true, expectedTotalIntervalCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiAllIntervalRoute, http.NoBody)
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
			handler := http.HandlerFunc(controller.AllIntervals)
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
				var res responseDTO.MultiIntervalsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Response total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeleteIntervalByName(t *testing.T) {
	interval := dtos.ToIntervalModel(addIntervalRequestData().Interval)
	noName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	schedulerManagerMock := &dbMock.SchedulerManager{}
	dbClientMock.On("DeleteIntervalByName", interval.Name).Return(nil)
	dbClientMock.On("IntervalActionsByIntervalName", 0, 1, interval.Name).Return([]models.IntervalAction{}, nil)
	dbClientMock.On("DeleteIntervalByName", notFoundName).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "interval doesn't exist in the database", nil))
	dbClientMock.On("IntervalActionsByIntervalName", 0, 1, notFoundName).Return([]models.IntervalAction{}, nil)
	schedulerManagerMock.On("DeleteIntervalByName", interval.Name).Return(nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.SchedulerManagerName: func(get di.Get) interface{} {
			return schedulerManagerMock
		},
	})

	controller := NewIntervalController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		intervalName       string
		expectedStatusCode int
	}{
		{"Valid - interval by name", interval.Name, http.StatusOK},
		{"Invalid - name parameter is empty", noName, http.StatusBadRequest},
		{"Invalid - interval not found by name", notFoundName, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiIntervalByNameRoute, testCase.intervalName)
			req, err := http.NewRequest(http.MethodDelete, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.intervalName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteIntervalByName)
			handler.ServeHTTP(recorder, req)
			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
			if testCase.expectedStatusCode == http.StatusOK {
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			} else {
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}

func TestPatchInterval(t *testing.T) {
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	schedulerManagerMock := &dbMock.SchedulerManager{}
	testReq := updateIntervalRequestData()
	model := models.Interval{
		Id:       *testReq.Interval.Id,
		Name:     *testReq.Interval.Name,
		Interval: *testReq.Interval.Interval,
	}

	valid := testReq
	dbClientMock.On("IntervalById", *valid.Interval.Id).Return(model, nil)
	dbClientMock.On("IntervalActionsByIntervalName", 0, 1, *valid.Interval.Name).Return([]models.IntervalAction{}, nil)
	dbClientMock.On("UpdateInterval", model).Return(nil)
	schedulerManagerMock.On("UpdateInterval", model).Return(nil)
	validWithNoReqID := testReq
	validWithNoReqID.RequestId = ""
	validWithNoId := testReq
	validWithNoId.Interval.Id = nil
	dbClientMock.On("IntervalByName", *validWithNoId.Interval.Name).Return(model, nil)
	validWithNoName := testReq
	validWithNoName.Interval.Name = nil

	invalidId := testReq
	invalidUUID := "invalidUUID"
	invalidId.Interval.Id = &invalidUUID

	emptyString := ""
	emptyId := testReq
	emptyId.Interval.Id = &emptyString
	emptyId.Interval.Name = nil
	emptyName := testReq
	emptyName.Interval.Id = nil
	emptyName.Interval.Name = &emptyString

	invalidNoIdAndName := testReq
	invalidNoIdAndName.Interval.Id = nil
	invalidNoIdAndName.Interval.Name = nil

	invalidNotFoundId := testReq
	invalidNotFoundId.Interval.Name = nil
	notFoundId := "12345678-1111-1234-5678-de9dac3fb9bc"
	invalidNotFoundId.Interval.Id = &notFoundId
	notFoundIdError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("IntervalById", *invalidNotFoundId.Interval.Id).Return(model, notFoundIdError)

	invalidNotFoundName := testReq
	invalidNotFoundName.Interval.Id = nil
	notFoundName := "notFoundName"
	invalidNotFoundName.Interval.Name = &notFoundName
	notFoundNameError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundName), nil)
	dbClientMock.On("IntervalByName", *invalidNotFoundName.Interval.Name).Return(model, notFoundNameError)

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.SchedulerManagerName: func(get di.Get) interface{} {
			return schedulerManagerMock
		},
	})
	controller := NewIntervalController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name                 string
		request              []requests.UpdateIntervalRequest
		expectedStatusCode   int
		expectedResponseCode int
	}{
		{"Valid", []requests.UpdateIntervalRequest{valid}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no requestId", []requests.UpdateIntervalRequest{validWithNoReqID}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no id", []requests.UpdateIntervalRequest{validWithNoId}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no name", []requests.UpdateIntervalRequest{validWithNoName}, http.StatusMultiStatus, http.StatusOK},
		{"Invalid - invalid id", []requests.UpdateIntervalRequest{invalidId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty id", []requests.UpdateIntervalRequest{emptyId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty name", []requests.UpdateIntervalRequest{emptyName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - no id and name", []requests.UpdateIntervalRequest{invalidNoIdAndName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found id", []requests.UpdateIntervalRequest{invalidNotFoundId}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found name", []requests.UpdateIntervalRequest{invalidNotFoundName}, http.StatusMultiStatus, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPatch, common.ApiIntervalRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchInterval)
			handler.ServeHTTP(recorder, req)

			if testCase.expectedStatusCode == http.StatusMultiStatus {
				var res []commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedResponseCode, res[0].StatusCode, "BaseResponse status code not as expected")
				if testCase.expectedResponseCode == http.StatusOK {
					assert.Empty(t, res[0].Message, "Message should be empty when it is successful")
				} else {
					assert.NotEmpty(t, res[0].Message, "Response message doesn't contain the error message")
				}
			} else {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedResponseCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}

		})
	}

}
