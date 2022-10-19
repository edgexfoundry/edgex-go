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

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
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

func addIntervalActionRequestData() requests.AddIntervalActionRequest {
	restAddress := dtos.NewRESTAddress(TestHost, TestPort, TestHTTPMethod)
	intervalAction := dtos.NewIntervalAction(TestIntervalActionName, TestIntervalName, restAddress)
	return requests.NewAddIntervalActionRequest(intervalAction)
}

func updateIntervalActionRequestData() requests.UpdateIntervalActionRequest {
	testUUID := ExampleUUID
	testIntervalActionName := TestIntervalActionName
	testIntervalName := TestIntervalName
	restAddress := dtos.NewRESTAddress(TestHost, TestPort, TestHTTPMethod)
	var req = requests.UpdateIntervalActionRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		Action: dtos.UpdateIntervalAction{
			Id:           &testUUID,
			Name:         &testIntervalActionName,
			IntervalName: &testIntervalName,
			Address:      &restAddress,
		},
	}

	return req
}

func TestAddIntervalAction(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	schedulerManagerMock := &dbMock.SchedulerManager{}

	valid := addIntervalActionRequestData()
	model := dtos.ToIntervalActionModel(valid.Action)
	dbClientMock.On("IntervalByName", model.IntervalName).Return(models.Interval{}, nil)
	dbClientMock.On("AddIntervalAction", model).Return(model, nil)
	schedulerManagerMock.On("AddIntervalAction", model).Return(nil)

	noName := valid
	noName.Action.Name = ""
	noRequestId := valid
	noRequestId.RequestId = ""

	duplicatedName := valid
	duplicatedName.Action.Name = "duplicatedName"
	model = dtos.ToIntervalActionModel(duplicatedName.Action)
	dbClientMock.On("AddIntervalAction", model).Return(model, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("intervalAction name %s already exists", model.Name), nil))

	invalidIntervalNotFound := valid
	intervalNotFoundName := "intervalNotFoundName"
	invalidIntervalNotFound.Action.IntervalName = intervalNotFoundName
	model = dtos.ToIntervalActionModel(invalidIntervalNotFound.Action)
	dbClientMock.On("AddIntervalAction", model).Return(model, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("interval '%s' does not exists", model.IntervalName), nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.SchedulerManagerName: func(get di.Get) interface{} {
			return schedulerManagerMock
		},
	})
	controller := NewIntervalActionController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name               string
		request            []requests.AddIntervalActionRequest
		expectedStatusCode int
	}{
		{"Valid", []requests.AddIntervalActionRequest{valid}, http.StatusCreated},
		{"Valid - no request Id", []requests.AddIntervalActionRequest{noRequestId}, http.StatusCreated},
		{"Invalid - no name", []requests.AddIntervalActionRequest{noName}, http.StatusBadRequest},
		{"Invalid - duplicated name", []requests.AddIntervalActionRequest{duplicatedName}, http.StatusConflict},
		{"Invalid - interval not found", []requests.AddIntervalActionRequest{invalidIntervalNotFound}, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiIntervalActionRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddIntervalAction)
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
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}

func TestAllIntervalActions(t *testing.T) {
	expectedTotalIntervalActionCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("IntervalActionTotalCount").Return(expectedTotalIntervalActionCount, nil)
	dbClientMock.On("AllIntervalActions", 0, 20).Return([]models.IntervalAction{}, nil)
	dbClientMock.On("AllIntervalActions", 0, 1).Return([]models.IntervalAction{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewIntervalActionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get intervalActions without offset and limit", "", "", false, expectedTotalIntervalActionCount, http.StatusOK},
		{"Valid - get intervalActions with offset and limit", "0", "1", false, expectedTotalIntervalActionCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", true, expectedTotalIntervalActionCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", true, expectedTotalIntervalActionCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiAllIntervalActionRoute, http.NoBody)
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
			handler := http.HandlerFunc(controller.AllIntervalActions)
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

func TestIntervalActionByName(t *testing.T) {
	action := dtos.ToIntervalActionModel(addIntervalActionRequestData().Action)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("IntervalActionByName", action.Name).Return(action, nil)
	dbClientMock.On("IntervalActionByName", notFoundName).Return(models.IntervalAction{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "intervalAction doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewIntervalActionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		actionName         string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - find intervalAction by name", action.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - intervalAction not found by name", notFoundName, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiIntervalActionByNameRoute, testCase.actionName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.actionName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.IntervalActionByName)
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
				var res responseDTO.IntervalActionResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.actionName, res.Action.Name, "Name not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeleteIntervalActionByName(t *testing.T) {
	action := dtos.ToIntervalActionModel(addIntervalActionRequestData().Action)
	noName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	schedulerManagerMock := &dbMock.SchedulerManager{}
	dbClientMock.On("DeleteIntervalActionByName", action.Name).Return(nil)
	schedulerManagerMock.On("DeleteIntervalActionByName", action.Name).Return(nil)
	dbClientMock.On("DeleteIntervalActionByName", notFoundName).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "intervalAction doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.SchedulerManagerName: func(get di.Get) interface{} {
			return schedulerManagerMock
		},
	})

	controller := NewIntervalActionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		actionName         string
		expectedStatusCode int
	}{
		{"Valid - intervalAction by name", action.Name, http.StatusOK},
		{"Invalid - name parameter is empty", noName, http.StatusBadRequest},
		{"Invalid - intervalAction not found by name", notFoundName, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiIntervalActionByNameRoute, testCase.actionName)
			req, err := http.NewRequest(http.MethodDelete, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.actionName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteIntervalActionByName)
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

func TestPatchIntervalAction(t *testing.T) {
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	schedulerManagerMock := &dbMock.SchedulerManager{}
	testReq := updateIntervalActionRequestData()
	model := models.IntervalAction{
		Id:           *testReq.Action.Id,
		Name:         *testReq.Action.Name,
		IntervalName: *testReq.Action.IntervalName,
		Address:      dtos.ToAddressModel(*testReq.Action.Address),
	}

	valid := testReq
	dbClientMock.On("IntervalActionById", *valid.Action.Id).Return(model, nil)
	dbClientMock.On("IntervalByName", *valid.Action.IntervalName).Return(models.Interval{}, nil)
	dbClientMock.On("UpdateIntervalAction", model).Return(nil)
	schedulerManagerMock.On("UpdateIntervalAction", model).Return(nil)
	validWithNoReqID := testReq
	validWithNoReqID.RequestId = ""
	validWithNoId := testReq
	validWithNoId.Action.Id = nil
	dbClientMock.On("IntervalActionByName", *validWithNoId.Action.Name).Return(model, nil)
	validWithNoName := testReq
	validWithNoName.Action.Name = nil

	invalidId := testReq
	invalidUUID := "invalidUUID"
	invalidId.Action.Id = &invalidUUID

	emptyString := ""
	emptyId := testReq
	emptyId.Action.Id = &emptyString
	emptyId.Action.Name = nil
	emptyName := testReq
	emptyName.Action.Id = nil
	emptyName.Action.Name = &emptyString

	invalidNoIdAndName := testReq
	invalidNoIdAndName.Action.Id = nil
	invalidNoIdAndName.Action.Name = nil

	invalidNotFoundId := testReq
	invalidNotFoundId.Action.Name = nil
	notFoundId := "12345678-1111-1234-5678-de9dac3fb9bc"
	invalidNotFoundId.Action.Id = &notFoundId
	notFoundIdError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("IntervalActionById", *invalidNotFoundId.Action.Id).Return(model, notFoundIdError)

	invalidNotFoundName := testReq
	invalidNotFoundName.Action.Id = nil
	notFoundName := "notFoundName"
	invalidNotFoundName.Action.Name = &notFoundName
	notFoundNameError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundName), nil)
	dbClientMock.On("IntervalActionByName", *invalidNotFoundName.Action.Name).Return(model, notFoundNameError)

	intervalNotFoundName := "intervalNotFoundName"
	intervalNotFoundNameError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", intervalNotFoundName), nil)
	dbClientMock.On("IntervalByName", intervalNotFoundName).Return(models.Interval{}, intervalNotFoundNameError)
	invalidIntervalNotFound := testReq
	invalidIntervalNotFound.Action.IntervalName = &intervalNotFoundName
	invalidIntervalNotFoundModel := model
	invalidIntervalNotFoundModel.IntervalName = intervalNotFoundName
	invalidIntervalNotFound.Action.Id = nil
	dbClientMock.On("IntervalActionByName", intervalNotFoundName).Return(invalidIntervalNotFoundModel, nil)
	dbClientMock.On("UpdateIntervalAction", invalidIntervalNotFoundModel).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", intervalNotFoundName), nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.SchedulerManagerName: func(get di.Get) interface{} {
			return schedulerManagerMock
		},
	})
	controller := NewIntervalActionController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name                 string
		request              []requests.UpdateIntervalActionRequest
		expectedStatusCode   int
		expectedResponseCode int
	}{
		{"Valid", []requests.UpdateIntervalActionRequest{valid}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no requestId", []requests.UpdateIntervalActionRequest{validWithNoReqID}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no id", []requests.UpdateIntervalActionRequest{validWithNoId}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no name", []requests.UpdateIntervalActionRequest{validWithNoName}, http.StatusMultiStatus, http.StatusOK},
		{"Invalid - invalid id", []requests.UpdateIntervalActionRequest{invalidId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty id", []requests.UpdateIntervalActionRequest{emptyId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty name", []requests.UpdateIntervalActionRequest{emptyName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - no id and name", []requests.UpdateIntervalActionRequest{invalidNoIdAndName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found id", []requests.UpdateIntervalActionRequest{invalidNotFoundId}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found name", []requests.UpdateIntervalActionRequest{invalidNotFoundName}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found interval name", []requests.UpdateIntervalActionRequest{invalidIntervalNotFound}, http.StatusMultiStatus, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPatch, common.ApiIntervalActionRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchIntervalAction)
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
