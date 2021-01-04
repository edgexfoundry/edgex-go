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
	"github.com/edgexfoundry/go-mod-core-contracts/v2"
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

var testDeviceLabels = []string{"MODBUS", "TEMP"}

func buildTestDeviceRequest() requests.AddDeviceRequest {
	var testAutoEvents = []dtos.AutoEvent{
		{Resource: "TestResource", Frequency: "300ms", OnChange: true},
	}
	var testProtocols = map[string]dtos.ProtocolProperties{
		"modbus-ip": {
			"Address": "localhost",
			"Port":    "1502",
			"UnitID":  "1",
		},
	}
	var testAddDeviceReq = requests.AddDeviceRequest{
		BaseRequest: common.BaseRequest{
			RequestId: ExampleUUID,
		},
		Device: dtos.Device{
			Id:             ExampleUUID,
			Name:           TestDeviceName,
			ServiceName:    TestDeviceServiceName,
			ProfileName:    TestDeviceProfileName,
			AdminState:     models.Locked,
			OperatingState: models.Up,
			Labels:         testDeviceLabels,
			Location:       "{40lat;45long}",
			AutoEvents:     testAutoEvents,
			Protocols:      testProtocols,
		},
	}

	return testAddDeviceReq
}

func buildTestUpdateDeviceRequest() requests.UpdateDeviceRequest {
	testUUID := ExampleUUID
	testName := TestDeviceName
	testDescription := TestDescription
	testServiceName := TestDeviceServiceName
	testProfileName := TestDeviceProfileName
	testAdminState := models.Unlocked
	testOperatingState := models.Up
	testLastReported := int64(123546789)
	testLastConnected := int64(123546789)
	testNotify := false
	var testAutoEvents = []dtos.AutoEvent{
		{Resource: "TestResource", Frequency: "300ms", OnChange: true},
	}
	var testProtocols = map[string]dtos.ProtocolProperties{
		"modbus-ip": {
			"Address": "localhost",
			"Port":    "1502",
			"UnitID":  "1",
		},
	}
	var testUpdateDeviceReq = requests.UpdateDeviceRequest{
		BaseRequest: common.BaseRequest{
			RequestId: ExampleUUID,
		},
		Device: dtos.UpdateDevice{
			Id:             &testUUID,
			Name:           &testName,
			Description:    &testDescription,
			ServiceName:    &testServiceName,
			ProfileName:    &testProfileName,
			AdminState:     &testAdminState,
			OperatingState: &testOperatingState,
			LastReported:   &testLastReported,
			LastConnected:  &testLastConnected,
			Labels:         []string{"MODBUS", "TEMP"},
			Location:       "{40lat;45long}",
			AutoEvents:     testAutoEvents,
			Protocols:      testProtocols,
			Notify:         &testNotify,
		},
	}

	return testUpdateDeviceReq
}

func TestAddDevice(t *testing.T) {
	testDevice := buildTestDeviceRequest()
	deviceModel := requests.AddDeviceReqToDeviceModels([]requests.AddDeviceRequest{testDevice})[0]
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}

	valid := testDevice
	dbClientMock.On("DeviceServiceNameExists", deviceModel.ServiceName).Return(true, nil)
	dbClientMock.On("DeviceProfileNameExists", deviceModel.ProfileName).Return(true, nil)
	dbClientMock.On("AddDevice", deviceModel).Return(deviceModel, nil)
	dbClientMock.On("DeviceServiceByName", deviceModel.ServiceName).Return(models.DeviceService{BaseAddress: testBaseAddress}, nil)

	notFoundService := testDevice
	notFoundService.Device.ServiceName = "notFoundService"
	dbClientMock.On("DeviceServiceNameExists", notFoundService.Device.ServiceName).Return(false, nil)
	notFoundProfile := testDevice
	notFoundProfile.Device.ProfileName = "notFoundProfile"
	dbClientMock.On("DeviceProfileNameExists", notFoundProfile.Device.ProfileName).Return(false, nil)

	noName := testDevice
	noName.Device.Name = ""
	noAdminState := testDevice
	noAdminState.Device.AdminState = ""
	noOperatingState := testDevice
	noOperatingState.Device.OperatingState = ""
	invalidAdminState := testDevice
	invalidAdminState.Device.AdminState = "invalidAdminState"
	invalidOperatingState := testDevice
	invalidOperatingState.Device.OperatingState = "invalidOperatingState"
	noServiceName := testDevice
	noServiceName.Device.ServiceName = ""
	noProfileName := testDevice
	noProfileName.Device.ProfileName = ""
	noProtocols := testDevice
	noProtocols.Device.Protocols = nil
	emptyProtocols := testDevice
	emptyProtocols.Device.Protocols = map[string]dtos.ProtocolProperties{}

	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceController(dic)
	assert.NotNil(t, controller)
	tests := []struct {
		name               string
		request            []requests.AddDeviceRequest
		expectedStatusCode int
	}{
		{"Valid", []requests.AddDeviceRequest{valid}, http.StatusCreated},
		{"Invalid - not found service", []requests.AddDeviceRequest{notFoundService}, http.StatusNotFound},
		{"Invalid - not found profile", []requests.AddDeviceRequest{notFoundProfile}, http.StatusNotFound},
		{"Invalid - no name", []requests.AddDeviceRequest{noName}, http.StatusBadRequest},
		{"Invalid - no adminState", []requests.AddDeviceRequest{noAdminState}, http.StatusBadRequest},
		{"Invalid - no operatingState", []requests.AddDeviceRequest{noOperatingState}, http.StatusBadRequest},
		{"Invalid - invalid adminState", []requests.AddDeviceRequest{invalidAdminState}, http.StatusBadRequest},
		{"Invalid - invalid operatingState", []requests.AddDeviceRequest{invalidOperatingState}, http.StatusBadRequest},
		{"Invalid - no service name", []requests.AddDeviceRequest{noServiceName}, http.StatusBadRequest},
		{"Invalid - no profile name", []requests.AddDeviceRequest{noProfileName}, http.StatusBadRequest},
		{"Invalid - no protocols", []requests.AddDeviceRequest{noProtocols}, http.StatusBadRequest},
		{"Invalid - empty protocols", []requests.AddDeviceRequest{emptyProtocols}, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, v2.ApiDeviceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDevice)
			handler.ServeHTTP(recorder, req)
			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Message is empty")
			} else {
				var res []common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, v2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}

func TestDeleteDeviceById(t *testing.T) {
	device := dtos.ToDeviceModel(buildTestDeviceRequest().Device)
	noId := ""
	notFoundId := "82eb2e26-1111-2222-ae4c-de9dac3fb9bc"
	invalidId := "invalidId"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteDeviceById", device.Id).Return(nil)
	dbClientMock.On("DeleteDeviceById", notFoundId).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceId           string
		expectedStatusCode int
	}{
		{"Valid - delete device by id", device.Id, http.StatusOK},
		{"Invalid - id parameter is empty", noId, http.StatusBadRequest},
		{"Invalid - device not found by id", notFoundId, http.StatusNotFound},
		{"Invalid - invalid uuid", invalidId, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", v2.ApiDeviceByIdRoute, testCase.deviceId)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{v2.Id: testCase.deviceId})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteDeviceById)
			handler.ServeHTTP(recorder, req)
			var res common.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
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

func TestDeleteDeviceByName(t *testing.T) {
	device := dtos.ToDeviceModel(buildTestDeviceRequest().Device)
	noName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteDeviceByName", device.Name).Return(nil)
	dbClientMock.On("DeleteDeviceByName", notFoundName).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device doesn't exist in the database", nil))
	dbClientMock.On("DeviceByName", notFoundName).Return(device, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device doesn't exist in the database", nil))
	dbClientMock.On("DeviceByName", device.Name).Return(device, nil)
	dbClientMock.On("DeviceServiceByName", device.ServiceName).Return(models.DeviceService{BaseAddress: testBaseAddress}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceName         string
		expectedStatusCode int
	}{
		{"Valid - delete device by name", device.Name, http.StatusOK},
		{"Invalid - name parameter is empty", noName, http.StatusBadRequest},
		{"Invalid - device not found by name", notFoundName, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", v2.ApiDeviceByNameRoute, testCase.deviceName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.deviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteDeviceByName)
			handler.ServeHTTP(recorder, req)
			var res common.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
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

func TestAllDeviceByServiceName(t *testing.T) {
	device := dtos.ToDeviceModel(buildTestDeviceRequest().Device)
	testServiceA := "testServiceA"
	testServiceB := "testServiceB"
	device1WithServiceA := device
	device1WithServiceA.ServiceName = testServiceA
	device2WithServiceA := device
	device2WithServiceA.ServiceName = testServiceA
	device3WithServiceB := device
	device3WithServiceB.ServiceName = testServiceB

	devices := []models.Device{device1WithServiceA, device2WithServiceA, device3WithServiceB}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DevicesByServiceName", 0, 5, testServiceA).Return([]models.Device{devices[0], devices[1]}, nil)
	dbClientMock.On("DevicesByServiceName", 1, 1, testServiceA).Return([]models.Device{devices[1]}, nil)
	dbClientMock.On("DevicesByServiceName", 4, 1, testServiceB).Return([]models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		serviceName        string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get devices with serviceName", "0", "5", testServiceA, false, 2, http.StatusOK},
		{"Valid - get devices with offset and no labels", "1", "1", testServiceA, false, 1, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", testServiceB, true, 0, http.StatusNotFound},
		{"Invalid - get devices without serviceName", "0", "10", "", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiDeviceByServiceNameRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(v2.Offset, testCase.offset)
			query.Add(v2.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.serviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DevicesByServiceName)
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
				var res responseDTO.MultiDevicesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Devices), "Device count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeviceIdExists(t *testing.T) {
	testId := ExampleUUID
	notFoundId := "82eb2e26-1111-0000-ae4c-de9dac3fb9bc"
	emptyId := ""
	invalidId := "invalidId"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceIdExists", testId).Return(true, nil)
	dbClientMock.On("DeviceIdExists", notFoundId).Return(false, nil)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceId           string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - check device by id", testId, false, http.StatusOK},
		{"Invalid - id parameter is empty", emptyId, true, http.StatusBadRequest},
		{"Invalid - device not found by id", notFoundId, false, http.StatusNotFound},
		{"Invalid - invalid uuid", invalidId, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", v2.ApiDeviceIdExistsRoute, testCase.deviceId)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{v2.Id: testCase.deviceId})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceIdExists)
			handler.ServeHTTP(recorder, req)
			var res common.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
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

func TestDeviceNameExists(t *testing.T) {
	testName := TestDeviceName
	notFoundName := "notFoundName"
	emptyName := ""

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceNameExists", testName).Return(true, nil)
	dbClientMock.On("DeviceNameExists", notFoundName).Return(false, nil)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceName         string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - check device by name", testName, false, http.StatusOK},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - device not found by name", notFoundName, false, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", v2.ApiDeviceNameExistsRoute, testCase.deviceName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.deviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceNameExists)
			handler.ServeHTTP(recorder, req)
			var res common.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
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

func TestPatchDevice(t *testing.T) {
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	testReq := buildTestUpdateDeviceRequest()
	dsModels := models.Device{
		Id:             *testReq.Device.Id,
		Name:           *testReq.Device.Name,
		Description:    *testReq.Device.Description,
		Labels:         testReq.Device.Labels,
		AdminState:     models.AdminState(*testReq.Device.AdminState),
		OperatingState: models.OperatingState(*testReq.Device.OperatingState),
		LastConnected:  *testReq.Device.LastConnected,
		LastReported:   *testReq.Device.LastReported,
		Location:       testReq.Device.Location,
		ServiceName:    *testReq.Device.ServiceName,
		ProfileName:    *testReq.Device.ProfileName,
		AutoEvents:     dtos.ToAutoEventModels(testReq.Device.AutoEvents),
		Protocols:      dtos.ToProtocolModels(testReq.Device.Protocols),
		Notify:         *testReq.Device.Notify,
	}

	valid := testReq
	dbClientMock.On("DeviceServiceNameExists", *valid.Device.ServiceName).Return(true, nil)
	dbClientMock.On("DeviceProfileNameExists", *valid.Device.ProfileName).Return(true, nil)
	dbClientMock.On("DeviceById", *valid.Device.Id).Return(dsModels, nil)
	dbClientMock.On("DeleteDeviceById", *valid.Device.Id).Return(nil)
	dbClientMock.On("AddDevice", mock.Anything).Return(dsModels, nil)
	dbClientMock.On("DeviceServiceByName", *valid.Device.ServiceName).Return(models.DeviceService{BaseAddress: testBaseAddress}, nil)

	validWithNoReqID := testReq
	validWithNoReqID.RequestId = ""
	validWithNoId := testReq
	validWithNoId.Device.Id = nil
	dbClientMock.On("DeviceByName", *validWithNoId.Device.Name).Return(dsModels, nil)
	validWithNoName := testReq
	validWithNoName.Device.Name = nil

	invalidId := testReq
	invalidUUID := "invalidUUID"
	invalidId.Device.Id = &invalidUUID

	emptyString := ""
	emptyId := testReq
	emptyId.Device.Id = &emptyString
	emptyName := testReq
	emptyName.Device.Id = nil
	emptyName.Device.Name = &emptyString

	invalidNoIdAndName := testReq
	invalidNoIdAndName.Device.Id = nil
	invalidNoIdAndName.Device.Name = nil

	invalidNotFoundId := testReq
	invalidNotFoundId.Device.Name = nil
	notFoundId := "12345678-1111-1234-5678-de9dac3fb9bc"
	invalidNotFoundId.Device.Id = &notFoundId
	notFoundIdError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("DeviceById", *invalidNotFoundId.Device.Id).Return(dsModels, notFoundIdError)

	invalidNotFoundName := testReq
	invalidNotFoundName.Device.Id = nil
	notFoundName := "notFoundName"
	invalidNotFoundName.Device.Name = &notFoundName
	notFoundNameError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundName), nil)
	dbClientMock.On("DeviceByName", *invalidNotFoundName.Device.Name).Return(dsModels, notFoundNameError)

	notFountServiceName := "notFoundService"
	notFoundService := testReq
	notFoundService.Device.ServiceName = &notFountServiceName
	dbClientMock.On("DeviceServiceNameExists", *notFoundService.Device.ServiceName).Return(false, nil)
	notFountProfileName := "notFoundProfile"
	notFoundProfile := testReq
	notFoundProfile.Device.ProfileName = &notFountProfileName
	dbClientMock.On("DeviceProfileNameExists", *notFoundProfile.Device.ProfileName).Return(false, nil)

	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name                 string
		request              []requests.UpdateDeviceRequest
		expectedStatusCode   int
		expectedResponseCode int
	}{
		{"Valid", []requests.UpdateDeviceRequest{valid}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no requestId", []requests.UpdateDeviceRequest{validWithNoReqID}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no id", []requests.UpdateDeviceRequest{validWithNoId}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no name", []requests.UpdateDeviceRequest{validWithNoName}, http.StatusMultiStatus, http.StatusOK},
		{"Invalid - invalid id", []requests.UpdateDeviceRequest{invalidId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty id", []requests.UpdateDeviceRequest{emptyId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty name", []requests.UpdateDeviceRequest{emptyName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found id", []requests.UpdateDeviceRequest{invalidNotFoundId}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found name", []requests.UpdateDeviceRequest{invalidNotFoundName}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - no id and name", []requests.UpdateDeviceRequest{invalidNoIdAndName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found service", []requests.UpdateDeviceRequest{notFoundService}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found profile", []requests.UpdateDeviceRequest{notFoundProfile}, http.StatusMultiStatus, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPatch, v2.ApiDeviceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchDevice)
			handler.ServeHTTP(recorder, req)

			if testCase.expectedStatusCode == http.StatusMultiStatus {
				var res []common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, v2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
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
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedResponseCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}

		})
	}

}

func TestAllDevices(t *testing.T) {
	device := dtos.ToDeviceModel(buildTestDeviceRequest().Device)
	devices := []models.Device{device, device, device}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AllDevices", 0, 10, []string(nil)).Return(devices, nil)
	dbClientMock.On("AllDevices", 0, 5, testDeviceLabels).Return([]models.Device{devices[0], devices[1]}, nil)
	dbClientMock.On("AllDevices", 1, 2, []string(nil)).Return([]models.Device{devices[1], devices[2]}, nil)
	dbClientMock.On("AllDevices", 4, 1, testDeviceLabels).Return([]models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		labels             string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get devices without labels", "0", "10", "", false, 3, http.StatusOK},
		{"Valid - get devices with labels", "0", "5", strings.Join(testDeviceLabels, ","), false, 2, http.StatusOK},
		{"Valid - get devices with offset and no labels", "1", "2", "", false, 2, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", strings.Join(testDeviceLabels, ","), true, 0, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiAllDeviceRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(v2.Offset, testCase.offset)
			query.Add(v2.Limit, testCase.limit)
			if len(testCase.labels) > 0 {
				query.Add(v2.Labels, testCase.labels)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllDevices)
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
				var res responseDTO.MultiDevicesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Devices), "Device count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeviceByName(t *testing.T) {
	device := dtos.ToDeviceModel(buildTestDeviceRequest().Device)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceByName", device.Name).Return(device, nil)
	dbClientMock.On("DeviceByName", notFoundName).Return(models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceName         string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - find device by name", device.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - device not found by name", notFoundName, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", v2.ApiDeviceByNameRoute, testCase.deviceName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.deviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceByName)
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
				var res responseDTO.DeviceResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.deviceName, res.Device.Name, "Name not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDevicesByProfileName(t *testing.T) {
	device := dtos.ToDeviceModel(buildTestDeviceRequest().Device)
	testProfileA := "testProfileA"
	testProfileB := "testServiceB"
	device1WithProfileA := device
	device1WithProfileA.ProfileName = testProfileA
	device2WithProfileA := device
	device2WithProfileA.ProfileName = testProfileA
	device3WithProfileB := device
	device3WithProfileB.ProfileName = testProfileB

	devices := []models.Device{device1WithProfileA, device2WithProfileA, device3WithProfileB}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DevicesByProfileName", 0, 5, testProfileA).Return([]models.Device{devices[0], devices[1]}, nil)
	dbClientMock.On("DevicesByProfileName", 1, 1, testProfileA).Return([]models.Device{devices[1]}, nil)
	dbClientMock.On("DevicesByProfileName", 4, 1, testProfileB).Return([]models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		profileName        string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get devices with profileName", "0", "5", testProfileA, false, 2, http.StatusOK},
		{"Valid - get devices with offset and limit", "1", "1", testProfileA, false, 1, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", testProfileB, true, 0, http.StatusNotFound},
		{"Invalid - get devices without profileName", "0", "10", "", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiDeviceByProfileNameRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(v2.Offset, testCase.offset)
			query.Add(v2.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.profileName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DevicesByProfileName)
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
				var res responseDTO.MultiDevicesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Devices), "Device count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
