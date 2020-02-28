/*******************************************************************************
 * Copyright 2019 VMware Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package notifications

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	notificationsConfig "github.com/edgexfoundry/edgex-go/internal/support/notifications/config"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces/mocks"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

// TestURI this is not really used since we are using the HTTP testing framework and not creating routes, but rather
// creating a specific handler which will accept all requests. Therefore, the URI is not important.
var TestURI = "/notification"
var TestId = "123e4567-e89b-12d3-a456-426655440000"
var TestSlug = "test-slug"
var TestAge = 1564594093
var TestLimit = 5
var TestTooLargeLimit = 100
var TestInvalidLimit = "invalid-limit"
var TestInvalidAge = "invalid age"
var TestSender = "System Management"
var TestStart int64 = 1564758450
var TestEnd int64 = 1564758650
var notificationId = "526c5c28-7a21-48a8-90f6-8009400441f4"
var invalidNotificationId = "..."
var badNotificationId = "11111111-7a21-48a8-90f6-8009400441f4"
var notificationServiceURI = clients.ApiBase + "/" + NOTIFICATIONSERVICE
var testError = errors.New("some error")

var TestLabels = []string{
	"test_label",
	"test_label2",
}

var TestCategories = []string{
	"test_category",
}

const (
	NOTIFICATIONSERVICE = "notification"
)

func TestGetNotificationById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequest(map[string]string{ID: TestId}),
			createMockNotificationLoader("GetNotificationById", TestId, nil),
			http.StatusOK,
		},
		{
			name:           "Notification not found",
			request:        createRequest(map[string]string{ID: TestId}),
			dbMock:         createMockNotificationLoader("GetNotificationById", TestId, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Other error from database",
			request:        createRequest(map[string]string{ID: TestId}),
			dbMock:         createMockNotificationLoader("GetNotificationById", TestId, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetNotificationByID(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetNotificationBySlug(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequest(map[string]string{SLUG: TestSlug}),
			createMockNotificationLoader("GetNotificationBySlug", TestSlug, nil),
			http.StatusOK,
		},
		{
			name:           "Notification not found",
			request:        createRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockNotificationLoader("GetNotificationBySlug", TestSlug, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Other error from database",
			request:        createRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockNotificationLoader("GetNotificationBySlug", TestSlug, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetNotificationBySlug(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func createMockNotificationLoader(methodName string, testID string, desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, testID).Return(contract.Notification{}, desiredError)
	} else {
		myMock.On(methodName, testID).Return(createNotifications(1)[0], nil)
	}
	return &myMock
}

func createMockNotificationDeleter(methodName string, testID string, desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, testID).Return(desiredError)
	} else {
		myMock.On(methodName, testID).Return(nil)
	}
	return &myMock
}

func createMockNotificationAgeDeleter(methodName string, testAge int, desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, testAge).Return(desiredError)
	} else {
		myMock.On(methodName, testAge).Return(nil)
	}
	return &myMock
}

func createMockNotificationSenderLoader(methodName string, sender string, limit int, desiredError error, ret interface{}) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, sender, limit).Return(ret, desiredError)
	} else {
		myMock.On(methodName, sender, limit).Return(ret, nil)
	}
	return &myMock
}

func createMockNotificationStartLoader(methodName string, start int64, limit int, desiredError error, ret interface{}) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, start, limit).Return(ret, desiredError)
	} else {
		myMock.On(methodName, start, limit).Return(ret, nil)
	}
	return &myMock
}

func createMockNotificationStartEndLoader(methodName string, start int64, end int64, limit int, desiredError error, ret interface{}) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, start, end, limit).Return(ret, desiredError)
	} else {
		myMock.On(methodName, start, end, limit).Return(ret, nil)
	}
	return &myMock
}

func createMockNotificationLabelsLoader(methodName string, labels []string, limit int, desiredError error, ret interface{}) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, labels, limit).Return(ret, desiredError)
	} else {
		myMock.On(methodName, labels, limit).Return(ret, nil)
	}
	return &myMock
}

func createMockNotificationNewestLoader(methodName string, limit int, desiredError error, ret interface{}) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, limit).Return(ret, desiredError)
	} else {
		myMock.On(methodName, limit).Return(ret, nil)
	}
	return &myMock
}

func createRequest(params map[string]string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, TestURI, nil)
	return mux.SetURLVars(req, params)
}

func createDeleteRequest(params map[string]string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, TestURI, nil)
	return mux.SetURLVars(req, params)
}

func createNotifications(howMany int) []contract.Notification {
	var notifications []contract.Notification
	for i := 0; i < howMany; i++ {
		notifications = append(notifications, contract.Notification{
			Slug:     "notice-test-123",
			Sender:   "System Management",
			Category: "SECURITY",
			Severity: "CRITICAL",
			Content:  "Hello, Notification!",
			Labels: []string{
				"cool",
				"test",
			},
		})
	}
	return notifications
}

func createNotificationBySeverityLevel(severityLevel string) contract.Notification {
	var notification contract.Notification

	switch severityLevel {
	case contract.Critical:
		notification = contract.Notification{
			ID:       notificationId,
			Slug:     "notice-critical-123",
			Sender:   "Sender A",
			Category: "SECURITY",
			Severity: contract.Critical,
			Content:  "Hello, Notification!",
			Status:   "NEW",
			Labels: []string{
				"first-label",
				"second-label",
			},
		}
	case contract.Normal:
		notification = contract.Notification{
			ID:       notificationId,
			Slug:     "notice-normal-123",
			Sender:   "Sender B",
			Category: "SECURITY",
			Severity: contract.Normal,
			Content:  "Hello, Notification!",
			Status:   "NEW",
			Labels: []string{
				"first-label",
				"second-label",
			},
		}
	default:
		//	...
	}

	return notification
}

func createInvalidNotification() contract.Notification {
	var notification contract.Notification

	notification = contract.Notification{
		ID:       invalidNotificationId,
		Slug:     "...",
		Sender:   "...",
		Category: "...",
		Severity: contract.Critical,
		Content:  "...",
		Status:   "...",
		Labels: []string{
			"...",
			"...",
		},
	}
	return notification
}

func createInvalidCategoriesAndLabelsNotification() contract.Notification {
	var notification contract.Notification

	notification = contract.Notification{
		ID:       notificationId,
		Slug:     "notice-critical-123",
		Sender:   "Sender A",
		Category: "...",
		Severity: contract.Critical,
		Content:  "Hello, Notification!",
		Status:   "NEW",
		Labels: []string{
			"first-bad-label",
			"second-bad-label",
		},
	}
	return notification
}

func createBadNotification() contract.Notification {
	var notification contract.Notification

	notification = contract.Notification{
		ID:       badNotificationId,
		Slug:     "notice-normal-123",
		Sender:   "Sender B",
		Category: "SECURITY",
		Severity: contract.Critical,
		Content:  "Hello, Notification!",
		Status:   "NEW",
		Labels: []string{
			"first-label",
			"second-label",
		},
	}
	return notification
}

// This function serves to update the unexported isValidated field (in "go-mod-core-contracts"),
// which can only be done by marshalling and unmarshalling to JSON.
func validateNotification(notification *contract.Notification) contract.Notification {
	b, _ := json.Marshal(notification)
	_ = notification.UnmarshalJSON(b)
	return *notification
}

func createNotificationHandlerRequestWithBody(
	httpMethod string,
	notification contract.Notification,
	pathParams map[string]string) *http.Request {

	// if your JSON marshalling fails you've got bigger problems
	body, _ := json.Marshal(notification)

	req := httptest.NewRequest(httpMethod, notificationServiceURI, bytes.NewReader(body))

	return mux.SetURLVars(req, pathParams)
}

type mockOutline struct {
	methodName string
	arg        []interface{}
	ret        []interface{}
}

func createMockWithOutlines(outlines []mockOutline) interfaces.DBClient {
	dbMock := mocks.DBClient{}

	for _, o := range outlines {
		dbMock.On(o.methodName, o.arg...).Return(o.ret...)
	}

	return &dbMock
}

func TestDeleteNotificationById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createDeleteRequest(map[string]string{ID: TestId}),
			dbMock:         createMockNotificationDeleter("DeleteNotificationById", TestId, nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unknown Error",
			request:        createDeleteRequest(map[string]string{ID: TestId}),
			dbMock:         createMockNotificationDeleter("DeleteNotificationById", TestId, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Notification not found",
			request:        createDeleteRequest(map[string]string{ID: TestId}),
			dbMock:         createMockNotificationDeleter("DeleteNotificationById", TestId, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restDeleteNotificationByID(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}
func TestDeleteNotificationBySlug(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createDeleteRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockNotificationDeleter("DeleteNotificationBySlug", TestSlug, nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unknown Error",
			request:        createDeleteRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockNotificationDeleter("DeleteNotificationBySlug", TestSlug, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Notification not found",
			request:        createDeleteRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockNotificationDeleter("DeleteNotificationBySlug", TestSlug, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restDeleteNotificationBySlug(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestDeleteNotificationsByAge(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createDeleteRequest(map[string]string{AGE: strconv.Itoa(TestAge)}),
			dbMock:         createMockNotificationAgeDeleter("DeleteNotificationsOld", TestAge, nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unknown Error",
			request:        createDeleteRequest(map[string]string{AGE: strconv.Itoa(TestAge)}),
			dbMock:         createMockNotificationAgeDeleter("DeleteNotificationsOld", TestAge, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Unknown Error",
			request:        createDeleteRequest(map[string]string{AGE: TestInvalidAge}),
			dbMock:         createMockNotificationAgeDeleter("DeleteNotificationsOld", TestAge, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restDeleteNotificationsByAge(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetNotificationsBySender(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createDeleteRequest(map[string]string{SENDER: TestSender, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationSenderLoader("GetNotificationBySender", TestSender, TestLimit, nil, createNotifications(1)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error converting string to integer",
			request:        createDeleteRequest(map[string]string{SENDER: TestSender, LIMIT: TestInvalidLimit}),
			dbMock:         createMockNotificationSenderLoader("GetNotificationBySender", TestSender, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Limit too large Error",
			request:        createDeleteRequest(map[string]string{SENDER: TestSender, LIMIT: strconv.Itoa(TestTooLargeLimit)}),
			dbMock:         createMockNotificationSenderLoader("GetNotificationBySender", TestSender, TestTooLargeLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "Not found",
			request:        createDeleteRequest(map[string]string{SENDER: TestSender, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationSenderLoader("GetNotificationBySender", TestSender, TestLimit, nil, []contract.Notification{}),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unknown Error",
			request:        createDeleteRequest(map[string]string{SENDER: TestSender, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationSenderLoader("GetNotificationBySender", TestSender, TestLimit, errors.New("Test error"), createNotifications(1)),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetNotificationsBySender(
				rr,
				tt.request,
				logger.NewMockClient(),
				tt.dbMock,
				notificationsConfig.ConfigurationStruct{Service: bootstrapConfig.ServiceInfo{MaxResultCount: 5}})
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetNotificationsByStart(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByStart", TestStart, TestLimit, nil, createNotifications(1)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error converting string to integer start",
			request:        createRequest(map[string]string{START: TestInvalidLimit, END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByStart", TestStart, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Error converting string to integer limit",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: TestInvalidLimit}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByStart", TestStart, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Limit too large Error",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestTooLargeLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByStart", TestStart, TestTooLargeLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "Not found",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByStart", TestStart, TestLimit, nil, []contract.Notification{}),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unknown Error",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByStart", TestStart, TestLimit, errors.New("Test error"), createNotifications(1)),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restNotificationByStart(
				rr,
				tt.request,
				logger.NewMockClient(),
				tt.dbMock,
				notificationsConfig.ConfigurationStruct{Service: bootstrapConfig.ServiceInfo{MaxResultCount: 5}})
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetNotificationsByEnd(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByEnd", TestEnd, TestLimit, nil, createNotifications(1)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error converting string to integer end",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: TestInvalidLimit, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByEnd", TestEnd, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Error converting string to integer limit",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: TestInvalidLimit}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByEnd", TestEnd, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Limit too large Error",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestTooLargeLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByEnd", TestEnd, TestTooLargeLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "Not found",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByEnd", TestEnd, TestLimit, nil, []contract.Notification{}),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unknown Error",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByEnd", TestEnd, TestLimit, errors.New("Test error"), createNotifications(1)),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restNotificationByEnd(
				rr,
				tt.request,
				logger.NewMockClient(),
				tt.dbMock,
				notificationsConfig.ConfigurationStruct{Service: bootstrapConfig.ServiceInfo{MaxResultCount: 5}})
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetNotificationsByStartEnd(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartEndLoader("GetNotificationsByStartEnd", TestStart, TestEnd, TestLimit, nil, createNotifications(1)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error converting string to integer start",
			request:        createRequest(map[string]string{START: TestInvalidLimit, END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartEndLoader("GetNotificationsByStartEnd", TestStart, TestEnd, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Error converting string to integer end",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: TestInvalidLimit, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartEndLoader("GetNotificationsByStartEnd", TestStart, TestEnd, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Error converting string to integer limit",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: TestInvalidLimit}),
			dbMock:         createMockNotificationStartEndLoader("GetNotificationsByStartEnd", TestStart, TestEnd, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Limit too large Error",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestTooLargeLimit)}),
			dbMock:         createMockNotificationStartEndLoader("GetNotificationsByStartEnd", TestStart, TestEnd, TestTooLargeLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "Not Found",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartEndLoader("GetNotificationsByStartEnd", TestStart, TestEnd, TestLimit, nil, []contract.Notification{}),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unknown Error",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartEndLoader("GetNotificationsByStartEnd", TestStart, TestEnd, TestLimit, errors.New("Test error"), createNotifications(1)),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restNotificationByStartEnd(
				rr,
				tt.request,
				logger.NewMockClient(),
				tt.dbMock,
				notificationsConfig.ConfigurationStruct{Service: bootstrapConfig.ServiceInfo{MaxResultCount: 5}})
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetNotificationsByLabels(t *testing.T) {
	labelsURL := strings.Join(TestLabels, ",")
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createRequest(map[string]string{LABELS: labelsURL, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationLabelsLoader("GetNotificationsByLabels", TestLabels, TestLimit, nil, createNotifications(1)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error converting string to integer limit",
			request:        createRequest(map[string]string{LABELS: labelsURL, LIMIT: TestInvalidLimit}),
			dbMock:         createMockNotificationLabelsLoader("GetNotificationsByLabels", TestLabels, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Limit too large Error",
			request:        createRequest(map[string]string{LABELS: labelsURL, LIMIT: strconv.Itoa(TestTooLargeLimit)}),
			dbMock:         createMockNotificationLabelsLoader("GetNotificationsByLabels", TestLabels, TestTooLargeLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "Not Found",
			request:        createRequest(map[string]string{LABELS: labelsURL, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationLabelsLoader("GetNotificationsByLabels", TestLabels, TestLimit, nil, []contract.Notification{}),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unknown Error",
			request:        createRequest(map[string]string{LABELS: labelsURL, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationLabelsLoader("GetNotificationsByLabels", TestLabels, TestLimit, errors.New("Test error"), createNotifications(1)),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restNotificationsByLabels(
				rr,
				tt.request,
				logger.NewMockClient(),
				tt.dbMock,
				notificationsConfig.ConfigurationStruct{Service: bootstrapConfig.ServiceInfo{MaxResultCount: 5}})
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetNotificationsNewest(t *testing.T) {
	labelsURL := strings.Join(TestLabels, ",")
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createRequest(map[string]string{LABELS: labelsURL, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationNewestLoader("GetNewNotifications", TestLimit, nil, createNotifications(1)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error converting string to integer limit",
			request:        createRequest(map[string]string{LABELS: labelsURL, LIMIT: TestInvalidLimit}),
			dbMock:         createMockNotificationNewestLoader("GetNewNotifications", TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Limit too large Error",
			request:        createRequest(map[string]string{LABELS: labelsURL, LIMIT: strconv.Itoa(TestTooLargeLimit)}),
			dbMock:         createMockNotificationNewestLoader("GetNewNotifications", TestTooLargeLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "Not Found",
			request:        createRequest(map[string]string{LABELS: labelsURL, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationNewestLoader("GetNewNotifications", TestLimit, nil, []contract.Notification{}),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unknown Error",
			request:        createRequest(map[string]string{LABELS: labelsURL, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationNewestLoader("GetNewNotifications", TestLimit, errors.New("Test error"), createNotifications(1)),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restNotificationsNew(
				rr,
				tt.request,
				logger.NewMockClient(),
				tt.dbMock,
				notificationsConfig.ConfigurationStruct{Service: bootstrapConfig.ServiceInfo{MaxResultCount: 5}})
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestNotificationHandler(t *testing.T) {

	notificationNormal := createNotificationBySeverityLevel(contract.Normal)
	notificationCritical := createNotificationBySeverityLevel(contract.Critical)
	notificationInvalid := createInvalidNotification()
	notificationBad := createBadNotification()
	notificationInvalidCategoriesAndLabels := createInvalidCategoriesAndLabelsNotification()

	var categories []string
	categories = append(categories, string(notificationNormal.Category))
	var labels = []string{"first-label", "second-label"}

	var badCategories []string
	badCategories = append(badCategories, string(notificationInvalidCategoriesAndLabels.Category))
	var badLabels = []string{"first-bad-label", "second-bad-label"}

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"ok normal notification",
			createNotificationHandlerRequestWithBody(http.MethodPost, notificationNormal, nil),
			createMockWithOutlines([]mockOutline{
				{"AddNotification", []interface{}{validateNotification(&notificationNormal)}, []interface{}{notificationId, nil}},
				{"GetNotificationById", []interface{}{notificationId}, []interface{}{notificationNormal, nil}},
				{"GetSubscriptionByCategoriesLabels", []interface{}{categories, labels}, []interface{}{[]contract.Subscription{}, nil}},
				{"MarkNotificationProcessed", []interface{}{validateNotification(&notificationNormal)}, []interface{}{nil}},
			}),
			http.StatusAccepted,
		},
		{
			"ok critical notification",
			createNotificationHandlerRequestWithBody(http.MethodPost, notificationCritical, nil),
			createMockWithOutlines([]mockOutline{
				{"AddNotification", []interface{}{validateNotification(&notificationCritical)}, []interface{}{notificationId, nil}},
				{"GetNotificationById", []interface{}{notificationId}, []interface{}{notificationCritical, nil}},
				{"GetSubscriptionByCategoriesLabels", []interface{}{categories, labels}, []interface{}{[]contract.Subscription{}, nil}},
				{"MarkNotificationProcessed", []interface{}{validateNotification(&notificationCritical)}, []interface{}{nil}},
			}),
			http.StatusAccepted,
		},
		{
			"notification validation error",
			createNotificationHandlerRequestWithBody(http.MethodPost, notificationInvalid, nil),
			createMockWithOutlines([]mockOutline{
				{"AddNotification", []interface{}{validateNotification(&notificationInvalid)}, []interface{}{notificationId, nil}},
				{"GetNotificationById", []interface{}{invalidNotificationId}, []interface{}{notificationInvalid, nil}},
				{"GetSubscriptionByCategoriesLabels", []interface{}{categories, labels}, []interface{}{[]contract.Subscription{}, nil}},
				{"MarkNotificationProcessed", []interface{}{validateNotification(&notificationInvalid)}, []interface{}{nil}},
			}),
			http.StatusBadRequest,
		},
		{
			"add notification error",
			createNotificationHandlerRequestWithBody(http.MethodPost, notificationCritical, nil),
			createMockWithOutlines([]mockOutline{
				{"AddNotification", []interface{}{validateNotification(&notificationCritical)}, []interface{}{invalidNotificationId, testError}},
				{"GetNotificationById", []interface{}{notificationId}, []interface{}{notificationCritical, nil}},
				{"GetSubscriptionByCategoriesLabels", []interface{}{categories, labels}, []interface{}{[]contract.Subscription{}, nil}},
				{"MarkNotificationProcessed", []interface{}{validateNotification(&notificationCritical)}, []interface{}{nil}},
			}),
			http.StatusConflict,
		},
		{
			"get notification error",
			createNotificationHandlerRequestWithBody(http.MethodPost, notificationCritical, nil),
			createMockWithOutlines([]mockOutline{
				{"AddNotification", []interface{}{validateNotification(&notificationCritical)}, []interface{}{notificationId, nil}},
				{"GetNotificationById", []interface{}{notificationId}, []interface{}{notificationBad, testError}},
				{"GetSubscriptionByCategoriesLabels", []interface{}{categories, labels}, []interface{}{[]contract.Subscription{}, nil}},
				{"MarkNotificationProcessed", []interface{}{validateNotification(&notificationCritical)}, []interface{}{nil}},
			}),
			http.StatusInternalServerError,
		},
		{
			"distribute and mark notification ok",
			createNotificationHandlerRequestWithBody(http.MethodPost, notificationCritical, nil),
			createMockWithOutlines([]mockOutline{
				{"AddNotification", []interface{}{validateNotification(&notificationCritical)}, []interface{}{notificationId, nil}},
				{"GetNotificationById", []interface{}{notificationId}, []interface{}{notificationCritical, nil}},
				{"GetSubscriptionByCategoriesLabels", []interface{}{categories, labels}, []interface{}{[]contract.Subscription{}, nil}},
				{"MarkNotificationProcessed", []interface{}{validateNotification(&notificationCritical)}, []interface{}{nil}},
			}),
			http.StatusAccepted,
		},
		{
			"distribute and mark notification processed",
			createNotificationHandlerRequestWithBody(http.MethodPost, notificationCritical, nil),
			createMockWithOutlines([]mockOutline{
				{"AddNotification", []interface{}{validateNotification(&notificationCritical)}, []interface{}{notificationId, nil}},
				{"GetNotificationById", []interface{}{notificationId}, []interface{}{notificationCritical, nil}},
				{"GetSubscriptionByCategoriesLabels", []interface{}{categories, labels}, []interface{}{[]contract.Subscription{}, testError}},
				{"MarkNotificationProcessed", []interface{}{validateNotification(&notificationCritical)}, []interface{}{testError}},
			}),
			http.StatusOK,
		},
		{
			"distribute and mark notification error",
			createNotificationHandlerRequestWithBody(http.MethodPost, notificationInvalidCategoriesAndLabels, nil),
			createMockWithOutlines([]mockOutline{
				{"AddNotification", []interface{}{validateNotification(&notificationInvalidCategoriesAndLabels)}, []interface{}{notificationId, nil}},
				{"GetNotificationById", []interface{}{notificationId}, []interface{}{notificationInvalidCategoriesAndLabels, nil}},
				{"GetSubscriptionByCategoriesLabels", []interface{}{badCategories, badLabels}, []interface{}{[]contract.Subscription{}, testError}},
				{"MarkNotificationProcessed", []interface{}{validateNotification(&notificationInvalidCategoriesAndLabels)}, []interface{}{testError}},
			}),
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			notificationHandler(rr,
				tt.request,
				logger.NewMockClient(),
				tt.dbMock,
				notificationsConfig.ConfigurationStruct{Service: bootstrapConfig.ServiceInfo{MaxResultCount: 5}})
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}
