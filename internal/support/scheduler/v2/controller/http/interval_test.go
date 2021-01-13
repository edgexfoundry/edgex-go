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

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
	schedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	v2MetadataContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		schedulerContainer.ConfigurationName: func(get di.Get) interface{} {
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

func addIntervalRequestData() requests.AddIntervalRequest {
	var testAddIntervalReq = requests.AddIntervalRequest{
		BaseRequest: common.BaseRequest{
			RequestId: ExampleUUID,
		},
		Interval: dtos.Interval{
			Id:        ExampleUUID,
			Name:      TestIntervalName,
			Start:     TestIntervalStart,
			End:       TestIntervalEnd,
			Frequency: TestIntervalFrequency,
			RunOnce:   TestIntervalRunOnce,
		},
	}

	return testAddIntervalReq
}

func TestAddInterval(t *testing.T) {
	expectedRequestId := ExampleUUID
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}

	valid := addIntervalRequestData()
	model := dtos.ToIntervalModel(valid.Interval)
	dbClientMock.On("AddInterval", model).Return(model, nil)

	noName := addIntervalRequestData()
	noName.Interval.Name = ""
	noRequestId := addIntervalRequestData()
	noRequestId.RequestId = ""

	duplicatedName := addIntervalRequestData()
	duplicatedName.Interval.Name = "duplicatedName"
	model = dtos.ToIntervalModel(duplicatedName.Interval)
	dbClientMock.On("AddInterval", model).Return(model, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("interval name %s already exists", model.Name), nil))

	dic.Update(di.ServiceConstructorMap{
		v2MetadataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewIntervalController(dic)
	assert.NotNil(t, controller)
	tests := []struct {
		name               string
		request            []requests.AddIntervalRequest
		expectedStatusCode int
	}{
		{"Valid", []requests.AddIntervalRequest{valid}, http.StatusCreated},
		{"Valid - no request Id", []requests.AddIntervalRequest{noRequestId}, http.StatusCreated},
		{"Invalid - no name", []requests.AddIntervalRequest{noName}, http.StatusBadRequest},
		{"Invalid - duplicated name", []requests.AddIntervalRequest{duplicatedName}, http.StatusConflict},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, v2.ApiIntervalRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddInterval)
			handler.ServeHTTP(recorder, req)
			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Message is empty")
			} else {
				var res []common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, v2.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}
