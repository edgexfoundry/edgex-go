//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testDeviceCommandRequest = requests.AddDeviceCommandRequest{
	BaseRequest: commonDTO.BaseRequest{
		RequestId:   ExampleUUID,
		Versionable: commonDTO.NewVersionable(),
	},
	ProfileName: TestDeviceProfileName,
	DeviceCommand: dtos.DeviceCommand{
		Name:      "TestDeviceCommandNewName",
		IsHidden:  false,
		ReadWrite: common.ReadWrite_RW,
		ResourceOperations: []dtos.ResourceOperation{{
			DeviceResource: TestDeviceResourceName,
			Mappings:       testMappings,
		}},
	},
}

func TestAddDeviceProfileDeviceCommand(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	expectedRequestId := ExampleUUID

	valid := testDeviceCommandRequest
	noRequestId := testDeviceCommandRequest
	noRequestId.RequestId = ""
	duplicateName := testDeviceCommandRequest
	duplicateName.DeviceCommand.Name = TestDeviceCommandName
	notFoundDeviceResource := testDeviceCommandRequest
	notFoundDeviceResource.DeviceCommand.ResourceOperations = []dtos.ResourceOperation{{
		DeviceResource: "notFoundName",
		Mappings:       testMappings,
	}}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceProfileByName", valid.ProfileName).Return(deviceProfile, nil)
	dbClientMock.On("UpdateDeviceProfile", mock.Anything).Return(nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceCommandController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		Request            []requests.AddDeviceCommandRequest
		expectedStatusCode int
	}{
		{"Valid - AddDeviceCommandRequest", []requests.AddDeviceCommandRequest{valid}, http.StatusCreated},
		{"Valid - No requestId", []requests.AddDeviceCommandRequest{noRequestId}, http.StatusCreated},
		{"invalid - Duplicate deviceCommand name", []requests.AddDeviceCommandRequest{duplicateName}, http.StatusBadRequest},
		{"invalid - Not Exist deviceResource", []requests.AddDeviceCommandRequest{notFoundDeviceResource}, http.StatusBadRequest},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceProfileDeviceCommandRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfileDeviceCommand)
			handler.ServeHTTP(recorder, req)

			var res []commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
			if res[0].RequestId != "" {
				assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
			}
			assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
		})
	}
}

func TestAddDeviceProfileDeviceCommand_BadRequest(t *testing.T) {
	noDeviceCommandName := testDeviceCommandRequest
	noDeviceCommandName.DeviceCommand.Name = ""
	invalidReadWrite := testDeviceCommandRequest
	invalidReadWrite.DeviceCommand.ReadWrite = "invalid"
	noResourceOperations := testDeviceCommandRequest
	noResourceOperations.DeviceCommand.ResourceOperations = []dtos.ResourceOperation{{}}

	dic := mockDic()
	controller := NewDeviceCommandController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name    string
		Request []requests.AddDeviceCommandRequest
	}{
		{"invalid - No deviceCommand name", []requests.AddDeviceCommandRequest{noDeviceCommandName}},
		{"invalid - Invalid ReadWrite", []requests.AddDeviceCommandRequest{invalidReadWrite}},
		{"invalid - No ResourceOperations", []requests.AddDeviceCommandRequest{noResourceOperations}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceProfileDeviceCommandRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfileDeviceCommand)
			handler.ServeHTTP(recorder, req)

			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, http.StatusBadRequest, res.StatusCode, "BaseResponse status code not as expected")
			assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
		})
	}
}
