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

	v2MetadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testProvisionWatcherName = "TestProvisionWatcher"
var testProvisionWatcherLabels = []string{"test", "temp"}
var testProvisionWatcherIdentifiers = map[string]string{
	"address": "localhost",
	"port":    "3[0-9]{2}",
}
var testProvisionWatcherBlockingIdentifiers = map[string][]string{
	"port": {"397", "398", "399"},
}
var testProvisionWatcherAutoEvents = []dtos.AutoEvent{
	{Resource: "TestResource", Frequency: "300ms", OnChange: true},
}

func buildTestAddProvisionWatcherRequest() requests.AddProvisionWatcherRequest {
	return requests.AddProvisionWatcherRequest{
		BaseRequest: common.BaseRequest{
			RequestId: ExampleUUID,
		},
		ProvisionWatcher: dtos.ProvisionWatcher{
			Id:                  ExampleUUID,
			Name:                testProvisionWatcherName,
			Labels:              testProvisionWatcherLabels,
			Identifiers:         testProvisionWatcherIdentifiers,
			BlockingIdentifiers: testProvisionWatcherBlockingIdentifiers,
			ProfileName:         TestDeviceProfileName,
			ServiceName:         TestDeviceServiceName,
			AdminState:          models.Unlocked,
			AutoEvents:          testProvisionWatcherAutoEvents,
		},
	}
}

func TestProvisionWatcherController_AddProvisionWatcher_Created(t *testing.T) {
	validReq := buildTestAddProvisionWatcherRequest()
	pwModel := requests.AddProvisionWatcherReqToProvisionWatcherModels([]requests.AddProvisionWatcherRequest{validReq})[0]
	expectedRequestId := ExampleUUID

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddProvisionWatcher", pwModel).Return(pwModel, nil)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	reqWithNoID := validReq
	reqWithNoID.RequestId = ""

	tests := []struct {
		name    string
		Request []requests.AddProvisionWatcherRequest
	}{
		{"Valid - AddProvisionWatcherRequest", []requests.AddProvisionWatcherRequest{validReq}},
		{"Valid - no RequestId", []requests.AddProvisionWatcherRequest{reqWithNoID}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddProvisionWatcher)
			handler.ServeHTTP(recorder, req)
			var res []common.BaseWithIdResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contractsV2.ApiVersion, res[0].ApiVersion, "API version not as expected")
			if res[0].RequestId != "" {
				assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
			}
			assert.Equal(t, http.StatusCreated, res[0].StatusCode, "BaseResonse status code not as expected")
			assert.Empty(t, res[0].Message, "Message should be empty when it's successful")
		})
	}
}

func TestProvisionWatcherController_AddProvisionWatcher_BadRequest(t *testing.T) {
	dic := mockDic()

	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	provisionWatcher := buildTestAddProvisionWatcherRequest()
	badRequestId := provisionWatcher
	badRequestId.RequestId = "niv3sl"
	noName := provisionWatcher
	noName.ProvisionWatcher.Name = ""

	tests := []struct {
		name    string
		Request []requests.AddProvisionWatcherRequest
	}{
		{"Invalid - Bad requestId", []requests.AddProvisionWatcherRequest{badRequestId}},
		{"Invalid - Bad name", []requests.AddProvisionWatcherRequest{noName}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddProvisionWatcher)
			handler.ServeHTTP(recorder, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.NotEmpty(t, string(recorder.Body.Bytes()), "Message is empty")
		})
	}
}

func TestProvisionWatcherController_AddProvisionWatcher_Duplicated(t *testing.T) {
	expectedRequestId := ExampleUUID

	duplicateIdRequest := buildTestAddProvisionWatcherRequest()
	duplicateIdModel := requests.AddProvisionWatcherReqToProvisionWatcherModels([]requests.AddProvisionWatcherRequest{duplicateIdRequest})[0]
	duplicateIdDBError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("provision watcher id %s already exists", duplicateIdModel.Id), nil)

	duplicateNameRequest := buildTestAddProvisionWatcherRequest()
	duplicateNameRequest.ProvisionWatcher.Id = "" // The infrastructure layer will generate id when the id field is empty
	duplicateNameModel := requests.AddProvisionWatcherReqToProvisionWatcherModels([]requests.AddProvisionWatcherRequest{duplicateNameRequest})[0]
	duplicateNameDBError := errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("provision watcher name %s already exists", duplicateNameModel.Name), nil)

	dic := mockDic()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddProvisionWatcher", duplicateNameModel).Return(duplicateNameModel, duplicateNameDBError)
	dbClientMock.On("AddProvisionWatcher", duplicateIdModel).Return(duplicateIdModel, duplicateIdDBError)
	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewProvisionWatcherController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name          string
		request       []requests.AddProvisionWatcherRequest
		expectedError errors.CommonEdgeX
	}{
		{"duplicate id", []requests.AddProvisionWatcherRequest{duplicateIdRequest}, duplicateIdDBError},
		{"duplicate name", []requests.AddProvisionWatcherRequest{duplicateNameRequest}, duplicateNameDBError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiProvisionWatcherRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddProvisionWatcher)
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
