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
	"github.com/stretchr/testify/require"
)

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
	var testAddDeviceServiceReq = requests.AddDeviceRequest{
		BaseRequest: common.BaseRequest{
			RequestId: ExampleUUID,
		},
		Device: dtos.Device{
			Id:             ExampleUUID,
			Name:           TestDeviceName,
			ServiceName:    TestDeviceServiceName,
			ProfileName:    TestDeviceProfileName,
			AdminState:     models.Locked,
			OperatingState: models.Enabled,
			Labels:         []string{"MODBUS", "TEMP"},
			Location:       "{40lat;45long}",
			AutoEvents:     testAutoEvents,
			Protocols:      testProtocols,
		},
	}

	return testAddDeviceServiceReq
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
	dbClientMock.On("AllDeviceByServiceName", 0, 5, testServiceA).Return([]models.Device{devices[0], devices[1]}, nil)
	dbClientMock.On("AllDeviceByServiceName", 1, 1, testServiceA).Return([]models.Device{devices[1]}, nil)
	dbClientMock.On("AllDeviceByServiceName", 4, 1, testServiceB).Return([]models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
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
			req, err := http.NewRequest(http.MethodGet, v2.ApiAllDeviceByServiceNameRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(v2.Offset, testCase.offset)
			query.Add(v2.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.serviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllDeviceByServiceName)
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
