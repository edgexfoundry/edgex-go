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

var testDeviceResourceRequest = requests.AddDeviceResourceRequest{
	BaseRequest: commonDTO.BaseRequest{
		RequestId:   ExampleUUID,
		Versionable: commonDTO.NewVersionable(),
	},
	ProfileName: TestDeviceProfileName,
	Resource: dtos.DeviceResource{
		Name:        "TestDeviceResourceNewName",
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ValueType: common.ValueTypeInt16,
			ReadWrite: common.ReadWrite_RW,
		},
	},
}

func TestDeviceResourceByProfileNameAndResourceName(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	emptyName := ""
	profileNotFoundName := "profileNotFoundName"
	resourceNotFoundName := "resourceNotFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceProfileByName", deviceProfile.Name).Return(deviceProfile, nil)
	dbClientMock.On("DeviceProfileByName", profileNotFoundName).Return(models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device profile doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceResourceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		profileName        string
		resourceName       string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - find device resource by profileName and resourceName", deviceProfile.Name, TestDeviceResourceName, false, http.StatusOK},
		{"Invalid - profile name is empty", emptyName, TestDeviceResourceName, true, http.StatusBadRequest},
		{"Invalid - profile name is empty", deviceProfile.Name, emptyName, true, http.StatusBadRequest},
		{"Invalid - device profile not", profileNotFoundName, TestDeviceResourceName, true, http.StatusNotFound},
		{"Invalid - resource not found", deviceProfile.Name, resourceNotFoundName, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiDeviceResourceByProfileAndResourceRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.ProfileName: testCase.profileName, common.ResourceName: testCase.resourceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceResourceByProfileNameAndResourceName)
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
				var res responseDTO.DeviceResourceResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.resourceName, res.Resource.Name, "Resource name not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestAddDeviceProfileResource(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	expectedRequestId := ExampleUUID

	valid := testDeviceResourceRequest
	noRequestId := testDeviceResourceRequest
	noRequestId.RequestId = ""
	duplicateName := testDeviceResourceRequest
	duplicateName.Resource.Name = TestDeviceResourceName
	noProfileName := testDeviceResourceRequest
	noProfileName.ProfileName = ""
	noDeviceResourcePropertyType := testDeviceResourceRequest
	noDeviceResourcePropertyType.Resource.Properties.ValueType = ""
	notFoundProfileName := testDeviceResourceRequest
	notFoundProfileName.ProfileName = "notFoundName"
	notFoundDBError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile %s does not exists", notFoundProfileName.ProfileName), nil)

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceProfileByName", valid.ProfileName).Return(deviceProfile, nil)
	dbClientMock.On("DeviceProfileByName", notFoundProfileName.ProfileName).Return(deviceProfile, notFoundDBError)
	dbClientMock.On("UpdateDeviceProfile", mock.Anything).Return(nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceResourceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		Request            []requests.AddDeviceResourceRequest
		expectedStatusCode int
	}{
		{"Valid - AddDeviceProfileResourceRequest", []requests.AddDeviceResourceRequest{valid}, http.StatusCreated},
		{"Valid - No requestId", []requests.AddDeviceResourceRequest{noRequestId}, http.StatusCreated},
		{"invalid - duplicate name", []requests.AddDeviceResourceRequest{duplicateName}, http.StatusBadRequest},
		{"invalid - Not Exist device profile", []requests.AddDeviceResourceRequest{notFoundProfileName}, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceProfileResourceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfileResource)
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

func TestAddDeviceProfileResource_BadRequest(t *testing.T) {
	noProfileName := testDeviceResourceRequest
	noProfileName.ProfileName = ""
	noDeviceResourcePropertyType := testDeviceResourceRequest
	noDeviceResourcePropertyType.Resource.Properties.ValueType = ""

	dic := mockDic()

	controller := NewDeviceResourceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name    string
		Request []requests.AddDeviceResourceRequest
	}{
		{"invalid - no device profile name", []requests.AddDeviceResourceRequest{noProfileName}},
		{"invalid - No deviceResource property type", []requests.AddDeviceResourceRequest{noDeviceResourcePropertyType}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceProfileResourceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfileResource)
			handler.ServeHTTP(recorder, req)
			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, http.StatusBadRequest, res.StatusCode, "BaseResponse status code not as expected")
			assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
		})
	}
}
