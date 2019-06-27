package metadata

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
)

// TestURI this is not really used since we are using the HTTP testing framework and not creating routes, but rather
// creating a specific handler which will accept all requests. Therefore, the URI is not important.
var TestURI = "/addressable"
var TestAddress = "TestAddress"
var TestPort = 8080
var TestPublisher = "TestPublisher"
var TestTopic = "TestTopic"

// ErrorPathParam path parameter value which will trigger the 'mux.Vars' function to throw an error due to the '%' not being followed by a valid hexadecimal number.
var ErrorPathParam = "%zz"

// ErrorPortPathParam path parameter used to trigger an error in the `restGetAddressableByPort` function where the port variable is expected to be a number.
var ErrorPortPathParam = "abc"

func TestGetAddressablesByAddress(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequest(ADDRESS, TestAddress),
			createMockAddressLoaderStringArg(1, "GetAddressablesByAddress", TestAddress),
			http.StatusOK,
		},
		{
			"OK(Multiple matches)",
			createRequest(ADDRESS, TestAddress),
			createMockAddressLoaderStringArg(3, "GetAddressablesByAddress", TestAddress),
			http.StatusOK,
		},
		{

			"OK(No matches)",
			createRequest(ADDRESS, TestAddress),
			createMockAddressLoaderStringArg(0, "GetAddressablesByAddress", TestAddress),
			http.StatusOK,
		},
		{
			"Invalid ADDRESS path parameter",
			createRequest(ADDRESS, ErrorPathParam),
			createMockAddressLoaderStringArg(1, "GetAddressablesByAddress", TestAddress),
			http.StatusBadRequest,
		},
		{
			"Internal Server Error",
			createRequest(ADDRESS, TestAddress),
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
			createRequest(PUBLISHER, TestPublisher),
			createMockAddressLoaderStringArg(1, "GetAddressablesByPublisher", TestPublisher),
			http.StatusOK,
		},
		{
			"OK(Multiple matches)",
			createRequest(PUBLISHER, TestPublisher), createMockAddressLoaderStringArg(3, "GetAddressablesByPublisher", TestPublisher),
			http.StatusOK,
		},
		{
			"OK(No matches)",
			createRequest(PUBLISHER, TestPublisher),
			createMockAddressLoaderStringArg(0, "GetAddressablesByPublisher", TestPublisher),
			http.StatusOK,
		},
		{
			"Invalid PUBLISHER path parameter",
			createRequest(PUBLISHER, ErrorPathParam),
			createMockAddressLoaderStringArg(1, "GetAddressablesByPublisher", TestPublisher),
			http.StatusBadRequest,
		},
		{
			"Internal Server Error",
			createRequest(PUBLISHER, TestPublisher),
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
			createRequest(PORT, strconv.Itoa(TestPort)),
			createMockAddressLoaderForPort(1, "GetAddressablesByPort"),
			http.StatusOK,
		},
		{
			"OK(Multiple matches)",
			createRequest(PORT, strconv.Itoa(TestPort)),
			createMockAddressLoaderForPort(3, "GetAddressablesByPort"),
			http.StatusOK,
		},
		{
			"OK(No matches)",
			createRequest(PORT, strconv.Itoa(TestPort)),
			createMockAddressLoaderForPort(0, "GetAddressablesByPort"),
			http.StatusOK,
		},
		{
			"Invalid PORT path parameter",
			createRequest(PORT, ErrorPathParam),
			createMockAddressLoaderForPort(1, "GetAddressablesByPort"),
			http.StatusBadRequest,
		},
		{
			"Non-integer PORT path parameter",
			createRequest(PORT, ErrorPortPathParam),
			createMockAddressLoaderForPort(1, "GetAddressablesByPort"),
			http.StatusBadRequest,
		},
		{
			"Internal Server Error",
			createRequest(PORT, strconv.Itoa(TestPort)),
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
			createRequest(TOPIC, TestTopic),
			createMockAddressLoaderStringArg(1, "GetAddressablesByTopic", TestTopic),
			http.StatusOK,
		},
		{
			"OK(Multiple matches)",
			createRequest(TOPIC, TestTopic),
			createMockAddressLoaderStringArg(3, "GetAddressablesByTopic", TestTopic),
			http.StatusOK,
		},
		{
			"OK(No matches)",
			createRequest(TOPIC, TestTopic),
			createMockAddressLoaderStringArg(0, "GetAddressablesByTopic", TestTopic),
			http.StatusOK,
		},
		{
			"Invalid TOPIC path parameter",
			createRequest(TOPIC, ErrorPathParam),
			createMockAddressLoaderStringArg(1, "GetAddressablesByTopic", TestTopic),
			http.StatusBadRequest,
		},
		{
			"Internal Server Error",
			createRequest(TOPIC, TestTopic),
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

func createRequest(pathParamName string, pathParamValue string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, TestURI, nil)
	return mux.SetURLVars(req, map[string]string{pathParamName: pathParamValue})
}

func createMockAddressLoaderStringArg(howMany int, methodName string, arg string) interfaces.DBClient {
	var addressables []contract.Addressable
	for i := 0; i < howMany; i++ {
		addressables = append(addressables, contract.Addressable{
			User:       "User" + strconv.Itoa(i),
			Protocol:   "http",
			Id:         "address" + strconv.Itoa(i),
			HTTPMethod: "POST",
		})
	}

	myMock := &mocks.DBClient{}
	myMock.On(methodName, arg).Return(addressables, nil)
	return myMock
}

func createMockAddressLoaderForPort(howMany int, methodName string) interfaces.DBClient {
	var addressables []contract.Addressable
	for i := 0; i < howMany; i++ {
		addressables = append(addressables, contract.Addressable{
			User:       "User" + strconv.Itoa(i),
			Protocol:   "http",
			Id:         "address" + strconv.Itoa(i),
			HTTPMethod: "POST",
		})
	}

	myMock := &mocks.DBClient{}
	myMock.On(methodName, TestPort).Return(addressables, nil)
	return myMock
}

func createErrorMockAddressLoaderStringArg(methodName string, arg string) interfaces.DBClient {
	myMock := &mocks.DBClient{}
	myMock.On(methodName, arg).Return(nil, errors.New("test error"))
	return myMock
}

func createErrorMockAddressLoaderPortExecutor(methodName string) interfaces.DBClient {
	myMock := &mocks.DBClient{}
	myMock.On(methodName, TestPort).Return(nil, errors.New("test error"))
	return myMock
}
