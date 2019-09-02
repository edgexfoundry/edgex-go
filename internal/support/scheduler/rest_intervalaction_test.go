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
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

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

func createRequestIntervalAction() *http.Request {
	req := httptest.NewRequest(http.MethodGet, TestIntervalActionURI, nil)
	return mux.SetURLVars(req, map[string]string{})
}
