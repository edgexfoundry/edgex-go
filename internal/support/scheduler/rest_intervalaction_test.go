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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces/mocks"
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
			name:           "OK",
			request:        createRequestIntervalAction(),
			dbMock:         createMockIntervalActionLoaderAllSuccess(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Exceeded max limit",
			request:        createRequestIntervalAction(),
			dbMock:         createMockIntervalActionLoaderAllExceedErr(),
			expectedStatus: http.StatusRequestEntityTooLarge,
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
			rr := httptest.NewRecorder()
			restGetIntervalAction(rr, tt.request, logger.NewMockClient(), tt.dbMock)
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
			name:           "Error IntervalActionNameInUes",
			request:        createRequestIntervalActionAdd(intervalActionForAdd),
			dbMock:         createMockIntervalActionLoaderAddNameInUse(),
			scClient:       createMockIntervalActionLoaderSCAddSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
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
			scClient = tt.scClient
			rr := httptest.NewRecorder()
			restAddIntervalAction(rr, tt.request, logger.NewMockClient(), tt.dbMock)
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
	return &myMock
}

func createMockIntervalActionLoaderAllExceedErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("IntervalActions").Return(createIntervalActions(200), nil)
	return &myMock
}

func createMockIntervalActionLoaderAllErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("IntervalActions").Return([]contract.IntervalAction{}, errors.New("test error"))
	return &myMock
}

func createMockIntervalActionLoaderAddSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	intervalAction := createIntervalActions(1)[0]
	b, _ := json.Marshal(intervalAction)
	_ = intervalAction.UnmarshalJSON(b)
	validateIntervalAction(&intervalActionForAdd)

	myMock.On("IntervalActionByName", intervalActionForAdd.Name).Return(intervalAction, nil)
	myMock.On("IntervalByName", intervalActionForAdd.Interval).Return(createIntervals(1)[0], nil)
	myMock.On("AddIntervalAction", intervalActionForAdd).Return(intervalActionForAdd.ID, nil)
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
	_ = intervalAction.UnmarshalJSON(b)
	validateIntervalAction(&intervalActionForAdd)

	myMock.On("IntervalActionByName", intervalActionForAdd.Name).Return(intervalAction, nil)
	myMock.On("IntervalByName", intervalActionForAdd.Interval).Return(createIntervals(1)[0], nil)
	myMock.On("AddIntervalAction", intervalActionForAdd).Return(intervalActionForAdd.ID, errors.New("test error"))
	return &myMock
}

func validateIntervalAction(intervalAction *contract.IntervalAction) {
	b, _ := json.Marshal(intervalAction)
	_ = intervalAction.UnmarshalJSON(b)
}

func createMockIntervalActionLoaderSCAddSuccess() interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}
	validateIntervalAction(&otherIntervalActionForAdd)
	myMock.On("QueryIntervalActionByName", intervalActionForAdd.Name).Return(otherIntervalActionForAdd, nil)
	myMock.On("AddIntervalActionToQueue", intervalActionForAdd).Return(nil)

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
