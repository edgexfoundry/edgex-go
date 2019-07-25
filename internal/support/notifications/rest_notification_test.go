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

func TestGetNotificationById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequest(ID, TestId),
			createMockNotificationLoaderForId(nil),
			http.StatusOK,
		},
		{
			name:           "Notification not found",
			request:        createRequest(ID, TestId),
			dbMock:         createMockNotificationLoaderForId(db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Other error from database",
			request:        createRequest(ID, TestId),
			dbMock:         createMockNotificationLoaderForId(errors.New("Test error")),
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
func TestMain(m *testing.M) {
	Configuration = &ConfigurationStruct{}
	LoggingClient = logger.NewMockClient()
	os.Exit(m.Run())
}

func createMockNotificationLoaderStringArg(howMany int, methodName string, arg string) interfaces.DBClient {
	notifications := createNotifications(howMany)

	myMock := mocks.DBClient{}
	myMock.On(methodName, arg).Return(notifications, nil)
	return &myMock
}
func createMockNotificationLoaderForId(desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On("GetNotificationById", TestId).Return(contract.Notification{}, desiredError)
	} else {
		myMock.On("GetNotificationById", TestId).Return(createNotifications(1)[0], nil)
	}
	return &myMock
}

func createMockNotificationDeleterForId(desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On("DeleteNotificationById", TestId).Return(desiredError)
	} else {
		myMock.On("DeleteNotificationById", TestId).Return(nil)
	}
	return &myMock
}

func createRequest(pathParamName string, pathParamValue string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, TestURI, nil)
	return mux.SetURLVars(req, map[string]string{pathParamName: pathParamValue})
}
func createDeleteRequest(pathParamName string, pathParamValue string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, TestURI, nil)
	return mux.SetURLVars(req, map[string]string{pathParamName: pathParamValue})
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
			request:        createDeleteRequest(ID, TestId),
			dbMock:         createMockNotificationDeleterForId(nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unknown Error",
			request:        createDeleteRequest(ID, TestId),
			dbMock:         createMockNotificationDeleterForId(errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Notification not found",
			request:        createDeleteRequest(ID, TestId),
			dbMock:         createMockNotificationDeleterForId(db.ErrNotFound),
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
