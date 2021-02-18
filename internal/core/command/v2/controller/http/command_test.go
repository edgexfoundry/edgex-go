//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	commandContainer "github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	V2Container "github.com/edgexfoundry/go-mod-bootstrap/v2/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mockProtocol = "http"
	mockHost     = "127.0.0.1"
	mockPort     = 66666

	testProfileName       = "testProfile"
	testResourceName      = "testResource"
	testDeviceName        = "testDevice"
	testDeviceServiceName = "testDeviceService"
	testCommandName       = "testCommand"
	testPathPrefix        = v2.ApiDeviceRoute + "/" + v2.Name + "/" + testDeviceName + "/" + v2.Command + "/"
	testUrl               = "http://localhost:48082"
	testBaseAddress       = "http://localhost:49990"
	testQueryStrings      = "a=1&b=2&ds-pushevent=no"
)

// NewMockDIC function returns a mock bootstrap di Container
func NewMockDIC() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		commandContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{
					Protocol:       mockProtocol,
					Host:           mockHost,
					Port:           mockPort,
					MaxResultCount: 20,
				},
			}
		},
		container.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
}

func buildTestSettings() map[string]string {
	var settings = make(map[string]string)
	settings["AHU-TargetTemperature"] = "28.5"
	settings["AHU-TargetBand"] = "4.0"
	return settings
}

func buildDeviceCoreCommands(device dtos.Device, deviceProfile dtos.DeviceProfile) dtos.DeviceCoreCommand {
	coreCommands := make([]dtos.CoreCommand, len(deviceProfile.CoreCommands))
	for i, c := range deviceProfile.CoreCommands {
		coreCommands[i] = dtos.CoreCommand{
			Name: c.Name,
			Get:  c.Get,
			Set:  c.Set,
			Url:  testUrl,
			Path: testPathPrefix + c.Name,
		}
	}
	return dtos.DeviceCoreCommand{
		DeviceName:   device.Name,
		ProfileName:  deviceProfile.Name,
		CoreCommands: coreCommands,
	}
}

func buildDeviceResponse() responseDTO.DeviceResponse {
	device := dtos.Device{
		Name:        testDeviceName,
		ProfileName: testProfileName,
		ServiceName: testDeviceServiceName,
	}
	deviceResponse := responseDTO.DeviceResponse{
		Device: device,
	}
	return deviceResponse
}

func buildMultiDevicesResponse() responseDTO.MultiDevicesResponse {
	devices := []dtos.Device{
		{Name: testDeviceName + "1", ProfileName: testProfileName, ServiceName: testDeviceServiceName},
		{Name: testDeviceName + "2", ProfileName: testProfileName, ServiceName: testDeviceServiceName},
	}
	return responseDTO.MultiDevicesResponse{
		Devices: devices,
	}
}

func buildCommands() []dtos.Command {
	c1 := dtos.Command{
		Name: "command1",
		Get:  true,
		Set:  false,
	}
	c2 := dtos.Command{
		Name: "command2",
		Get:  true,
		Set:  false,
	}
	var commands []dtos.Command
	commands = append(commands, c1, c2)
	return commands
}

func buildDeviceProfileResponse() responseDTO.DeviceProfileResponse {
	commands := buildCommands()
	profile := dtos.DeviceProfile{
		Name:         testProfileName,
		CoreCommands: commands,
	}
	deviceResponse := responseDTO.DeviceProfileResponse{
		Profile: profile,
	}
	return deviceResponse
}

func buildDeviceServiceResponse() responseDTO.DeviceServiceResponse {
	service := dtos.DeviceService{
		Name:        testDeviceServiceName,
		BaseAddress: testBaseAddress,
	}
	return responseDTO.DeviceServiceResponse{
		Service: service,
	}
}

func buildEvent() dtos.Event {
	event := dtos.NewEvent(testProfileName, testDeviceName)
	event.AddSimpleReading(testResourceName, v2.ValueTypeUint16, uint16(45))
	id, _ := uuid.NewUUID()
	event.Id = id.String()
	event.Readings[0].Id = id.String()
	return event
}

func buildEventResponse() responseDTO.EventResponse {
	return responseDTO.EventResponse{
		Event: buildEvent(),
	}
}

func TestAllCommands(t *testing.T) {
	expectedMultiDevicesResponse := buildMultiDevicesResponse()
	expectedDeviceProfileResponse := buildDeviceProfileResponse()
	deviceCoreCommand1 := buildDeviceCoreCommands(expectedMultiDevicesResponse.Devices[0], expectedDeviceProfileResponse.Profile)
	deviceCoreCommand2 := buildDeviceCoreCommands(expectedMultiDevicesResponse.Devices[1], expectedDeviceProfileResponse.Profile)
	expectedMultiDeviceCoreCommandsResponse := responseDTO.MultiDeviceCoreCommandsResponse{
		DeviceCoreCommands: []dtos.DeviceCoreCommand{deviceCoreCommand1, deviceCoreCommand2},
	}

	dcMock := &mocks.DeviceClient{}
	dcMock.On("AllDevices", context.Background(), []string(nil), 0, 20).Return(expectedMultiDevicesResponse, nil)
	dcMock.On("AllDevices", context.Background(), []string(nil), 0, 2).Return(expectedMultiDevicesResponse, nil)
	dcMock.On("AllDevices", context.Background(), []string(nil), 3, 10).Return(responseDTO.MultiDevicesResponse{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "bounds out of range", nil))

	dpcMock := &mocks.DeviceProfileClient{}
	dpcMock.On("DeviceProfileByName", context.Background(), testProfileName).Return(expectedDeviceProfileResponse, nil)

	dic := NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		V2Container.MetadataDeviceClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceClient
			return dcMock
		},
		V2Container.MetadataDeviceProfileClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceProfileClient
			return dpcMock
		},
	})
	cc := NewCommandController(dic)
	assert.NotNil(t, cc)

	tests := []struct {
		name               string
		offset             string
		limit              string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get commands without offset and limit", "", "", false, len(expectedMultiDeviceCoreCommandsResponse.DeviceCoreCommands), http.StatusOK},
		{"Valid - get commands with offset and limit", "0", "2", false, len(expectedMultiDeviceCoreCommandsResponse.DeviceCoreCommands), http.StatusOK},
		{"Invalid - bounds out of range", "3", "10", true, 0, http.StatusRequestedRangeNotSatisfiable},
		{"Invalid - invalid offset format", "aaa", "10", true, 0, http.StatusBadRequest},
		{"Invalid - invalid limit format", "0", "aaa", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiAllDeviceRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(cc.AllCommands)
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
				var res responseDTO.MultiDeviceCoreCommandsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.DeviceCoreCommands), "Device count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestCommandsByDeviceName(t *testing.T) {
	var nonExistDeviceName = "nonExistDevice"

	expectedDeviceResponse := buildDeviceResponse()
	expectedDeviceProfileResponse := buildDeviceProfileResponse()
	expectedDeviceCoreCommand := buildDeviceCoreCommands(expectedDeviceResponse.Device, expectedDeviceProfileResponse.Profile)

	dcMock := &mocks.DeviceClient{}
	dcMock.On("DeviceByName", context.Background(), testDeviceName).Return(expectedDeviceResponse, nil)
	dcMock.On("DeviceByName", context.Background(), nonExistDeviceName).Return(responseDTO.DeviceResponse{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "fail to query device by name", nil))

	dpcMock := &mocks.DeviceProfileClient{}
	dpcMock.On("DeviceProfileByName", context.Background(), testProfileName).Return(expectedDeviceProfileResponse, nil)

	dic := NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		V2Container.MetadataDeviceClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceClient
			return dcMock
		},
		V2Container.MetadataDeviceProfileClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceProfileClient
			return dpcMock
		},
	})
	cc := NewCommandController(dic)
	assert.NotNil(t, cc)

	tests := []struct {
		name               string
		deviceName         string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get coreCommands with deviceName", testDeviceName, false, len(expectedDeviceCoreCommand.CoreCommands), http.StatusOK},
		{"Invalid - get coreCommands with empty deviceName", "", true, 0, http.StatusBadRequest},
		{"Invalid - get coreCommands with non exist deviceName", nonExistDeviceName, true, 0, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiDeviceByNameRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.deviceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(cc.CommandsByDeviceName)
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
				var res responseDTO.DeviceCoreCommandResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.DeviceCoreCommand.CoreCommands), "Device count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestIssueGetCommand(t *testing.T) {
	var nonExistName = "nonExist"

	expectedEventResponse := buildEventResponse()
	expectedDeviceResponse := buildDeviceResponse()
	expectedDeviceServiceResponse := buildDeviceServiceResponse()

	dcMock := &mocks.DeviceClient{}
	dcMock.On("DeviceByName", context.Background(), testDeviceName).Return(expectedDeviceResponse, nil)
	dcMock.On("DeviceByName", context.Background(), nonExistName).Return(responseDTO.DeviceResponse{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "fail to query device by name", nil))

	dscMock := &mocks.DeviceServiceClient{}
	dscMock.On("DeviceServiceByName", context.Background(), testDeviceServiceName).Return(expectedDeviceServiceResponse, nil)

	dsccMock := &mocks.DeviceServiceCommandClient{}
	dsccMock.On("GetCommand", context.Background(), testBaseAddress, testDeviceName, testCommandName, testQueryStrings).Return(expectedEventResponse, nil)
	dsccMock.On("GetCommand", context.Background(), testBaseAddress, testDeviceName, testCommandName, "").Return(expectedEventResponse, nil)
	dsccMock.On("GetCommand", context.Background(), testBaseAddress, testDeviceName, nonExistName, testQueryStrings).Return(responseDTO.EventResponse{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to query device service by name", nil))

	dic := NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		V2Container.MetadataDeviceClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceClient
			return dcMock
		},
		V2Container.MetadataDeviceServiceClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceProfileClient
			return dscMock
		},
		V2Container.DeviceServiceCommandClientName: func(get di.Get) interface{} { // add v2 API DeviceServiceCommandClient
			return dsccMock
		},
	})
	cc := NewCommandController(dic)
	assert.NotNil(t, cc)

	tests := []struct {
		name               string
		deviceName         string
		commandName        string
		queryStrings       string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - execute read command with valid deviceName, commandName, and query strings", testDeviceName, testCommandName, testQueryStrings, false, http.StatusOK},
		{"Valid - empty query strings", testDeviceName, testCommandName, "", false, http.StatusOK},
		{"Invalid - execute read command with invalid deviceName", nonExistName, testCommandName, testQueryStrings, true, http.StatusNotFound},
		{"Invalid - execute read command with invalid commandName", testDeviceName, nonExistName, testQueryStrings, true, http.StatusBadRequest},
		{"Invalid - empty device name", "", nonExistName, testQueryStrings, true, http.StatusBadRequest},
		{"Invalid - empty command name", testDeviceName, "", testQueryStrings, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiDeviceNameCommandNameRoute, http.NoBody)
			req.URL.RawQuery = testCase.queryStrings
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.deviceName, v2.Command: testCase.commandName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(cc.IssueGetCommandByName)
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
				var res responseDTO.EventResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestIssueSetCommand(t *testing.T) {
	var nonExistName = "nonExist"

	expectedBaseResponse := common.NewBaseResponse("", "", http.StatusOK)
	expectedDeviceResponse := buildDeviceResponse()
	expectedDeviceServiceResponse := buildDeviceServiceResponse()

	dcMock := &mocks.DeviceClient{}
	dcMock.On("DeviceByName", context.Background(), testDeviceName).Return(expectedDeviceResponse, nil)
	dcMock.On("DeviceByName", context.Background(), nonExistName).Return(responseDTO.DeviceResponse{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "fail to query device by name", nil))

	dscMock := &mocks.DeviceServiceClient{}
	dscMock.On("DeviceServiceByName", context.Background(), testDeviceServiceName).Return(expectedDeviceServiceResponse, nil)

	testSettings := buildTestSettings()
	testSettingsJsonStr, _ := json.Marshal(testSettings)
	dsccMock := &mocks.DeviceServiceCommandClient{}
	dsccMock.On("SetCommand", context.Background(), testBaseAddress, testDeviceName, testCommandName, testQueryStrings, testSettings).Return(expectedBaseResponse, nil)
	dsccMock.On("SetCommand", context.Background(), testBaseAddress, testDeviceName, testCommandName, "", testSettings).Return(expectedBaseResponse, nil)
	dsccMock.On("SetCommand", context.Background(), testBaseAddress, testDeviceName, testCommandName, testQueryStrings, "").Return(common.BaseResponse{}, errors.NewCommonEdgeX(errors.KindServerError, "no request body provided for PUT command", nil))
	dsccMock.On("SetCommand", context.Background(), testBaseAddress, testDeviceName, nonExistName, testQueryStrings, testSettings).Return(common.BaseResponse{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "no corresponding PUT command", nil))

	dic := NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		V2Container.MetadataDeviceClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceClient
			return dcMock
		},
		V2Container.MetadataDeviceServiceClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceProfileClient
			return dscMock
		},
		V2Container.DeviceServiceCommandClientName: func(get di.Get) interface{} { // add v2 API DeviceServiceCommandClient
			return dsccMock
		},
	})
	cc := NewCommandController(dic)
	assert.NotNil(t, cc)

	tests := []struct {
		name               string
		deviceName         string
		commandName        string
		queryStrings       string
		settings           []byte
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - execute set command with valid deviceName, commandName, query strings, and settings", testDeviceName, testCommandName, testQueryStrings, testSettingsJsonStr, false, http.StatusOK},
		{"Valid - empty query strings", testDeviceName, testCommandName, "", testSettingsJsonStr, false, http.StatusOK},
		{"Invalid - execute set command with invalid deviceName", nonExistName, testCommandName, testQueryStrings, testSettingsJsonStr, true, http.StatusNotFound},
		{"Invalid - execute set command with invalid commandName", testDeviceName, nonExistName, testQueryStrings, testSettingsJsonStr, true, http.StatusBadRequest},
		{"Invalid - empty device name", "", testCommandName, testQueryStrings, testSettingsJsonStr, true, http.StatusBadRequest},
		{"Invalid - empty command name", testDeviceName, "", testQueryStrings, testSettingsJsonStr, true, http.StatusBadRequest},
		{"Invalid - empty settings", testDeviceName, testCommandName, testQueryStrings, []byte{}, true, http.StatusInternalServerError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPut, v2.ApiDeviceNameCommandNameRoute, bytes.NewBuffer(testCase.settings))
			req.URL.RawQuery = testCase.queryStrings
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.deviceName, v2.Command: testCase.commandName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(cc.IssueSetCommandByName)
			handler.ServeHTTP(recorder, req)

			// Assert
			var res common.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)
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
