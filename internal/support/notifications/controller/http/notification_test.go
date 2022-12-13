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

	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

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
	model.Status = models.Processed
	dbClientMock.On("UpdateNotification", model).Return(nil)
	dbClientMock.On("SubscriptionsByCategoriesAndLabels", 0, -1, []string{testNotificationCategory}, testNotificationLabels).Return([]models.Subscription{}, nil)

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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			req, err := http.NewRequest(http.MethodPost, common.ApiNotificationRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.AddNotification)
			handler.ServeHTTP(recorder, req)
			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Message is empty")
			} else {
				var res []commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
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
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
			reqPath := fmt.Sprintf("%s/%s", common.ApiNotificationByIdRoute, testCase.notificationId)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Id: testCase.notificationId})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.NotificationById)
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
				var res responseDTO.NotificationResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
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
	expectedNotificationCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationCountByCategory", testCategory).Return(expectedNotificationCount, nil)
	dbClientMock.On("NotificationsByCategory", 0, 20, testCategory).Return([]models.Notification{}, nil)
	dbClientMock.On("NotificationsByCategory", 0, 1, testCategory).Return([]models.Notification{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get notifications without offset, and limit", "", "", testCategory, false, expectedNotificationCount, http.StatusOK},
		{"Valid - get notifications with offset, and limit", "0", "1", testCategory, false, expectedNotificationCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testCategory, true, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testCategory, true, expectedNotificationCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationByCategoryRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Category: testCase.category})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.NotificationsByCategory)
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
				var res responseDTO.MultiNotificationsResponse
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

func TestNotificationsByLabel(t *testing.T) {
	testLabel := "label"
	expectedNotificationCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationCountByLabel", testLabel).Return(expectedNotificationCount, nil)
	dbClientMock.On("NotificationsByLabel", 0, 20, testLabel).Return([]models.Notification{}, nil)
	dbClientMock.On("NotificationsByLabel", 0, 1, testLabel).Return([]models.Notification{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get notifications without offset, and limit", "", "", testLabel, false, expectedNotificationCount, http.StatusOK},
		{"Valid - get notifications with offset, and limit", "0", "1", testLabel, false, expectedNotificationCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testLabel, true, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testLabel, true, expectedNotificationCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationByLabelRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Label: testCase.label})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.NotificationsByLabel)
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
				var res responseDTO.MultiNotificationsResponse
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

func TestNotificationsByStatus(t *testing.T) {
	testStatus := models.New
	expectedNotificationCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationCountByStatus", testStatus).Return(expectedNotificationCount, nil)
	dbClientMock.On("NotificationsByStatus", 0, 20, testStatus).Return([]models.Notification{}, nil)
	dbClientMock.On("NotificationsByStatus", 0, 1, testStatus).Return([]models.Notification{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
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
		expectedTotalCount uint32
		expectedStatusCode int
	}{
		{"Valid - get notifications without offset, and limit", "", "", testStatus, false, expectedNotificationCount, http.StatusOK},
		{"Valid - get notifications with offset, and limit", "0", "1", testStatus, false, expectedNotificationCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testStatus, true, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testStatus, true, expectedNotificationCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationByStatusRoute, http.NoBody)
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
			handler := http.HandlerFunc(controller.NotificationsByStatus)
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
				var res responseDTO.MultiNotificationsResponse
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

func TestNotificationsByTimeRange(t *testing.T) {
	expectedNotificationCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationCountByTimeRange", 0, 100).Return(expectedNotificationCount, nil)
	dbClientMock.On("NotificationsByTimeRange", 0, 100, 0, 10).Return([]models.Notification{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	nc := NewNotificationController(dic)
	assert.NotNil(t, nc)

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
		{"Valid - with proper start/end/offset/limit", "0", "100", "0", "10", false, 0, expectedNotificationCount, http.StatusOK},
		{"Invalid - invalid start format", "aaa", "100", "0", "10", true, 0, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - invalid end format", "0", "bbb", "0", "10", true, 0, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - empty start", "", "100", "0", "10", true, 0, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - empty end", "0", "", "0", "10", true, 0, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - end before start", "10", "0", "0", "10", true, 0, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - invalid offset format", "0", "100", "aaa", "10", true, 0, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "0", "100", "0", "aaa", true, 0, expectedNotificationCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationByTimeRangeRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			req = mux.SetURLVars(req, map[string]string{common.Start: testCase.start, common.End: testCase.end})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(nc.NotificationsByTimeRange)
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
				var res responseDTO.MultiNotificationsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.expectedCount, len(res.Notifications), "Device count not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Response total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeleteNotificationById(t *testing.T) {
	notification := dtos.ToNotificationModel(buildTestAddNotificationRequest().Notification)
	noId := ""
	notFoundId := "notFoundName"

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteNotificationById", notification.Id).Return(nil)
	dbClientMock.On("DeleteNotificationById", notFoundId).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "subscription doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewNotificationController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		notificationId     string
		expectedStatusCode int
	}{
		{"Valid - delete notification by id", notification.Id, http.StatusOK},
		{"Invalid - id parameter is empty", noId, http.StatusBadRequest},
		{"Invalid - notification not found by id", notFoundId, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiNotificationByIdRoute, testCase.notificationId)
			req, err := http.NewRequest(http.MethodDelete, reqPath, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Id: testCase.notificationId})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.DeleteNotificationById)
			handler.ServeHTTP(recorder, req)
			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
			if testCase.expectedStatusCode == http.StatusOK {
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			} else {
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}

func TestNotificationsBySubscriptionName(t *testing.T) {
	subscription := models.Subscription{
		Name:       testSubscriptionName,
		Categories: testSubscriptionCategories,
		Labels:     testSubscriptionLabels,
	}
	expectedNotificationCount := uint32(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("SubscriptionByName", subscription.Name).Return(subscription, nil)
	dbClientMock.On("NotificationCountByCategoriesAndLabels", subscription.Categories, subscription.Labels).Return(expectedNotificationCount, nil)
	dbClientMock.On("NotificationsByCategoriesAndLabels", 0, 20, subscription.Categories, subscription.Labels).Return([]models.Notification{}, nil)
	dbClientMock.On("NotificationsByCategoriesAndLabels", 0, 1, subscription.Categories, subscription.Labels).Return([]models.Notification{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	controller := NewNotificationController(dic)
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
		{"Valid - get notifications without offset, and limit", "", "", subscription.Name, false, expectedNotificationCount, http.StatusOK},
		{"Valid - get notifications with offset, and limit", "0", "1", subscription.Name, false, expectedNotificationCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", subscription.Name, true, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", subscription.Name, true, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - empty subscriptionName", "1", "0", "", true, expectedNotificationCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationBySubscriptionNameRoute, http.NoBody)
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
			handler := http.HandlerFunc(controller.NotificationsBySubscriptionName)
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
				var res responseDTO.MultiNotificationsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestCleanupNotificationByAge(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("CleanupNotificationsByAge", int64(0)).Return(nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	nc := NewNotificationController(dic)
	assert.NotNil(t, nc)

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
			req, err := http.NewRequest(http.MethodDelete, common.ApiNotificationCleanupByAgeRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Age: testCase.age})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(nc.CleanupNotificationsByAge)
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

func TestCleanupNotification(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("CleanupNotificationsByAge", int64(0)).Return(nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	nc := NewNotificationController(dic)
	assert.NotNil(t, nc)

	tests := []struct {
		name               string
		age                string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid", "0", false, http.StatusAccepted},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, common.ApiNotificationCleanupRoute, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(nc.CleanupNotifications)
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
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestDeleteNotificationByAge(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteProcessedNotificationsByAge", int64(0)).Return(nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	nc := NewNotificationController(dic)
	assert.NotNil(t, nc)

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
			req, err := http.NewRequest(http.MethodDelete, common.ApiNotificationByAgeRoute, http.NoBody)
			req = mux.SetURLVars(req, map[string]string{common.Age: testCase.age})
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(nc.DeleteProcessedNotificationsByAge)
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
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
