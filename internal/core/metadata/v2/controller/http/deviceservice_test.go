//
// Copyright (C) 2020 IOTech Ltd
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
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDeviceServiceName = "TestDeviceService"
var testDeviceServiceLabels = []string{"hvac", "thermostat"}
var testBaseAddress = "http://home-device-service:49990"

func buildTestDeviceServiceRequest() requests.AddDeviceServiceRequest {
	var testAddDeviceServiceReq = requests.AddDeviceServiceRequest{
		BaseRequest: common.BaseRequest{
			RequestID: ExampleUUID,
		},
		Service: dtos.DeviceService{
			Name:           testDeviceServiceName,
			Description:    TestDescription,
			Labels:         testDeviceServiceLabels,
			AdminState:     models.Unlocked,
			OperatingState: models.Enabled,
			BaseAddress:    testBaseAddress,
		},
	}

	return testAddDeviceServiceReq
}

func buildTestDBClient(dsModel models.DeviceService, errKind errors.ErrKind, errorMessage string) *dbMock.DBClient {
	dbClientMock := &dbMock.DBClient{}
	if len(errKind) > 0 {
		err := errors.NewCommonEdgeX(errKind, errorMessage, nil)
		dbClientMock.On("AddDeviceService", dsModel).Return(dsModel, err)
	} else {
		dbClientMock.On("AddDeviceService", dsModel).Return(dsModel, nil)
	}
	return dbClientMock
}

func TestAddDeviceService(t *testing.T) {
	validReq := buildTestDeviceServiceRequest()
	dsModels := requests.AddDeviceServiceReqToDeviceServiceModels([]requests.AddDeviceServiceRequest{validReq})
	expectedRequestId := ExampleUUID
	normalMessage := fmt.Sprintf("Add device service %s successfully", testDeviceServiceName)
	duplicateServiceNameMessage := fmt.Sprintf("device service %s already exists", testDeviceServiceName)

	reqWithNoID := validReq
	reqWithNoID.RequestID = ""
	reqWithInvalidId := validReq
	reqWithInvalidId.RequestID = "InvalidUUID"
	reqWithNoName := validReq
	reqWithNoName.Service.Name = ""

	tests := []struct {
		name                   string
		isValidRequest         bool
		dbClientMock           *dbMock.DBClient
		Request                []requests.AddDeviceServiceRequest
		expectedHttpStatusCode int
		expectedMessage        string
	}{
		{
			"Request Normal",
			true,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{validReq},
			http.StatusCreated,
			normalMessage,
		},
		{
			"Request without requestId",
			true,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{reqWithNoID},
			http.StatusCreated,
			normalMessage,
		},
		{
			"Request with duplicate service name",
			true,
			buildTestDBClient(dsModels[0], errors.KindDuplicateName, duplicateServiceNameMessage),
			[]requests.AddDeviceServiceRequest{validReq},
			http.StatusConflict,
			duplicateServiceNameMessage,
		},
		{
			"Request with invalid requestId",
			false,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{reqWithInvalidId},
			http.StatusBadRequest,
			"",
		},
		{
			"Request without service name",
			false,
			buildTestDBClient(dsModels[0], "", ""),
			[]requests.AddDeviceServiceRequest{reqWithNoName},
			http.StatusBadRequest,
			"",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {

			dic := mockDic()
			dic.Update(di.ServiceConstructorMap{
				v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
					return testCase.dbClientMock
				},
			})

			controller := NewDeviceServiceController(dic)
			require.NotNil(t, controller)

			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, contractsV2.ApiDeviceServiceRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddDeviceService)
			handler.ServeHTTP(recorder, req)
			if testCase.isValidRequest {
				var res []common.BaseWithIdResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)

				// Assert
				require.NoError(t, err)
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, contractsV2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestID != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestID, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedHttpStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
				assert.Contains(t, res[0].Message, testCase.expectedMessage, "Message not as expected")
			} else {
				assert.Equal(t, testCase.expectedHttpStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.NotEmpty(t, string(recorder.Body.Bytes()), "Message is empty")
			}
		})
	}
}
