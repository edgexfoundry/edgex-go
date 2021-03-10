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

	v2SchedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func addIntervalActionRequestData() requests.AddIntervalActionRequest {
	restAddress := dtos.NewRESTAddress(TestHost, TestPort, TestHTTPMethod)
	intervalAction := dtos.NewIntervalAction(TestIntervalActionName, TestIntervalName, restAddress)
	return requests.NewAddIntervalActionRequest(intervalAction)
}

func TestAddIntervalAction(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}

	valid := addIntervalActionRequestData()
	model := dtos.ToIntervalActionModel(valid.Action)
	dbClientMock.On("IntervalByName", model.IntervalName).Return(models.Interval{}, nil)
	dbClientMock.On("AddIntervalAction", model).Return(model, nil)

	noName := valid
	noName.Action.Name = ""
	noRequestId := valid
	noRequestId.RequestId = ""

	duplicatedName := valid
	duplicatedName.Action.Name = "duplicatedName"
	model = dtos.ToIntervalActionModel(duplicatedName.Action)
	dbClientMock.On("AddIntervalAction", model).Return(model, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("intervalAction name %s already exists", model.Name), nil))

	dic.Update(di.ServiceConstructorMap{
		v2SchedulerContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewIntervalActionController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name               string
		request            []requests.AddIntervalActionRequest
		expectedStatusCode int
	}{
		{"Valid", []requests.AddIntervalActionRequest{valid}, http.StatusCreated},
		{"Valid - no request Id", []requests.AddIntervalActionRequest{noRequestId}, http.StatusCreated},
		{"Invalid - no name", []requests.AddIntervalActionRequest{noName}, http.StatusBadRequest},
		{"Invalid - duplicated name", []requests.AddIntervalActionRequest{duplicatedName}, http.StatusConflict},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, v2.ApiIntervalActionRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddIntervalAction)
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
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}
