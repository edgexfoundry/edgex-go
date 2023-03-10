//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces/mocks"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDeviceProfileLabels = []string{"MODBUS", "TEMP"}
var testAttributes = map[string]interface{}{
	"TestAttribute": "TestAttributeValue",
}
var testMappings = map[string]string{"0": "off", "1": "on"}
var testTags = map[string]any{
	"TestTagKey": "TestTagValue",
}
var testOptional = map[string]any{"TestOptionalKey": "TestOptionalValue"}

func buildTestDeviceProfileRequest() requests.DeviceProfileRequest {
	var testDeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ValueType: common.ValueTypeInt16,
			ReadWrite: common.ReadWrite_RW,
			Units:     TestUnits,
			Optional:  testOptional,
		},
		Tags: testTags,
	}, {
		Name:        TestDeviceResourceName + "-dup",
		Description: TestDescription,
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ValueType: common.ValueTypeInt16,
			ReadWrite: common.ReadWrite_RW,
			Units:     TestUnits,
			Optional:  testOptional,
		},
		Tags: testTags,
	}}
	var testDeviceCommands = []dtos.DeviceCommand{{
		Name:      TestDeviceCommandName,
		ReadWrite: common.ReadWrite_RW,
		ResourceOperations: []dtos.ResourceOperation{{
			DeviceResource: TestDeviceResourceName,
			Mappings:       testMappings,
		}},
	}}

	var testDeviceProfileReq = requests.DeviceProfileRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   ExampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		Profile: dtos.DeviceProfile{
			DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{
				Id:           ExampleUUID,
				Name:         TestDeviceProfileName,
				Manufacturer: TestManufacturer,
				Description:  TestDescription,
				Model:        TestModel,
				Labels:       testDeviceProfileLabels,
			},
			DeviceResources: testDeviceResources,
			DeviceCommands:  testDeviceCommands,
		},
	}

	return testDeviceProfileReq
}

func buildTestDeviceProfileBasicInfoRequest() requests.DeviceProfileBasicInfoRequest {
	testUUID := ExampleUUID
	testName := TestDeviceProfileName
	testDescription := TestDescription
	testManufacturer := TestManufacturer
	testModel := TestModel

	var testBasicInfoReq = requests.DeviceProfileBasicInfoRequest{
		BaseRequest: commonDTO.BaseRequest{
			Versionable: commonDTO.NewVersionable(),
			RequestId:   ExampleUUID,
		},
		BasicInfo: dtos.UpdateDeviceProfileBasicInfo{
			Id:           &testUUID,
			Name:         &testName,
			Manufacturer: &testManufacturer,
			Description:  &testDescription,
			Model:        &testModel,
			Labels:       testDeviceProfileLabels,
		},
	}

	return testBasicInfoReq
}

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
				Service: bootstrapConfig.ServiceInfo{
					RequestTimeout: "30s",
					MaxResultCount: 30,
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
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
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(http.MethodPost, common.ApiDeviceProfileRoute+"/uploadfile", body)
	req.Header.Set(common.ContentType, "multipart/form-data; boundary="+boundary)
	return req, nil
}

func TestAddDeviceProfile_Created(t *testing.T) {
	deviceProfileRequest := buildTestDeviceProfileRequest()
	deviceProfileModel := requests.DeviceProfileReqToDeviceProfileModel(deviceProfileRequest)
	expectedRequestId := ExampleUUID

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddDeviceProfile", deviceProfileModel).Return(deviceProfileModel, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceProfileRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfile)
			handler.ServeHTTP(recorder, req)
			var res []commonDTO.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
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
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ValueType: "INT16",
			ReadWrite: common.ReadWrite_RW,
		},
		Tags: testTags,
	}}
	noDeviceResourcePropertyType := deviceProfile
	noDeviceResourcePropertyType.Profile.DeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ReadWrite: common.ReadWrite_RW,
		},
		Tags: testTags,
	}}
	noCommandName := deviceProfile
	noCommandName.Profile.DeviceCommands = []dtos.DeviceCommand{{
		ReadWrite: common.ReadWrite_RW,
	}}
	noCommandReadWrite := deviceProfile
	noCommandReadWrite.Profile.DeviceCommands = []dtos.DeviceCommand{{
		Name: TestDeviceCommandName,
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
		{"Invalid - No command readWrite", []requests.DeviceProfileRequest{noCommandReadWrite}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceProfileRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfile)
			handler.ServeHTTP(recorder, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
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
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddDeviceProfile", duplicateNameModel).Return(duplicateNameModel, duplicateNameDBError)
	dbClientMock.On("AddDeviceProfile", duplicateIdModel).Return(duplicateIdModel, duplicateIdDBError)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceProfileRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfile)
			handler.ServeHTTP(recorder, req)
			var res []commonDTO.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
			assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
			assert.Equal(t, http.StatusConflict, res[0].StatusCode, "BaseResponse status code not as expected")
			assert.Contains(t, res[0].Message, testCase.expectedError.Message(), "Message not as expected")
		})
	}
}

func TestAddDeviceProfile_UnitsOfMeasure_Validation(t *testing.T) {
	deviceProfileRequest := buildTestDeviceProfileRequest()
	deviceProfileModel := requests.DeviceProfileReqToDeviceProfileModel(deviceProfileRequest)
	expectedRequestId := ExampleUUID

	emptyUnits := ""
	invalidUnits := "invalid"
	validReq := deviceProfileRequest
	emptyUnitsReq := buildTestDeviceProfileRequest()
	for i := range emptyUnitsReq.Profile.DeviceResources {
		emptyUnitsReq.Profile.DeviceResources[i].Properties.Units = emptyUnits
	}
	noUnitsModel := requests.DeviceProfileReqToDeviceProfileModel(emptyUnitsReq)
	invalidUnitsReq := buildTestDeviceProfileRequest()
	for i := range invalidUnitsReq.Profile.DeviceResources {
		invalidUnitsReq.Profile.DeviceResources[i].Properties.Units = invalidUnits
	}

	dic := mockDic()
	container.ConfigurationFrom(dic.Get).Writable.UoM.Validation = true
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddDeviceProfile", deviceProfileModel).Return(deviceProfileModel, nil)
	dbClientMock.On("AddDeviceProfile", noUnitsModel).Return(noUnitsModel, nil)
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

	controller := NewDeviceProfileController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		Request            []requests.DeviceProfileRequest
		expectedStatusCode int
	}{
		{"valid - expected units", []requests.DeviceProfileRequest{validReq}, http.StatusCreated},
		{"valid - units not provided", []requests.DeviceProfileRequest{emptyUnitsReq}, http.StatusCreated},
		{"invalid - unexpected units", []requests.DeviceProfileRequest{invalidUnitsReq}, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceProfileRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceProfile)
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
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ValueType: "INT16",
			ReadWrite: common.ReadWrite_RW,
		},
		Tags: testTags,
	}}
	noDeviceResourcePropertyType := deviceProfileRequest
	noDeviceResourcePropertyType.Profile.DeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ReadWrite: common.ReadWrite_RW,
		},
		Tags: testTags,
	}}
	noCommandName := deviceProfileRequest
	noCommandName.Profile.DeviceCommands = []dtos.DeviceCommand{{
		ReadWrite: common.ReadWrite_RW,
	}}
	noCommandReadWrite := deviceProfileRequest
	noCommandReadWrite.Profile.DeviceCommands = []dtos.DeviceCommand{{
		Name: TestDeviceCommandName,
	}}
	notFound := deviceProfileRequest
	notFound.Profile.Name = "testDevice"
	notFoundDeviceProfileModel := dtos.ToDeviceProfileModel(notFound.Profile)
	notFoundDBError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile %s does not exists", notFound.Profile.Name), nil)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("UpdateDeviceProfile", deviceProfileModel).Return(nil)
	dbClientMock.On("UpdateDeviceProfile", notFoundDeviceProfileModel).Return(notFoundDBError)
	dbClientMock.On("DeviceCountByProfileName", deviceProfileModel.Name).Return(uint32(1), nil)
	dbClientMock.On("DevicesByProfileName", 0, -1, deviceProfileModel.Name).Return([]models.Device{{ServiceName: testDeviceServiceName}}, nil)
	dbClientMock.On("DeviceServiceByName", testDeviceServiceName).Return(models.DeviceService{}, nil)
	dbClientMock.On("DeviceProfileByName", deviceProfileModel.Name).Return(deviceProfileModel, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		{"Invalid - No command readWrite", []requests.DeviceProfileRequest{noCommandReadWrite}, http.StatusBadRequest},
		{"Valid - No requestId", []requests.DeviceProfileRequest{noRequestId}, http.StatusOK},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceProfileRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.UpdateDeviceProfile)
			handler.ServeHTTP(recorder, req)

			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
			} else {
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
			}
		})
	}
}

func TestUpdateDeviceProfile_StrictProfileChanges(t *testing.T) {
	valid := buildTestDeviceProfileRequest()

	dic := mockDic()
	configuration := container.ConfigurationFrom(dic.Get)
	configuration.Writable.ProfileChange.StrictDeviceProfileChanges = true
	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	controller := NewDeviceProfileController(dic)
	require.NotNil(t, controller)

	jsonData, err := json.Marshal(valid)
	require.NoError(t, err)

	reader := strings.NewReader(string(jsonData))
	req, err := http.NewRequest(http.MethodPost, common.ApiDeviceProfileRoute, reader)
	require.NoError(t, err)

	// Act
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.UpdateDeviceProfile)
	handler.ServeHTTP(recorder, req)

	var res commonDTO.BaseWithIdResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusLocked, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusLocked, res.StatusCode, "HTTP status code not as expected")
	assert.NotEmpty(t, recorder.Body.String(), "Message is empty")

}

func TestPatchDeviceProfileBasicInfo(t *testing.T) {
	expectedRequestId := ExampleUUID
	testReq := buildTestDeviceProfileBasicInfoRequest()
	dpModel := requests.DeviceProfileReqToDeviceProfileModel(buildTestDeviceProfileRequest())

	valid := testReq
	noRequestId := testReq
	noRequestId.RequestId = ""
	noName := testReq
	noName.BasicInfo.Name = nil
	noId := testReq
	noId.BasicInfo.Id = nil

	noIdAndName := testReq
	noIdAndName.BasicInfo.Id = nil
	noIdAndName.BasicInfo.Name = nil
	emptyName := testReq
	emptyString := ""
	emptyName.BasicInfo.Name = &emptyString
	notFound := testReq
	notFound.BasicInfo.Id = nil
	notFoundName := "testDevice"
	notFound.BasicInfo.Name = &notFoundName

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileById", *valid.BasicInfo.Id).Return(dpModel, nil)
	dbClientMock.On("DeviceProfileByName", *valid.BasicInfo.Name).Return(dpModel, nil)
	dbClientMock.On("DeviceProfileByName", notFoundName).Return(dpModel, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "not found", nil))
	dbClientMock.On("UpdateDeviceProfile", mock.Anything).Return(nil)
	dbClientMock.On("DeviceCountByProfileName", *valid.BasicInfo.Name).Return(uint32(1), nil)
	dbClientMock.On("DevicesByProfileName", 0, -1, *valid.BasicInfo.Name).Return([]models.Device{{ServiceName: testDeviceServiceName}}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewDeviceProfileController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		request            []requests.DeviceProfileBasicInfoRequest
		expectedStatusCode int
	}{
		{"valid", []requests.DeviceProfileBasicInfoRequest{valid}, http.StatusOK},
		{"valid - no request id", []requests.DeviceProfileBasicInfoRequest{noRequestId}, http.StatusOK},
		{"valid - no name", []requests.DeviceProfileBasicInfoRequest{noName}, http.StatusOK},
		{"invalid - no id and name", []requests.DeviceProfileBasicInfoRequest{noIdAndName}, http.StatusBadRequest},
		{"invalid - empty name", []requests.DeviceProfileBasicInfoRequest{emptyName}, http.StatusBadRequest},
		{"invalid - device profile not found", []requests.DeviceProfileBasicInfoRequest{notFound}, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPatch, common.ApiDeviceProfileBasicInfoRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.PatchDeviceProfileBasicInfo)
			handler.ServeHTTP(recorder, req)

			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
			} else {
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
			}
		})
	}
}

func TestAddDeviceProfileByYaml_Created(t *testing.T) {
	deviceProfileDTO := buildTestDeviceProfileRequest().Profile
	deviceProfileModel := dtos.ToDeviceProfileModel(deviceProfileDTO)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddDeviceProfile", deviceProfileModel).Return(deviceProfileModel, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
	var res commonDTO.BaseWithIdResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
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
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ValueType: "INT16",
			ReadWrite: common.ReadWrite_RW,
		},
		Tags: testTags,
	}}
	noDeviceResourcePropertyType := deviceProfile
	noDeviceResourcePropertyType.DeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ReadWrite: common.ReadWrite_RW,
		},
		Tags: testTags,
	}}
	noCommandName := deviceProfile
	noCommandName.DeviceCommands = []dtos.DeviceCommand{{
		ReadWrite: common.ReadWrite_RW,
	}}
	noCommandReadWrite := deviceProfile
	noCommandReadWrite.DeviceCommands = []dtos.DeviceCommand{{
		Name: TestDeviceCommandName,
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
		{"Invalid - No command readWrite", noCommandReadWrite},
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
			var res commonDTO.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			assert.NoError(t, err)

			// Assert
			assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, http.StatusBadRequest, res.StatusCode, "HTTP status code not as expected")
			assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
		})
	}
}

func TestAddDeviceProfileByYaml_Duplicated(t *testing.T) {
	deviceProfileDTO := buildTestDeviceProfileRequest().Profile
	deviceProfileModel := dtos.ToDeviceProfileModel(deviceProfileDTO)
	dbError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile %s already exists", TestDeviceProfileName), nil)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddDeviceProfile", deviceProfileModel).Return(deviceProfileModel, dbError)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
	var res commonDTO.BaseWithIdResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusConflict, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
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
	var res commonDTO.BaseWithIdResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
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
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ValueType: "INT16",
			ReadWrite: common.ReadWrite_RW,
		},
		Tags: testTags,
	}}
	noDeviceResourcePropertyType := deviceProfile
	noDeviceResourcePropertyType.DeviceResources = []dtos.DeviceResource{{
		Name:        TestDeviceResourceName,
		Description: TestDescription,
		Attributes:  testAttributes,
		Properties: dtos.ResourceProperties{
			ReadWrite: common.ReadWrite_RW,
		},
		Tags: testTags,
	}}
	noCommandName := deviceProfile
	noCommandName.DeviceCommands = []dtos.DeviceCommand{{
		ReadWrite: common.ReadWrite_RW,
	}}
	noCommandReadWrite := deviceProfile
	noCommandReadWrite.DeviceCommands = []dtos.DeviceCommand{{
		Name: TestDeviceCommandName,
	}}
	notFound := deviceProfile
	notFound.Name = "testDevice"
	notFoundDeviceProfileModel := dtos.ToDeviceProfileModel(notFound)
	notFoundDBError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile %s does not exists", notFoundDeviceProfileModel.Name), nil)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("UpdateDeviceProfile", validDeviceProfileModel).Return(nil)
	dbClientMock.On("UpdateDeviceProfile", notFoundDeviceProfileModel).Return(notFoundDBError)
	dbClientMock.On("DeviceCountByProfileName", validDeviceProfileModel.Name).Return(uint32(1), nil)
	dbClientMock.On("DevicesByProfileName", 0, -1, validDeviceProfileModel.Name).Return([]models.Device{{ServiceName: testDeviceServiceName}}, nil)
	dbClientMock.On("DeviceServiceByName", testDeviceServiceName).Return(models.DeviceService{}, nil)
	dbClientMock.On("DeviceProfileByName", validDeviceProfileModel.Name).Return(validDeviceProfileModel, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		{"Invalid - No command readWrite", noCommandReadWrite, http.StatusBadRequest},
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
			var res commonDTO.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			assert.NoError(t, err)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "HTTP status code not as expected")
			assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
		})
	}
}

func TestUpdateDeviceProfileByYaml_StrictProfileChanges(t *testing.T) {
	valid := buildTestDeviceProfileRequest().Profile

	dic := mockDic()
	configuration := container.ConfigurationFrom(dic.Get)
	configuration.Writable.ProfileChange.StrictDeviceProfileChanges = true
	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	controller := NewDeviceProfileController(dic)
	require.NotNil(t, controller)

	validBytes, err := yaml.Marshal(valid)
	require.NoError(t, err)
	req, err := createDeviceProfileRequestWithFile(validBytes)
	require.NoError(t, err)

	// Act
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.UpdateDeviceProfileByYaml)
	handler.ServeHTTP(recorder, req)

	var res commonDTO.BaseWithIdResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusLocked, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusLocked, res.StatusCode, "HTTP status code not as expected")
	assert.NotEmpty(t, recorder.Body.String(), "Message is empty")
}

func TestDeviceProfileByName(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileByName", deviceProfile.Name).Return(deviceProfile, nil)
	dbClientMock.On("DeviceProfileByName", notFoundName).Return(models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device profile doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			reqPath := fmt.Sprintf("%s/%s/%s", common.ApiDeviceProfileRoute, common.Name, testCase.deviceProfileName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceProfileName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceProfileByName)
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
				var res responseDTO.DeviceProfileResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
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
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DevicesByProfileName", 0, 1, deviceProfile.Name).Return([]models.Device{}, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 0, 1, deviceProfile.Name).Return([]models.ProvisionWatcher{}, nil)
	dbClientMock.On("DeleteDeviceProfileByName", deviceProfile.Name).Return(nil)
	dbClientMock.On("DevicesByProfileName", 0, 1, notFoundName).Return([]models.Device{}, nil)
	dbClientMock.On("ProvisionWatchersByProfileName", 0, 1, notFoundName).Return([]models.ProvisionWatcher{}, nil)
	dbClientMock.On("DeleteDeviceProfileByName", notFoundName).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device profile doesn't exist in the database", nil))
	dbClientMock.On("DeleteDeviceProfileByName", deviceExists).Return(errors.NewCommonEdgeX(
		errors.KindStatusConflict, "fail to delete the device profile when associated device exists", nil))
	dbClientMock.On("DeleteDeviceProfileByName", provisionWatcherExists).Return(errors.NewCommonEdgeX(
		errors.KindStatusConflict, "fail to delete the device profile when associated provisionWatcher exists", nil))
	dbClientMock.On("ProvisionWatchersByProfileName", 0, 1, provisionWatcherExists).Return([]models.ProvisionWatcher{models.ProvisionWatcher{}}, nil)
	dbClientMock.On("DeviceProfileByName", mock.Anything).Return(models.DeviceProfile{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		{"Invalid - associated device exists", deviceExists, true, http.StatusConflict},
		{"Invalid - associated provisionWatcher Exists", provisionWatcherExists, true, http.StatusConflict},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s/%s", common.ApiDeviceProfileRoute, common.Name, testCase.deviceProfileName)
			req, err := http.NewRequest(http.MethodDelete, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.deviceProfileName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteDeviceProfileByName)
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

func TestDeleteDeviceProfileByName_StrictProfileChanges(t *testing.T) {
	dic := mockDic()
	configuration := container.ConfigurationFrom(dic.Get)
	configuration.Writable.ProfileChange.StrictDeviceProfileDeletes = true
	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	controller := NewDeviceProfileController(dic)
	require.NotNil(t, controller)

	req, err := http.NewRequest(http.MethodDelete, common.ApiDeviceProfileByNameRoute, http.NoBody)
	req = mux.SetURLVars(req, map[string]string{common.Name: TestDeviceProfileName})
	require.NoError(t, err)

	// Act
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.DeleteDeviceProfileByName)
	handler.ServeHTTP(recorder, req)

	var res commonDTO.BaseResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusLocked, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusLocked, res.StatusCode, "BaseResponse status code not as expected")
	assert.NotEmpty(t, res.Message, "Message is empty")
}

func TestAllDeviceProfiles(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	deviceProfiles := []models.DeviceProfile{deviceProfile, deviceProfile, deviceProfile}
	expectedTotalProfileCount := uint32(3)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileCountByLabels", []string(nil)).Return(expectedTotalProfileCount, nil)
	dbClientMock.On("DeviceProfileCountByLabels", testDeviceProfileLabels).Return(expectedTotalProfileCount, nil)
	dbClientMock.On("AllDeviceProfiles", 0, 10, []string(nil)).Return(deviceProfiles, nil)
	dbClientMock.On("AllDeviceProfiles", 0, 5, testDeviceProfileLabels).Return([]models.DeviceProfile{deviceProfiles[0], deviceProfiles[1]}, nil)
	dbClientMock.On("AllDeviceProfiles", 1, 2, []string(nil)).Return([]models.DeviceProfile{deviceProfiles[1], deviceProfiles[2]}, nil)
	dbClientMock.On("AllDeviceProfiles", 4, 1, testDeviceProfileLabels).Return([]models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get device profiles without labels", "0", "10", "", false, 3, expectedTotalProfileCount, http.StatusOK},
		{"Valid - get device profiles with labels", "0", "5", strings.Join(testDeviceProfileLabels, ","), false, 2, expectedTotalProfileCount, http.StatusOK},
		{"Valid - get device profiles with offset and no labels", "1", "2", "", false, 2, expectedTotalProfileCount, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", strings.Join(testDeviceProfileLabels, ","), true, 0, expectedTotalProfileCount, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiAllDeviceProfileRoute, http.NoBody)
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
			handler := http.HandlerFunc(controller.AllDeviceProfiles)
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
				var res responseDTO.MultiDeviceProfilesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Profiles), "Profile count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeviceProfilesByModel(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	deviceProfiles := []models.DeviceProfile{deviceProfile, deviceProfile, deviceProfile}
	expectedTotalProfileCount := uint32(3)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileCountByModel", TestModel).Return(expectedTotalProfileCount, nil)
	dbClientMock.On("DeviceProfilesByModel", 0, 10, TestModel).Return(deviceProfiles, nil)
	dbClientMock.On("DeviceProfilesByModel", 1, 2, TestModel).Return([]models.DeviceProfile{deviceProfiles[1], deviceProfiles[2]}, nil)
	dbClientMock.On("DeviceProfilesByModel", 4, 1, TestModel).Return([]models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get device profiles by model", "0", "10", TestModel, false, 3, expectedTotalProfileCount, http.StatusOK},
		{"Valid - get device profiles by model with offset and limit", "1", "2", TestModel, false, 2, expectedTotalProfileCount, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", TestModel, true, 0, expectedTotalProfileCount, http.StatusNotFound},
		{"Invalid - model is empty", "0", "10", "", true, 0, expectedTotalProfileCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiDeviceProfileByModelRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Model: testCase.model})
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceProfilesByModel)
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
				var res responseDTO.MultiDeviceProfilesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Profiles), "Profile count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeviceProfilesByManufacturer(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	deviceProfiles := []models.DeviceProfile{deviceProfile, deviceProfile, deviceProfile}
	expectedTotalProfileCount := uint32(3)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileCountByManufacturer", TestManufacturer).Return(expectedTotalProfileCount, nil)
	dbClientMock.On("DeviceProfilesByManufacturer", 0, 10, TestManufacturer).Return(deviceProfiles, nil)
	dbClientMock.On("DeviceProfilesByManufacturer", 1, 2, TestManufacturer).Return([]models.DeviceProfile{deviceProfiles[1], deviceProfiles[2]}, nil)
	dbClientMock.On("DeviceProfilesByManufacturer", 4, 1, TestManufacturer).Return([]models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get device profiles by manufacturer", "0", "10", TestManufacturer, false, 3, expectedTotalProfileCount, http.StatusOK},
		{"Valid - get device profiles by manufacturer with offset and limit", "1", "2", TestManufacturer, false, 2, expectedTotalProfileCount, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", TestManufacturer, true, 0, expectedTotalProfileCount, http.StatusNotFound},
		{"Invalid - manufacturer is empty", "0", "10", "", true, 0, expectedTotalProfileCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiDeviceProfileByManufacturerRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Manufacturer: testCase.manufacturer})
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceProfilesByManufacturer)
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
				var res responseDTO.MultiDeviceProfilesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Profiles), "Profile count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeviceProfilesByManufacturerAndModel(t *testing.T) {
	deviceProfile := dtos.ToDeviceProfileModel(buildTestDeviceProfileRequest().Profile)
	deviceProfiles := []models.DeviceProfile{deviceProfile, deviceProfile, deviceProfile}
	expectedTotalProfileCount := uint32(3)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfilesByManufacturerAndModel", 0, 10, TestManufacturer, TestModel).Return(deviceProfiles, expectedTotalProfileCount, nil)
	dbClientMock.On("DeviceProfilesByManufacturerAndModel", 1, 2, TestManufacturer, TestModel).Return([]models.DeviceProfile{deviceProfiles[1], deviceProfiles[2]}, expectedTotalProfileCount, nil)
	dbClientMock.On("DeviceProfilesByManufacturerAndModel", 4, 1, TestManufacturer, TestModel).Return([]models.DeviceProfile{}, expectedTotalProfileCount, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get device profiles by manufacturer and model", "0", "10", TestManufacturer, TestModel, false, 3, expectedTotalProfileCount, http.StatusOK},
		{"Valid - get device profiles by manufacturer with offset and limit", "1", "2", TestManufacturer, TestModel, false, 2, expectedTotalProfileCount, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", TestManufacturer, TestModel, true, 0, expectedTotalProfileCount, http.StatusNotFound},
		{"Invalid - manufacturer is empty", "0", "10", "", TestModel, true, 0, expectedTotalProfileCount, http.StatusBadRequest},
		{"Invalid - model is empty", "0", "10", TestManufacturer, "", true, 0, expectedTotalProfileCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiDeviceProfileByManufacturerAndModelRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Manufacturer: testCase.manufacturer, common.Model: testCase.model})
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeviceProfilesByManufacturerAndModel)
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
				var res responseDTO.MultiDeviceProfilesResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Profiles), "Profile count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
