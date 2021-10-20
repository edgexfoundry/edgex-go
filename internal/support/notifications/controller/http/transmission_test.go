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

	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			reqPath := fmt.Sprintf("%s/%s", common.ApiTransmissionByIdRoute, testCase.transmissionId)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Id: testCase.transmissionId})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.TransmissionById)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.TransmissionResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.transmissionId, res.Transmission.Id, "ID is not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestTransmissionsByTimeRange(t *testing.T) {
	expectedTransmissionCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("TransmissionCountByTimeRange", 0, 100).Return(expectedTransmissionCount, nil)
	dbClientMock.On("TransmissionsByTimeRange", 0, 100, 0, 10).Return([]models.Transmission{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - with proper start/end/offset/limit", "0", "100", "0", "10", false, 0, expectedTransmissionCount, http.StatusOK},
		{"Invalid - invalid start format", "aaa", "100", "0", "10", true, 0, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - invalid end format", "0", "bbb", "0", "10", true, 0, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - empty start", "", "100", "0", "10", true, 0, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - empty end", "0", "", "0", "10", true, 0, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - end before start", "10", "0", "0", "10", true, 0, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - invalid offset format", "0", "100", "aaa", "10", true, 0, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "0", "100", "0", "aaa", true, 0, expectedTransmissionCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiTransmissionByTimeRangeRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Start: testCase.start, common.End: testCase.end})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(tc.TransmissionsByTimeRange)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiTransmissionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Transmissions), "Transmission count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Response total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestAllTransmissions(t *testing.T) {
	trans := transmissionData()
	transmissions := []models.Transmission{trans, trans, trans}
	expectedTransmissionCount := uint32(len(transmissions))

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("TransmissionTotalCount").Return(expectedTransmissionCount, nil)
	dbClientMock.On("AllTransmissions", 0, 20).Return(transmissions, nil)
	dbClientMock.On("AllTransmissions", 1, 2).Return([]models.Transmission{transmissions[1], transmissions[2]}, nil)
	dbClientMock.On("AllTransmissions", 4, 1).Return([]models.Transmission{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get transmissions without offset and limit", "", "", false, 3, expectedTransmissionCount, http.StatusOK},
		{"Valid - get transmissions with offset and limit", "1", "2", false, 2, expectedTransmissionCount, http.StatusOK},
		{"Invalid - offset out of range", "4", "1", true, 0, expectedTransmissionCount, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiAllTransmissionRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AllTransmissions)
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
				var res responseDTO.MultiTransmissionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Transmissions), "Transmission count is not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Response total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestTransmissionsByStatus(t *testing.T) {
	testStatus := models.New
	expectedTransmissionCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("TransmissionCountByStatus", testStatus).Return(expectedTransmissionCount, nil)
	dbClientMock.On("TransmissionsByStatus", 0, 20, testStatus).Return([]models.Transmission{}, nil)
	dbClientMock.On("TransmissionsByStatus", 0, 1, testStatus).Return([]models.Transmission{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get transmissions without offset, and limit", "", "", testStatus, false, expectedTransmissionCount, http.StatusOK},
		{"Valid - get transmissions with offset, and limit", "0", "1", testStatus, false, expectedTransmissionCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testStatus, true, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testStatus, true, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - status is empty", "0", "1", "", true, expectedTransmissionCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiTransmissionByStatusRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Status: testCase.status})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.TransmissionsByStatus)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiTransmissionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Response total count not as expected")
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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			req, err := http.NewRequest(http.MethodDelete, common.ApiTransmissionByAgeRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Age: testCase.age})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(tc.DeleteProcessedTransmissionsByAge)
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
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestTransmissionsBySubscriptionName(t *testing.T) {
	testName := "testName"
	expectedTransmissionCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("TransmissionCountBySubscriptionName", testName).Return(expectedTransmissionCount, nil)
	dbClientMock.On("TransmissionsBySubscriptionName", 0, 20, testName).Return([]models.Transmission{}, nil)
	dbClientMock.On("TransmissionsBySubscriptionName", 0, 1, testName).Return([]models.Transmission{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get transmissions without offset, and limit", "", "", testName, false, expectedTransmissionCount, http.StatusOK},
		{"Valid - get transmissions with offset, and limit", "0", "1", testName, false, expectedTransmissionCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testName, true, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testName, true, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - subscription name is empty", "0", "1", "", true, expectedTransmissionCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiTransmissionBySubscriptionNameRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Name: testCase.subscriptionName})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.TransmissionsBySubscriptionName)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiTransmissionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Response total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestTransmissionsByNotificationId(t *testing.T) {
	testId := "id"
	expectedTransmissionCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("TransmissionCountByNotificationId", testId).Return(expectedTransmissionCount, nil)
	dbClientMock.On("TransmissionsByNotificationId", 0, 20, testId).Return([]models.Transmission{}, nil)
	dbClientMock.On("TransmissionsByNotificationId", 0, 1, testId).Return([]models.Transmission{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewTransmissionController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		notificationId     string
		errorExpected      bool
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get transmissions without offset, and limit", "", "", testId, false, expectedTransmissionCount, http.StatusOK},
		{"Valid - get transmissions with offset, and limit", "0", "1", testId, false, expectedTransmissionCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testId, true, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testId, true, expectedTransmissionCount, http.StatusBadRequest},
		{"Invalid - notification id is empty", "0", "1", "", true, expectedTransmissionCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiTransmissionByNotificationIdRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Id: testCase.notificationId})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.TransmissionsByNotificationId)
			handler.ServeHTTP(recorder, req)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiTransmissionsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Response total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
