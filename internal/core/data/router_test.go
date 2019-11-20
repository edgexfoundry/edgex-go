/*
 * ******************************************************************************
 *  Copyright 2019 Dell Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software distributed under the License
 *  is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 *  or implied. See the License for the specific language governing permissions and limitations under
 *  the License.
 *  ******************************************************************************
 */

package data

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"
)

// TestSuccessfulConfig configuration used to avoid MaxResultCount errors.
var TestSuccessfulConfig = config.ServiceInfo{MaxResultCount: 5}

// ErrorQueryParam query parameter value which will trigger the ' url.ParseQuery' function to throw an error due to the '%' not being followed by a valid hexadecimal number.
var ErrorQueryParam = "%zz"

var TestURI = "/valuedescriptor"

// TestError error used to simulate non EdgeX type errors.
var TestError = errors.New("test error")

// TestNonExistentValueDescriptorName name used when attempting to simulate no matching results.
var TestNonExistentValueDescriptorName = "Non-existent ValueDescriptor"

// TestErrorValueDescriptorName name used when attempting to simulate an error.
var TestErrorValueDescriptorName = "TestErrorValueDescriptor"

var TestValueDescriptor1 = contract.ValueDescriptor{
	Name:        "TestValueDescriptor1",
	Description: "Test Value descriptor associated with no readings",
}
var TestValueDescriptor2 = contract.ValueDescriptor{
	Name:        "TestValueDescriptor2",
	Description: "Test Value descriptor which associated with 1 reading",
}
var TestValueDescriptor3 = contract.ValueDescriptor{
	Name:        "TestValueDescriptor3",
	Description: "Test Value descriptor which associated with 2 reading",
}

func TestRestValueDescriptorsUsageHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		config         config.ServiceInfo
		expectedResult map[string]bool
		expectedStatus int
	}{
		{
			"OK response with no query parameter",
			createRequest(),
			createMockLoader(),
			TestSuccessfulConfig,
			map[string]bool{
				TestValueDescriptor1.Name: false,
				TestValueDescriptor2.Name: true,
				TestValueDescriptor3.Name: true,
			},
			http.StatusOK,
		},
		{
			"OK response one name with as name query parameter",
			createRequestWithQueryParameter(map[string][]string{NAMES: {TestValueDescriptor1.Name}}),
			createMockLoader(),
			TestSuccessfulConfig,
			map[string]bool{
				TestValueDescriptor1.Name: false,
			},
			http.StatusOK,
		},
		{
			"OK response with multiple names as query parameter",
			createRequestWithQueryParameter(map[string][]string{NAMES: {TestValueDescriptor1.Name, TestValueDescriptor2.Name, TestValueDescriptor3.Name}}),
			createMockLoader(),
			TestSuccessfulConfig,
			map[string]bool{
				TestValueDescriptor1.Name: false,
				TestValueDescriptor2.Name: true,
				TestValueDescriptor3.Name: true,
			},
			http.StatusOK,
		},
		{
			"OK - No matching value descriptors without query parameter",
			createRequest(),
			createMockLoaderNoValueDescriptors(),
			TestSuccessfulConfig,
			map[string]bool{},
			http.StatusOK,
		},
		{
			"OK - No matching value descriptors with one name as query parameter",
			createRequestWithQueryParameter(map[string][]string{NAMES: {TestNonExistentValueDescriptorName}}),
			createMockLoader(),
			TestSuccessfulConfig,
			map[string]bool{},
			http.StatusOK,
		},
		{
			"OK - No matching value descriptors with multiple names as query parameter",
			createRequestWithQueryParameter(map[string][]string{NAMES: {TestNonExistentValueDescriptorName, TestNonExistentValueDescriptorName}}),
			createMockLoader(),
			TestSuccessfulConfig,
			map[string]bool{},
			http.StatusOK,
		},
		{
			"Invalid name query parameter",
			createRequestWithInvalidParameter(map[string][]string{}),
			createMockLoader(),
			TestSuccessfulConfig,
			nil,
			http.StatusBadRequest,
		},
		{
			"Invalid extra query parameter",
			createRequestWithInvalidParameter(map[string][]string{NAMES: {TestNonExistentValueDescriptorName, TestNonExistentValueDescriptorName}}),
			createMockLoader(),
			TestSuccessfulConfig,
			nil,
			http.StatusBadRequest,
		}, {
			"GetValueDescriptorsExecutor Error",
			createRequest(),
			createMockLoaderValueDescriptorsExecutorError(),
			TestSuccessfulConfig,
			nil,
			http.StatusInternalServerError,
		},
		{
			"GetReadingsExecutor Error",
			createRequest(),
			createMockLoaderReadingsExecutorError(),
			TestSuccessfulConfig,
			nil,
			http.StatusInternalServerError,
		},
		{
			"Error Limit Exceeded",
			createRequestWithQueryParameter(map[string][]string{NAMES: {TestValueDescriptor1.Name, TestValueDescriptor2.Name, TestValueDescriptor3.Name}}),
			createMockLoader(),
			config.ServiceInfo{MaxResultCount: 1},
			nil,
			http.StatusRequestEntityTooLarge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 10}}
			defer func() { Configuration = &ConfigurationStruct{} }()

			httpErrorHandler = errorconcept.NewErrorHandler(logger.NewMockClient())
			Configuration.Service = tt.config
			rr := httptest.NewRecorder()
			restValueDescriptorsUsageHandler(rr, tt.request, logger.NewMockClient(), tt.dbMock)
			response := rr.Result()
			b, _ := ioutil.ReadAll(response.Body)
			var r []map[string]bool
			_ = json.Unmarshal(b, &r)
			observed := convertResponseToMap(r)
			if tt.expectedResult != nil && !reflect.DeepEqual(tt.expectedResult, observed) {
				t.Errorf("Observed result doesn't match expected.\nExpected: %v\nActual: %v\n", tt.expectedResult, observed)
			}

			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func createRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, TestURI, nil)
}

func createRequestWithQueryParameter(queryParam map[string][]string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, TestURI, nil)
	q := req.URL.Query()
	for k, v := range queryParam {
		q.Add(k, strings.Join(v, ","))
	}
	req.URL.RawQuery = q.Encode()
	return req
}

func createRequestWithInvalidParameter(queryParam map[string][]string) *http.Request {
	req := createRequestWithQueryParameter(queryParam)
	req.URL.RawQuery = req.URL.RawQuery + ErrorQueryParam
	return req
}

func createReadings(howMany int) []contract.Reading {
	var readings []contract.Reading
	for i := 0; i < howMany; i++ {
		readings = append(readings, contract.Reading{
			Id: "reading" + strconv.Itoa(i),
		})
	}
	return readings
}

func createMockLoader() interfaces.DBClient {
	dbMock := mocks.DBClient{}
	dbMock.On("ValueDescriptors").Return([]contract.ValueDescriptor{TestValueDescriptor1, TestValueDescriptor2, TestValueDescriptor3}, nil)
	dbMock.On("ValueDescriptorsByName", []string{TestValueDescriptor1.Name}).Return([]contract.ValueDescriptor{TestValueDescriptor1}, nil)
	dbMock.On("ValueDescriptorsByName", []string{TestValueDescriptor1.Name, TestValueDescriptor2.Name}).Return([]contract.ValueDescriptor{TestValueDescriptor1, TestValueDescriptor2}, nil)
	dbMock.On("ValueDescriptorsByName", []string{TestValueDescriptor1.Name, TestValueDescriptor2.Name, TestValueDescriptor3.Name}).Return([]contract.ValueDescriptor{TestValueDescriptor1, TestValueDescriptor2, TestValueDescriptor3}, nil)
	dbMock.On("ValueDescriptorsByName", []string{TestNonExistentValueDescriptorName}).Return(nil, nil)
	dbMock.On("ValueDescriptorsByName", []string{TestNonExistentValueDescriptorName, TestNonExistentValueDescriptorName}).Return(nil, nil)
	dbMock.On("ReadingsByValueDescriptor", TestValueDescriptor1.Name, ValueDescriptorUsageReadLimit).Return(createReadings(0), nil)
	dbMock.On("ReadingsByValueDescriptor", TestValueDescriptor2.Name, ValueDescriptorUsageReadLimit).Return(createReadings(1), nil)
	dbMock.On("ReadingsByValueDescriptor", TestValueDescriptor3.Name, ValueDescriptorUsageReadLimit).Return(createReadings(2), nil)

	return &dbMock
}

func createMockLoaderNoValueDescriptors() interfaces.DBClient {
	dbMock := mocks.DBClient{}
	dbMock.On("ValueDescriptors").Return([]contract.ValueDescriptor{}, nil)

	return &dbMock
}

func createMockLoaderValueDescriptorsExecutorError() interfaces.DBClient {
	dbMock := mocks.DBClient{}
	dbMock.On("ValueDescriptors").Return(nil, TestError)
	dbMock.On("ValueDescriptorsByName", []string{TestErrorValueDescriptorName}).Return(nil, TestError)
	dbMock.On("ValueDescriptorsByName", []string{TestErrorValueDescriptorName}).Return(nil, TestError)

	return &dbMock
}
func createMockLoaderReadingsExecutorError() interfaces.DBClient {
	dbMock := mocks.DBClient{}
	dbMock.On("ValueDescriptors").Return([]contract.ValueDescriptor{TestValueDescriptor1, TestValueDescriptor2, TestValueDescriptor3}, nil)
	dbMock.On("ValueDescriptorsByName", []string{TestValueDescriptor1.Name}).Return([]contract.ValueDescriptor{TestValueDescriptor1}, nil)
	dbMock.On("ValueDescriptorsByName", []string{TestValueDescriptor1.Name, TestValueDescriptor2.Name}).Return([]contract.ValueDescriptor{TestValueDescriptor1, TestValueDescriptor2}, nil)
	dbMock.On("ValueDescriptorsByName", []string{TestValueDescriptor1.Name, TestValueDescriptor2.Name, TestValueDescriptor3.Name}).Return([]contract.ValueDescriptor{TestValueDescriptor1, TestValueDescriptor2, TestValueDescriptor3}, nil)
	dbMock.On("ValueDescriptorsByName", []string{TestNonExistentValueDescriptorName}).Return(nil, nil)
	dbMock.On("ValueDescriptorsByName", []string{TestNonExistentValueDescriptorName, TestNonExistentValueDescriptorName}).Return(nil, nil)
	dbMock.On("ReadingsByValueDescriptor", TestValueDescriptor1.Name, ValueDescriptorUsageReadLimit).Return(nil, TestError)
	dbMock.On("ReadingsByValueDescriptor", TestValueDescriptor2.Name, ValueDescriptorUsageReadLimit).Return(nil, TestError)
	dbMock.On("ReadingsByValueDescriptor", TestValueDescriptor3.Name, ValueDescriptorUsageReadLimit).Return(nil, TestError)

	return &dbMock
}

// convertResponseToMap converts the response from the restValueDescriptorsUsageHandler to a map where the key is the
// ValueDescriptor name and the value is a bool specifying the usage state. This is made to make testing easier and
// more readable.
func convertResponseToMap(resp []map[string]bool) map[string]bool {
	convertedMap := map[string]bool{}
	for _, m := range resp {
		for k, v := range m {
			convertedMap[k] = v
		}
	}

	return convertedMap
}
