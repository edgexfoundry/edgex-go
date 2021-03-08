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

	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testNotificationCategory    = "category"
	testNotificationContent     = "content"
	testNotificationContentType = "text/plain"
	testNotificationDescription = "description"
	testNotificationLabels      = []string{"label1", "label2"}
	testNotificationSender      = "sender"
	testNotificationSeverity    = models.Normal
)

func buildTestAddNotificationRequest() requests.AddNotificationRequest {
	notification := dtos.NewNotification(testNotificationLabels, testNotificationCategory, testNotificationContent,
		testNotificationSender, testNotificationSeverity)
	notification.ContentType = testNotificationContentType
	notification.Description = testNotificationDescription
	return requests.NewAddNotificationRequest(notification)
}

func TestAddNotification(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}

	validRequest := buildTestAddNotificationRequest()
	model := dtos.ToNotificationModel(validRequest.Notification)
	dbClientMock.On("AddNotification", model).Return(model, nil)

	noRequestId := validRequest
	noRequestId.RequestId = ""
	invalidReqId := validRequest
	invalidReqId.RequestId = "abc"

	noCategoryAndLabels := validRequest
	noCategoryAndLabels.Notification.Category = ""
	noCategoryAndLabels.Notification.Labels = nil

	noContent := validRequest
	noContent.Notification.Content = ""

	noSender := validRequest
	noSender.Notification.Sender = ""

	noSeverity := validRequest
	noSeverity.Notification.Severity = ""
	invalidSeverity := validRequest
	invalidSeverity.Notification.Severity = "foo"

	invalidStatus := validRequest
	invalidStatus.Notification.Status = "foo"

	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewNotificationController(dic)
	assert.NotNil(t, controller)
	tests := []struct {
		name               string
		request            []requests.AddNotificationRequest
		expectedStatusCode int
	}{
		{"valid", []requests.AddNotificationRequest{validRequest}, http.StatusCreated},
		{"valid - no request Id", []requests.AddNotificationRequest{noRequestId}, http.StatusCreated},
		{"invalid, request ID is not an UUID", []requests.AddNotificationRequest{invalidReqId}, http.StatusBadRequest},
		{"invalid, no category and labels", []requests.AddNotificationRequest{noCategoryAndLabels}, http.StatusBadRequest},
		{"invalid, no content", []requests.AddNotificationRequest{noContent}, http.StatusBadRequest},
		{"invalid, no sender", []requests.AddNotificationRequest{noSender}, http.StatusBadRequest},
		{"invalid, no severity", []requests.AddNotificationRequest{noSeverity}, http.StatusBadRequest},
		{"invalid, unsupported severity level", []requests.AddNotificationRequest{invalidSeverity}, http.StatusBadRequest},
		{"invalid, unsupported status", []requests.AddNotificationRequest{invalidStatus}, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, v2.ApiNotificationRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddNotification)
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
					assert.Equal(t, testCase.request[0].RequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}

func TestNotificationById(t *testing.T) {
	request := buildTestAddNotificationRequest()
	notification := dtos.ToNotificationModel(request.Notification)
	emptyId := ""
	notFoundId := "1208bbca-8521-434a-a923-66255a68ba00"
	invalidId := "invalidId"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationById", notification.Id).Return(notification, nil)
	dbClientMock.On("NotificationById", notFoundId).Return(models.Notification{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "notification doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewNotificationController(dic)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		notificationId     string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - find notification by ID", notification.Id, false, http.StatusOK},
		{"Invalid - ID parameter is empty", emptyId, true, http.StatusBadRequest},
		{"Invalid - ID parameter is not a valid UUID", invalidId, true, http.StatusBadRequest},
		{"Invalid - notification not found by ID", notFoundId, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", v2.ApiNotificationByIdRoute, testCase.notificationId)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{v2.Id: testCase.notificationId})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.NotificationById)
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
				var res responseDTO.NotificationResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, v2.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.notificationId, res.Notification.Id, "ID is not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestNotificationsByCategory(t *testing.T) {
	testCategory := "category"
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationsByCategory", 0, 20, testCategory).Return([]models.Notification{}, nil)
	dbClientMock.On("NotificationsByCategory", 0, 1, testCategory).Return([]models.Notification{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewNotificationController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		category           string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - get notifications without offset, and limit", "", "", testCategory, false, http.StatusOK},
		{"Valid - get notifications with offset, and limit", "0", "1", testCategory, false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testCategory, true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testCategory, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiNotificationByCategoryRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Category: testCase.category})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.NotificationsByCategory)
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
				var res responseDTO.MultiNotificationsResponse
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

func TestNotificationsByLabel(t *testing.T) {
	testLabel := "label"
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationsByLabel", 0, 20, testLabel).Return([]models.Notification{}, nil)
	dbClientMock.On("NotificationsByLabel", 0, 1, testLabel).Return([]models.Notification{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewNotificationController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		label              string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - get notifications without offset, and limit", "", "", testLabel, false, http.StatusOK},
		{"Valid - get notifications with offset, and limit", "0", "1", testLabel, false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testLabel, true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testLabel, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiNotificationByLabelRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{v2.Label: testCase.label})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.NotificationsByLabel)
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
				var res responseDTO.MultiNotificationsResponse
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

func TestNotificationsByStatus(t *testing.T) {
	testStatus := models.New
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationsByStatus", 0, 20, testStatus).Return([]models.Notification{}, nil)
	dbClientMock.On("NotificationsByStatus", 0, 1, testStatus).Return([]models.Notification{}, nil)
	dic.Update(di.ServiceConstructorMap{
		v2NotificationsContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewNotificationController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		offset             string
		limit              string
		status             string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - get notifications without offset, and limit", "", "", testStatus, false, http.StatusOK},
		{"Valid - get notifications with offset, and limit", "0", "1", testStatus, false, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testStatus, true, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testStatus, true, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiNotificationByStatusRoute, http.NoBody)
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
			handler := http.HandlerFunc(controller.NotificationsByStatus)
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
				var res responseDTO.MultiNotificationsResponse
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
