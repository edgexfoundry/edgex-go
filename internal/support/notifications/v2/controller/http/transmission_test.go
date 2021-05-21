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
	"testing"

	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func transmissionData() models.Transmission {
	transmissionId := "1208bbca-8521-434a-a923-66255a68ba11"
	notificationId := "1208bbca-8521-434a-a923-66255a68ba22"
	return models.Transmission{
		Id:               transmissionId,
		SubscriptionName: testSubscriptionName,
		Channel:          models.RESTAddress{},
		NotificationId:   notificationId,
	}
}

func TestTransmissionById(t *testing.T) {
	trans := transmissionData()
	emptyId := ""
	notFoundId := "1208bbca-8521-434a-a923-000000000000"
	invalidId := "invalidId"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("TransmissionById", trans.Id).Return(trans, nil)
	dbClientMock.On("TransmissionById", notFoundId).Return(models.Transmission{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "transmission doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewTransmissionController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		transmissionId     string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - find transmission by ID", trans.Id, false, http.StatusOK},
		{"Invalid - ID parameter is empty", emptyId, true, http.StatusBadRequest},
		{"Invalid - ID parameter is not a valid UUID", invalidId, true, http.StatusBadRequest},
		{"Invalid - transmission not found by ID", notFoundId, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", v2.ApiTransmissionByIdRoute, testCase.transmissionId)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{v2.Id: testCase.transmissionId})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.TransmissionById)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.TransmissionResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.transmissionId, res.Transmission.Id, "ID is not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestTransmissionsByTimeRange(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("TransmissionsByTimeRange", 0, 100, 0, 10).Return([]models.Transmission{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	tc := NewTransmissionController(dic)
	assert.NotNil(t, tc)

	tests := []struct {
		name               string
		start              string
		end                string
		offset             string
		limit              string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - with proper start/end/offset/limit", "0", "100", "0", "10", false, 0, http.StatusOK},
		{"Invalid - invalid start format", "aaa", "100", "0", "10", true, 0, http.StatusBadRequest},
		{"Invalid - invalid end format", "0", "bbb", "0", "10", true, 0, http.StatusBadRequest},
		{"Invalid - empty start", "", "100", "0", "10", true, 0, http.StatusBadRequest},
		{"Invalid - empty end", "0", "", "0", "10", true, 0, http.StatusBadRequest},
		{"Invalid - end before start", "10", "0", "0", "10", true, 0, http.StatusBadRequest},
		{"Invalid - invalid offset format", "0", "100", "aaa", "10", true, 0, http.StatusBadRequest},
		{"Invalid - invalid limit format", "0", "100", "0", "aaa", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiTransmissionByTimeRangeRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(v2.Offset, testCase.offset)
			query.Add(v2.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Start: testCase.start, v2.End: testCase.end})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(tc.TransmissionsByTimeRange)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiTransmissionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Transmissions), "Transmission count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestAllTransmissions(t *testing.T) {
	trans := transmissionData()
	transmissions := []models.Transmission{trans, trans, trans}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AllTransmissions", 0, 20).Return(transmissions, nil)
	dbClientMock.On("AllTransmissions", 1, 2).Return([]models.Transmission{transmissions[1], transmissions[2]}, nil)
	dbClientMock.On("AllTransmissions", 4, 1).Return([]models.Transmission{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewTransmissionController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - get transmissions without offset and limit", "", "", false, 3, http.StatusOK},
		{"Valid - get transmissions with offset and limit", "1", "2", false, 2, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", true, 0, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiAllTransmissionRoute, http.NoBody)
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
			handler := http.HandlerFunc(controller.AllTransmissions)
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
				var res responseDTO.MultiTransmissionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Transmissions), "Transmission count is not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestTransmissionsByStatus(t *testing.T) {
	testStatus := models.New
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("TransmissionsByStatus", 0, 20, testStatus).Return([]models.Transmission{}, nil)
	dbClientMock.On("TransmissionsByStatus", 0, 1, testStatus).Return([]models.Transmission{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewTransmissionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		status             string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - get transmissions without offset, and limit", "", "", testStatus, false, http.StatusOK},
		{"Valid - get transmissions with offset, and limit", "0", "1", testStatus, false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testStatus, true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testStatus, true, http.StatusBadRequest},
		{"Invalid - status is empty", "0", "1", "", true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiTransmissionByStatusRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Status: testCase.status})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.TransmissionsByStatus)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiTransmissionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeleteTransmissionsByAge(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteProcessedTransmissionsByAge", int64(0)).Return(nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	tc := NewTransmissionController(dic)
	assert.NotNil(t, tc)

	tests := []struct {
		name               string
		age                string
		errorExpected      bool
		expectedCount      int
		expectedStatusCode int
	}{
		{"Valid - age with proper format", "0", false, 0, http.StatusAccepted},
		{"Invalid - age with unparsable format", "aaa", true, 0, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, v2.ApiTransmissionByAgeRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{v2.Age: testCase.age})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(tc.DeleteProcessedTransmissionsByAge)
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
				var res common.BaseResponse
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

func TestTransmissionsBySubscriptionName(t *testing.T) {
	testName := "testName"
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("TransmissionsBySubscriptionName", 0, 20, testName).Return([]models.Transmission{}, nil)
	dbClientMock.On("TransmissionsBySubscriptionName", 0, 1, testName).Return([]models.Transmission{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewTransmissionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		subscriptionName   string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - get transmissions without offset, and limit", "", "", testName, false, http.StatusOK},
		{"Valid - get transmissions with offset, and limit", "0", "1", testName, false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testName, true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testName, true, http.StatusBadRequest},
		{"Invalid - subscription name is empty", "0", "1", "", true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiTransmissionBySubscriptionNameRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Name: testCase.subscriptionName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.TransmissionsBySubscriptionName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res common.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiTransmissionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
