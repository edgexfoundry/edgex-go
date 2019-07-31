/*********************************************************************
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

package scheduler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

// TestURI this is not really used since we are using the HTTP testing framework and not creating routes, but rather
// creating a specific handler which will accept all requests. Therefore, the URI is not important.
var TestURI = "/interval"
var TestId = "123e4567-e89b-12d3-a456-426655440000"
var TestIncorrectId = "123e4567-e89b-12d3-a456-4266554400%0"

func TestIntervalById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequest(ID, TestId),
			createMockIntervalLoaderForId(nil),
			http.StatusOK,
		},
		{
			name:           "Interval not found",
			request:        createRequest(ID, TestId),
			dbMock:         createMockIntervalLoaderForId(db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Error QueryUnescape",
			request:        createRequest(ID, TestIncorrectId),
			dbMock:         createMockIntervalLoaderForId(nil),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Other error from database",
			request:        createRequest(ID, TestId),
			dbMock:         createMockIntervalLoaderForId(errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetIntervalByID)
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

func createMockIntervalLoaderForId(desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On("IntervalById", TestId).Return(contract.Interval{}, desiredError)
	} else {
		myMock.On("IntervalById", TestId).Return(createIntervals(1)[0], nil)
	}
	return &myMock
}

func createMockIntervalDeleterForId(desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On("DeleteIntervalById", TestId).Return(desiredError)
	} else {
		myMock.On("DeleteIntervalById", TestId).Return(nil)
	}
	return &myMock
}

func createMockSCDeleterForId(interval contract.Interval, desiredError error) interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}

	if desiredError != nil {
		myMock.On("QueryIntervalByID", TestId).Return(contract.Interval{}, desiredError)
		myMock.On("RemoveIntervalInQueue", TestId).Return(desiredError)
	} else {
		myMock.On("RemoveIntervalInQueue", TestId).Return(nil)
		myMock.On("QueryIntervalByID", TestId).Return(interval, nil)
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

func createIntervals(howMany int) []contract.Interval {
	var intervals []contract.Interval
	for i := 0; i < howMany; i++ {
		intervals = append(intervals, contract.Interval{
			ID:        TestId,
			Name:      "hourly",
			Start:     "20160101T000000",
			End:       "",
			Frequency: "PT1H!",
		})
	}
	return intervals
}

func TestDeleteIntervalById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		scMock         interfaces.SchedulerQueueClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createDeleteRequest(ID, TestId),
			dbMock:         createMockIntervalDeleterForId(nil),
			scMock:         createMockSCDeleterForId(createIntervals(1)[0], nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Interval not found",
			request:        createDeleteRequest(ID, TestId),
			dbMock:         createMockIntervalDeleterForId(db.ErrNotFound),
			scMock:         createMockSCDeleterForId(createIntervals(1)[0], nil),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Error QueryUnescape",
			request:        createDeleteRequest(ID, TestIncorrectId),
			dbMock:         createMockIntervalDeleterForId(nil),
			scMock:         createMockSCDeleterForId(createIntervals(1)[0], nil),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unknown Error",
			request:        createDeleteRequest(ID, TestId),
			dbMock:         createMockIntervalDeleterForId(errors.New("Test error")),
			scMock:         createMockSCDeleterForId(createIntervals(1)[0], nil),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			scClient = tt.scMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restDeleteIntervalByID)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}
