//
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/labstack/echo/v4"
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
			e := echo.New()
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiNotificationRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			err = controller.AddNotification(c)
			require.NoError(t, err)

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
			e := echo.New()
			reqPath := fmt.Sprintf("%s/%s", common.ApiNotificationByIdRoute, testCase.notificationId)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Id)
			c.SetParamValues(testCase.notificationId)
			err = controller.NotificationById(c)
			require.NoError(t, err)

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
	expectedNotificationCount := int64(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationCountByCategory", testCategory, "").Return(expectedNotificationCount, nil)
	dbClientMock.On("NotificationsByCategory", 0, 20, "", testCategory).Return([]models.Notification{}, nil)
	dbClientMock.On("NotificationsByCategory", 0, 1, "", testCategory).Return([]models.Notification{}, nil)
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
		expectedTotalCount int64
		expectedStatusCode int
	}{
		{"Valid - get notifications without offset, and limit", "", "", testCategory, false, expectedNotificationCount, http.StatusOK},
		{"Valid - get notifications with offset, and limit", "0", "1", testCategory, false, expectedNotificationCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testCategory, true, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testCategory, true, expectedNotificationCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationByCategoryRoute, http.NoBody)
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
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Category)
			c.SetParamValues(testCase.category)
			err = controller.NotificationsByCategory(c)
			require.NoError(t, err)

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
	expectedNotificationCount := int64(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationCountByLabel", testLabel, "").Return(expectedNotificationCount, nil)
	dbClientMock.On("NotificationsByLabel", 0, 20, "", testLabel).Return([]models.Notification{}, nil)
	dbClientMock.On("NotificationsByLabel", 0, 1, "", testLabel).Return([]models.Notification{}, nil)
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
		expectedTotalCount int64
		expectedStatusCode int
	}{
		{"Valid - get notifications without offset, and limit", "", "", testLabel, false, expectedNotificationCount, http.StatusOK},
		{"Valid - get notifications with offset, and limit", "0", "1", testLabel, false, expectedNotificationCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testLabel, true, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testLabel, true, expectedNotificationCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationByLabelRoute, http.NoBody)
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
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Label)
			c.SetParamValues(testCase.label)
			err = controller.NotificationsByLabel(c)
			require.NoError(t, err)

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
	expectedNotificationCount := int64(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationCountByStatus", testStatus, "").Return(expectedNotificationCount, nil)
	dbClientMock.On("NotificationsByStatus", 0, 20, "", testStatus).Return([]models.Notification{}, nil)
	dbClientMock.On("NotificationsByStatus", 0, 1, "", testStatus).Return([]models.Notification{}, nil)
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
		expectedTotalCount int64
		expectedStatusCode int
	}{
		{"Valid - get notifications without offset, and limit", "", "", testStatus, false, expectedNotificationCount, http.StatusOK},
		{"Valid - get notifications with offset, and limit", "0", "1", testStatus, false, expectedNotificationCount, http.StatusOK},
		{"Invalid - invalid offset format", "aaa", "1", testStatus, true, expectedNotificationCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "1", "aaa", testStatus, true, expectedNotificationCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationByStatusRoute, http.NoBody)
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
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Status)
			c.SetParamValues(testCase.status)
			err = controller.NotificationsByStatus(c)
			require.NoError(t, err)

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
	expectedNotificationCount := int64(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("NotificationCountByTimeRange", int64(0), int64(100), "").Return(expectedNotificationCount, nil)
	dbClientMock.On("NotificationsByTimeRange", int64(0), int64(100), 0, 10, "").Return([]models.Notification{}, nil)
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
		expectedTotalCount int64
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
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationByTimeRangeRoute, http.NoBody)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Start, common.End)
			c.SetParamValues(testCase.start, testCase.end)
			err = nc.NotificationsByTimeRange(c)
			require.NoError(t, err)

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
	dbClientMock.On("NotificationById", notification.Id).Return(notification, nil)
	dbClientMock.On("DeleteNotificationById", notification.Id).Return(nil)
	dbClientMock.On("NotificationById", notFoundId).Return(notification, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "subscription doesn't exist in the database", nil))
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
			e := echo.New()
			reqPath := fmt.Sprintf("%s/%s", common.ApiNotificationByIdRoute, testCase.notificationId)
			req, err := http.NewRequest(http.MethodDelete, reqPath, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Id)
			c.SetParamValues(testCase.notificationId)
			err = controller.DeleteNotificationById(c)
			require.NoError(t, err)
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
	expectedNotificationCount := int64(0)
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("SubscriptionByName", subscription.Name).Return(subscription, nil)
	dbClientMock.On("NotificationCountByCategoriesAndLabels", subscription.Categories, subscription.Labels, "").Return(expectedNotificationCount, nil)
	dbClientMock.On("NotificationsByCategoriesAndLabels", 0, 20, subscription.Categories, subscription.Labels, "").Return([]models.Notification{}, int64(0), nil)
	dbClientMock.On("NotificationsByCategoriesAndLabels", 0, 1, subscription.Categories, subscription.Labels, "").Return([]models.Notification{}, int64(0), nil)
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
		expectedTotalCount int64
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
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationBySubscriptionNameRoute, http.NoBody)
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
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Name)
			c.SetParamValues(testCase.subscriptionName)
			err = controller.NotificationsBySubscriptionName(c)
			require.NoError(t, err)

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
			e := echo.New()
			req, err := http.NewRequest(http.MethodDelete, common.ApiNotificationCleanupByAgeRoute, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Age)
			c.SetParamValues(testCase.age)
			err = nc.CleanupNotificationsByAge(c)
			require.NoError(t, err)

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
			e := echo.New()
			req, err := http.NewRequest(http.MethodDelete, common.ApiNotificationCleanupRoute, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			err = nc.CleanupNotifications(c)
			require.NoError(t, err)

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
			e := echo.New()
			req, err := http.NewRequest(http.MethodDelete, common.ApiNotificationByAgeRoute, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Age)
			c.SetParamValues(testCase.age)
			err = nc.DeleteProcessedNotificationsByAge(c)
			require.NoError(t, err)

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

func TestNotificationsByQueryConditions(t *testing.T) {
	start := int64(1556985370116)
	end := int64(1756985370116)
	expectedTotalCount := int64(10)
	queryCondition := requests.NotificationQueryCondition{Category: []string{"test"}, Start: start, End: end}
	queryConditionNoCategory := requests.NotificationQueryCondition{Category: []string{}, Start: start, End: end}
	queryConditionWithoutCategoryField := requests.NotificationQueryCondition{Start: start, End: end}
	queryConditionWithoutStartField := requests.NotificationQueryCondition{Category: []string{"test"}, End: end}
	queryConditionWithoutEndField := requests.NotificationQueryCondition{Category: []string{"test"}, Start: start}
	queryConditionDbError := requests.NotificationQueryCondition{Category: []string{"dbError"}, Start: start, End: end}
	var notifications []models.Notification
	methodNotificationsByQueryConditions := "NotificationsByQueryConditions"
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On(methodNotificationsByQueryConditions, 0, 20, queryCondition, "").Return(notifications, nil)
	dbClientMock.On(methodNotificationsByQueryConditions, 0, 20, queryConditionNoCategory, "").Return(notifications, nil)
	dbClientMock.On(methodNotificationsByQueryConditions, 0, 20, queryConditionWithoutCategoryField, "").Return(
		notifications, nil)
	dbClientMock.On(methodNotificationsByQueryConditions, 0, 20, queryConditionWithoutStartField, "").Return(
		notifications, nil)
	dbClientMockDefaultEndValue := queryConditionWithoutEndField
	dbClientMockDefaultEndValue.End = defaultEnd
	dbClientMock.On(methodNotificationsByQueryConditions, 0, 20, dbClientMockDefaultEndValue, "").Return(
		notifications, nil)
	dbClientMock.On(methodNotificationsByQueryConditions, 0, 20, queryConditionDbError, "").Return([]models.Notification{},
		errors.NewCommonEdgeX(errors.KindDatabaseError, "DB error", nil))
	dbClientMock.On("NotificationCountByQueryConditions", mock.Anything, "").Return(expectedTotalCount, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	nc := NewNotificationController(dic)
	assert.NotNil(t, nc)
	tests := []struct {
		name               string
		offset             string
		limit              string
		condition          *requests.NotificationQueryCondition
		errorExpected      bool
		expectedTotalCount int64
		expectedStatusCode int
	}{
		{"Valid", "0", "20", &queryCondition, false, expectedTotalCount, http.StatusOK},
		{"Valid - category no specified", "0", "20", &queryConditionNoCategory, false, expectedTotalCount, http.StatusOK},
		{"Valid - without category field", "0", "20", &queryConditionWithoutCategoryField, false, expectedTotalCount, http.StatusOK},
		{"Valid - without start field", "0", "20", &queryConditionWithoutStartField, false, expectedTotalCount, http.StatusOK},
		{"Valid - without end field", "0", "20", &queryConditionWithoutEndField, false, expectedTotalCount, http.StatusOK},
		{"Invalid - empty request body", "0", "20", nil, true, 0, http.StatusBadRequest},
		{"Invalid - offset with unparsable format", "aaa", "20", &queryCondition, true, 0, http.StatusBadRequest},
		{"Invalid - limit with unparsable format", "0", "aaa", &queryCondition, true, 0, http.StatusBadRequest},
		{"Invalid - database error", "0", "20", &queryConditionDbError, true, 0, http.StatusInternalServerError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var reader io.Reader
			if testCase.condition != nil {
				reqDTO := requests.GetNotificationRequest{}
				reqDTO.ApiVersion = common.ApiVersion
				reqDTO.QueryCondition = *testCase.condition
				byteData, err := json.Marshal(reqDTO)
				require.NoError(t, err)
				reader = strings.NewReader(string(byteData))
			} else {
				reader = http.NoBody
			}

			req, err := http.NewRequest(http.MethodGet, common.ApiNotificationRoute, reader)
			req.Header.Set(common.ContentType, common.ContentTypeJSON)
			require.NoError(t, err)
			query := req.URL.Query()
			query.Add(common.Offset, testCase.offset)
			query.Add(common.Limit, testCase.limit)
			req.URL.RawQuery = query.Encode()

			// Act
			recorder := httptest.NewRecorder()
			c := echo.New().NewContext(req, recorder)
			err = nc.NotificationsByQueryConditions(c)
			require.NoError(t, err)

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
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Total count not as expected")
			}
		})

	}
}

func TestDeleteNotificationByIds(t *testing.T) {
	ids := []string{"1793b2b5-1873-44da-9dbc-73a8bcb1f567", "f76e4602-419d-4a90-a7e4-4110c781eb0e"}
	var noId []string
	dbError := []string{"dbError"}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteNotificationByIds", ids).Return(nil)
	dbClientMock.On("DeleteNotificationByIds", dbError).Return(errors.NewCommonEdgeX(
		errors.KindDatabaseError, "DB error", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewNotificationController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		ids                []string
		expectedStatusCode int
	}{
		{"Valid - delete notification by ids", ids, http.StatusOK},
		{"Invalid - ids parameter is empty", noId, http.StatusBadRequest},
		{"Invalid - database error", dbError, http.StatusInternalServerError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiNotificationByIdsRoute, testCase.ids)
			req, err := http.NewRequest(http.MethodDelete, reqPath, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := echo.New().NewContext(req, recorder)
			c.SetParamNames(common.Ids)
			c.SetParamValues(strings.Join(testCase.ids[:], common.CommaSeparator))
			err = controller.DeleteNotificationByIds(c)
			require.NoError(t, err)

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

func TestUpdateNotificationAckStatusByIds(t *testing.T) {
	ids := []string{"1793b2b5-1873-44da-9dbc-73a8bcb1f567", "f76e4602-419d-4a90-a7e4-4110c781eb0e"}
	var noId []string
	dbError := []string{"dbError"}

	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("UpdateNotificationAckStatusByIds", true, ids).Return(nil)
	dbClientMock.On("UpdateNotificationAckStatusByIds", true, dbError).Return(errors.NewCommonEdgeX(
		errors.KindDatabaseError, "DB error", nil))
	dbClientMock.On("UpdateNotificationAckStatusByIds", false, ids).Return(nil)
	dbClientMock.On("UpdateNotificationAckStatusByIds", false, dbError).Return(errors.NewCommonEdgeX(
		errors.KindDatabaseError, "DB error", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	controller := NewNotificationController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		controller         func(c echo.Context) error
		ids                []string
		expectedStatusCode int
	}{
		{"Valid - acknowledge succeeded", controller.AcknowledgeNotificationByIds, ids, http.StatusOK},
		{"Invalid - ids parameter is empty", controller.AcknowledgeNotificationByIds, noId, http.StatusBadRequest},
		{"Invalid - database error", controller.AcknowledgeNotificationByIds, dbError, http.StatusInternalServerError},
		{"Valid - unacknowledge succeeded", controller.UnacknowledgeNotificationByIds, ids, http.StatusOK},
		{"Invalid - ids parameter is empty", controller.UnacknowledgeNotificationByIds, noId, http.StatusBadRequest},
		{"Invalid - database error", controller.UnacknowledgeNotificationByIds, dbError, http.StatusInternalServerError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			reqPath := fmt.Sprintf("%s/%s", common.ApiNotificationAcknowledgeByIdsRoute, testCase.ids)
			req, err := http.NewRequest(http.MethodPut, reqPath, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := echo.New().NewContext(req, recorder)
			c.SetParamNames(common.Ids)
			c.SetParamValues(strings.Join(testCase.ids[:], common.CommaSeparator))
			err = testCase.controller(c)
			require.NoError(t, err)

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
