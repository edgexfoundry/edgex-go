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

package reading

import (
	goErrors "errors"
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/data/operators/reading/mocks"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
)

// TestErrorReadingName is the name used for testing database errors.
var TestErrorReadingName = "TestReadingError"

// TestLoaderLimit is the limit to use when invoking the mock Loader.
var TestLoaderLimit = 5

// TestError error used to simulate non EdgeX type errors.
var TestError = goErrors.New("testing error")

var TestReading = contract.Reading{
	Name:        "TestReadingName",
	Id:          "TestReadingID",
	Value:       "TestReadingValue",
	Device:      "TestReadingDevice",
	Modified:    111,
	Created:     222,
	Origin:      333,
	Pushed:      444,
	BinaryValue: []byte{1, 2, 3, 4},
}

// Name used for readings which are associated with the same ValueDescriptor
var TestReadingNameForSameValueDescriptor = "TestReadingName2"
var TestReading3 = contract.Reading{
	Name:        TestReadingNameForSameValueDescriptor,
	Id:          "TestReadingID2Dup",
	Value:       "TestReadingValue2Dup",
	Device:      "TestReadingDevice2Dup",
	Modified:    111,
	Created:     222,
	Origin:      333,
	Pushed:      444,
	BinaryValue: []byte{1, 2, 3, 4},
}

var TestReading2 = contract.Reading{
	Name:        TestReadingNameForSameValueDescriptor,
	Id:          "TestReadingID2",
	Value:       "TestReadingValue2",
	Device:      "TestReadingDevice2",
	Modified:    111,
	Created:     222,
	Origin:      333,
	Pushed:      444,
	BinaryValue: []byte{1, 2, 3, 4},
}

func TestGetReadingsExecutor(t *testing.T) {
	tests := []struct {
		name              string
		readingName       string
		loader            Loader
		config            bootstrapConfig.ServiceInfo
		expectedResult    []contract.Reading
		expectError       bool
		expectedErrorType error
	}{
		{
			"Get one Reading",
			TestReading.Name,
			createMockLoader(),
			bootstrapConfig.ServiceInfo{MaxResultCount: 5},
			[]contract.Reading{TestReading},
			false,
			nil,
		},

		{
			"Get multiple Readings",
			TestReading2.Name,
			createMockLoader(),
			bootstrapConfig.ServiceInfo{MaxResultCount: 5},
			[]contract.Reading{TestReading2, TestReading3},
			false,
			nil,
		},
		{
			"Database error",
			TestErrorReadingName,
			createMockLoader(),
			bootstrapConfig.ServiceInfo{MaxResultCount: 5},
			nil,
			true,
			TestError,
		}, {
			"MaxResultCount exceeded error",
			TestReading2.Name,
			createMockLoader(),
			bootstrapConfig.ServiceInfo{MaxResultCount: 1},
			nil,
			true,
			errors.ErrLimitExceeded{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewGetReadingsNameExecutor(test.readingName, TestLoaderLimit, test.loader, logger.MockLogger{}, test.config)
			observed, err := op.Execute()
			if test.expectError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if test.expectError && test.expectedErrorType != nil {
				eet := reflect.TypeOf(test.expectedErrorType)
				aet := reflect.TypeOf(err)
				if !aet.AssignableTo(eet) {
					t.Errorf("Expected error of type %v, but got an error of type %v", eet, aet)
				}
			}

			if !reflect.DeepEqual(test.expectedResult, observed) {
				t.Errorf("Observed result doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedResult, observed)
			}

		})
	}
}

func createMockLoader() Loader {
	mock := &mocks.Loader{}
	mock.On("ReadingsByValueDescriptor", TestReading.Name, TestLoaderLimit).Return([]contract.Reading{TestReading}, nil)
	mock.On("ReadingsByValueDescriptor", TestReading2.Name, TestLoaderLimit).Return([]contract.Reading{TestReading2, TestReading3}, nil)
	mock.On("ReadingsByValueDescriptor", TestErrorReadingName, TestLoaderLimit).Return(nil, TestError)

	return mock
}
