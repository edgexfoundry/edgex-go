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

package intervalaction

import (
	"errors"
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/operators/intervalaction/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var Id = "83cb038b-5a94-4707-985d-13effec62de2"
var Name = "hourly"
var OtherName = "other"
var Error = errors.New("test error")
var ErrorNotFound = db.ErrNotFound
var TestLimit = 20
var TestServiceConfig = config.ServiceInfo{
	MaxResultCount: TestLimit,
}
var Intervals = []contract.Interval{
	{
		ID:        Id,
		Name:      "hourly",
		Start:     "20160101T000000",
		End:       "",
		Frequency: "PT1H",
	},
}

var IntervalActions = []contract.IntervalAction{
	{
		ID:         Id,
		Name:       "scrub pushed records",
		Interval:   "hourly",
		Parameters: "",
		Target:     "test target",
	},
	{
		ID:         Id,
		Name:       "scrub pushed records 2",
		Interval:   "hourly",
		Parameters: "",
		Target:     "test target",
	},
	{
		ID:         Id,
		Name:       "scrub pushed records 3",
		Interval:   "hourly",
		Parameters: "",
	},
	{
		ID:         Id,
		Name:       "scrub pushed records 4",
		Interval:   "",
		Target:     "test target",
		Parameters: "",
	},
	{
		ID:         Id,
		Interval:   "hourly",
		Parameters: "",
		Target:     "test target",
	},
	{
		ID:         Id,
		Name:       "scrub pushed records 6",
		Interval:   "hourly",
		Parameters: "",
	},
}

var ValidIntervalAction = IntervalActions[0]
var OtherValidIntervalAction = IntervalActions[1]
var InvalidIntervalAction = IntervalActions[2]
var IntervalActionNoInterval = IntervalActions[3]
var IntervalActionNoName = IntervalActions[4]
var IntervalActionNoTarget = IntervalActions[5]

func createMockIntervalActionsSuccess() IntervalActionLoader {
	dbMock := mocks.IntervalActionLoader{}
	dbMock.On("IntervalActions").Return(IntervalActions, nil)
	return &dbMock
}

func createMockIntervalActionsExceedErr() IntervalActionLoader {
	dbMock := mocks.IntervalActionLoader{}
	dbMock.On("IntervalActions").Return(createIntervalActions(21), nil)
	return &dbMock
}

func createMockIntervalActionsFail() IntervalActionLoader {
	dbMock := mocks.IntervalActionLoader{}
	dbMock.On("IntervalActions").Return([]contract.IntervalAction{}, Error)
	return &dbMock
}

func TestAllExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         IntervalActionLoader
		expectedResult []contract.IntervalAction
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockIntervalActionsSuccess(),
			expectedResult: IntervalActions,
			expectedError:  false,
		},
		{
			name:           "Exceed limit error",
			mockDb:         createMockIntervalActionsExceedErr(),
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:           "Unexpected error",
			mockDb:         createMockIntervalActionsFail(),
			expectedResult: []contract.IntervalAction{},
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewAllExecutor(test.mockDb, TestServiceConfig)
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

func createIntervalActions(howMany int) []contract.IntervalAction {
	var intervals []contract.IntervalAction
	for i := 0; i < howMany; i++ {
		intervals = append(intervals, contract.IntervalAction{
			ID:         Id,
			Name:       "scrub pushed records",
			Interval:   "hourly",
			Parameters: "",
		})
	}
	return intervals
}
