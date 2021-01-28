//
// Copyright (C) 2020-2021 IOTech Ltd
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

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDeviceProfileLabels = []string{"MODBUS", "TEMP"}
var testAttributes = map[string]string{
	"TestAttribute": "TestAttributeValue",
}

func buildTestDeviceProfileRequest() requests.DeviceProfileRequest {
	var testDeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.PropertyValue{
			Type:      contractsV2.ValueTypeInt16,
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

	var testDeviceProfileReq = requests.DeviceProfileRequest{
		BaseRequest: common.BaseRequest{
			RequestId: ExampleUUID,
		},
		Profile: dtos.DeviceProfile{
			Id:              ExampleUUID,
			Name:            TestDeviceProfileName,
			Manufacturer:    TestManufacturer,
			Description:     TestDescription,
			Model:           TestModel,
			Labels:          testDeviceProfileLabels,
			DeviceResources: testDeviceResources,
			DeviceCommands:  testDeviceCommands,
			CoreCommands:    testCoreCommands,
		},
	}

	return testDeviceProfileReq
}

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		metadataContainer.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
				Service: bootstrapConfig.ServiceInfo{
					MaxResultCount: 30,
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
	deviceProfileModel := requests.DeviceProfileReqToDeviceProfileModel(deviceProfileRequest)
	expectedRequestId := ExampleUUID

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
	noRequestId.RequestId = ""

	tests := []struct {
		name    string
		Request []requests.DeviceProfileRequest
	}{
		{"Valid - AddDeviceProfileRequest", []requests.DeviceProfileRequest{valid}},
		{"Valid - No requestId", []requests.DeviceProfileRequest{noRequestId}},
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
			if res[0].RequestId != "" {
				assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
			}
			assert.Equal(t, http.StatusCreated, res[0].StatusCode, "BaseResponse status code not as expected")
			assert.Empty(t, res[0].Message, "Message should be empty when it is successful")
		})
	}
}

func TestAddDeviceProfile_BadRequest(t *testing.T) {
	dic := mockDic()

	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	deviceProfile := buildTestDeviceProfileRequest()
	badRequestId := deviceProfile
	badRequestId.RequestId = "niv3sl"
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
		Request []requests.DeviceProfileRequest
	}{
		{"Invalid - Bad requestId", []requests.DeviceProfileRequest{badRequestId}},
		{"Invalid - Bad name", []requests.DeviceProfileRequest{noName}},
		{"Invalid - No deviceResource", []requests.DeviceProfileRequest{noDeviceResource}},
		{"Invalid - No deviceResource name", []requests.DeviceProfileRequest{noDeviceResourceName}},
		{"Invalid - No deviceResource property type", []requests.DeviceProfileRequest{noDeviceResourcePropertyType}},
		{"Invalid - No command name", []requests.DeviceProfileRequest{noCommandName}},
		{"Invalid - No command Get", []requests.DeviceProfileRequest{noCommandGet}},
		{"Invalid - No command Put", []requests.DeviceProfileRequest{noCommandPut}},
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
	duplicateIdModel := requests.DeviceProfileReqToDeviceProfileModel(duplicateIdRequest)
	duplicateIdDBError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile id %s exists", duplicateIdModel.Id), nil)

	duplicateNameRequest := buildTestDeviceProfileRequest()
	duplicateNameRequest.Profile.Id = "" // The infrastructure layer will generate id when the id field is empty
	duplicateNameModel := requests.DeviceProfileReqToDeviceProfileModel(duplicateNameRequest)
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
		request       []requests.DeviceProfileRequest
		expectedError errors.CommonEdgeX
	}{
		{"duplicate id", []requests.DeviceProfileRequest{duplicateIdRequest}, duplicateIdDBError},
		{"duplicate name", []requests.DeviceProfileRequest{duplicateNameRequest}, duplicateNameDBError},
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
			assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
			assert.Equal(t, http.StatusConflict, res[0].StatusCode, "BaseResponse status code not as expected")
			assert.Contains(t, res[0].Message, testCase.expectedError.Message(), "Message not as expected")
		})
	}
}

func TestUpdateDeviceProfile(t *testing.T) {
	deviceProfileRequest := buildTestDeviceProfileRequest()
	deviceProfileModel := requests.DeviceProfileReqToDeviceProfileModel(deviceProfileRequest)
	expectedRequestId := ExampleUUID

	valid := deviceProfileRequest
	noRequestId := deviceProfileRequest
	noRequestId.RequestId = ""
	noName := deviceProfileRequest
	noName.Profile.Name = ""
	noDeviceResource := deviceProfileRequest
	noDeviceResource.Profile.DeviceResources = []dtos.DeviceResource{}
	noDeviceResourceName := deviceProfileRequest
	noDeviceResourceName.Profile.DeviceResources = []dtos.DeviceResource{{
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.PropertyValue{
			Type:      "INT16",
			ReadWrite: "RW",
		},
	}}
	noDeviceResourcePropertyType := deviceProfileRequest
	noDeviceResourcePropertyType.Profile.DeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Tag:         TestTag,
		Attributes:  testAttributes,
		Properties: dtos.PropertyValue{
			ReadWrite: "RW",
		},
	}}
	noCommandName := deviceProfileRequest
	noCommandName.Profile.CoreCommands = []dtos.Command{{
		Get: true,
		Put: true,
	}}
	noCommandGet := deviceProfileRequest
	noCommandGet.Profile.CoreCommands = []dtos.Command{{
		Name: TestProfileResourceName,
		Get:  false,
	}}
	noCommandPut := deviceProfileRequest
	noCommandPut.Profile.CoreCommands = []dtos.Command{{
		Name: TestProfileResourceName,
		Put:  false,
	}}
	notFound := deviceProfileRequest
	notFound.Profile.Name = "testDevice"
	notFoundDeviceProfileModel := dtos.ToDeviceProfileModel(notFound.Profile)
	notFoundDBError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile %s does not exists", notFound.Profile.Name), nil)

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("UpdateDeviceProfile", deviceProfileModel).Return(nil)
	dbClientMock.On("UpdateDeviceProfile", notFoundDeviceProfileModel).Return(notFoundDBError)
	dbClientMock.On("DevicesByProfileName", 0, -1, deviceProfileModel.Name).Return([]models.Device{{ServiceName: testDeviceServiceName}}, nil)
	dbClientMock.On("DeviceServiceByName", testDeviceServiceName).Return(models.DeviceService{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceProfileController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		request            []requests.DeviceProfileRequest
		expectedStatusCode int
	}{
		{"Valid - DeviceProfileRequest", []requests.DeviceProfileRequest{valid}, http.StatusOK},
		{"Invalid - No name", []requests.DeviceProfileRequest{noName}, http.StatusBadRequest},
		{"Invalid - No deviceResource", []requests.DeviceProfileRequest{noDeviceResource}, http.StatusBadRequest},
		{"Invalid - No deviceResource name", []requests.DeviceProfileRequest{noDeviceResourceName}, http.StatusBadRequest},
		{"Invalid - No deviceResource property type", []requests.DeviceProfileRequest{noDeviceResourcePropertyType}, http.StatusBadRequest},
		{"Invalid - No command name", []requests.DeviceProfileRequest{noCommandName}, http.StatusBadRequest},
		{"Invalid - No command Get", []requests.DeviceProfileRequest{noCommandGet}, http.StatusBadRequest},
		{"Invalid - No command Put", []requests.DeviceProfileRequest{noCommandPut}, http.StatusBadRequest},
		{"Valid - No requestId", []requests.DeviceProfileRequest{noRequestId}, http.StatusOK},
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
			handler := http.HandlerFunc(controller.UpdateDeviceProfile)
			handler.ServeHTTP(recorder, req)

			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, string(recorder.Body.Bytes()), "Message is empty")
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
				assert.NotEmpty(t, string(recorder.Body.Bytes()), "Message is empty")
			}
		})
	}
}

func TestAddDeviceProfileByYaml_Created(t *testing.T) {
	deviceProfileDTO := buildTestDeviceProfileRequest().Profile
	deviceProfileModel := dtos.ToDeviceProfileModel(deviceProfileDTO)

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
	assert.Empty(t, res.Message, "Message should be empty when it is successful")
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
	dbClientMock.On("DevicesByProfileName", 0, -1, validDeviceProfileModel.Name).Return([]models.Device{{ServiceName: testDeviceServiceName}}, nil)
	dbClientMock.On("DeviceServiceByName", testDeviceServiceName).Return(models.DeviceService{}, nil)
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

func TestDeviceProfileByName(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceProfileByName", deviceProfile.Name).Return(deviceProfile, nil)
	dbClientMock.On("DeviceProfileByName", notFoundName).Return(models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device profile doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceProfileName  string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - find device profile by name", deviceProfile.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - device profile not found by name", notFoundName, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s/%s", contractsV2.ApiDeviceProfileRoute, contractsV2.Name, testCase.deviceProfileName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{contractsV2.Name: testCase.deviceProfileName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceProfileByName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.DeviceProfileResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.deviceProfileName, res.Profile.Name, "Event Id not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeleteDeviceProfileByName(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	noName := ""
	notFoundName := "notFoundName"
	deviceExists := "deviceExists"
	provisionWatcherExists := "provisionWatcherExists"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DevicesByProfileName", 0, 1, deviceProfile.Name).Return([]models.Device{}, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 0, 1, deviceProfile.Name).Return([]models.ProvisionWatcher{}, nil)
	dbClientMock.On("DeleteDeviceProfileByName", deviceProfile.Name).Return(nil)
	dbClientMock.On("DevicesByProfileName", 0, 1, notFoundName).Return([]models.Device{}, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 0, 1, notFoundName).Return([]models.ProvisionWatcher{}, nil)
	dbClientMock.On("DeleteDeviceProfileByName", notFoundName).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device profile doesn't exist in the database", nil))
	dbClientMock.On("DevicesByProfileName", 0, 1, deviceExists).Return([]models.Device{models.Device{}}, nil)
	dbClientMock.On("DevicesByProfileName", 0, 1, provisionWatcherExists).Return([]models.Device{}, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 0, 1, provisionWatcherExists).Return([]models.ProvisionWatcher{models.ProvisionWatcher{}}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceProfileController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceProfileName  string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - delete device profile by name", deviceProfile.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", noName, true, http.StatusBadRequest},
		{"Invalid - device profile not found by name", notFoundName, true, http.StatusNotFound},
		{"Invalid - associated device exists", deviceExists, true, http.StatusBadRequest},
		{"Invalid - associated provisionWatcher Exists", provisionWatcherExists, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s/%s", contractsV2.ApiDeviceProfileRoute, contractsV2.Name, testCase.deviceProfileName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{contractsV2.Name: testCase.deviceProfileName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteDeviceProfileByName)
			handler.ServeHTTP(recorder, req)
			var res common.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
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

func TestAllDeviceProfiles(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	deviceProfiles := []models.DeviceProfile{deviceProfile, deviceProfile, deviceProfile}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AllDeviceProfiles", 0, 10, []string(nil)).Return(deviceProfiles, nil)
	dbClientMock.On("AllDeviceProfiles", 0, 5, testDeviceProfileLabels).Return([]models.DeviceProfile{deviceProfiles[0], deviceProfiles[1]}, nil)
	dbClientMock.On("AllDeviceProfiles", 1, 2, []string(nil)).Return([]models.DeviceProfile{deviceProfiles[1], deviceProfiles[2]}, nil)
	dbClientMock.On("AllDeviceProfiles", 4, 1, testDeviceProfileLabels).Return([]models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		labels             string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get device profiles without labels", "0", "10", "", false, 3, http.StatusOK},
		{"Valid - get device profiles with labels", "0", "5", strings.Join(testDeviceProfileLabels, ","), false, 2, http.StatusOK},
		{"Valid - get device profiles with offset and no labels", "1", "2", "", false, 2, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", strings.Join(testDeviceProfileLabels, ","), true, 0, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, contractsV2.ApiAllDeviceProfileRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(contractsV2.Offset, testCase.offset)
			query.Add(contractsV2.Limit, testCase.limit)
			if len(testCase.labels) > 0 {
				query.Add(contractsV2.Labels, testCase.labels)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllDeviceProfiles)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiDeviceProfilesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Profiles), "Profile count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeviceProfilesByModel(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	deviceProfiles := []models.DeviceProfile{deviceProfile, deviceProfile, deviceProfile}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceProfilesByModel", 0, 10, TestModel).Return(deviceProfiles, nil)
	dbClientMock.On("DeviceProfilesByModel", 1, 2, TestModel).Return([]models.DeviceProfile{deviceProfiles[1], deviceProfiles[2]}, nil)
	dbClientMock.On("DeviceProfilesByModel", 4, 1, TestModel).Return([]models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		model              string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get device profiles by model", "0", "10", TestModel, false, 3, http.StatusOK},
		{"Valid - get device profiles by model with offset and limit", "1", "2", TestModel, false, 2, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", TestModel, true, 0, http.StatusNotFound},
		{"Invalid - model is empty", "0", "10", "", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, contractsV2.ApiDeviceProfileByModelRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{contractsV2.Model: testCase.model})
			query := req.URL.Query()
			query.Add(contractsV2.Offset, testCase.offset)
			query.Add(contractsV2.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceProfilesByModel)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiDeviceProfilesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Profiles), "Profile count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeviceProfilesByManufacturer(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	deviceProfiles := []models.DeviceProfile{deviceProfile, deviceProfile, deviceProfile}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceProfilesByManufacturer", 0, 10, TestManufacturer).Return(deviceProfiles, nil)
	dbClientMock.On("DeviceProfilesByManufacturer", 1, 2, TestManufacturer).Return([]models.DeviceProfile{deviceProfiles[1], deviceProfiles[2]}, nil)
	dbClientMock.On("DeviceProfilesByManufacturer", 4, 1, TestManufacturer).Return([]models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		manufacturer       string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get device profiles by manufacturer", "0", "10", TestManufacturer, false, 3, http.StatusOK},
		{"Valid - get device profiles by manufacturer with offset and limit", "1", "2", TestManufacturer, false, 2, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", TestManufacturer, true, 0, http.StatusNotFound},
		{"Invalid - manufacturer is empty", "0", "10", "", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, contractsV2.ApiDeviceProfileByManufacturerRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{contractsV2.Manufacturer: testCase.manufacturer})
			query := req.URL.Query()
			query.Add(contractsV2.Offset, testCase.offset)
			query.Add(contractsV2.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceProfilesByManufacturer)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiDeviceProfilesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Profiles), "Profile count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeviceProfilesByManufacturerAndModel(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	deviceProfiles := []models.DeviceProfile{deviceProfile, deviceProfile, deviceProfile}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeviceProfilesByManufacturerAndModel", 0, 10, TestManufacturer, TestModel).Return(deviceProfiles, nil)
	dbClientMock.On("DeviceProfilesByManufacturerAndModel", 1, 2, TestManufacturer, TestModel).Return([]models.DeviceProfile{deviceProfiles[1], deviceProfiles[2]}, nil)
	dbClientMock.On("DeviceProfilesByManufacturerAndModel", 4, 1, TestManufacturer, TestModel).Return([]models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		manufacturer       string
		model              string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get device profiles by manufacturer and model", "0", "10", TestManufacturer, TestModel, false, 3, http.StatusOK},
		{"Valid - get device profiles by manufacturer with offset and limit", "1", "2", TestManufacturer, TestModel, false, 2, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", TestManufacturer, TestModel, true, 0, http.StatusNotFound},
		{"Invalid - manufacturer is empty", "0", "10", "", TestModel, true, 0, http.StatusBadRequest},
		{"Invalid - model is empty", "0", "10", TestManufacturer, "", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, contractsV2.ApiDeviceProfileByManufacturerAndModelRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{contractsV2.Manufacturer: testCase.manufacturer, contractsV2.Model: testCase.model})
			query := req.URL.Query()
			query.Add(contractsV2.Offset, testCase.offset)
			query.Add(contractsV2.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceProfilesByManufacturerAndModel)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiDeviceProfilesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, contractsV2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Profiles), "Profile count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
