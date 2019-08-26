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
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

var intervalActionForAdd = contract.IntervalAction{
	ID:         TestId,
	Name:       "scrub pushed records for add",
	Interval:   "hourly",
	Parameters: "",
	Target:     "test target",
}
var otherIntervalActionForAdd = contract.IntervalAction{
	ID:         TestId,
	Name:       "scrub pushed records for add 2",
	Interval:   "hourly",
	Parameters: "",
	Target:     "test target",
}
var InvalidIntervalActionForAdd = contract.IntervalAction{
	ID:         TestId,
	Name:       "scrub pushed records for add 2",
	Interval:   "hourly",
	Parameters: "",
}

var TestIntervalActionURI = "/" + INTERVALACTION

func TestGetIntervalAction(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestIntervalAction(),
			createMockIntervalActionLoaderAllSuccess(),
			http.StatusOK,
		},
		{
			name:           "Unexpected Error",
			request:        createRequestIntervalAction(),
			dbMock:         createMockIntervalActionLoaderAllErr(),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetIntervalAction)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestAddIntervalAction(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		scClient       interfaces.SchedulerQueueClient
		expectedStatus int
	}{

		{
			name:           "OK",
			request:        createRequestIntervalActionAdd(intervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderAddSuccess(),
			scClient:       createMockIntervalActionLoaderSCAddSuccess(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error Decoding",
			request:        createRequestIntervalActionAdd(InvalidIntervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderAddSuccess(),
			scClient:       createMockIntervalActionLoaderSCAddSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error IntervalActionNameInUes",
			request:        createRequestIntervalActionAdd(intervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderAddNameInUse(),
			scClient:       createMockIntervalActionLoaderSCAddSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error IntervalNotFound",
			request:        createRequestIntervalActionAdd(intervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderAddIntervalNotFoundErr(),
			scClient:       createMockIntervalActionLoaderSCAddSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error Unknown",
			request:        createRequestIntervalActionAdd(intervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderAddErr(),
			scClient:       createMockIntervalActionLoaderSCAddSuccess(),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			scClient = tt.scClient
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restAddIntervalAction)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestUpdateIntervalAction(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		scClient       interfaces.SchedulerQueueClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createRequestIntervalActionUpdate(intervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderUpdateSuccess(),
			scClient:       createMockIntervalActionLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error Decoding",
			request:        createRequestIntervalActionAdd(InvalidIntervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderUpdateSuccess(),
			scClient:       createMockIntervalActionLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error IntervalActionNameInUse",
			request:        createRequestIntervalActionAdd(intervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderUpdateNameUsed(),
			scClient:       createMockIntervalActionLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error IntervalActionNotFound",
			request:        createRequestIntervalActionAdd(intervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderUpdateNotFound(),
			scClient:       createMockIntervalActionLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Error Unknown",
			request:        createRequestIntervalActionAdd(intervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderUpdateErr(),
			scClient:       createMockIntervalActionLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			scClient = tt.scClient
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restUpdateIntervalAction)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func createMockIntervalActionLoaderAllSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("IntervalActions").Return(createIntervalActions(1), nil)
	myMock.On("IntervalActionsWithLimit", mock.Anything).Return(createIntervalActions(1), nil)
	return &myMock
}

func createMockIntervalActionLoaderAllErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("IntervalActions").Return([]contract.IntervalAction{}, errors.New("test error"))
	myMock.On("IntervalActionsWithLimit", mock.Anything).Return([]contract.IntervalAction{}, errors.New("test error"))
	return &myMock
}

func createMockIntervalActionLoaderAddSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	intervalAction := createIntervalActions(1)[0]
	b, _ := json.Marshal(intervalAction)
	intervalAction.UnmarshalJSON(b)
	validateIntervalAction(&intervalActionForAdd)

	myMock.On("IntervalActionByName", intervalActionForAdd.Name).Return(intervalAction, nil)
	myMock.On("IntervalByName", intervalActionForAdd.Interval).Return(createIntervals(1)[0], nil)
	myMock.On("AddIntervalAction", intervalActionForAdd).Return(intervalActionForAdd.ID, nil)
	return &myMock
}

func createMockIntervalActionLoaderUpdateSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	intervalAction := createIntervalActions(1)[0]

	validateIntervalAction(&intervalActionForAdd)
	validateIntervalAction(&intervalAction)

	myMock.On("IntervalActionById", intervalActionForAdd.ID).Return(intervalAction, nil)
	myMock.On("IntervalActionByName", intervalActionForAdd.Name).Return(intervalAction, nil)
	myMock.On("IntervalByName", intervalActionForAdd.Interval).Return(contract.Interval{}, nil)
	myMock.On("UpdateIntervalAction", intervalActionForAdd).Return(nil)
	return &myMock
}

func createMockIntervalActionLoaderUpdateNotFound() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("IntervalActionById", intervalActionForAdd.ID).Return(contract.IntervalAction{}, db.ErrNotFound)
	myMock.On("IntervalActionByName", intervalActionForAdd.Name).Return(contract.IntervalAction{}, db.ErrNotFound)
	return &myMock
}

func createMockIntervalActionLoaderUpdateNameUsed() interfaces.DBClient {
	myMock := mocks.DBClient{}
	intervalAction := createIntervalActions(1)[0]

	validateIntervalAction(&intervalActionForAdd)
	validateIntervalAction(&intervalAction)

	myMock.On("IntervalActionById", intervalActionForAdd.ID).Return(intervalAction, nil)
	myMock.On("IntervalActionByName", intervalActionForAdd.Name).Return(intervalActionForAdd, nil)
	myMock.On("IntervalByName", intervalActionForAdd.Interval).Return(createIntervals(1)[0], nil)
	return &myMock
}

func createMockIntervalActionLoaderUpdateNameStillUsed() interfaces.DBClient {
	myMock := mocks.DBClient{}
	interval := createIntervals(1)[0]

	myMock.On("IntervalById", intervalForAdd.ID).Return(interval, nil)
	myMock.On("IntervalByName", intervalForAdd.Name).Return(contract.Interval{}, db.ErrNotFound)
	myMock.On("IntervalActionsByIntervalName", interval.Name).Return(createIntervalActions(1), nil)
	myMock.On("UpdateInterval", intervalForAdd).Return(nil)
	return &myMock
}

func createMockIntervalActionLoaderAddNameInUse() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("IntervalActionByName", intervalActionForAdd.Name).Return(intervalActionForAdd, nil)
	return &myMock
}

func createMockIntervalActionLoaderAddIntervalNotFoundErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("IntervalActionByName", intervalActionForAdd.Name).Return(otherIntervalActionForAdd, nil)
	myMock.On("IntervalByName", intervalActionForAdd.Interval).Return(contract.Interval{}, db.ErrNotFound)
	return &myMock
}

func createMockIntervalActionLoaderAddErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	intervalAction := createIntervalActions(1)[0]
	b, _ := json.Marshal(intervalAction)
	intervalAction.UnmarshalJSON(b)
	validateIntervalAction(&intervalActionForAdd)

	myMock.On("IntervalActionByName", intervalActionForAdd.Name).Return(intervalAction, nil)
	myMock.On("IntervalByName", intervalActionForAdd.Interval).Return(createIntervals(1)[0], nil)
	myMock.On("AddIntervalAction", intervalActionForAdd).Return(intervalActionForAdd.ID, errors.New("test error"))
	return &myMock
}

func createMockIntervalActionLoaderUpdateErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	intervalAction := createIntervalActions(1)[0]

	validateIntervalAction(&intervalActionForAdd)
	validateIntervalAction(&intervalAction)

	myMock.On("IntervalActionById", intervalActionForAdd.ID).Return(intervalAction, nil)
	myMock.On("IntervalActionByName", intervalActionForAdd.Name).Return(intervalAction, nil)
	myMock.On("IntervalByName", intervalActionForAdd.Interval).Return(contract.Interval{}, nil)
	myMock.On("UpdateIntervalAction", intervalActionForAdd).Return(errors.New("test error"))
	return &myMock
}

func validateIntervalAction(intervalAction *contract.IntervalAction) {
	b, _ := json.Marshal(intervalAction)
	intervalAction.UnmarshalJSON(b)
}

func createMockIntervalActionLoaderSCAddSuccess() interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}
	validateIntervalAction(&otherIntervalActionForAdd)
	myMock.On("QueryIntervalActionByName", intervalActionForAdd.Name).Return(otherIntervalActionForAdd, nil)
	myMock.On("AddIntervalActionToQueue", intervalActionForAdd).Return(nil)

	return &myMock
}

func createMockIntervalActionLoaderSCUpdateSuccess() interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}
	myMock.On("UpdateIntervalActionQueue", intervalActionForAdd).Return(nil)
	return &myMock
}

func createMockIntervalActionLoadeSCAddErr() interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}
	myMock.On("AddIntervalToQueue", intervalForAdd).Return(nil)
	return &myMock
}

func createMockIntervalActionLoadeSCUpdateErr() interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}
	myMock.On("UpdateIntervalActionQueue", intervalForAdd).Return(nil)
	return &myMock
}

func createRequestIntervalAction() *http.Request {
	req := httptest.NewRequest(http.MethodGet, TestIntervalActionURI, nil)
	return mux.SetURLVars(req, map[string]string{})
}

func createRequestIntervalActionAdd(intervalAction contract.IntervalAction) *http.Request {
	b, _ := json.Marshal(intervalAction)
	req := httptest.NewRequest(http.MethodPost, TestIntervalActionURI, bytes.NewBuffer(b))
	return mux.SetURLVars(req, map[string]string{})
}

func createRequestIntervalActionUpdate(intervalAction contract.IntervalAction) *http.Request {
	b, _ := json.Marshal(intervalAction)
	req := httptest.NewRequest(http.MethodPut, TestIntervalActionURI, bytes.NewBuffer(b))
	return mux.SetURLVars(req, map[string]string{})
}
