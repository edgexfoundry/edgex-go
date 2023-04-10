//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	messagingMocks "github.com/edgexfoundry/go-mod-messaging/v3/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces/mocks"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDeviceLabels = []string{"MODBUS", "TEMP"}

var testProperties = map[string]any{
	"TestProperty1": "property1",
	"TestProperty2": true,
	"TestProperty3": 123.45,
}

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
			Properties:     testProperties,
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
			Labels:         []string{"MODBUS", "TEMP"},
			Location:       "{40lat;45long}",
			AutoEvents:     testAutoEvents,
			Protocols:      testProtocols,
			Properties:     testProperties,
		},
	}

	return testUpdateDeviceReq
}

func TestAddDevice(t *testing.T) {
	testDevice := buildTestDeviceRequest()
	deviceModel := dtos.ToDeviceModel(testDevice.Device)
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}

	valid := testDevice
	dbClientMock.On("DeviceServiceNameExists", deviceModel.ServiceName).Return(true, nil)
	dbClientMock.On("AddDevice", deviceModel).Return(deviceModel, nil)

	notFoundProfile := testDevice
	notFoundProfile.Device.ProfileName = "notFoundProfile"
	notFoundProfileDeviceModel := requests.AddDeviceReqToDeviceModels([]requests.AddDeviceRequest{notFoundProfile})[0]
	dbClientMock.On("AddDevice", notFoundProfileDeviceModel).Return(notFoundProfileDeviceModel,
		edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, fmt.Sprintf("device profile '%s' does not exists",
			notFoundProfile.Device.ProfileName), nil))

	notFoundService := testDevice
	notFoundService.Device.ServiceName = "notFoundService"
	notFoundServiceDeviceModel := requests.AddDeviceReqToDeviceModels([]requests.AddDeviceRequest{notFoundService})[0]
	dbClientMock.On("DeviceServiceNameExists", notFoundService.Device.ServiceName).Return(false, nil)
	dbClientMock.On("AddDevice", notFoundServiceDeviceModel).Return(notFoundServiceDeviceModel,
		edgexErr.NewCommonEdgeX(edgexErr.KindContractInvalid, fmt.Sprintf("device service '%s' does not exists", notFoundService.Device.ServiceName), nil))

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

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceController(dic)
	assert.NotNil(t, controller)
	tests := []struct {
		name                 string
		request              []requests.AddDeviceRequest
		expectedStatusCode   int
		expectedResponseCode int
		expectedValidation   bool
		expectedSystemEvent  bool
	}{
		{"Valid", []requests.AddDeviceRequest{valid}, http.StatusMultiStatus, http.StatusCreated, true, true},
		{"Invalid - not found profile", []requests.AddDeviceRequest{notFoundProfile}, http.StatusMultiStatus, http.StatusNotFound, true, false},
		{"Invalid - no name", []requests.AddDeviceRequest{noName}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - no adminState", []requests.AddDeviceRequest{noAdminState}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - no operatingState", []requests.AddDeviceRequest{noOperatingState}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - invalid adminState", []requests.AddDeviceRequest{invalidAdminState}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - invalid operatingState", []requests.AddDeviceRequest{invalidOperatingState}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - no service name", []requests.AddDeviceRequest{noServiceName}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - no profile name", []requests.AddDeviceRequest{noProfileName}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - no protocols", []requests.AddDeviceRequest{noProtocols}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - empty protocols", []requests.AddDeviceRequest{emptyProtocols}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - invalid protocols", []requests.AddDeviceRequest{invalidProtocols}, http.StatusMultiStatus, http.StatusInternalServerError, true, false},
		{"Invalid - not found device service", []requests.AddDeviceRequest{notFoundService}, http.StatusMultiStatus, http.StatusBadRequest, false, false},
		{"Invalid - device service unavailable", []requests.AddDeviceRequest{valid}, http.StatusMultiStatus, http.StatusServiceUnavailable, true, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			var responseEnvelope types.MessageEnvelope
			mockMessaging := &messagingMocks.MessageClient{}
			if testCase.expectedValidation {
				if testCase.expectedResponseCode == http.StatusInternalServerError {
					mockMessaging.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
						requestEnvelope, ok := args.Get(0).(types.MessageEnvelope)
						require.True(t, ok)
						responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, "validation failed")
					}).Return(&responseEnvelope, nil)
				} else if testCase.expectedResponseCode == http.StatusServiceUnavailable {
					mockMessaging.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&responseEnvelope, errors.New("timed out"))
				} else {
					mockMessaging.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
						requestEnvelope, ok := args.Get(0).(types.MessageEnvelope)
						require.True(t, ok)
						responseEnvelope, err = types.NewMessageEnvelopeForResponse(nil, requestEnvelope.RequestID, requestEnvelope.CorrelationID, common.ContentTypeJSON)
						require.NoError(t, err)
					}).Return(&responseEnvelope, nil)
				}
			}

			var wg sync.WaitGroup
			if testCase.expectedSystemEvent {
				wg.Add(1)
				mockMessaging.On("Publish", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					wg.Done()
				}).Return(nil)
			}

			dic.Update(di.ServiceConstructorMap{
				bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
					return mockMessaging
				},
			})

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDevice)
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
				if testCase.expectedResponseCode == http.StatusCreated {
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

			wg.Wait()
			mockMessaging.AssertExpectations(t)
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
	dbClientMock.On("DeleteDeviceByName", notFoundName).Return(edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "device doesn't exist in the database", nil))
	dbClientMock.On("DeviceByName", notFoundName).Return(device, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "device doesn't exist in the database", nil))
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
	dbClientMock.On("DevicesByServiceName", 4, 1, testServiceB).Return([]models.Device{}, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
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
		Location:       testReq.Device.Location,
		ServiceName:    *testReq.Device.ServiceName,
		ProfileName:    *testReq.Device.ProfileName,
		AutoEvents:     dtos.ToAutoEventModels(testReq.Device.AutoEvents),
		Protocols:      dtos.ToProtocolModels(testReq.Device.Protocols),
		Properties:     testProperties,
	}

	valid := testReq
	dbClientMock.On("DeviceServiceNameExists", *valid.Device.ServiceName).Return(true, nil)
	dbClientMock.On("DeviceById", *valid.Device.Id).Return(dsModels, nil)
	dbClientMock.On("UpdateDevice", dsModels).Return(nil)

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
	notFoundIdError := edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("DeviceById", *invalidNotFoundId.Device.Id).Return(dsModels, notFoundIdError)

	invalidNotFoundName := testReq
	invalidNotFoundName.Device.Id = nil
	notFoundName := "notFoundName"
	invalidNotFoundName.Device.Name = &notFoundName
	notFoundNameError := edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundName), nil)
	dbClientMock.On("DeviceByName", *invalidNotFoundName.Device.Name).Return(dsModels, notFoundNameError)

	notFoundServiceName := "notFoundService"
	notFoundService := testReq
	notFoundService.Device.ServiceName = &notFoundServiceName
	dbClientMock.On("DeviceServiceNameExists", *notFoundService.Device.ServiceName).Return(false, nil)

	notFountProfileName := "notFoundProfile"
	notFoundProfile := testReq
	notFoundProfile.Device.ProfileName = &notFountProfileName
	notFoundProfileDeviceModel := dsModels
	notFoundProfileDeviceModel.ProfileName = notFountProfileName
	dbClientMock.On("UpdateDevice", notFoundProfileDeviceModel).Return(
		edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist,
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
		expectedValidation   bool
		expectedSystemEvent  bool
	}{
		{"Valid", []requests.UpdateDeviceRequest{valid}, http.StatusMultiStatus, http.StatusOK, true, true},
		{"Valid - no requestId", []requests.UpdateDeviceRequest{validWithNoReqID}, http.StatusMultiStatus, http.StatusOK, true, true},
		{"Valid - no id", []requests.UpdateDeviceRequest{validWithNoId}, http.StatusMultiStatus, http.StatusOK, true, true},
		{"Valid - no name", []requests.UpdateDeviceRequest{validWithNoName}, http.StatusMultiStatus, http.StatusOK, true, true},
		{"Invalid - invalid id", []requests.UpdateDeviceRequest{invalidId}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - empty id", []requests.UpdateDeviceRequest{emptyId}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - empty name", []requests.UpdateDeviceRequest{emptyName}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - no id and name", []requests.UpdateDeviceRequest{invalidNoIdAndName}, http.StatusBadRequest, http.StatusBadRequest, false, false},
		{"Invalid - not found id", []requests.UpdateDeviceRequest{invalidNotFoundId}, http.StatusMultiStatus, http.StatusNotFound, false, false},
		{"Invalid - not found name", []requests.UpdateDeviceRequest{invalidNotFoundName}, http.StatusMultiStatus, http.StatusNotFound, false, false},
		{"Invalid - not found profile", []requests.UpdateDeviceRequest{notFoundProfile}, http.StatusMultiStatus, http.StatusNotFound, true, false},
		{"Invalid - invalid protocols", []requests.UpdateDeviceRequest{invalidProtocols}, http.StatusMultiStatus, http.StatusInternalServerError, true, false},
		{"Invalid - not found device service", []requests.UpdateDeviceRequest{notFoundService}, http.StatusMultiStatus, http.StatusBadRequest, false, false},
		{"Invalid - device service unavailable", []requests.UpdateDeviceRequest{valid}, http.StatusMultiStatus, http.StatusServiceUnavailable, true, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			var responseEnvelope types.MessageEnvelope
			mockMessaging := &messagingMocks.MessageClient{}
			if testCase.expectedValidation {
				if testCase.expectedResponseCode == http.StatusInternalServerError {
					mockMessaging.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
						requestEnvelope, ok := args.Get(0).(types.MessageEnvelope)
						require.True(t, ok)
						responseEnvelope = types.NewMessageEnvelopeWithError(requestEnvelope.RequestID, "validation failed")
					}).Return(&responseEnvelope, nil)
				} else if testCase.expectedResponseCode == http.StatusServiceUnavailable {
					mockMessaging.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&responseEnvelope, errors.New("timed out"))
				} else {
					mockMessaging.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
						requestEnvelope, ok := args.Get(0).(types.MessageEnvelope)
						require.True(t, ok)
						responseEnvelope, err = types.NewMessageEnvelopeForResponse(nil, requestEnvelope.RequestID, requestEnvelope.CorrelationID, common.ContentTypeJSON)
						require.NoError(t, err)
					}).Return(&responseEnvelope, nil)
				}
			}

			var wg sync.WaitGroup
			if testCase.expectedSystemEvent {
				wg.Add(1)
				mockMessaging.On("Publish", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					wg.Done()
				}).Return(nil)
			}

			dic.Update(di.ServiceConstructorMap{
				bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
					return mockMessaging
				},
			})

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

			wg.Wait()
			mockMessaging.AssertExpectations(t)
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
	dbClientMock.On("AllDevices", 4, 1, testDeviceLabels).Return([]models.Device{}, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
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
	dbClientMock.On("DeviceByName", notFoundName).Return(models.Device{}, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "device doesn't exist in the database", nil))
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
	dbClientMock.On("DevicesByProfileName", 4, 1, testProfileB).Return([]models.Device{}, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
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
