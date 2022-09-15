//
// Copyright (C) 2020-2022 IOTech Ltd
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
	"github.com/stretchr/testify/require"
)

var testDeviceLabels = []string{"MODBUS", "TEMP"}

func buildTestDeviceRequest() requests.AddDeviceRequest {
	var testAutoEvents = []dtos.AutoEvent{
		{SourceName: "TestResource", Interval: "300ms", OnChange: true},
	}
	var testProtocols = map[string]dtos.ProtocolProperties{
		"modbus-ip": {
			"Address": "localhost",
			"Port":    "1502",
			"UnitID":  "1",
		},
	}
	var testAddDeviceReq = requests.AddDeviceRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
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
		{SourceName: "TestResource", Interval: "300ms", OnChange: true},
	}
	var testProtocols = map[string]dtos.ProtocolProperties{
		"modbus-ip": {
			"Address": "localhost",
			"Port":    "1502",
			"UnitID":  "1",
		},
	}
	var testUpdateDeviceReq = requests.UpdateDeviceRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
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

func mockValidationHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var res commonDTO.BaseResponse
	var device requests.AddDeviceRequest
	err := json.NewDecoder(r.Body).Decode(&device)
	if err != nil {
		http.Error(w, "json decoding failed", http.StatusBadRequest)
		return
	}

	if _, ok := device.Device.Protocols["modbus-ip"]; !ok {
		w.WriteHeader(http.StatusInternalServerError)
		res = commonDTO.NewBaseResponse("", "validation failed", http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		res = commonDTO.NewBaseResponse("", "", http.StatusOK)
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "json encoding failed", http.StatusInternalServerError)
	}
	w.Write(resBytes) //nolint:errcheck
}

func TestAddDevice(t *testing.T) {
	mockDeviceServiceServer := httptest.NewServer(http.HandlerFunc(mockValidationHandler))
	defer mockDeviceServiceServer.Close()

	testDevice := buildTestDeviceRequest()
	deviceModel := requests.AddDeviceReqToDeviceModels([]requests.AddDeviceRequest{testDevice})[0]
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}

	valid := testDevice
	dbClientMock.On("DeviceServiceNameExists", deviceModel.ServiceName).Return(true, nil)
	dbClientMock.On("DeviceProfileNameExists", deviceModel.ProfileName).Return(true, nil)
	dbClientMock.On("AddDevice", deviceModel).Return(deviceModel, nil)
	dbClientMock.On("DeviceServiceByName", deviceModel.ServiceName).Return(models.DeviceService{BaseAddress: mockDeviceServiceServer.URL}, nil)
	dbClientMock.On("DeviceServiceByName", "unavailable").Return(models.DeviceService{BaseAddress: "http://unavailable"}, nil)

	notFoundService := testDevice
	notFoundService.Device.ServiceName = "notFoundService"
	dbClientMock.On("DeviceServiceByName", notFoundService.Device.ServiceName).Return(models.DeviceService{},
		errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device service '%s' does not exists",
			notFoundService.Device.ServiceName), nil))

	notFoundProfile := testDevice
	notFoundProfile.Device.ProfileName = "notFoundProfile"
	notFoundProfileDeviceModel := requests.AddDeviceReqToDeviceModels([]requests.AddDeviceRequest{notFoundProfile})[0]
	dbClientMock.On("AddDevice", notFoundProfileDeviceModel).Return(notFoundProfileDeviceModel,
		errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile '%s' does not exists",
			notFoundProfile.Device.ProfileName), nil))

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
	invalidProtocols := testDevice
	invalidProtocols.Device.Protocols = map[string]dtos.ProtocolProperties{"others": {}}
	serviceUnavailable := testDevice
	testDevice.Device.ServiceName = "unavailable"

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		{"Invalid - invalid protocols", []requests.AddDeviceRequest{invalidProtocols}, http.StatusInternalServerError},
		{"Valid - device service unavailable", []requests.AddDeviceRequest{serviceUnavailable}, http.StatusCreated},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDevice)
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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			reqPath := fmt.Sprintf("%s/%s", common.ApiDeviceByNameRoute, testCase.deviceName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteDeviceByName)
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
	expectedTotalCountServiceA := uint32(2)

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceCountByServiceName", testServiceA).Return(expectedTotalCountServiceA, nil)
	dbClientMock.On("DevicesByServiceName", 0, 5, testServiceA).Return([]models.Device{devices[0], devices[1]}, nil)
	dbClientMock.On("DevicesByServiceName", 1, 1, testServiceA).Return([]models.Device{devices[1]}, nil)
	dbClientMock.On("DevicesByServiceName", 4, 1, testServiceB).Return([]models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get devices with serviceName", "0", "5", testServiceA, false, 2, expectedTotalCountServiceA, http.StatusOK},
		{"Valid - get devices with offset and no labels", "1", "1", testServiceA, false, 1, expectedTotalCountServiceA, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", testServiceB, true, 0, 0, http.StatusNotFound},
		{"Invalid - get devices without serviceName", "0", "10", "", true, 0, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiDeviceByServiceNameRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.serviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DevicesByServiceName)
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
				var res responseDTO.MultiDevicesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Devices), "Device count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			reqPath := fmt.Sprintf("%s/%s", common.ApiDeviceNameExistsRoute, testCase.deviceName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceNameExists)
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

func TestPatchDevice(t *testing.T) {
	mockDeviceServiceServer := httptest.NewServer(http.HandlerFunc(mockValidationHandler))
	defer mockDeviceServiceServer.Close()

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
	dbClientMock.On("UpdateDevice", dsModels).Return(nil)
	dbClientMock.On("DeviceServiceByName", *valid.Device.ServiceName).Return(models.DeviceService{BaseAddress: mockDeviceServiceServer.URL}, nil)

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
	emptyId.Device.Name = nil
	emptyName := testReq
	emptyName.Device.Id = nil
	emptyName.Device.Name = &emptyString

	invalidProtocols := testReq
	invalidProtocols.Device.Protocols = map[string]dtos.ProtocolProperties{"others": {}}

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
	dbClientMock.On("DeviceServiceByName", *notFoundService.Device.ServiceName).Return(models.DeviceService{},
		errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device service '%s' does not exists",
			*notFoundService.Device.ServiceName), nil))

	notFountProfileName := "notFoundProfile"
	notFoundProfile := testReq
	notFoundProfile.Device.ProfileName = &notFountProfileName
	notFoundProfileDeviceModel := dsModels
	notFoundProfileDeviceModel.ProfileName = notFountProfileName
	dbClientMock.On("DeviceById", *notFoundProfile.Device.Id).Return(notFoundProfileDeviceModel, nil)
	dbClientMock.On("UpdateDevice", notFoundProfileDeviceModel).Return(
		errors.NewCommonEdgeX(errors.KindEntityDoesNotExist,
			fmt.Sprintf("device profile '%s' does not exists", notFountProfileName), nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		{"Invalid - invalid protocols", []requests.UpdateDeviceRequest{invalidProtocols}, http.StatusMultiStatus, http.StatusInternalServerError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPatch, common.ApiDeviceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchDevice)
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

func TestAllDevices(t *testing.T) {
	device := dtos.ToDeviceModel(buildTestDeviceRequest().Device)
	devices := []models.Device{device, device, device}
	expectedDeviceTotalCount := uint32(len(devices))

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceCountByLabels", []string(nil)).Return(expectedDeviceTotalCount, nil)
	dbClientMock.On("DeviceCountByLabels", testDeviceLabels).Return(expectedDeviceTotalCount, nil)
	dbClientMock.On("AllDevices", 0, 10, []string(nil)).Return(devices, nil)
	dbClientMock.On("AllDevices", 0, 5, testDeviceLabels).Return([]models.Device{devices[0], devices[1]}, nil)
	dbClientMock.On("AllDevices", 1, 2, []string(nil)).Return([]models.Device{devices[1], devices[2]}, nil)
	dbClientMock.On("AllDevices", 4, 1, testDeviceLabels).Return([]models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get devices without labels", "0", "10", "", false, 3, expectedDeviceTotalCount, http.StatusOK},
		{"Valid - get devices with labels", "0", "5", strings.Join(testDeviceLabels, ","), false, 2, expectedDeviceTotalCount, http.StatusOK},
		{"Valid - get devices with offset and no labels", "1", "2", "", false, 2, expectedDeviceTotalCount, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", strings.Join(testDeviceLabels, ","), true, 0, expectedDeviceTotalCount, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiAllDeviceRoute, http.NoBody)
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
			handler := http.HandlerFunc(controller.AllDevices)
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
				var res responseDTO.MultiDevicesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Devices), "Device count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			reqPath := fmt.Sprintf("%s/%s", common.ApiDeviceByNameRoute, testCase.deviceName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceByName)
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
				var res responseDTO.DeviceResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
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
	expectedTotalCountProfileA := uint32(2)

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceCountByProfileName", testProfileA).Return(expectedTotalCountProfileA, nil)
	dbClientMock.On("DevicesByProfileName", 0, 5, testProfileA).Return([]models.Device{devices[0], devices[1]}, nil)
	dbClientMock.On("DevicesByProfileName", 1, 1, testProfileA).Return([]models.Device{devices[1]}, nil)
	dbClientMock.On("DevicesByProfileName", 4, 1, testProfileB).Return([]models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get devices with profileName", "0", "5", testProfileA, false, 2, expectedTotalCountProfileA, http.StatusOK},
		{"Valid - get devices with offset and limit", "1", "1", testProfileA, false, 1, expectedTotalCountProfileA, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", testProfileB, true, 0, 0, http.StatusNotFound},
		{"Invalid - get devices without profileName", "0", "10", "", true, 0, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiDeviceByProfileNameRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.profileName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DevicesByProfileName)
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
				var res responseDTO.MultiDevicesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Devices), "Device count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
