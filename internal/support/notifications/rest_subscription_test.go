/*******************************************************************************
 * Copyright 2019 VMware Technologies Inc.
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
 *
 *******************************************************************************/

package notifications

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

var TestSubscriptionURI = "/subscription"

func TestGetSubscriptionById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "Notification not found",
			request:        createSubscriptionRequest(map[string]string{ID: TestId}),
			dbMock:         createMocSubscriptionLoader("GetSubscriptionById", TestId, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			"OK",
			createSubscriptionRequest(map[string]string{ID: TestId}),
			createMocSubscriptionLoader("GetSubscriptionById", TestId, nil),
			http.StatusOK,
		},

		{
			name:           "Other error from database",
			request:        createSubscriptionRequest(map[string]string{ID: TestId}),
			dbMock:         createMocSubscriptionLoader("GetSubscriptionById", TestId, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetSubscriptionByID)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func createSubscriptionRequest(params map[string]string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, TestSubscriptionURI, nil)
	return mux.SetURLVars(req, params)
}

func createSubscriptionDeleteRequest(params map[string]string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, TestURI, nil)
	return mux.SetURLVars(req, params)
}

func createMocSubscriptionLoader(methodName string, testID string, desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, testID).Return(contract.Subscription{}, desiredError)
	} else {
		myMock.On(methodName, testID).Return(createSubscriptions(1)[0], nil)
	}
	return &myMock
}

func createMockSubscriptionDeleter(methodName string, testID string, desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, testID).Return(desiredError)
	} else {
		myMock.On(methodName, testID).Return(nil)
	}
	return &myMock
}

func createSubscriptions(howMany int) []contract.Subscription {
	var notifications []contract.Subscription
	for i := 0; i < howMany; i++ {
		notifications = append(notifications, contract.Subscription{
			Slug:     "notice-test-123",
			Receiver: "System Admin",
			SubscribedCategories: []contract.NotificationsCategory{
				"SECURITY",
				"HW_HEALTH",
				"SW_HEALTH",
			},
			SubscribedLabels: []string{
				"Dell",
				"IoT",
				"test",
			},
			Channels: []contract.Channel{
				{
					Type: "REST",
					Url:  "http://abc.def/alert",
				},
				{
					Type: "EMAIL",
					MailAddresses: []string{
						"cloud@abc.def",
						"jack@abc.def",
					},
				},
			},
		})
	}
	return notifications
}

func TestDeleteSubscriptionById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createSubscriptionDeleteRequest(map[string]string{ID: TestId}),
			dbMock:         createMockSubscriptionDeleter("DeleteSubscriptionById", TestId, nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unknown Error",
			request:        createSubscriptionDeleteRequest(map[string]string{ID: TestId}),
			dbMock:         createMockSubscriptionDeleter("DeleteSubscriptionById", TestId, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Subscription not found",
			request:        createSubscriptionDeleteRequest(map[string]string{ID: TestId}),
			dbMock:         createMockSubscriptionDeleter("DeleteSubscriptionById", TestId, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restDeleteSubscriptionByID)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetSubscriptionBySlug(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "Notification not found",
			request:        createSubscriptionRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMocSubscriptionLoader("GetSubscriptionBySlug", TestSlug, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			"OK",
			createSubscriptionRequest(map[string]string{SLUG: TestSlug}),
			createMocSubscriptionLoader("GetSubscriptionBySlug", TestSlug, nil),
			http.StatusOK,
		},

		{
			name:           "Other error from database",
			request:        createSubscriptionRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMocSubscriptionLoader("GetSubscriptionBySlug", TestSlug, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetSubscriptionBySlug)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestDeleteSubscriptionBySlug(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createSubscriptionDeleteRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockSubscriptionDeleter("DeleteSubscriptionBySlug", TestSlug, nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unknown Error",
			request:        createSubscriptionDeleteRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockSubscriptionDeleter("DeleteSubscriptionBySlug", TestSlug, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Subscription not found",
			request:        createSubscriptionDeleteRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockSubscriptionDeleter("DeleteSubscriptionBySlug", TestSlug, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restDeleteSubscriptionBySlug)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}
