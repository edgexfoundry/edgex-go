//
// Copyright (C) 2021-2023 IOTech Ltd
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
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces/mocks"

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
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ValueType: common.ValueTypeInt16,
			ReadWrite: common.ReadWrite_RW,
		},
		Tags: testTags,
	},
}

func buildTestUpdateDeviceResourceRequest() requests.UpdateDeviceResourceRequest {
	testDeviceResourceName := TestDeviceResourceName
	testDescription := TestDescription
	testIsHidden := false
	var testUpdateDeviceResourceReq = requests.UpdateDeviceResourceRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		ProfileName: TestDeviceProfileName,
		Resource: dtos.UpdateDeviceResource{
			Name:        &testDeviceResourceName,
			Description: &testDescription,
			IsHidden:    &testIsHidden,
		},
	}
	return testUpdateDeviceResourceReq
}

func TestDeviceResourceByProfileNameAndResourceName(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	emptyName := ""
	profileNotFoundName := "profileNotFoundName"
	resourceNotFoundName := "resourceNotFoundName"

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
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
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileByName", valid.ProfileName).Return(deviceProfile, nil)
	dbClientMock.On("DeviceProfileByName", notFoundProfileName.ProfileName).Return(deviceProfile, notFoundDBError)
	dbClientMock.On("UpdateDeviceProfile", mock.Anything).Return(nil)
	dbClientMock.On("DevicesByProfileName", 0, -1, TestDeviceProfileName).Return([]models.Device{}, nil)
	dbClientMock.On("DeviceCountByProfileName", TestDeviceProfileName).Return(uint32(1), nil)
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

func TestAddDeviceProfileResource_UnitsOfMeasure_Validation(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	expectedRequestId := ExampleUUID

	emptyUnitsReq := testDeviceResourceRequest
	validReq := emptyUnitsReq
	validReq.Resource.Properties.Units = TestUnits
	invalidUnitsReq := emptyUnitsReq
	invalidUnitsReq.Resource.Properties.Units = "invalid"

	dic := mockDic()
	container.ConfigurationFrom(dic.Get).Writable.UoM.Validation = true
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileByName", validReq.ProfileName).Return(deviceProfile, nil)
	dbClientMock.On("UpdateDeviceProfile", mock.Anything).Return(nil)
	dbClientMock.On("DevicesByProfileName", 0, -1, validReq.ProfileName).Return([]models.Device{}, nil)
	dbClientMock.On("DeviceCountByProfileName", validReq.ProfileName).Return(uint32(1), nil)
	uomMock := &mocks.UnitsOfMeasure{}
	uomMock.On("Validate", TestUnits).Return(true)
	uomMock.On("Validate", "").Return(true)
	uomMock.On("Validate", "invalid").Return(false)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.UnitsOfMeasureInterfaceName: func(get di.Get) interface{} {
			return uomMock
		},
	})

	controller := NewDeviceResourceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		Request            []requests.AddDeviceResourceRequest
		expectedStatusCode int
	}{
		{"valid - units not provided", []requests.AddDeviceResourceRequest{emptyUnitsReq}, http.StatusCreated},
		{"valid - expected units", []requests.AddDeviceResourceRequest{validReq}, http.StatusCreated},
		{"invalid - unexpected units", []requests.AddDeviceResourceRequest{invalidUnitsReq}, http.StatusBadRequest},
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
			// Assert
			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
			if res[0].RequestId != "" {
				assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
			}
			assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
		})
	}
}
func TestPatchDeviceProfileResource(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	testReq := buildTestUpdateDeviceResourceRequest()
	deviceExistsProfileName := "deviceExists"
	deviceExistsProfile := deviceProfile
	deviceExistsProfile.Name = deviceExistsProfileName
	expectedRequestId := ExampleUUID
	emptyString := ""
	notFound := "notFoundName"

	valid := testReq
	validWithNoReqID := testReq
	validWithNoReqID.RequestId = emptyString
	deviceExists := testReq
	deviceExists.ProfileName = deviceExistsProfileName
	notFoundResource := testReq
	notFoundResource.Resource.Name = &notFound
	notFoundProfile := testReq
	notFoundProfile.ProfileName = notFound
	notFoundDBError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile %s does not exists", notFoundProfile.ProfileName), nil)

	noProfileName := testReq
	noProfileName.ProfileName = emptyString

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileByName", valid.ProfileName).Return(deviceProfile, nil)
	dbClientMock.On("DevicesByProfileName", 0, mock.Anything, valid.ProfileName).Return([]models.Device{}, nil)
	dbClientMock.On("DeviceCountByProfileName", TestDeviceProfileName).Return(uint32(1), nil)
	dbClientMock.On("UpdateDeviceProfile", mock.Anything).Return(nil)

	dbClientMock.On("DeviceProfileByName", deviceExistsProfileName).Return(deviceExistsProfile, nil)
	dbClientMock.On("DevicesByProfileName", 0, -1, deviceExistsProfileName).Return([]models.Device{{}}, nil)
	dbClientMock.On("DeviceCountByProfileName", deviceExistsProfileName).Return(uint32(1), nil)

	dbClientMock.On("DeviceProfileByName", notFound).Return(deviceProfile, notFoundDBError)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceResourceController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name                 string
		Request              []requests.UpdateDeviceResourceRequest
		expectedStatusCode   int
		expectedResponseCode int
	}{
		{"Valid - PatchDeviceProfileResourceRequest", []requests.UpdateDeviceResourceRequest{valid}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - PatchDeviceProfileResourceRequest no requestId", []requests.UpdateDeviceResourceRequest{validWithNoReqID}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - Profile is in use by other device", []requests.UpdateDeviceResourceRequest{deviceExists}, http.StatusMultiStatus, http.StatusOK},
		{"Invalid - Not found device resource", []requests.UpdateDeviceResourceRequest{notFoundResource}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - Not found device profile", []requests.UpdateDeviceResourceRequest{notFoundProfile}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - No profile name", []requests.UpdateDeviceResourceRequest{noProfileName}, http.StatusBadRequest, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))

			req, err := http.NewRequest(http.MethodPatch, common.ApiDeviceProfileResourceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchDeviceProfileResource)
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

func TestDeleteDeviceResourceByName(t *testing.T) {
	dpModel := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	notFoundName := "notFoundName"
	deviceExists := "deviceExists"

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DevicesByProfileName", 0, mock.Anything, TestDeviceProfileName).Return([]models.Device{}, nil)
	dbClientMock.On("DeviceCountByProfileName", TestDeviceProfileName).Return(uint32(1), nil)
	dbClientMock.On("DeviceProfileByName", TestDeviceProfileName).Return(dpModel, nil)
	dbClientMock.On("UpdateDeviceProfile", mock.Anything).Return(nil)

	dbClientMock.On("DevicesByProfileName", 0, 1, deviceExists).Return([]models.Device{models.Device{}}, nil)

	dbClientMock.On("DevicesByProfileName", 0, 1, notFoundName).Return([]models.Device{}, nil)
	dbClientMock.On("DeviceProfileByName", notFoundName).Return(models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "not found", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceResourceController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		profileName        string
		resourceName       string
		expectedStatusCode int
	}{
		{"valid", TestDeviceProfileName, TestDeviceResourceName + "-dup", http.StatusOK},
		{"invalid - empty profile name", "", TestDeviceResourceName, http.StatusBadRequest},
		{"invalid - empty resource name", "", TestDeviceResourceName, http.StatusBadRequest},
		{"invalid - profile is in use by other device", deviceExists, TestDeviceResourceName, http.StatusConflict},
		{"invalid - profile not found", notFoundName, TestDeviceResourceName, http.StatusNotFound},
		{"invalid - resource not found in profile", TestDeviceProfileName, notFoundName, http.StatusNotFound},
		{"invalid - device resource is referenced by device command", TestDeviceProfileName, TestDeviceResourceName, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, common.ApiDeviceProfileResourceByNameRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.profileName, common.ResourceName: testCase.resourceName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteDeviceResourceByName)
			handler.ServeHTTP(recorder, req)

			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
			if testCase.expectedStatusCode != http.StatusOK {
				assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
			}
		})
	}

}

func TestDeleteDeviceResourceByName_StrictProfileChanges(t *testing.T) {
	dic := mockDic()
	configuration := container.ConfigurationFrom(dic.Get)
	configuration.Writable.ProfileChange.StrictDeviceProfileChanges = true
	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	controller := NewDeviceResourceController(dic)
	require.NotNil(t, controller)

	req, err := http.NewRequest(http.MethodDelete, common.ApiDeviceProfileResourceByNameRoute, http.NoBody)
	req = mux.SetURLVars(req, map[string]string{common.Name: TestDeviceProfileName, common.ResourceName: TestDeviceResourceName})
	require.NoError(t, err)

	// Act
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.DeleteDeviceResourceByName)
	handler.ServeHTTP(recorder, req)

	var res commonDTO.BaseResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusLocked, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusLocked, res.StatusCode, "BaseResponse status code not as expected")
}
