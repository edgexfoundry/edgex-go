//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
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
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mockProtocol = "http"
	mockHost     = "127.0.0.1"
	mockPort     = 66666

	testProfileName = "testProfileName"
	testDeviceName  = "testDeviceName"
	testPathPrefix  = v2.ApiDeviceRoute + "/" + v2.Name + "/" + testDeviceName + "/" + v2.Command + "/"
	testUrl         = "http://localhost:48082"
)

// NewMockDIC function returns a mock bootstrap di Container
func NewMockDIC() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		commandContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{
					Protocol: mockProtocol,
					Host:     mockHost,
					Port:     mockPort,
				},
			}
		},
		container.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
}

func buildCoreCommands(commands []dtos.Command) []dtos.CoreCommand {
	coreCommands := make([]dtos.CoreCommand, len(commands))
	for i, c := range commands {
		coreCommands[i] = dtos.CoreCommand{
			Name:       c.Name,
			DeviceName: testDeviceName,
			Get:        c.Get,
			Put:        c.Put,
			Url:        testUrl,
			Path:       testPathPrefix + c.Name,
		}
	}
	return coreCommands
}

func buildDeviceResponse() responseDTO.DeviceResponse {
	device := dtos.Device{
		Name:        testDeviceName,
		ProfileName: testProfileName,
	}
	deviceResponse := responseDTO.DeviceResponse{
		Device: device,
	}
	return deviceResponse
}

func buildCommands() []dtos.Command {
	c1 := dtos.Command{
		Name: "command1",
		Get:  true,
		Put:  false,
	}
	c2 := dtos.Command{
		Name: "command2",
		Get:  true,
		Put:  false,
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

func TestCommandsByDeviceName(t *testing.T) {
	var nonExistDeviceName = "nonExistDevice"

	expectedDeviceResponse := buildDeviceResponse()
	expectedDeviceProfileResponse := buildDeviceProfileResponse()
	expectedCoreCommands := buildCoreCommands(expectedDeviceProfileResponse.Profile.CoreCommands)

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
		{"Valid - get coreCommands with deviceName", testDeviceName, false, len(expectedCoreCommands), http.StatusOK},
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
				var res responseDTO.MultiCoreCommandsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.CoreCommands), "Device count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
