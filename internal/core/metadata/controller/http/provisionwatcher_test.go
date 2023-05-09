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

	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces/mocks"
)

var testProvisionWatcherName = "TestProvisionWatcher"
var testProvisionWatcherLabels = []string{"test", "temp"}
var testProvisionWatcherIdentifiers = map[string]string{
	"address": "localhost",
	"port":    "3[0-9]{2}",
}
var testProvisionWatcherBlockingIdentifiers = map[string][]string{
	"port": {"397", "398", "399"},
}
var testProvisionWatcherAutoEvents = []dtos.AutoEvent{
	{SourceName: "TestResource", Interval: "300ms", OnChange: true},
}

func buildTestAddProvisionWatcherRequest() requests.AddProvisionWatcherRequest {
	return requests.AddProvisionWatcherRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		ProvisionWatcher: dtos.ProvisionWatcher{
			Id:                  ExampleUUID,
			Name:                testProvisionWatcherName,
			ServiceName:         TestDeviceServiceName,
			Labels:              testProvisionWatcherLabels,
			Identifiers:         testProvisionWatcherIdentifiers,
			BlockingIdentifiers: testProvisionWatcherBlockingIdentifiers,
			AdminState:          models.Unlocked,
			DiscoveredDevice: dtos.DiscoveredDevice{
				ProfileName: TestDeviceProfileName,
				AdminState:  models.Unlocked,
				AutoEvents:  testProvisionWatcherAutoEvents,
				Properties:  testProperties,
			},
		},
	}
}

func buildTestUpdateProvisionWatcherRequest() requests.UpdateProvisionWatcherRequest {
	testUUID := ExampleUUID
	testName := testProvisionWatcherName
	testServiceName := TestDeviceServiceName
	testProfileName := TestDeviceProfileName
	testAdminState := models.Unlocked

	var testUpdateProvisionWatcherReq = requests.UpdateProvisionWatcherRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		ProvisionWatcher: dtos.UpdateProvisionWatcher{
			Id:                  &testUUID,
			Name:                &testName,
			ServiceName:         &testServiceName,
			Labels:              testProvisionWatcherLabels,
			Identifiers:         testProvisionWatcherIdentifiers,
			BlockingIdentifiers: testProvisionWatcherBlockingIdentifiers,
			AdminState:          &testAdminState,
			DiscoveredDevice: dtos.UpdateDiscoveredDevice{
				ProfileName: &testProfileName,
				AdminState:  &testAdminState,
				AutoEvents:  testProvisionWatcherAutoEvents,
				Properties:  testProperties,
			},
		},
	}

	return testUpdateProvisionWatcherReq
}

func TestProvisionWatcherController_AddProvisionWatcher_Created(t *testing.T) {
	validReq := buildTestAddProvisionWatcherRequest()
	pwModel := requests.AddProvisionWatcherReqToProvisionWatcherModels([]requests.AddProvisionWatcherRequest{validReq})[0]
	expectedRequestId := ExampleUUID

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddProvisionWatcher", pwModel).Return(pwModel, nil)
	dbClientMock.On("DeviceServiceByName", pwModel.ServiceName).Return(models.DeviceService{}, nil)
	dbClientMock.On("DeviceServiceNameExists", pwModel.ServiceName).Return(true, nil)
	dbClientMock.On("DeviceProfileNameExists", pwModel.DiscoveredDevice.ProfileName).Return(true, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	reqWithNoID := validReq
	reqWithNoID.RequestId = ""

	tests := []struct {
		name    string
		Request []requests.AddProvisionWatcherRequest
	}{
		{"Valid - AddProvisionWatcherRequest", []requests.AddProvisionWatcherRequest{validReq}},
		{"Valid - no RequestId", []requests.AddProvisionWatcherRequest{reqWithNoID}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddProvisionWatcher)
			handler.ServeHTTP(recorder, req)
			var res []commonDTO.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API version not as expected")
			if res[0].RequestId != "" {
				assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
			}
			assert.Equal(t, http.StatusCreated, res[0].StatusCode, "BaseResonse status code not as expected")
			assert.Empty(t, res[0].Message, "Message should be empty when it's successful")
		})
	}
}

func TestProvisionWatcherController_AddProvisionWatcher_BadRequest(t *testing.T) {
	dic := mockDic()
	dbClientMock := &mocks.DBClient{}

	provisionWatcher := buildTestAddProvisionWatcherRequest()
	badRequestId := provisionWatcher
	badRequestId.RequestId = "niv3sl"
	noName := provisionWatcher
	noName.ProvisionWatcher.Name = ""

	notFountServiceName := "notFoundService"
	notFoundService := provisionWatcher
	notFoundService.ProvisionWatcher.ServiceName = notFountServiceName
	dbClientMock.On("DeviceServiceNameExists", notFoundService.ProvisionWatcher.ServiceName).Return(false, nil)
	notFountProfileName := "notFoundProfile"
	notFoundProfile := provisionWatcher
	notFoundProfile.ProvisionWatcher.DiscoveredDevice.ProfileName = notFountProfileName
	notFoundServiceProvisionWatcherModel := requests.AddProvisionWatcherReqToProvisionWatcherModels(
		[]requests.AddProvisionWatcherRequest{notFoundService})[0]
	dbClientMock.On("AddProvisionWatcher", notFoundServiceProvisionWatcherModel).Return(
		notFoundServiceProvisionWatcherModel,
		errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device service '%s' does not exists",
			notFountServiceName), nil))
	notFoundProfileProvisionWatcherModel := requests.AddProvisionWatcherReqToProvisionWatcherModels(
		[]requests.AddProvisionWatcherRequest{notFoundProfile})[0]
	dbClientMock.On("AddProvisionWatcher", notFoundProfileProvisionWatcherModel).Return(
		notFoundProfileProvisionWatcherModel,
		errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile '%s' does not exists",
			notFountProfileName), nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name                 string
		Request              []requests.AddProvisionWatcherRequest
		expectedStatusCode   int
		expectedResponseCode int
	}{
		{"Invalid - Bad requestId", []requests.AddProvisionWatcherRequest{badRequestId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - Bad name", []requests.AddProvisionWatcherRequest{noName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found service", []requests.AddProvisionWatcherRequest{notFoundService}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found profile", []requests.AddProvisionWatcherRequest{notFoundProfile}, http.StatusMultiStatus, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddProvisionWatcher)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.expectedStatusCode == http.StatusMultiStatus {
				var res []commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedResponseCode, res[0].StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
			} else {
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
			}
		})
	}
}

func TestProvisionWatcherController_AddProvisionWatcher_Duplicated(t *testing.T) {
	expectedRequestId := ExampleUUID

	duplicateIdRequest := buildTestAddProvisionWatcherRequest()
	duplicateIdModel := requests.AddProvisionWatcherReqToProvisionWatcherModels([]requests.AddProvisionWatcherRequest{duplicateIdRequest})[0]
	duplicateIdDBError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("provision watcher id %s already exists", duplicateIdModel.Id), nil)

	duplicateNameRequest := buildTestAddProvisionWatcherRequest()
	duplicateNameRequest.ProvisionWatcher.Id = "" // The infrastructure layer will generate id when the id field is empty
	duplicateNameModel := requests.AddProvisionWatcherReqToProvisionWatcherModels([]requests.AddProvisionWatcherRequest{duplicateNameRequest})[0]
	duplicateNameDBError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("provision watcher name %s already exists", duplicateNameModel.Name), nil)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddProvisionWatcher", duplicateNameModel).Return(duplicateNameModel, duplicateNameDBError)
	dbClientMock.On("AddProvisionWatcher", duplicateIdModel).Return(duplicateIdModel, duplicateIdDBError)
	dbClientMock.On("DeviceServiceNameExists", duplicateIdRequest.ProvisionWatcher.ServiceName).Return(true, nil)
	dbClientMock.On("DeviceProfileNameExists", duplicateIdRequest.ProvisionWatcher.DiscoveredDevice.ProfileName).Return(true, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name          string
		request       []requests.AddProvisionWatcherRequest
		expectedError errors.CommonEdgeX
	}{
		{"duplicate id", []requests.AddProvisionWatcherRequest{duplicateIdRequest}, duplicateIdDBError},
		{"duplicate name", []requests.AddProvisionWatcherRequest{duplicateNameRequest}, duplicateNameDBError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddProvisionWatcher)
			handler.ServeHTTP(recorder, req)
			var res []commonDTO.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
			assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
			assert.Equal(t, http.StatusConflict, res[0].StatusCode, "BaseResponse status code not as expected")
			assert.Contains(t, res[0].Message, testCase.expectedError.Message(), "Message not as expected")
		})
	}
}

func TestProvisionWatcherController_ProvisionWatcherByName(t *testing.T) {
	provisionWatcher := dtos.ToProvisionWatcherModel(buildTestAddProvisionWatcherRequest().ProvisionWatcher)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ProvisionWatcherByName", provisionWatcher.Name).Return(provisionWatcher, nil)
	dbClientMock.On("ProvisionWatcherByName", notFoundName).Return(models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "provision watcher doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name                 string
		provisionWatcherName string
		errorExpected        bool
		expectedStatusCode   int
	}{
		{"Valid - find provision watcher by name", provisionWatcher.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - provision watcher not found by name", notFoundName, true, http.StatusNotFound},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiProvisionWatcherByNameRoute, testCase.provisionWatcherName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.provisionWatcherName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ProvisionWatcherByName)
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
				var res responseDTO.ProvisionWatcherResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.provisionWatcherName, res.ProvisionWatcher.Name, "Name not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestProvisionWatcherController_ProvisionWatchersByServiceName(t *testing.T) {
	provisionWatcher := dtos.ToProvisionWatcherModel(buildTestAddProvisionWatcherRequest().ProvisionWatcher)
	testServiceA := "testServiceA"
	testServiceB := "testServiceB"
	pw1WithServiceA := provisionWatcher
	pw1WithServiceA.ServiceName = testServiceA
	pw2WithServiceA := provisionWatcher
	pw2WithServiceA.ServiceName = testServiceA
	pw3WithServiceB := provisionWatcher
	pw3WithServiceB.ServiceName = testServiceB

	provisionWatchers := []models.ProvisionWatcher{pw1WithServiceA, pw2WithServiceA, pw3WithServiceB}
	expectedTotalCountServiceA := uint32(2)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ProvisionWatcherCountByServiceName", testServiceA).Return(expectedTotalCountServiceA, nil)
	dbClientMock.On("ProvisionWatchersByServiceName", 0, 5, testServiceA).Return([]models.ProvisionWatcher{provisionWatchers[0], provisionWatchers[1]}, nil)
	dbClientMock.On("ProvisionWatchersByServiceName", 1, 1, testServiceA).Return([]models.ProvisionWatcher{provisionWatchers[1]}, nil)
	dbClientMock.On("ProvisionWatchersByServiceName", 4, 1, testServiceB).Return([]models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		serviceName        string
		errorExpected      bool
		expectedCount      int
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get provision watchers with serviceName", "0", "5", testServiceA, false, 2, expectedTotalCountServiceA, http.StatusOK},
		{"Valid - get provision watchers with offset and limit", "1", "1", testServiceA, false, 1, expectedTotalCountServiceA, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", testServiceB, true, 0, expectedTotalCountServiceA, http.StatusNotFound},
		{"Invalid - get provision watchers without serviceName", "0", "10", "", true, 0, expectedTotalCountServiceA, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiProvisionWatcherByServiceNameRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.serviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ProvisionWatchersByServiceName)
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
				var res responseDTO.MultiProvisionWatchersResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.ProvisionWatchers), "ProvisionWatcher count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestProvisionWatcherController_ProvisionWatchersByProfileName(t *testing.T) {
	provisionWatcher := dtos.ToProvisionWatcherModel(buildTestAddProvisionWatcherRequest().ProvisionWatcher)
	testProfileA := "testProfileA"
	testProfileB := "testProfileB"
	pw1WithProfileA := provisionWatcher
	pw1WithProfileA.DiscoveredDevice.ProfileName = testProfileA
	pw2WithProfileA := provisionWatcher
	pw2WithProfileA.DiscoveredDevice.ProfileName = testProfileA
	pw3WithProfileB := provisionWatcher
	pw3WithProfileB.DiscoveredDevice.ProfileName = testProfileB

	provisionWatchers := []models.ProvisionWatcher{pw1WithProfileA, pw2WithProfileA, pw3WithProfileB}
	expectedTotalPWCountProfileA := uint32(2)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ProvisionWatcherCountByProfileName", testProfileA).Return(expectedTotalPWCountProfileA, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 0, 5, testProfileA).Return([]models.ProvisionWatcher{provisionWatchers[0], provisionWatchers[1]}, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 0, 5, testProfileA).Return([]models.ProvisionWatcher{provisionWatchers[0], provisionWatchers[1]}, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 1, 1, testProfileA).Return([]models.ProvisionWatcher{provisionWatchers[1]}, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 4, 1, testProfileB).Return([]models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		profileName        string
		errorExpected      bool
		expectedCount      int
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get provision watchers with profileName", "0", "5", testProfileA, false, 2, expectedTotalPWCountProfileA, http.StatusOK},
		{"Valid - get provision watchers with offset and limit", "1", "1", testProfileA, false, 1, expectedTotalPWCountProfileA, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", testProfileB, true, 0, expectedTotalPWCountProfileA, http.StatusNotFound},
		{"Invalid - get provision watchers without profileName", "0", "10", "", true, 0, expectedTotalPWCountProfileA, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiProvisionWatcherByProfileNameRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.profileName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ProvisionWatchersByProfileName)
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
				var res responseDTO.MultiProvisionWatchersResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.ProvisionWatchers), "ProvisionWatcher count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestProvisionWatcherController_AllProvisionWatchers(t *testing.T) {
	provisionWatcher := dtos.ToProvisionWatcherModel(buildTestAddProvisionWatcherRequest().ProvisionWatcher)
	provisionWatchers := []models.ProvisionWatcher{provisionWatcher, provisionWatcher, provisionWatcher}
	expectedTotalPWCount := uint32(len(provisionWatchers))

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ProvisionWatcherCountByLabels", []string(nil)).Return(expectedTotalPWCount, nil)
	dbClientMock.On("ProvisionWatcherCountByLabels", testProvisionWatcherLabels).Return(expectedTotalPWCount, nil)
	dbClientMock.On("AllProvisionWatchers", 0, 10, []string(nil)).Return(provisionWatchers, nil)
	dbClientMock.On("AllProvisionWatchers", 0, 5, testProvisionWatcherLabels).Return([]models.ProvisionWatcher{provisionWatchers[0], provisionWatchers[1]}, nil)
	dbClientMock.On("AllProvisionWatchers", 1, 2, []string(nil)).Return([]models.ProvisionWatcher{provisionWatchers[1], provisionWatchers[2]}, nil)
	dbClientMock.On("AllProvisionWatchers", 4, 1, testProvisionWatcherLabels).Return([]models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewProvisionWatcherController(dic)
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
		{"Valid - get provision watchers without labels", "0", "10", "", false, 3, expectedTotalPWCount, http.StatusOK},
		{"Valid - get provision watchers with labels", "0", "5", strings.Join(testProvisionWatcherLabels, ","), false, 2, expectedTotalPWCount, http.StatusOK},
		{"Valid - get provision watchers with offset and no labels", "1", "2", "", false, 2, expectedTotalPWCount, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", strings.Join(testProvisionWatcherLabels, ","), true, 0, expectedTotalPWCount, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiAllProvisionWatcherRoute, http.NoBody)
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
			handler := http.HandlerFunc(controller.AllProvisionWatchers)
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
				var res responseDTO.MultiProvisionWatchersResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.ProvisionWatchers), "ProvisionWatcher count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestProvisionWatcherController_DeleteProvisionWatcherByName(t *testing.T) {
	provisionWatcher := dtos.ToProvisionWatcherModel(buildTestAddProvisionWatcherRequest().ProvisionWatcher)
	noName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ProvisionWatcherByName", provisionWatcher.Name).Return(provisionWatcher, nil)
	dbClientMock.On("ProvisionWatcherByName", notFoundName).Return(provisionWatcher, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "provision watcher doesn't exist in the database", nil))
	dbClientMock.On("DeleteProvisionWatcherByName", provisionWatcher.Name).Return(nil)
	dbClientMock.On("DeviceServiceByName", provisionWatcher.ServiceName).Return(models.DeviceService{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewProvisionWatcherController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name                 string
		provisionWatcherName string
		expectedStatusCode   int
	}{
		{"Valid - delete provision watcher by name", provisionWatcher.Name, http.StatusOK},
		{"Invalid - name parameter is empty", noName, http.StatusBadRequest},
		{"Invalid - provision watcher not found by name", notFoundName, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiProvisionWatcherByNameRoute, testCase.provisionWatcherName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.provisionWatcherName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteProvisionWatcherByName)
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

func TestProvisionWatcherController_PatchProvisionWatcher(t *testing.T) {
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	testReq := buildTestUpdateProvisionWatcherRequest()
	pwModels := models.ProvisionWatcher{
		Id:                  *testReq.ProvisionWatcher.Id,
		Name:                *testReq.ProvisionWatcher.Name,
		ServiceName:         *testReq.ProvisionWatcher.ServiceName,
		Labels:              testReq.ProvisionWatcher.Labels,
		Identifiers:         testReq.ProvisionWatcher.Identifiers,
		BlockingIdentifiers: testReq.ProvisionWatcher.BlockingIdentifiers,
		AdminState:          models.AdminState(*testReq.ProvisionWatcher.AdminState),
		DiscoveredDevice: models.DiscoveredDevice{
			ProfileName: *testReq.ProvisionWatcher.DiscoveredDevice.ProfileName,
			AutoEvents:  dtos.ToAutoEventModels(testReq.ProvisionWatcher.DiscoveredDevice.AutoEvents),
			Properties:  testProperties,
			AdminState:  models.AdminState(*testReq.ProvisionWatcher.AdminState),
		},
	}

	valid := testReq
	dbClientMock.On("DeviceServiceNameExists", *valid.ProvisionWatcher.ServiceName).Return(true, nil)
	dbClientMock.On("DeviceProfileNameExists", *valid.ProvisionWatcher.DiscoveredDevice.ProfileName).Return(true, nil)
	dbClientMock.On("ProvisionWatcherByName", *valid.ProvisionWatcher.Name).Return(pwModels, nil)
	dbClientMock.On("UpdateProvisionWatcher", pwModels).Return(nil)
	dbClientMock.On("DeviceServiceByName", *valid.ProvisionWatcher.ServiceName).Return(models.DeviceService{}, nil)
	validWithNoReqID := testReq
	validWithNoReqID.RequestId = ""
	validWithNoId := testReq
	validWithNoId.ProvisionWatcher.Id = nil
	dbClientMock.On("ProvisionWatcherByName", *validWithNoId.ProvisionWatcher.Name).Return(pwModels, nil)
	validWithNoName := testReq
	validWithNoName.ProvisionWatcher.Name = nil
	dbClientMock.On("ProvisionWatcherById", *validWithNoName.ProvisionWatcher.Id).Return(pwModels, nil)

	invalidId := testReq
	invalidUUID := "invalidUUID"
	invalidId.ProvisionWatcher.Id = &invalidUUID

	emptyString := ""
	emptyId := testReq
	emptyId.ProvisionWatcher.Id = &emptyString
	emptyId.ProvisionWatcher.Name = nil
	emptyName := testReq
	emptyName.ProvisionWatcher.Id = nil
	emptyName.ProvisionWatcher.Name = &emptyString

	invalidNoIdAndName := testReq
	invalidNoIdAndName.ProvisionWatcher.Id = nil
	invalidNoIdAndName.ProvisionWatcher.Name = nil

	invalidNotFoundId := testReq
	invalidNotFoundId.ProvisionWatcher.Name = nil
	notFoundId := "12345678-1111-1234-5678-de9dac3fb9bc"
	invalidNotFoundId.ProvisionWatcher.Id = &notFoundId
	notFoundIdError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("ProvisionWatcherById", *invalidNotFoundId.ProvisionWatcher.Id).Return(pwModels, notFoundIdError)

	invalidNotFoundName := testReq
	notFoundName := "notFoundName"
	invalidNotFoundName.ProvisionWatcher.Name = &notFoundName
	invalidNotFoundName.ProvisionWatcher.Id = nil
	notFoundNameError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundName), nil)
	dbClientMock.On("ProvisionWatcherByName", *invalidNotFoundName.ProvisionWatcher.Name).Return(pwModels, notFoundNameError)

	notFountServiceName := "notFoundService"
	notFoundService := testReq
	notFoundService.ProvisionWatcher.Id = nil
	notFoundService.ProvisionWatcher.Name = &notFountServiceName
	notFoundServiceProvisionWatcherModel := pwModels
	notFoundServiceProvisionWatcherModel.Name = notFountServiceName
	dbClientMock.On("ProvisionWatcherByName", notFountServiceName).Return(notFoundServiceProvisionWatcherModel, nil)
	dbClientMock.On("UpdateProvisionWatcher", notFoundServiceProvisionWatcherModel).Return(
		errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device service '%s' does not exists",
			notFountServiceName), nil))

	notFountProfileName := "notFoundProfile"
	notFoundProfile := testReq
	notFoundProfile.ProvisionWatcher.Id = nil
	notFoundProfile.ProvisionWatcher.Name = &notFountProfileName
	notFoundProfileProvisionWatcherModel := pwModels
	notFoundProfileProvisionWatcherModel.Name = notFountProfileName
	dbClientMock.On("ProvisionWatcherByName", notFountProfileName).Return(notFoundProfileProvisionWatcherModel, nil)
	dbClientMock.On("UpdateProvisionWatcher", notFoundProfileProvisionWatcherModel).Return(
		errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile '%s' does not exists",
			notFountProfileName), nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewProvisionWatcherController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name                 string
		request              []requests.UpdateProvisionWatcherRequest
		expectedStatusCode   int
		expectedResponseCode int
	}{
		{"Valid", []requests.UpdateProvisionWatcherRequest{valid}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no requestId", []requests.UpdateProvisionWatcherRequest{validWithNoReqID}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no id", []requests.UpdateProvisionWatcherRequest{validWithNoId}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no name", []requests.UpdateProvisionWatcherRequest{validWithNoName}, http.StatusMultiStatus, http.StatusOK},
		{"Invalid - invalid id", []requests.UpdateProvisionWatcherRequest{invalidId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty id", []requests.UpdateProvisionWatcherRequest{emptyId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty name", []requests.UpdateProvisionWatcherRequest{emptyName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found id", []requests.UpdateProvisionWatcherRequest{invalidNotFoundId}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found name", []requests.UpdateProvisionWatcherRequest{invalidNotFoundName}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - no id and name", []requests.UpdateProvisionWatcherRequest{invalidNoIdAndName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found service", []requests.UpdateProvisionWatcherRequest{notFoundService}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found profile", []requests.UpdateProvisionWatcherRequest{notFoundProfile}, http.StatusMultiStatus, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchProvisionWatcher)
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
