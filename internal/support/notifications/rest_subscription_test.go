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
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces/mocks"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

var TestSubscriptionURI = "/subscription"

var subscriptionForAdd = contract.Subscription{

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
}

var subscriptionForAddInvalid = contract.Subscription{

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
				"cloud\n",
				"jack\r",
			},
		},
	},
}

func TestSubscriptionsAll(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createSubscriptionRequest(map[string]string{}),
			dbMock:         createMockSubscriptionAllLoader("GetSubscriptions", nil),
			expectedStatus: http.StatusOK,
		},

		{
			name:           "Other error from database",
			request:        createSubscriptionRequest(map[string]string{}),
			dbMock:         createMockSubscriptionAllLoader("GetSubscriptions", errors.New("test error")),
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetSubscriptions(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetSubscriptionById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "Subscription not found",
			request:        createSubscriptionRequest(map[string]string{ID: TestId}),
			dbMock:         createMockSubscriptionLoader("GetSubscriptionById", TestId, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "OK",
			request:        createSubscriptionRequest(map[string]string{ID: TestId}),
			dbMock:         createMockSubscriptionLoader("GetSubscriptionById", TestId, nil),
			expectedStatus: http.StatusOK,
		},

		{
			name:           "Other error from database",
			request:        createSubscriptionRequest(map[string]string{ID: TestId}),
			dbMock:         createMockSubscriptionLoader("GetSubscriptionById", TestId, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetSubscriptionByID(rr, tt.request, logger.NewMockClient(), tt.dbMock)
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

func createMockSubscriptionAllLoader(methodName string, desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName).Return([]contract.Subscription{}, desiredError)
	} else {
		myMock.On(methodName).Return(createSubscriptions(1), nil)
	}
	return &myMock
}

func createMockSubscriptionLoader(methodName string, testID string, desiredError error) interfaces.DBClient {
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

func createMockSubscriptionLoaderCollection(methodName string, arg interface{}, desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, arg).Return([]contract.Subscription{}, desiredError)
	} else {
		myMock.On(methodName, arg).Return(createSubscriptions(1), nil)
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
			rr := httptest.NewRecorder()
			restDeleteSubscriptionByID(rr, tt.request, logger.NewMockClient(), tt.dbMock)
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
			name:           "Subscription not found",
			request:        createSubscriptionRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockSubscriptionLoader("GetSubscriptionBySlug", TestSlug, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "OK",
			request:        createSubscriptionRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockSubscriptionLoader("GetSubscriptionBySlug", TestSlug, nil),
			expectedStatus: http.StatusOK,
		},

		{
			name:           "Other error from database",
			request:        createSubscriptionRequest(map[string]string{SLUG: TestSlug}),
			dbMock:         createMockSubscriptionLoader("GetSubscriptionBySlug", TestSlug, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetSubscriptionBySlug(rr, tt.request, logger.NewMockClient(), tt.dbMock)
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
			rr := httptest.NewRecorder()
			restDeleteSubscriptionBySlug(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetSubscriptionsByCategories(t *testing.T) {

	categoriesURL := strings.Join(TestCategories, ",")

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "Subscription not found",
			request:        createSubscriptionRequest(map[string]string{CATEGORIES: categoriesURL}),
			dbMock:         createMockSubscriptionLoaderCollection("GetSubscriptionByCategories", TestCategories, db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "OK",
			request:        createSubscriptionRequest(map[string]string{CATEGORIES: categoriesURL}),
			dbMock:         createMockSubscriptionLoaderCollection("GetSubscriptionByCategories", TestCategories, nil),
			expectedStatus: http.StatusOK,
		},

		{
			name:           "Other error from database",
			request:        createSubscriptionRequest(map[string]string{CATEGORIES: categoriesURL}),
			dbMock:         createMockSubscriptionLoaderCollection("GetSubscriptionByCategories", TestCategories, errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetSubscriptionsByCategories(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestAddSubscription(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createRequestSubscriptionAdd(subscriptionForAdd),
			dbMock:         createMockSubscriptionLoaderAddSuccess(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid Email",
			request:        createRequestSubscriptionAdd(subscriptionForAddInvalid),
			dbMock:         createMockSubscriptionLoaderAddSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unexpected Error",
			request:        createRequestSubscriptionAdd(subscriptionForAdd),
			dbMock:         createMockSubscriptionLoaderAddErr(),
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restAddSubscription(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func createRequestSubscriptionAdd(subscription contract.Subscription) *http.Request {
	b, _ := json.Marshal(subscription)
	req := httptest.NewRequest(http.MethodPost, TestURI, bytes.NewBuffer(b))
	return mux.SetURLVars(req, map[string]string{})
}

func createMockSubscriptionLoaderAddSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddSubscription", subscriptionForAdd).Return(subscriptionForAdd.ID, nil)
	return &myMock
}

func createMockSubscriptionLoaderAddErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("AddSubscription", subscriptionForAdd).Return(subscriptionForAdd.ID, errors.New("test error"))
	return &myMock
}

func TestUpdateSubscription(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createRequestSubscriptionUpdate(subscriptionForAdd),
			dbMock:         createMockSubscriptionLoaderUpdateSuccess(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Email",
			request:        createRequestSubscriptionUpdate(subscriptionForAddInvalid),
			dbMock:         createMockSubscriptionLoaderUpdateSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GetSubscriptionBySlug Not Found Error",
			request:        createRequestSubscriptionUpdate(subscriptionForAdd),
			dbMock:         createMockSubscriptionLoaderUpdateGetNotFoundErr(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "GetSubscriptionBySlug Unexpected Error",
			request:        createRequestSubscriptionUpdate(subscriptionForAdd),
			dbMock:         createMockSubscriptionLoaderUpdateGetErr(),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Unexpected Error",
			request:        createRequestSubscriptionUpdate(subscriptionForAdd),
			dbMock:         createMockSubscriptionLoaderUpdateErr(),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restUpdateSubscription(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func createRequestSubscriptionUpdate(subscription contract.Subscription) *http.Request {
	b, _ := json.Marshal(subscription)
	req := httptest.NewRequest(http.MethodPut, TestURI, bytes.NewBuffer(b))
	return mux.SetURLVars(req, map[string]string{})
}

func createMockSubscriptionLoaderUpdateSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("GetSubscriptionBySlug", subscriptionForAdd.Slug).Return(subscriptionForAdd, nil)
	myMock.On("UpdateSubscription", subscriptionForAdd).Return(nil)
	return &myMock
}

func createMockSubscriptionLoaderUpdateErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("GetSubscriptionBySlug", subscriptionForAdd.Slug).Return(subscriptionForAdd, nil)
	myMock.On("UpdateSubscription", subscriptionForAdd).Return(errors.New("test error"))
	return &myMock
}

func createMockSubscriptionLoaderUpdateGetNotFoundErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("GetSubscriptionBySlug", subscriptionForAdd.Slug).Return(subscriptionForAdd, db.ErrNotFound)
	myMock.On("UpdateSubscription", subscriptionForAdd).Return(errors.New("test error"))
	return &myMock
}

func createMockSubscriptionLoaderUpdateGetErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("GetSubscriptionBySlug", subscriptionForAdd.Slug).Return(subscriptionForAdd, errors.New("test error"))
	myMock.On("UpdateSubscription", subscriptionForAdd).Return(errors.New("test error"))
	return &myMock
}
