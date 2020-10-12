//
// Copyright (C) 2020 IOTech Ltd
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

	v2MetadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testDeviceServiceName = "TestDeviceService"
var testDeviceServiceLabels = []string{"hvac", "thermostat"}
var testBaseAddress = "http://home-device-service:49990"

func buildTestDeviceServiceRequest() requests.AddDeviceServiceRequest {
	var testAddDeviceServiceReq = requests.AddDeviceServiceRequest{
		BaseRequest: common.BaseRequest{
			RequestID: ExampleUUID,
		},
		Service: dtos.DeviceService{
			Name:           testDeviceServiceName,
			Description:    TestDescription,
			Labels:         testDeviceServiceLabels,
			AdminState:     models.Unlocked,
			OperatingState: models.Enabled,
			BaseAddress:    testBaseAddress,
		},
	}

	return testAddDeviceServiceReq
}

func buildTestUpdateDeviceServiceRequest() requests.UpdateDeviceServiceRequest {
	testUUID := ExampleUUID
	testAdminState := models.Unlocked
	testOperatingState := models.Enabled
	var testUpdateDeviceServiceReq = requests.UpdateDeviceServiceRequest{
		BaseRequest: common.BaseRequest{
			RequestID: ExampleUUID,
		},
		Service: dtos.UpdateDeviceService{
			Id:             &testUUID,
			Name:           &testDeviceServiceName,
			Labels:         testDeviceServiceLabels,
			AdminState:     &testAdminState,
			OperatingState: &testOperatingState,
			BaseAddress:    &testBaseAddress,
		},
	}

	return testUpdateDeviceServiceReq
}

func buildTestDBClient(dsModel models.DeviceService, errKind errors.ErrKind, errorMessage string) *dbMock.DBClient {
	dbClientMock := &dbMock.DBClient{}
	if len(errKind) > 0 {
		err := errors.NewCommonEdgeX(errKind, errorMessage, nil)
		dbClientMock.On("AddDeviceService", dsModel).Return(dsModel, err)
	} else {
		dbClientMock.On("AddDeviceService", dsModel).Return(dsModel, nil)
	}
	return dbClientMock
}

func TestAddDeviceService(t *testing.T) {
	validReq := buildTestDeviceServiceRequest()
	dsModels := requests.AddDeviceServiceReqToDeviceServiceModels([]requests.AddDeviceServiceRequest{validReq})
	expectedRequestId := ExampleUUID
	normalMessage := fmt.Sprintf("Add device service %s successfully", testDeviceServiceName)
	duplicateServiceNameMessage := fmt.Sprintf("device service %s already exists", testDeviceServiceName)

	reqWithNoID := validReq
	reqWithNoID.RequestID = ""
	reqWithInvalidId := validReq
	reqWithInvalidId.RequestID = "InvalidUUID"
	reqWithNoName := validReq
	reqWithNoName.Service.Name = ""

	tests := []struct {
		name                   string
		isValidRequest         bool
		dbClientMock           *dbMock.DBClient
		Request                []requests.AddDeviceServiceRequest
		expectedHttpStatusCode int
		expectedMessage        string
	}{
		{
			"Request Normal",
			true,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{validReq},
			http.StatusCreated,
			normalMessage,
		},
		{
			"Request without requestId",
			true,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{reqWithNoID},
			http.StatusCreated,
			normalMessage,
		},
		{
			"Request with duplicate service name",
			true,
			buildTestDBClient(dsModels[0], errors.KindDuplicateName, duplicateServiceNameMessage),
			[]requests.AddDeviceServiceRequest{validReq},
			http.StatusConflict,
			duplicateServiceNameMessage,
		},
		{
			"Request with invalid requestId",
			false,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{reqWithInvalidId},
			http.StatusBadRequest,
			"",
		},
		{
			"Request without service name",
			false,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{reqWithNoName},
			http.StatusBadRequest,
			"",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {

			dic := mockDic()
			dic.Update(di.ServiceConstructorMap{
				v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
					return testCase.dbClientMock
				},
			})

			controller := NewDeviceServiceController(dic)
			require.NotNil(t, controller)

			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiDeviceServiceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceService)
			handler.ServeHTTP(recorder, req)
			if testCase.isValidRequest {
				var res []common.BaseWithIdResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)

				// Assert
				require.NoError(t, err)
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, contractsV2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestID != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestID, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedHttpStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
				assert.Contains(t, res[0].Message, testCase.expectedMessage, "Message not as expected")
			} else {
				assert.Equal(t, testCase.expectedHttpStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.NotEmpty(t, string(recorder.Body.Bytes()), "Message is empty")
			}
		})
	}
}

func TestGetDeviceServiceByName(t *testing.T) {
	deviceService := dtos.ToDeviceServiceModel(buildTestDeviceServiceRequest().Service)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("GetDeviceServiceByName", deviceService.Name).Return(deviceService, nil)
	dbClientMock.On("GetDeviceServiceByName", notFoundName).Return(models.DeviceService{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device service doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceServiceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceServiceName  string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - find device service by name", deviceService.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - device service not found by name", notFoundName, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s/%s", contractsV2.ApiDeviceProfileRoute, contractsV2.Name, testCase.deviceServiceName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{contractsV2.Name: testCase.deviceServiceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.GetDeviceServiceByName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.DeviceServiceResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.deviceServiceName, res.Service.Name, "Name not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestPatchDeviceService(t *testing.T) {
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	testReq := buildTestUpdateDeviceServiceRequest()
	dsModels := models.DeviceService{
		Id:             *testReq.Service.Id,
		Name:           *testReq.Service.Name,
		Labels:         testReq.Service.Labels,
		AdminState:     models.AdminState(*testReq.Service.AdminState),
		OperatingState: models.OperatingState(*testReq.Service.OperatingState),
		BaseAddress:    *testReq.Service.BaseAddress,
	}

	valid := testReq
	dbClientMock.On("GetDeviceServiceById", *valid.Service.Id).Return(dsModels, nil)
	dbClientMock.On("DeleteDeviceServiceById", *valid.Service.Id).Return(nil)
	dbClientMock.On("AddDeviceService", mock.Anything).Return(dsModels, nil)
	validWithNoReqID := testReq
	validWithNoReqID.RequestID = ""
	validWithNoId := testReq
	validWithNoId.Service.Id = nil
	dbClientMock.On("GetDeviceServiceByName", *validWithNoId.Service.Name).Return(dsModels, nil)
	validWithNoName := testReq
	validWithNoName.Service.Name = nil
	validNotFoundId := testReq
	notFoundId := "12345678-0000-1234-5678-de9dac3fb9bc"
	validNotFoundId.Service.Id = &notFoundId
	notFoundIdError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("GetDeviceServiceById", *validNotFoundId.Service.Id).Return(dsModels, notFoundIdError)

	invalidNoIdAndName := testReq
	invalidNoIdAndName.Service.Id = nil
	invalidNoIdAndName.Service.Name = nil

	invalidNotFoundId := testReq
	invalidNotFoundId.Service.Name = nil
	notFoundId = "12345678-1111-1234-5678-de9dac3fb9bc"
	invalidNotFoundId.Service.Id = &notFoundId
	notFoundIdError = errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("GetDeviceServiceById", *invalidNotFoundId.Service.Id).Return(dsModels, notFoundIdError)

	invalidNotFoundName := testReq
	invalidNotFoundName.Service.Id = nil
	notFoundName := "notFoundName"
	invalidNotFoundName.Service.Name = &notFoundName
	notFoundNameError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundName), nil)
	dbClientMock.On("GetDeviceServiceByName", *invalidNotFoundName.Service.Name).Return(dsModels, notFoundNameError)

	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceServiceController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name               string
		request            []requests.UpdateDeviceServiceRequest
		expectedStatusCode int
	}{
		{"Valid", []requests.UpdateDeviceServiceRequest{valid}, http.StatusOK},
		{"Valid - no requestId", []requests.UpdateDeviceServiceRequest{validWithNoReqID}, http.StatusOK},
		{"Valid - no id", []requests.UpdateDeviceServiceRequest{validWithNoId}, http.StatusOK},
		{"Valid - no name", []requests.UpdateDeviceServiceRequest{validWithNoName}, http.StatusOK},
		{"Valid - not found id", []requests.UpdateDeviceServiceRequest{validNotFoundId}, http.StatusOK},
		{"Invalid - no id and name", []requests.UpdateDeviceServiceRequest{invalidNoIdAndName}, http.StatusBadRequest},
		{"Invalid - not found id", []requests.UpdateDeviceServiceRequest{invalidNotFoundId}, http.StatusNotFound},
		{"Invalid - not found name", []requests.UpdateDeviceServiceRequest{invalidNotFoundName}, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiDeviceServiceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchDeviceService)
			handler.ServeHTTP(recorder, req)

			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Message is empty")
			} else {
				var res []common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, contractsV2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestID != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestID, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}

		})
	}
}
