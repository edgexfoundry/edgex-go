//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/infrastructure/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

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

var testAddNotificationRequest = dtos.Notification{
	Versionable: common.NewVersionable(),
	Category:    testNotificationCategory,
	Content:     testNotificationContent,
	ContentType: testNotificationContentType,
	Description: testNotificationDescription,
	Labels:      testNotificationLabels,
	Sender:      testNotificationSender,
	Severity:    testNotificationSeverity,
}

func TestAddNotification(t *testing.T) {
	dic := mockDic()
	dbClientMock := &dbMock.DBClient{}

	valid := requests.NewAddNotificationRequest(testAddNotificationRequest)
	model := dtos.ToNotificationModel(valid.Notification)
	dbClientMock.On("AddNotification", model).Return(model, nil)

	noRequestId := requests.NewAddNotificationRequest(testAddNotificationRequest)
	noRequestId.RequestId = ""
	invalidReqId := requests.NewAddNotificationRequest(testAddNotificationRequest)
	invalidReqId.RequestId = "abc"

	noCategoryAndLabels := requests.NewAddNotificationRequest(testAddNotificationRequest)
	noCategoryAndLabels.Notification.Category = ""
	noCategoryAndLabels.Notification.Labels = nil

	noContent := requests.NewAddNotificationRequest(testAddNotificationRequest)
	noContent.Notification.Content = ""

	noSender := requests.NewAddNotificationRequest(testAddNotificationRequest)
	noSender.Notification.Sender = ""

	noSeverity := requests.NewAddNotificationRequest(testAddNotificationRequest)
	noSeverity.Notification.Severity = ""
	invalidSeverity := requests.NewAddNotificationRequest(testAddNotificationRequest)
	invalidSeverity.Notification.Severity = "foo"

	invalidStatus := requests.NewAddNotificationRequest(testAddNotificationRequest)
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
		{"valid", []requests.AddNotificationRequest{valid}, http.StatusCreated},
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
