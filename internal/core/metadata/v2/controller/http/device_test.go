//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v2MetadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/di"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

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
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiDeviceRoute, reader)
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
				if res[0].RequestId != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}
