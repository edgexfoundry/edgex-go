/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package metadata

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
)

// AddressableTestURI this is not really used since we are using the HTTP testing framework and not creating routes, but rather
// creating a specific handler which will accept all requests. Therefore, the URI is not important.
var AddressableTestURI = "/addressable"
var TestAddress = "TestAddress"
var TestPort = 8080
var TestPublisher = "TestPublisher"
var TestTopic = "TestTopic"
var TestName = "AddressableName"
var TestId = "123e4567-e89b-12d3-a456-426655440000"

// ErrorPathParam path parameter value which will trigger the 'mux.Vars' function to throw an error due to the '%' not being followed by a valid hexadecimal number.
var ErrorPathParam = "%zz"

// ErrorPortPathParam path parameter used to trigger an error in the `restGetAddressableByPort` function where the port variable is expected to be a number.
var ErrorPortPathParam = "abc"

func TestMain(m *testing.M) {
	LoggingClient = logger.NewMockClient()
	httpErrorHandler = errorconcept.NewErrorHandler(LoggingClient)
	os.Exit(m.Run())
}

func TestGetAllAddressables(t *testing.T) {
	Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 10}}
	defer func() { Configuration = &ConfigurationStruct{} }()

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createAddressableRequest(http.MethodGet, "", ""),
			createMockAddressLoader(5, nil),
			http.StatusOK,
		},
		{
			"OK(No Addressables)",
			createAddressableRequest(http.MethodGet, "", ""),
			createMockAddressLoader(0, nil),
			http.StatusOK,
		},
		{
			"Error Limit Exceeded",
			createAddressableRequest(http.MethodGet, "", ""),
			createMockAddressLoader(11, nil),
			http.StatusRequestEntityTooLarge,
		},
		{
			"Error Unknown",
			createAddressableRequest(http.MethodGet, "", ""),
			createMockAddressLoader(0, errors.New("Some error")),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetAllAddressables)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestAddAddressable(t *testing.T) {
	noName := createAddressables(1)[0]
	noName.Name = ""

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:    "OK",
			request: createAddressableRequestWithBody(http.MethodPost, createAddressables(1)[0], ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"AddAddressable", mock.Anything, TestId, nil}}),
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Missing name field on addressable",
			request: createAddressableRequestWithBody(http.MethodPost, noName, ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"AddAddressable", mock.Anything, TestId, nil}}),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Duplicated addressable",
			request: createAddressableRequestWithBody(http.MethodPost, createAddressables(1)[0], ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"AddAddressable", mock.Anything, "", db.ErrNotUnique}}),
			expectedStatus: http.StatusConflict,
		},
		{
			name:    "Other error from database",
			request: createAddressableRequestWithBody(http.MethodPost, createAddressables(1)[0], ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"AddAddressable", mock.Anything, "", errors.New("some error")}}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Bad JSON parse",
			request:        createAddressableRequest(http.MethodPost, ID, TestId),
			dbMock:         nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restAddAddressable)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestUpdateAddressable(t *testing.T) {
	successNewName := createAddressables(1)[0]
	successNewName.Name = "something different"

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			name:    "OK",
			request: createAddressableRequestWithBody(http.MethodPut, createAddressables(1)[0], ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableById", createAddressables(1)[0].Id, createAddressables(1)[0], nil},
				{"UpdateAddressable", mock.Anything, nil, nil}}),
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Unsuccessful database call",
			request: createAddressableRequestWithBody(http.MethodPut, createAddressables(1)[0], ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableById", createAddressables(1)[0].Id, createAddressables(1)[0], nil},
				{"UpdateAddressable", mock.Anything, errors.New("some error"), nil}}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "Unsuccessful device service database call, updated name",
			request: createAddressableRequestWithBody(http.MethodPut, successNewName, ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableById", createAddressables(1)[0].Id, createAddressables(1)[0], nil},
				{"UpdateAddressable", mock.Anything, nil, nil},
				{"GetDeviceServicesByAddressableId", mock.Anything, nil, errors.New("some error")}}),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "Addressable in use",
			request: createAddressableRequestWithBody(http.MethodPut, successNewName, ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableById", createAddressables(1)[0].Id, createAddressables(1)[0], nil},
				{"UpdateAddressable", mock.Anything, nil, nil},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{{}}, nil}}),
			expectedStatus: http.StatusConflict,
		},
		{
			name:    "Addressable not found",
			request: createAddressableRequestWithBody(http.MethodPut, createAddressables(1)[0], ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableById", createAddressables(1)[0].Id, contract.Addressable{}, db.ErrNotFound}}),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Bad JSON parse",
			request:        createAddressableRequest(http.MethodPut, ID, TestId),
			dbMock:         nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restUpdateAddressable)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetAddressableByName(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createAddressableRequest(http.MethodGet, NAME, TestName),
			createMockAddressLoaderForName(nil),
			http.StatusOK,
		},
		{
			name:           "Bad escape character",
			request:        createAddressableRequest(http.MethodGet, NAME, TestName+"%zz"),
			dbMock:         createMockAddressLoaderForName(nil),
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "Addressable not found",
			request:        createAddressableRequest(http.MethodGet, NAME, TestName),
			dbMock:         createMockAddressLoaderForName(db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Other error from database",
			request:        createAddressableRequest(http.MethodGet, NAME, TestName),
			dbMock:         createMockAddressLoaderForName(errors.New("Test error")),
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetAddressableByName)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetAddressableById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createAddressableRequest(http.MethodGet, ID, TestId),
			createMockAddressLoaderForId(nil),
			http.StatusOK,
		},
		{
			name:           "Addressable not found",
			request:        createAddressableRequest(http.MethodGet, ID, TestId),
			dbMock:         createMockAddressLoaderForId(db.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Other error from database",
			request:        createAddressableRequest(http.MethodGet, ID, TestId),
			dbMock:         createMockAddressLoaderForId(errors.New("Test error")),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetAddressableById)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetAddressablesByAddress(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createAddressableRequest(http.MethodGet, ADDRESS, TestAddress),
			createMockAddressLoaderStringArg(1, "GetAddressablesByAddress", TestAddress),
			http.StatusOK,
		},
		{
			"OK(Multiple matches)",
			createAddressableRequest(http.MethodGet, ADDRESS, TestAddress),
			createMockAddressLoaderStringArg(3, "GetAddressablesByAddress", TestAddress),
			http.StatusOK,
		},
		{

			"OK(No matches)",
			createAddressableRequest(http.MethodGet, ADDRESS, TestAddress),
			createMockAddressLoaderStringArg(0, "GetAddressablesByAddress", TestAddress),
			http.StatusOK,
		},
		{
			"Invalid ADDRESS path parameter",
			createAddressableRequest(http.MethodGet, ADDRESS, ErrorPathParam),
			createMockAddressLoaderStringArg(1, "GetAddressablesByAddress", TestAddress),
			http.StatusBadRequest,
		},
		{
			"Internal Server Error",
			createAddressableRequest(http.MethodGet, ADDRESS, TestAddress),
			createErrorMockAddressLoaderStringArg("GetAddressablesByAddress", TestAddress),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetAddressableByAddress)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetAddressablesByPublisher(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{"OK",
			createAddressableRequest(http.MethodGet, PUBLISHER, TestPublisher),
			createMockAddressLoaderStringArg(1, "GetAddressablesByPublisher", TestPublisher),
			http.StatusOK,
		},
		{
			"OK(Multiple matches)",
			createAddressableRequest(http.MethodGet, PUBLISHER, TestPublisher), createMockAddressLoaderStringArg(3, "GetAddressablesByPublisher", TestPublisher),
			http.StatusOK,
		},
		{
			"OK(No matches)",
			createAddressableRequest(http.MethodGet, PUBLISHER, TestPublisher),
			createMockAddressLoaderStringArg(0, "GetAddressablesByPublisher", TestPublisher),
			http.StatusOK,
		},
		{
			"Invalid PUBLISHER path parameter",
			createAddressableRequest(http.MethodGet, PUBLISHER, ErrorPathParam),
			createMockAddressLoaderStringArg(1, "GetAddressablesByPublisher", TestPublisher),
			http.StatusBadRequest,
		},
		{
			"Internal Server Error",
			createAddressableRequest(http.MethodGet, PUBLISHER, TestPublisher),
			createErrorMockAddressLoaderStringArg("GetAddressablesByPublisher", TestPublisher),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetAddressableByPublisher)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetAddressablesByPort(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createAddressableRequest(http.MethodGet, PORT, strconv.Itoa(TestPort)),
			createMockAddressLoaderForPort(1, "GetAddressablesByPort"),
			http.StatusOK,
		},
		{
			"OK(Multiple matches)",
			createAddressableRequest(http.MethodGet, PORT, strconv.Itoa(TestPort)),
			createMockAddressLoaderForPort(3, "GetAddressablesByPort"),
			http.StatusOK,
		},
		{
			"OK(No matches)",
			createAddressableRequest(http.MethodGet, PORT, strconv.Itoa(TestPort)),
			createMockAddressLoaderForPort(0, "GetAddressablesByPort"),
			http.StatusOK,
		},
		{
			"Invalid PORT path parameter",
			createAddressableRequest(http.MethodGet, PORT, ErrorPathParam),
			createMockAddressLoaderForPort(1, "GetAddressablesByPort"),
			http.StatusBadRequest,
		},
		{
			"Non-integer PORT path parameter",
			createAddressableRequest(http.MethodGet, PORT, ErrorPortPathParam),
			createMockAddressLoaderForPort(1, "GetAddressablesByPort"),
			http.StatusBadRequest,
		},
		{
			"Internal Server Error",
			createAddressableRequest(http.MethodGet, PORT, strconv.Itoa(TestPort)),
			createErrorMockAddressLoaderPortExecutor("GetAddressablesByPort"),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetAddressableByPort)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetAddressablesByTopic(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createAddressableRequest(http.MethodGet, TOPIC, TestTopic),
			createMockAddressLoaderStringArg(1, "GetAddressablesByTopic", TestTopic),
			http.StatusOK,
		},
		{
			"OK(Multiple matches)",
			createAddressableRequest(http.MethodGet, TOPIC, TestTopic),
			createMockAddressLoaderStringArg(3, "GetAddressablesByTopic", TestTopic),
			http.StatusOK,
		},
		{
			"OK(No matches)",
			createAddressableRequest(http.MethodGet, TOPIC, TestTopic),
			createMockAddressLoaderStringArg(0, "GetAddressablesByTopic", TestTopic),
			http.StatusOK,
		},
		{
			"Invalid TOPIC path parameter",
			createAddressableRequest(http.MethodGet, TOPIC, ErrorPathParam),
			createMockAddressLoaderStringArg(1, "GetAddressablesByTopic", TestTopic),
			http.StatusBadRequest,
		},
		{
			"Internal Server Error",
			createAddressableRequest(http.MethodGet, TOPIC, TestTopic),
			createErrorMockAddressLoaderStringArg("GetAddressablesByTopic", TestTopic),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetAddressableByTopic)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestDeleteAddressableById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createAddressableRequest(http.MethodDelete, ID, TestId),
			createMockWithOutlines([]mockOutline{
				{"GetAddressableById", TestId, createAddressables(1)[0], nil},
				{"GetAddressableByName", mock.Anything, contract.Addressable{}, errors.New("some error")},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{}, nil},
				{"DeleteAddressableById", mock.Anything, nil, nil}}),
			http.StatusOK,
		},
		{
			name:    "Addressable not found",
			request: createAddressableRequest(http.MethodDelete, ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", TestId, contract.Addressable{}, db.ErrNotFound},
				{"GetAddressableById", TestId, contract.Addressable{}, db.ErrNotFound}}),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "Addressable in use",
			request: createAddressableRequest(http.MethodDelete, ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableById", TestId, createAddressables(1)[0], nil},
				{"GetAddressableByName", mock.Anything, contract.Addressable{}, errors.New("some error")},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{{}}, nil},
				{"DeleteAddressableById", mock.Anything, nil, nil}}),
			expectedStatus: http.StatusConflict,
		},
		{
			name:    "Other error from database",
			request: createAddressableRequest(http.MethodDelete, ID, TestId),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", TestId, contract.Addressable{}, errors.New("some error")},
				{"GetAddressableById", TestId, contract.Addressable{}, errors.New("some error")}}),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restDeleteAddressableById)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestDeleteAddressableByName(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createAddressableRequest(http.MethodDelete, NAME, TestName),
			createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", TestName, createAddressables(1)[0], nil},
				{"GetAddressableById", mock.Anything, contract.Addressable{}, errors.New("some error")},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{}, nil},
				{"DeleteAddressableById", mock.Anything, nil, nil}}),
			http.StatusOK,
		},
		{
			name:    "Addressable not found",
			request: createAddressableRequest(http.MethodDelete, NAME, TestName),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", TestName, contract.Addressable{}, db.ErrNotFound},
				{"GetAddressableById", TestName, contract.Addressable{}, db.ErrNotFound}}),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "Addressable in use",
			request: createAddressableRequest(http.MethodDelete, NAME, TestName),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", TestName, createAddressables(1)[0], nil},
				{"GetAddressableById", mock.Anything, contract.Addressable{}, errors.New("some error")},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{{}}, nil},
				{"DeleteAddressableById", mock.Anything, nil, nil}}),
			expectedStatus: http.StatusConflict,
		},
		{
			name:    "Other error from database",
			request: createAddressableRequest(http.MethodDelete, NAME, TestName),
			dbMock: createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", TestName, contract.Addressable{}, errors.New("some error")},
				{"GetAddressableById", TestName, contract.Addressable{}, errors.New("some error")}}),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restDeleteAddressableByName)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func createAddressableRequest(httpMethod string, pathParamName string, pathParamValue string) *http.Request {
	req := httptest.NewRequest(httpMethod, AddressableTestURI, nil)
	return mux.SetURLVars(req, map[string]string{pathParamName: pathParamValue})
}

func createAddressableRequestWithBody(httpMethod string, addressable contract.Addressable, pathParamName string, pathParamValue string) *http.Request {
	// if your JSON marshalling fails you've got bigger problems
	body, _ := json.Marshal(addressable)

	req := httptest.NewRequest(httpMethod, AddressableTestURI, bytes.NewReader(body))

	return mux.SetURLVars(req, map[string]string{pathParamName: pathParamValue})
}

func createAddressables(howMany int) []contract.Addressable {
	var addressables []contract.Addressable
	for i := 0; i < howMany; i++ {
		addressables = append(addressables, contract.Addressable{
			Name:       "Name" + strconv.Itoa(i),
			User:       "User" + strconv.Itoa(i),
			Protocol:   "http",
			Id:         "address" + strconv.Itoa(i),
			HTTPMethod: "POST",
		})
	}
	return addressables
}

func createMockAddressLoaderStringArg(howMany int, methodName string, arg string) interfaces.DBClient {
	addressables := createAddressables(howMany)

	myMock := mocks.DBClient{}
	myMock.On(methodName, arg).Return(addressables, nil)
	return &myMock
}

func createMockAddressLoaderForPort(howMany int, methodName string) interfaces.DBClient {
	addressables := createAddressables(howMany)

	myMock := mocks.DBClient{}
	myMock.On(methodName, TestPort).Return(addressables, nil)
	return &myMock
}

func createMockAddressLoaderForName(desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On("GetAddressableByName", TestName).Return(contract.Addressable{}, desiredError)
	} else {
		myMock.On("GetAddressableByName", TestName).Return(createAddressables(1)[0], nil)
	}
	return &myMock
}

func createMockAddressLoaderForId(desiredError error) interfaces.DBClient {
	myMock := mocks.DBClient{}

	if desiredError != nil {
		myMock.On("GetAddressableById", TestId).Return(contract.Addressable{}, desiredError)
	} else {
		myMock.On("GetAddressableById", TestId).Return(createAddressables(1)[0], nil)
	}
	return &myMock
}

func createErrorMockAddressLoaderStringArg(methodName string, arg string) interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On(methodName, arg).Return(nil, errors.New("test error"))
	return &myMock
}

func createErrorMockAddressLoaderPortExecutor(methodName string) interfaces.DBClient {
	myMock := mocks.DBClient{}
	myMock.On(methodName, TestPort).Return(nil, errors.New("test error"))
	return &myMock
}

func createMockAddressLoader(howMany int, err error) interfaces.DBClient {
	addressables := createAddressables(howMany)

	dbMock := mocks.DBClient{}
	dbMock.On("GetAddressables").Return(addressables, err)
	return &dbMock
}

type mockOutline struct {
	methodName string
	arg        interface{}
	ret        interface{}
	err        error
}

func createMockWithOutlines(outlines []mockOutline) interfaces.DBClient {
	dbMock := mocks.DBClient{}

	for _, o := range outlines {
		dbMock.On(o.methodName, o.arg).Return(o.ret, o.err)
	}

	return &dbMock
}
