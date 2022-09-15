//
// Copyright (C) 2020-2021 IOTech Ltd
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

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces/mocks"

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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testDeviceServiceName = "TestDeviceService"
var testDeviceServiceLabels = []string{"hvac", "thermostat"}
var testBaseAddress = "http://home-device-service:49990"

func buildTestDeviceServiceRequest() requests.AddDeviceServiceRequest {
	var testAddDeviceServiceReq = requests.AddDeviceServiceRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		Service: dtos.DeviceService{
			Id:          ExampleUUID,
			Name:        testDeviceServiceName,
			Description: TestDescription,
			Labels:      testDeviceServiceLabels,
			AdminState:  models.Unlocked,
			BaseAddress: testBaseAddress,
		},
	}

	return testAddDeviceServiceReq
}

func buildTestUpdateDeviceServiceRequest() requests.UpdateDeviceServiceRequest {
	testUUID := ExampleUUID
	testAdminState := models.Unlocked
	var testUpdateDeviceServiceReq = requests.UpdateDeviceServiceRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		Service: dtos.UpdateDeviceService{
			Id:          &testUUID,
			Name:        &testDeviceServiceName,
			Labels:      testDeviceServiceLabels,
			AdminState:  &testAdminState,
			BaseAddress: &testBaseAddress,
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

	reqWithNoID := validReq
	reqWithNoID.RequestId = ""
	reqWithInvalidId := validReq
	reqWithInvalidId.RequestId = "InvalidUUID"
	reqWithNoName := validReq
	reqWithNoName.Service.Name = ""

	tests := []struct {
		name                   string
		isValidRequest         bool
		dbClientMock           *dbMock.DBClient
		Request                []requests.AddDeviceServiceRequest
		expectedHttpStatusCode int
	}{
		{
			"Request Normal",
			true,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{validReq},
			http.StatusCreated,
		},
		{
			"Request without requestId",
			true,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{reqWithNoID},
			http.StatusCreated,
		},
		{
			"Request with duplicate service name",
			true,
			buildTestDBClient(dsModels[0], errors.KindDuplicateName, "duplicate service name error"),
			[]requests.AddDeviceServiceRequest{validReq},
			http.StatusConflict,
		},
		{
			"Request with invalid requestId",
			false,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{reqWithInvalidId},
			http.StatusBadRequest,
		},
		{
			"Request without service name",
			false,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{reqWithNoName},
			http.StatusBadRequest,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {

			dic := mockDic()
			dic.Update(di.ServiceConstructorMap{
				container.DBClientInterfaceName: func(get di.Get) interface{} {
					return testCase.dbClientMock
				},
			})

			controller := NewDeviceServiceController(dic)
			require.NotNil(t, controller)

			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceServiceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceService)
			handler.ServeHTTP(recorder, req)
			if testCase.isValidRequest {
				var res []commonDTO.BaseWithIdResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)

				// Assert
				require.NoError(t, err)
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedHttpStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
				if testCase.expectedHttpStatusCode == http.StatusCreated {
					assert.Empty(t, res[0].Message, "Message should be empty when it is successful")
				} else {
					assert.NotEmpty(t, res[0].Message, "Response message doesn't contain the error message")
				}
			} else {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedHttpStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedHttpStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}

func TestDeviceServiceByName(t *testing.T) {
	deviceService := dtos.ToDeviceServiceModel(buildTestDeviceServiceRequest().Service)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceServiceByName", deviceService.Name).Return(deviceService, nil)
	dbClientMock.On("DeviceServiceByName", notFoundName).Return(models.DeviceService{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device service doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			reqPath := fmt.Sprintf("%s/%s", common.ApiDeviceServiceByNameRoute, testCase.deviceServiceName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceServiceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceServiceByName)
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
				var res responseDTO.DeviceServiceResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
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
		Id:          *testReq.Service.Id,
		Name:        *testReq.Service.Name,
		Labels:      testReq.Service.Labels,
		AdminState:  models.AdminState(*testReq.Service.AdminState),
		BaseAddress: *testReq.Service.BaseAddress,
	}

	valid := testReq
	dbClientMock.On("DeviceServiceById", *valid.Service.Id).Return(dsModels, nil)
	dbClientMock.On("UpdateDeviceService", mock.Anything).Return(nil)
	validWithNoReqID := testReq
	validWithNoReqID.RequestId = ""
	validWithNoId := testReq
	validWithNoId.Service.Id = nil
	dbClientMock.On("DeviceServiceByName", *validWithNoId.Service.Name).Return(dsModels, nil)
	validWithNoName := testReq
	validWithNoName.Service.Name = nil

	invalidId := testReq
	invalidUUID := "invalidUUID"
	invalidId.Service.Id = &invalidUUID

	emptyString := ""
	emptyId := testReq
	emptyId.Service.Id = &emptyString
	emptyId.Service.Name = nil
	emptyName := testReq
	emptyName.Service.Id = nil
	emptyName.Service.Name = &emptyString

	invalidNoIdAndName := testReq
	invalidNoIdAndName.Service.Id = nil
	invalidNoIdAndName.Service.Name = nil

	invalidNotFoundId := testReq
	invalidNotFoundId.Service.Name = nil
	notFoundId := "12345678-1111-1234-5678-de9dac3fb9bc"
	invalidNotFoundId.Service.Id = &notFoundId
	notFoundIdError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("DeviceServiceById", *invalidNotFoundId.Service.Id).Return(dsModels, notFoundIdError)

	invalidNotFoundName := testReq
	invalidNotFoundName.Service.Id = nil
	notFoundName := "notFoundName"
	invalidNotFoundName.Service.Name = &notFoundName
	notFoundNameError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundName), nil)
	dbClientMock.On("DeviceServiceByName", *invalidNotFoundName.Service.Name).Return(dsModels, notFoundNameError)

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceServiceController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name                 string
		request              []requests.UpdateDeviceServiceRequest
		expectedStatusCode   int
		expectedResponseCode int
	}{
		{"Valid", []requests.UpdateDeviceServiceRequest{valid}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no requestId", []requests.UpdateDeviceServiceRequest{validWithNoReqID}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no id", []requests.UpdateDeviceServiceRequest{validWithNoId}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no name", []requests.UpdateDeviceServiceRequest{validWithNoName}, http.StatusMultiStatus, http.StatusOK},
		{"Invalid - invalid id", []requests.UpdateDeviceServiceRequest{invalidId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty id", []requests.UpdateDeviceServiceRequest{emptyId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty name", []requests.UpdateDeviceServiceRequest{emptyName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - no id and name", []requests.UpdateDeviceServiceRequest{invalidNoIdAndName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found id", []requests.UpdateDeviceServiceRequest{invalidNotFoundId}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found name", []requests.UpdateDeviceServiceRequest{invalidNotFoundName}, http.StatusMultiStatus, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceServiceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchDeviceService)
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

func TestAllDeviceServices(t *testing.T) {
	deviceServices := []models.DeviceService{
		{
			Name:   "ds1",
			Labels: testDeviceServiceLabels,
		},
		{
			Name:   "ds2",
			Labels: testDeviceServiceLabels,
		},
		{
			Name: "ds3",
		},
	}
	expectedTotalDeviceServiceCount := uint32(len(deviceServices))

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceServiceCountByLabels", []string(nil)).Return(expectedTotalDeviceServiceCount, nil)
	dbClientMock.On("DeviceServiceCountByLabels", testDeviceServiceLabels).Return(expectedTotalDeviceServiceCount, nil)
	dbClientMock.On("AllDeviceServices", 0, 10, []string(nil)).Return(deviceServices, nil)
	dbClientMock.On("AllDeviceServices", 0, 5, testDeviceServiceLabels).Return([]models.DeviceService{deviceServices[0], deviceServices[1]}, nil)
	dbClientMock.On("AllDeviceServices", 1, 2, []string(nil)).Return([]models.DeviceService{deviceServices[1], deviceServices[2]}, nil)
	dbClientMock.On("AllDeviceServices", 4, 1, testDeviceServiceLabels).Return([]models.DeviceService{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceServiceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		labels             string
		errorExpected      bool
		expectedCount      int
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get device services without labels", "0", "10", "", false, 3, expectedTotalDeviceServiceCount, http.StatusOK},
		{"Valid - get device services with labels", "0", "5", strings.Join(testDeviceServiceLabels, ","), false, 2, expectedTotalDeviceServiceCount, http.StatusOK},
		{"Valid - get device services with offset and no labels", "1", "2", "", false, 2, expectedTotalDeviceServiceCount, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", strings.Join(testDeviceServiceLabels, ","), true, 0, expectedTotalDeviceServiceCount, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiAllDeviceServiceRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			if len(testCase.labels) > 0 {
				query.Add(common.Labels, testCase.labels)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllDeviceServices)
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
				var res responseDTO.MultiDeviceServicesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Services), "Service count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeleteDeviceServiceByName(t *testing.T) {
	deviceService := dtos.ToDeviceServiceModel(buildTestDeviceServiceRequest().Service)
	noName := ""
	notFoundName := "notFoundName"
	deviceExists := "deviceExists"
	provisionWatcherExists := "provisionWatcherExists"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DevicesByServiceName", 0, 1, deviceService.Name).Return([]models.Device{}, nil)
	dbClientMock.On("ProvisionWatchersByServiceName", 0, 1, deviceService.Name).Return([]models.ProvisionWatcher{}, nil)
	dbClientMock.On("DeleteDeviceServiceByName", deviceService.Name).Return(nil)
	dbClientMock.On("DevicesByServiceName", 0, 1, notFoundName).Return([]models.Device{}, nil)
	dbClientMock.On("ProvisionWatchersByServiceName", 0, 1, notFoundName).Return([]models.ProvisionWatcher{}, nil)
	dbClientMock.On("DeleteDeviceServiceByName", notFoundName).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device service doesn't exist in the database", nil))
	dbClientMock.On("DeleteDeviceServiceByName", deviceExists).Return(errors.NewCommonEdgeX(
		errors.KindStatusConflict, "fail to delete the device service when associated device exists", nil))
	dbClientMock.On("DeleteDeviceServiceByName", provisionWatcherExists).Return(errors.NewCommonEdgeX(
		errors.KindStatusConflict, "fail to delete the device service when associated provisionWatcher exists", nil))
	dbClientMock.On("ProvisionWatchersByServiceName", 0, 1, provisionWatcherExists).Return([]models.ProvisionWatcher{models.ProvisionWatcher{}}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceServiceController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceServiceName  string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - delete device service by name", deviceService.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", noName, true, http.StatusBadRequest},
		{"Invalid - device service not found by name", notFoundName, true, http.StatusNotFound},
		{"Invalid - associated device exists", deviceExists, true, http.StatusConflict},
		{"Invalid - associated provisionWatcher Exists", provisionWatcherExists, true, http.StatusConflict},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiDeviceServiceByNameRoute, testCase.deviceServiceName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceServiceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteDeviceServiceByName)
			handler.ServeHTTP(recorder, req)
			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
			if testCase.errorExpected {
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
