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
	customError "github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

// TestURI this is not really used since we are using the HTTP testing framework and not creating routes, but rather
// creating a specific handler which will accept all requests. Therefore, the URI is not important.
var TestURI = "/interval"
var TestId = "123e4567-e89b-12d3-a456-426655440000"
var TestName = "hourly"
var TestIncorrectId = "123e4567-e89b-12d3-a456-4266554400%0"
var TestIncorrectName = "hourly%b"

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
			createMockIntervalLoader("IntervalById", nil, TestId),
			http.StatusOK,
		},
		{
			name:           "Interval not found",
			request:        createRequest(ID, TestId),
			dbMock:         createMockIntervalLoader("IntervalById", db.ErrNotFound, TestId),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Error QueryUnescape",
			request:        createRequest(ID, TestIncorrectId),
			dbMock:         createMockIntervalLoader("IntervalById", nil, TestId),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Other error from database",
			request:        createRequest(ID, TestId),
			dbMock:         createMockIntervalLoader("IntervalById", errors.New("Test error"), TestId),
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

func createMockIntervalLoader(methodName string, desiredError error, arg string) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On(methodName, arg).Return(contract.Interval{}, desiredError)
	} else {
		myMock.On(methodName, arg).Return(createIntervals(1)[0], nil)
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

func createMockIntervalDeleterForName(desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On("DeleteIntervalById", TestId).Return(desiredError)
		myMock.On("IntervalByName", TestName).Return(contract.Interval{}, desiredError)
		myMock.On("IntervalActionsByIntervalName", TestName).Return([]contract.IntervalAction{}, desiredError)
	} else {
		myMock.On("DeleteIntervalById", TestId).Return(nil)
		myMock.On("IntervalByName", TestName).Return(createIntervals(1)[0], nil)
		myMock.On("IntervalActionsByIntervalName", TestName).Return([]contract.IntervalAction{}, nil)
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

func createMockSCDeleterForName(interval contract.Interval, desiredError error) interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}

	if desiredError != nil {
		myMock.On("QueryIntervalByName", TestName).Return(contract.Interval{}, desiredError)
		myMock.On("RemoveIntervalInQueue", TestId).Return(desiredError)
	} else {
		myMock.On("RemoveIntervalInQueue", TestId).Return(nil)
		myMock.On("QueryIntervalByName", TestName).Return(interval, nil)
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

func createIntervalActions(howMany int) []contract.IntervalAction {
	var intervals []contract.IntervalAction
	for i := 0; i < howMany; i++ {
		intervals = append(intervals, contract.IntervalAction{
			ID:         TestId,
			Name:       "scrub pushed records",
			Interval:   "hourly",
			Parameters: "",
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

func TestIntervalByName(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequest(NAME, TestName),
			createMockIntervalLoader("IntervalByName", nil, TestName),
			http.StatusOK,
		},
		{
			name:           "Interval not found",
			request:        createRequest(NAME, TestName),
			dbMock:         createMockIntervalLoader("IntervalByName", db.ErrNotFound, TestName),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Error QueryUnescape",
			request:        createRequest(NAME, TestIncorrectName),
			dbMock:         createMockIntervalLoader("IntervalByName", nil, TestName),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error ErrServiceClient",
			request:        createRequest(NAME, TestName),
			dbMock:         createMockIntervalLoader("IntervalByName", customError.NewErrServiceClient(500, []byte{}), TestName),
			expectedStatus: 500,
		},
		{
			name:           "Other error from database",
			request:        createRequest(NAME, TestName),
			dbMock:         createMockIntervalLoader("IntervalByName", errors.New("Test error"), TestName),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetIntervalByName)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func createMockNameDeleterSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("DeleteIntervalById", TestId).Return(nil)
	myMock.On("IntervalByName", TestName).Return(createIntervals(1)[0], nil)
	myMock.On("IntervalActionsByIntervalName", TestName).Return([]contract.IntervalAction{}, nil)
	return &myMock
}

func createMockNameDeleterNotFoundErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("DeleteIntervalById", TestId).Return(nil)
	myMock.On("IntervalByName", TestName).Return(createIntervals(1)[0], db.ErrNotFound)
	myMock.On("IntervalActionsByIntervalName", TestName).Return([]contract.IntervalAction{}, nil)
	return &myMock
}

func createMockNameDeleterErrServiceClient() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("DeleteIntervalById", TestId).Return(nil)
	myMock.On("IntervalByName", TestName).Return(createIntervals(1)[0], customError.ErrServiceClient{})
	myMock.On("IntervalActionsByIntervalName", TestName).Return([]contract.IntervalAction{}, nil)
	return &myMock
}

func createMockNameDeleterErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("DeleteIntervalById", TestId).Return(errors.New("Test error"))
	myMock.On("IntervalByName", TestName).Return(createIntervals(1)[0], customError.ErrServiceClient{StatusCode: 500})
	myMock.On("IntervalActionsByIntervalName", TestName).Return([]contract.IntervalAction{}, nil)
	return &myMock
}

func createMockNameSCDeleterSuccess() interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}
	myMock.On("QueryIntervalByName", TestName).Return(createIntervals(1)[0], nil)
	myMock.On("RemoveIntervalInQueue", TestId).Return(nil)
	return &myMock
}

func createMockNameSCDeleterErrIntervalStillUsed() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("DeleteIntervalById", TestId).Return(nil)
	myMock.On("IntervalByName", TestName).Return(createIntervals(1)[0], nil)
	myMock.On("IntervalActionsByIntervalName", TestName).Return(createIntervalActions(1), nil)
	return &myMock
}

func TestDeleteIntervalByName(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		scMock         interfaces.SchedulerQueueClient
		expectedStatus int
	}{

		{
			name:           "OK",
			request:        createDeleteRequest(NAME, TestName),
			dbMock:         createMockNameDeleterSuccess(),
			scMock:         createMockNameSCDeleterSuccess(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Interval not found",
			request:        createDeleteRequest(NAME, TestName),
			dbMock:         createMockNameDeleterNotFoundErr(),
			scMock:         createMockNameSCDeleterSuccess(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Error QueryUnescape",
			request:        createDeleteRequest(NAME, TestIncorrectName),
			dbMock:         createMockNameDeleterSuccess(),
			scMock:         createMockNameSCDeleterSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error ErrServiceClient",
			request:        createDeleteRequest(NAME, TestName),
			dbMock:         createMockNameDeleterErrServiceClient(),
			scMock:         createMockNameSCDeleterSuccess(),
			expectedStatus: 500,
		},
		{
			name:           "ErrIntervalStillUsedByIntervalActions Error",
			request:        createDeleteRequest(NAME, TestName),
			dbMock:         createMockNameSCDeleterErrIntervalStillUsed(),
			scMock:         createMockNameSCDeleterSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unknown Error",
			request:        createDeleteRequest(NAME, TestName),
			dbMock:         createMockNameDeleterErr(),
			scMock:         createMockNameSCDeleterSuccess(),
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
			handler := http.HandlerFunc(restDeleteIntervalByName)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}
