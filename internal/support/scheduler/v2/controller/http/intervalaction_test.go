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
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
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

func TestAllIntervalActions(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AllIntervalActions", 0, 20).Return([]models.IntervalAction{}, nil)
	dbClientMock.On("AllIntervalActions", 0, 1).Return([]models.IntervalAction{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2SchedulerContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewIntervalActionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - get intervalActions without offset and limit", "", "", false, http.StatusOK},
		{"Valid - get intervalActions with offset and limit", "0", "1", false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiAllIntervalActionRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllIntervalActions)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiIntervalsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
