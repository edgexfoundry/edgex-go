//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	metadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	v2MetadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testLabels = []string{"MODBUS", "TEMP"}
var testAttributes = map[string]string{
	"TestAttribute": "TestAttributeValue",
}

func buildTestDeviceProfileRequest() requests.AddDeviceProfileRequest {
	var testDeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.PropertyValue{
			Type:      "INT16",
			ReadWrite: "RW",
		},
	}}
	var testDeviceCommands = []dtos.ProfileResource{{
		Name: TestProfileResourceName,
		Get: []dtos.ResourceOperation{{
			DeviceResource: TestDeviceResourceName,
		}},
		Set: []dtos.ResourceOperation{{
			DeviceResource: TestDeviceResourceName,
		}},
	}}
	var testCoreCommands = []dtos.Command{{
		Name: TestProfileResourceName,
		Get:  true,
		Put:  true,
	}}

	var testAddDeviceProfileReq = requests.AddDeviceProfileRequest{
		BaseRequest: common.BaseRequest{
			RequestID: ExampleUUID,
		},
		Profile: dtos.DeviceProfile{
			Id:              ExampleUUID,
			Name:            TestDeviceProfileName,
			Manufacturer:    TestManufacturer,
			Description:     TestDescription,
			Model:           TestModel,
			Labels:          testLabels,
			DeviceResources: testDeviceResources,
			DeviceCommands:  testDeviceCommands,
			CoreCommands:    testCoreCommands,
		},
	}

	return testAddDeviceProfileReq
}

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		metadataContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
			}
		},
		container.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
}

func createDeviceProfileRequestWithFile(fileContents []byte) (*http.Request, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "deviceProfile.yaml")
	if err != nil {
		return nil, err
	}
	_, err = part.Write(fileContents)
	if err != nil {
		return nil, err
	}
	boundary := writer.Boundary()

	err = writer.Close()
	if err != err {
		return nil, err
	}

	req, _ := http.NewRequest(http.MethodPost, contractsV2.ApiDeviceProfileRoute+"/uploadfile", body)
	req.Header.Set(clients.ContentType, "multipart/form-data; boundary="+boundary)
	return req, nil
}

func TestAddDeviceProfile_Created(t *testing.T) {
	deviceProfileRequest := buildTestDeviceProfileRequest()
	deviceProfileModel := requests.AddDeviceProfileReqToDeviceProfileModel(deviceProfileRequest)
	expectedRequestId := ExampleUUID
	expectedMessage := "Add device profiles successfully"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AddDeviceProfile", deviceProfileModel).Return(deviceProfileModel, nil)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	valid := deviceProfileRequest
	noRequestId := deviceProfileRequest
	noRequestId.RequestID = ""

	tests := []struct {
		name    string
		Request []requests.AddDeviceProfileRequest
	}{
		{"Valid - AddDeviceProfileRequest", []requests.AddDeviceProfileRequest{valid}},
		{"Valid - No requestId", []requests.AddDeviceProfileRequest{noRequestId}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiDeviceProfileRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfile)
			handler.ServeHTTP(recorder, req)
			var res []common.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contractsV2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
			if res[0].RequestID != "" {
				assert.Equal(t, expectedRequestId, res[0].RequestID, "RequestID not as expected")
			}
			assert.Equal(t, http.StatusCreated, res[0].StatusCode, "BaseResponse status code not as expected")
			assert.Equal(t, expectedMessage, res[0].Message, "Message not as expected")
		})
	}
}

func TestAddDeviceProfile_BadRequest(t *testing.T) {
	dic := mockDic()

	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	deviceProfile := buildTestDeviceProfileRequest()
	badRequestId := deviceProfile
	badRequestId.RequestID = "niv3sl"
	noName := deviceProfile
	noName.Profile.Name = ""
	noDeviceResource := deviceProfile
	noDeviceResource.Profile.DeviceResources = []dtos.DeviceResource{}
	noDeviceResourceName := deviceProfile
	noDeviceResourceName.Profile.DeviceResources = []dtos.DeviceResource{{
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.PropertyValue{
			Type:      "INT16",
			ReadWrite: "RW",
		},
	}}
	noDeviceResourcePropertyType := deviceProfile
	noDeviceResourcePropertyType.Profile.DeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.PropertyValue{
			ReadWrite: "RW",
		},
	}}
	noCommandName := deviceProfile
	noCommandName.Profile.CoreCommands = []dtos.Command{{
		Get: true,
		Put: true,
	}}
	noCommandGet := deviceProfile
	noCommandGet.Profile.CoreCommands = []dtos.Command{{
		Name: TestProfileResourceName,
		Get:  false,
	}}
	noCommandPut := deviceProfile
	noCommandPut.Profile.CoreCommands = []dtos.Command{{
		Name: TestProfileResourceName,
		Put:  false,
	}}

	tests := []struct {
		name    string
		Request []requests.AddDeviceProfileRequest
	}{
		{"Invalid - Bad requestId", []requests.AddDeviceProfileRequest{badRequestId}},
		{"Invalid - Bad name", []requests.AddDeviceProfileRequest{noName}},
		{"Invalid - No deviceResource", []requests.AddDeviceProfileRequest{noDeviceResource}},
		{"Invalid - No deviceResource name", []requests.AddDeviceProfileRequest{noDeviceResourceName}},
		{"Invalid - No deviceResource property type", []requests.AddDeviceProfileRequest{noDeviceResourcePropertyType}},
		{"Invalid - No command name", []requests.AddDeviceProfileRequest{noCommandName}},
		{"Invalid - No command Get", []requests.AddDeviceProfileRequest{noCommandGet}},
		{"Invalid - No command Put", []requests.AddDeviceProfileRequest{noCommandPut}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiDeviceProfileRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfile)
			handler.ServeHTTP(recorder, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.NotEmpty(t, string(recorder.Body.Bytes()), "Message is empty")
		})
	}
}

func TestAddDeviceProfile_Duplicated(t *testing.T) {
	expectedRequestId := ExampleUUID

	duplicateIdRequest := buildTestDeviceProfileRequest()
	duplicateIdModel := requests.AddDeviceProfileReqToDeviceProfileModel(duplicateIdRequest)
	duplicateIdDBError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile id %s exists", duplicateIdModel.Id), nil)

	duplicateNameRequest := buildTestDeviceProfileRequest()
	duplicateNameRequest.Profile.Id = "" // The infrastructure layer will generate id when the id field is empty
	duplicateNameModel := requests.AddDeviceProfileReqToDeviceProfileModel(duplicateNameRequest)
	duplicateNameDBError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile name %s exists", duplicateNameModel.Name), nil)

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AddDeviceProfile", duplicateNameModel).Return(duplicateNameModel, duplicateNameDBError)
	dbClientMock.On("AddDeviceProfile", duplicateIdModel).Return(duplicateIdModel, duplicateIdDBError)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name          string
		request       []requests.AddDeviceProfileRequest
		expectedError errors.CommonEdgeX
	}{
		{"duplicate id", []requests.AddDeviceProfileRequest{duplicateIdRequest}, duplicateIdDBError},
		{"duplicate name", []requests.AddDeviceProfileRequest{duplicateNameRequest}, duplicateNameDBError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiDeviceProfileRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfile)
			handler.ServeHTTP(recorder, req)
			var res []common.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contractsV2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
			assert.Equal(t, expectedRequestId, res[0].RequestID, "RequestID not as expected")
			assert.Equal(t, http.StatusConflict, res[0].StatusCode, "BaseResponse status code not as expected")
			assert.Contains(t, res[0].Message, testCase.expectedError.Message(), "Message not as expected")
		})
	}
}

func TestAddDeviceProfileByYaml_Created(t *testing.T) {
	deviceProfileDTO := buildTestDeviceProfileRequest().Profile
	deviceProfileModel := dtos.ToDeviceProfileModel(deviceProfileDTO)
	expectedMessage := "Add device profiles successfully"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AddDeviceProfile", deviceProfileModel).Return(deviceProfileModel, nil)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	valid, err := yaml.Marshal(deviceProfileDTO)
	require.NoError(t, err)
	req, err := createDeviceProfileRequestWithFile(valid)
	require.NoError(t, err)

	// Act
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.AddDeviceProfileByYaml)
	handler.ServeHTTP(recorder, req)
	var res common.BaseWithIdResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusCreated, res.StatusCode, "BaseResponse status code not as expected")
	assert.Equal(t, expectedMessage, res.Message, "Message not as expected")
}

func TestAddDeviceProfileByYaml_BadRequest(t *testing.T) {
	dic := mockDic()

	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	deviceProfile := buildTestDeviceProfileRequest().Profile
	noName := deviceProfile
	noName.Name = ""
	noDeviceResource := deviceProfile
	noDeviceResource.DeviceResources = []dtos.DeviceResource{}
	noDeviceResourceName := deviceProfile
	noDeviceResourceName.DeviceResources = []dtos.DeviceResource{{
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.PropertyValue{
			Type:      "INT16",
			ReadWrite: "RW",
		},
	}}
	noDeviceResourcePropertyType := deviceProfile
	noDeviceResourcePropertyType.DeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.PropertyValue{
			ReadWrite: "RW",
		},
	}}
	noCommandName := deviceProfile
	noCommandName.CoreCommands = []dtos.Command{{
		Get: true,
		Put: true,
	}}
	noCommandGet := deviceProfile
	noCommandGet.CoreCommands = []dtos.Command{{
		Name: TestProfileResourceName,
		Get:  false,
	}}
	noCommandPut := deviceProfile
	noCommandPut.CoreCommands = []dtos.Command{{
		Name: TestProfileResourceName,
		Put:  false,
	}}
	tests := []struct {
		name    string
		Request dtos.DeviceProfile
	}{
		{"Invalid - No name", noName},
		{"Invalid - No deviceResource", noDeviceResource},
		{"Invalid - No deviceResource name", noDeviceResourceName},
		{"Invalid - No deviceResource property type", noDeviceResourcePropertyType},
		{"Invalid - No command name", noCommandName},
		{"Invalid - No command Get", noCommandGet},
		{"Invalid - No command Put", noCommandPut},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			valid, err := yaml.Marshal(testCase.Request)
			require.NoError(t, err)
			req, err := createDeviceProfileRequestWithFile(valid)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfileByYaml)
			handler.ServeHTTP(recorder, req)
			var res common.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)

			// Assert
			assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, http.StatusBadRequest, res.StatusCode, "HTTP status code not as expected")
			assert.NotEmpty(t, string(recorder.Body.Bytes()), "Message is empty")
		})
	}
}

func TestAddDeviceProfileByYaml_Duplicated(t *testing.T) {
	deviceProfileDTO := buildTestDeviceProfileRequest().Profile
	deviceProfileModel := dtos.ToDeviceProfileModel(deviceProfileDTO)
	dbError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile %s already exists", TestDeviceProfileName), nil)

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AddDeviceProfile", deviceProfileModel).Return(deviceProfileModel, dbError)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	valid, err := yaml.Marshal(deviceProfileDTO)
	require.NoError(t, err)
	req, err := createDeviceProfileRequestWithFile(valid)
	require.NoError(t, err)

	// Act
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.AddDeviceProfileByYaml)
	handler.ServeHTTP(recorder, req)
	var res common.BaseWithIdResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusConflict, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusConflict, res.StatusCode, "BaseResponse status code not as expected")
	assert.Contains(t, res.Message, dbError.Message(), "Message not as expected")
}

func TestAddDeviceProfileByYaml_MissingFile(t *testing.T) {
	deviceProfileDTO := buildTestDeviceProfileRequest().Profile
	dic := mockDic()

	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	valid, err := yaml.Marshal(deviceProfileDTO)
	require.NoError(t, err)
	req, err := createDeviceProfileRequestWithFile(valid)
	require.NoError(t, err)

	req.MultipartForm = new(multipart.Form)
	req.MultipartForm.File = nil

	// Act
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.AddDeviceProfileByYaml)
	handler.ServeHTTP(recorder, req)
	var res common.BaseWithIdResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusBadRequest, res.StatusCode, "BaseResponse status code not as expected")
	assert.Contains(t, res.Message, "missing yaml file")
}

func TestUpdateDeviceProfileByYaml(t *testing.T) {
	deviceProfile := buildTestDeviceProfileRequest().Profile

	valid := deviceProfile
	validDeviceProfileModel := dtos.ToDeviceProfileModel(valid)
	noName := deviceProfile
	noName.Name = ""
	noDeviceResource := deviceProfile
	noDeviceResource.DeviceResources = []dtos.DeviceResource{}
	noDeviceResourceName := deviceProfile
	noDeviceResourceName.DeviceResources = []dtos.DeviceResource{{
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.PropertyValue{
			Type:      "INT16",
			ReadWrite: "RW",
		},
	}}
	noDeviceResourcePropertyType := deviceProfile
	noDeviceResourcePropertyType.DeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.PropertyValue{
			ReadWrite: "RW",
		},
	}}
	noCommandName := deviceProfile
	noCommandName.CoreCommands = []dtos.Command{{
		Get: true,
		Put: true,
	}}
	noCommandGet := deviceProfile
	noCommandGet.CoreCommands = []dtos.Command{{
		Name: TestProfileResourceName,
		Get:  false,
	}}
	noCommandPut := deviceProfile
	noCommandPut.CoreCommands = []dtos.Command{{
		Name: TestProfileResourceName,
		Put:  false,
	}}
	notFound := deviceProfile
	notFound.Name = "testDevice"
	notFoundDeviceProfileModel := dtos.ToDeviceProfileModel(notFound)
	notFoundDBError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile %s does not exists", notFoundDeviceProfileModel.Name), nil)

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("UpdateDeviceProfile", validDeviceProfileModel).Return(nil)
	dbClientMock.On("UpdateDeviceProfile", notFoundDeviceProfileModel).Return(notFoundDBError)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceProfileController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		request            dtos.DeviceProfile
		expectedStatusCode int
	}{
		{"Valid", valid, http.StatusOK},
		{"Invalid - No name", noName, http.StatusBadRequest},
		{"Invalid - No deviceResource", noDeviceResource, http.StatusBadRequest},
		{"Invalid - No deviceResource name", noDeviceResourceName, http.StatusBadRequest},
		{"Invalid - No deviceResource property type", noDeviceResourcePropertyType, http.StatusBadRequest},
		{"Invalid - No command name", noCommandName, http.StatusBadRequest},
		{"Invalid - No command Get", noCommandGet, http.StatusBadRequest},
		{"Invalid - No command Put", noCommandPut, http.StatusBadRequest},
		{"Not found", notFound, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			valid, err := yaml.Marshal(testCase.request)
			require.NoError(t, err)
			req, err := createDeviceProfileRequestWithFile(valid)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.UpdateDeviceProfileByYaml)
			handler.ServeHTTP(recorder, req)
			var res common.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "HTTP status code not as expected")
			assert.NotEmpty(t, string(recorder.Body.Bytes()), "Message is empty")
		})
	}
}
