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
	goErrors "errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	errors "github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces/mocks"
)

// TestURI this is not really used since we are using the HTTP testing framework and not creating routes, but rather
// creating a specific handler which will accept all requests. Therefore, the URI is not important.
var TestURI = "/interval"
var TestId = "123e4567-e89b-12d3-a456-426655440000"
var TestName = "hourly"
var TestOtherName = "weekly"
var TestIncorrectId = "123e4567-e89b-12d3-a456-4266554400%0"
var TestIncorrectName = "hourly%b"
var TestLimit = 5

var intervalForAdd = contract.Interval{
	ID:        TestId,
	Name:      TestOtherName,
	Start:     "20160101T000000",
	End:       "",
	Frequency: "PT1H",
}

var intervalForAddInvalidTime = contract.Interval{
	ID:        TestId,
	Name:      TestOtherName,
	Start:     "invalid",
	End:       "invalid",
	Frequency: "PT1H",
}

var intervalForAddInvalidFreq = contract.Interval{
	ID:        TestId,
	Name:      TestOtherName,
	Start:     "20160101T000000",
	End:       "",
	Frequency: "PT1HS",
}

var intervalForUpdateInvalidCron = contract.Interval{
	ID:        TestId,
	Name:      TestOtherName,
	Start:     "20160101T000000",
	End:       "",
	Frequency: "PT1H",
	Cron:      "invalid23",
}

func TestGetIntervals(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestIntervalAll(),
			createMockIntervalLoaderAllSuccess(),
			http.StatusOK,
		},
		{
			name:           "Unexpected Error",
			request:        createRequestIntervalAll(),
			dbMock:         createMockIntervalLoaderAllErr(),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetIntervals(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestAddInterval(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		scClient       interfaces.SchedulerQueueClient
		expectedStatus int
	}{
		{
			name:           "ErrInvalidTimeFormat",
			request:        createRequestIntervalAdd(intervalForAddInvalidTime),
			dbMock:         createMockIntervalLoaderAddSuccess(),
			scClient:       createMockIntervalLoaderSCAddSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "OK",
			request:        createRequestIntervalAdd(intervalForAdd),
			dbMock:         createMockIntervalLoaderAddSuccess(),
			scClient:       createMockIntervalLoaderSCAddSuccess(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ErrIntervalNameInUse",
			request:        createRequestIntervalAdd(intervalForAdd),
			dbMock:         createMockIntervalLoaderAddNameInUse(),
			scClient:       createMockIntervalLoaderSCAddSuccess(),
			expectedStatus: http.StatusBadRequest,
		},

		{
			name:           "ErrInvalidFrequencyFormat",
			request:        createRequestIntervalAdd(intervalForAddInvalidFreq),
			dbMock:         createMockIntervalLoaderAddSuccess(),
			scClient:       createMockIntervalLoaderSCAddSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unexpected Error",
			request:        createRequestIntervalAdd(intervalForAdd),
			dbMock:         createMockIntervalLoaderAddErr(),
			scClient:       createMockIntervalLoadeSCAddErr(),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restAddInterval(rr, tt.request, logger.NewMockClient(), tt.dbMock, tt.scClient)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestUpdateInterval(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		scClient       interfaces.SchedulerQueueClient
		expectedStatus int
	}{
		{
			name:           "OK",
			request:        createRequestIntervalUpdate(intervalForAdd),
			dbMock:         createMockIntervalLoaderUpdateSuccess(),
			scClient:       createMockIntervalLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ErrInvalidTimeFormat",
			request:        createRequestIntervalUpdate(intervalForAddInvalidTime),
			dbMock:         createMockIntervalLoaderUpdateSuccess(),
			scClient:       createMockIntervalLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ErrIntervalNotFound",
			request:        createRequestIntervalUpdate(intervalForAdd),
			dbMock:         createMockIntervalLoaderUpdateNotFound(),
			scClient:       createMockIntervalLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ErrInvalidCronFormat",
			request:        createRequestIntervalUpdate(intervalForUpdateInvalidCron),
			dbMock:         createMockIntervalLoaderUpdateInvalidCron(),
			scClient:       createMockIntervalLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ErrInvalidFrequencyFormat",
			request:        createRequestIntervalUpdate(intervalForAddInvalidFreq),
			dbMock:         createMockIntervalLoaderUpdateSuccess(),
			scClient:       createMockIntervalLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ErrIntervalNameInUse",
			request:        createRequestIntervalUpdate(intervalForAdd),
			dbMock:         createMockIntervalLoaderUpdateNameUsed(),
			scClient:       createMockIntervalLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ErrIntervalStillUsedByIntervalActions",
			request:        createRequestIntervalUpdate(intervalForAdd),
			dbMock:         createMockIntervalLoaderUpdateNameStillUsed(),
			scClient:       createMockIntervalLoaderSCUpdateSuccess(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unexpected Error",
			request:        createRequestIntervalUpdate(intervalForAdd),
			dbMock:         createMockIntervalLoaderUpdateErr(),
			scClient:       createMockIntervalLoadeSCUpdateErr(),
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restUpdateInterval(rr, tt.request, logger.NewMockClient(), tt.dbMock, tt.scClient)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

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
			dbMock:         createMockIntervalLoader("IntervalById", goErrors.New("test error"), TestId),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetIntervalByID(rr, tt.request, logger.NewMockClient(), tt.dbMock)
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
	os.Exit(m.Run())
}

func createMockIntervalLoaderAllSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("Intervals").Return(createIntervals(1), nil)
	myMock.On("IntervalsWithLimit", TestLimit).Return(createIntervals(1), nil)

	return &myMock
}

func createMockIntervalLoaderAllErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("Intervals").Return([]contract.Interval{}, goErrors.New("test error"))
	myMock.On("IntervalsWithLimit", TestLimit).Return([]contract.Interval{}, goErrors.New("test error"))

	return &myMock
}

func createMockIntervalLoaderAddSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	interval := createIntervals(1)[0]

	validateInterval(&intervalForAdd)

	myMock.On("IntervalByName", intervalForAdd.Name).Return(interval, nil)
	myMock.On("AddInterval", intervalForAdd).Return(intervalForAdd.ID, nil)
	return &myMock
}

func createMockIntervalLoaderUpdateSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	interval := createIntervals(1)[0]

	validateInterval(&intervalForAdd)
	validateInterval(&interval)

	myMock.On("IntervalById", intervalForAdd.ID).Return(interval, nil)
	myMock.On("IntervalByName", intervalForAdd.Name).Return(contract.Interval{}, db.ErrNotFound)
	myMock.On("IntervalActionsByIntervalName", interval.Name).Return([]contract.IntervalAction{}, nil)
	myMock.On("UpdateInterval", intervalForAdd).Return(nil)
	return &myMock
}

func createMockIntervalLoaderUpdateNotFound() interfaces.DBClient {
	myMock := mocks.DBClient{}
	interval := createIntervals(1)[0]

	myMock.On("IntervalById", intervalForAdd.ID).Return(interval, goErrors.New("test error"))
	myMock.On("IntervalByName", intervalForAdd.Name).Return(contract.Interval{}, db.ErrNotFound)
	myMock.On("IntervalActionsByIntervalName", interval.Name).Return([]contract.IntervalAction{}, nil)
	myMock.On("UpdateInterval", intervalForAdd).Return(nil)
	return &myMock
}

func createMockIntervalLoaderUpdateInvalidCron() interfaces.DBClient {
	myMock := mocks.DBClient{}
	interval := createIntervals(1)[0]

	myMock.On("IntervalById", intervalForUpdateInvalidCron.ID).Return(interval, nil)
	myMock.On("IntervalByName", intervalForUpdateInvalidCron.Name).Return(contract.Interval{}, db.ErrNotFound)
	myMock.On("IntervalActionsByIntervalName", interval.Name).Return([]contract.IntervalAction{}, nil)
	myMock.On("UpdateInterval", intervalForAdd).Return(nil)
	return &myMock
}

func createMockIntervalLoaderUpdateNameUsed() interfaces.DBClient {
	myMock := mocks.DBClient{}
	interval := createIntervals(1)[0]

	myMock.On("IntervalById", intervalForAdd.ID).Return(interval, nil)
	myMock.On("IntervalByName", intervalForAdd.Name).Return(interval, nil)
	myMock.On("IntervalActionsByIntervalName", interval.Name).Return([]contract.IntervalAction{}, nil)
	myMock.On("UpdateInterval", intervalForAdd).Return(nil)
	return &myMock
}

func createMockIntervalLoaderUpdateNameStillUsed() interfaces.DBClient {
	myMock := mocks.DBClient{}
	interval := createIntervals(1)[0]

	myMock.On("IntervalById", intervalForAdd.ID).Return(interval, nil)
	myMock.On("IntervalByName", intervalForAdd.Name).Return(contract.Interval{}, db.ErrNotFound)
	myMock.On("IntervalActionsByIntervalName", interval.Name).Return(createIntervalActions(1), nil)
	myMock.On("UpdateInterval", intervalForAdd).Return(nil)
	return &myMock
}

func createMockIntervalLoaderAddNameInUse() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("IntervalByName", intervalForAdd.Name).Return(intervalForAdd, nil)
	myMock.On("AddInterval", intervalForAdd).Return(intervalForAdd.ID, nil)
	return &myMock
}

func createMockIntervalLoaderAddErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("IntervalByName", intervalForAdd.Name).Return(contract.Interval{}, goErrors.New("test error"))
	myMock.On("AddInterval", intervalForAdd).Return(intervalForAdd.ID, goErrors.New("test error"))
	return &myMock
}

func createMockIntervalLoaderUpdateErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	interval := createIntervals(1)[0]
	myMock.On("IntervalById", intervalForAdd.ID).Return(interval, nil)
	myMock.On("IntervalByName", intervalForAdd.Name).Return(contract.Interval{}, goErrors.New("test error"))
	myMock.On("IntervalActionsByIntervalName", intervalForAdd.Name).Return([]contract.IntervalAction{}, nil)
	myMock.On("UpdateInterval", intervalForAdd).Return(nil)
	return &myMock
}

// this function serves to update the unexported isValidated field,
// which can only be done by marshalling and unmarshalling to JSON.
func validateInterval(interval *contract.Interval) {
	b, _ := json.Marshal(interval)
	_ = interval.UnmarshalJSON(b)
}

func createMockIntervalLoaderSCAddSuccess() interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}
	myMock.On("AddIntervalToQueue", intervalForAdd).Return(nil)
	return &myMock
}

func createMockIntervalLoaderSCUpdateSuccess() interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}
	myMock.On("UpdateIntervalInQueue", intervalForAdd).Return(nil)
	return &myMock
}

func createMockIntervalLoadeSCAddErr() interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}
	myMock.On("AddIntervalToQueue", intervalForAdd).Return(nil)
	return &myMock
}

func createMockIntervalLoadeSCUpdateErr() interfaces.SchedulerQueueClient {
	myMock := mocks.SchedulerQueueClient{}
	myMock.On("UpdateIntervalInQueue", intervalForAdd).Return(nil)
	return &myMock
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

func createRequestIntervalAll() *http.Request {
	req := httptest.NewRequest(http.MethodGet, TestURI, nil)
	return mux.SetURLVars(req, map[string]string{})
}

func createRequestIntervalAdd(interval contract.Interval) *http.Request {
	b, _ := json.Marshal(interval)
	req := httptest.NewRequest(http.MethodPost, TestURI, bytes.NewBuffer(b))
	return mux.SetURLVars(req, map[string]string{})
}

func createRequestIntervalUpdate(interval contract.Interval) *http.Request {
	b, _ := json.Marshal(interval)
	req := httptest.NewRequest(http.MethodPut, TestURI, bytes.NewBuffer(b))
	return mux.SetURLVars(req, map[string]string{})
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
			Frequency: "PT1H",
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
			dbMock:         createMockIntervalDeleterForId(goErrors.New("test error")),
			scMock:         createMockSCDeleterForId(createIntervals(1)[0], nil),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			rr := httptest.NewRecorder()
			restDeleteIntervalByID(rr, tt.request, logger.NewMockClient(), tt.dbMock, tt.scMock)
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
			name:    "Error ErrServiceClient",
			request: createRequest(NAME, TestName),
			dbMock: createMockIntervalLoader("IntervalByName", errors.NewErrServiceClient(500, []byte{}),
				TestName),
			expectedStatus: 500,
		},
		{
			name:           "Other error from database",
			request:        createRequest(NAME, TestName),
			dbMock:         createMockIntervalLoader("IntervalByName", goErrors.New("test error"), TestName),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetIntervalByName(rr, tt.request, logger.NewMockClient(), tt.dbMock)
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
	myMock.On("IntervalByName", TestName).Return(createIntervals(1)[0], errors.ErrServiceClient{})
	myMock.On("IntervalActionsByIntervalName", TestName).Return([]contract.IntervalAction{}, nil)
	return &myMock
}

func createMockNameDeleterErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("DeleteIntervalById", TestId).Return(goErrors.New("test error"))
	myMock.On("IntervalByName", TestName).Return(createIntervals(1)[0], errors.ErrServiceClient{StatusCode: 500})
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
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			rr := httptest.NewRecorder()
			restDeleteIntervalByName(rr, tt.request, logger.NewMockClient(), tt.dbMock, tt.scMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func createScrubDeleteRequest() *http.Request {
	req := httptest.NewRequest(http.MethodDelete, TestURI+"/scrub/", nil)
	return mux.SetURLVars(req, map[string]string{})
}

func createMockScrubDeleterSuccess() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("ScrubAllIntervals").Return(1, nil)
	return &myMock
}

func createMockScrubDeleterErr() interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On("ScrubAllIntervals").Return(0, goErrors.New("test error"))
	return &myMock
}

func TestScrubIntervals(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{

		{
			name:           "OK",
			request:        createScrubDeleteRequest(),
			dbMock:         createMockScrubDeleterSuccess(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unknown Error",
			request:        createScrubDeleteRequest(),
			dbMock:         createMockScrubDeleterErr(),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			rr := httptest.NewRecorder()
			restScrubAllIntervals(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}
