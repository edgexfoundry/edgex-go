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
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces/mocks"
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
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetNotificationByID)
			handler.ServeHTTP(rr, tt.request)
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
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetNotificationBySlug)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestMain(m *testing.M) {
	Configuration = &ConfigurationStruct{}
	LoggingClient = logger.NewMockClient()
	os.Exit(m.Run())
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
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restDeleteNotificationByID)
			handler.ServeHTTP(rr, tt.request)
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
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restDeleteNotificationBySlug)
			handler.ServeHTTP(rr, tt.request)
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
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restDeleteNotificationsByAge)
			handler.ServeHTTP(rr, tt.request)
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
			name:           "Unknown Error",
			request:        createDeleteRequest(map[string]string{SENDER: TestSender, LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationSenderLoader("GetNotificationBySender", TestSender, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 5}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetNotificationsBySender)
			handler.ServeHTTP(rr, tt.request)
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
			name:           "Error converting string to integer end",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: TestInvalidLimit, LIMIT: strconv.Itoa(TestLimit)}),
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
			name:           "Unknown Error",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartLoader("GetNotificationsByStart", TestStart, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 5}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restNotificationByStart)
			handler.ServeHTTP(rr, tt.request)
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
			name:           "Unknown Error",
			request:        createRequest(map[string]string{START: strconv.Itoa(int(TestStart)), END: strconv.Itoa(int(TestEnd)), LIMIT: strconv.Itoa(TestLimit)}),
			dbMock:         createMockNotificationStartEndLoader("GetNotificationsByStartEnd", TestStart, TestEnd, TestLimit, errors.New("Test error"), []contract.Notification{}),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 5}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restNotificationByStartEnd)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}
