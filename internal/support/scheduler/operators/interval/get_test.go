/*******************************************************************************
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

package interval

import (
	"errors"
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/operators/interval/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var Id = "83cb038b-5a94-4707-985d-13effec62de2"
var Name = "hourly"
var OtherName = "other"
var Error = errors.New("test error")
var ErrorNotFound = db.ErrNotFound
var TestLimit = 5
var SuccessfulDatabaseResult = []contract.Interval{
	{
		ID:        Id,
		Name:      "hourly",
		Start:     "20160101T000000",
		End:       "",
		Frequency: "PT1H",
	},
	{
		ID:        Id,
		Name:      "weekly",
		Start:     "20160101T000000",
		End:       "",
		Frequency: "PT1H",
	},
	{
		ID:        Id,
		Name:      "weekly2",
		Start:     "invalid",
		End:       "",
		Frequency: "PT1H",
	},
	{
		ID:        Id,
		Name:      "weekly3",
		Start:     "20160101T000000",
		End:       "invalid",
		Frequency: "PT1H",
	},
	{
		ID:        Id,
		Name:      "weekly4",
		Start:     "20160101T000000",
		End:       "",
		Frequency: "PT1H!",
	},
}

var SuccessfulIntervalActionResult = []contract.IntervalAction{
	{
		ID:         Id,
		Name:       "scrub pushed records",
		Interval:   "hourly",
		Parameters: "",
	},
}

func TestAllExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         IntervalLoader
		expectedResult []contract.Interval
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockIntervalsSuccess(),
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Unexpected error",
			mockDb:         createMockIntervalsFail(),
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewAllExecutor(test.mockDb, 0)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewAllExecutor(test.mockDb, TestLimit)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestIdExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         IntervalLoader
		expectedResult contract.Interval
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockIntervalLoaderStringArg("IntervalById", nil, SuccessfulDatabaseResult[0], Id),
			expectedResult: SuccessfulDatabaseResult[0],
			expectedError:  false,
		},
		{
			name:           "Unexpected error",
			mockDb:         createMockIntervalLoaderStringArg("IntervalById", Error, contract.Interval{}, Id),
			expectedResult: contract.Interval{},
			expectedError:  true,
		},
		{
			name:           "Interval not found error",
			mockDb:         createMockIntervalLoaderStringArg("IntervalById", ErrorNotFound, contract.Interval{}, Id),
			expectedResult: contract.Interval{},
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewIdExecutor(test.mockDb, Id)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func createMockIntervalsSuccess() IntervalLoader {
	dbMock := mocks.IntervalLoader{}
	dbMock.On("Intervals").Return(SuccessfulDatabaseResult, nil)
	dbMock.On("IntervalsWithLimit", TestLimit).Return(SuccessfulDatabaseResult, nil)
	return &dbMock
}

func createMockIntervalsFail() IntervalLoader {
	dbMock := mocks.IntervalLoader{}
	dbMock.On("Intervals").Return([]contract.Interval{}, Error)
	dbMock.On("IntervalsWithLimit", TestLimit).Return([]contract.Interval{}, Error)
	return &dbMock
}

func createMockIntervalLoaderStringArg(methodName string, err error, ret interface{}, arg string) IntervalLoader {
	dbMock := mocks.IntervalLoader{}
	dbMock.On(methodName, arg).Return(ret, err)
	return &dbMock
}

func TestNameExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         IntervalLoader
		expectedResult contract.Interval
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockIntervalLoaderStringArg("IntervalByName", nil, SuccessfulDatabaseResult[0], Name),
			expectedResult: SuccessfulDatabaseResult[0],
			expectedError:  false,
		},
		{
			name:           "Unexpected error",
			mockDb:         createMockIntervalLoaderStringArg("IntervalByName", Error, contract.Interval{}, Name),
			expectedResult: contract.Interval{},
			expectedError:  true,
		},
		{
			name:           "Interval not found error",
			mockDb:         createMockIntervalLoaderStringArg("IntervalByName", ErrorNotFound, contract.Interval{}, Name),
			expectedResult: contract.Interval{},
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewNameExecutor(test.mockDb, Name)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}
