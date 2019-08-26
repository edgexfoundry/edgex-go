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
	"reflect"
	"testing"

	intervalErrors "github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/operators/intervalaction/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

//var InvalidFreqInterval = SuccessfulIntervalActionResult[4]

func TestAddExecutor(t *testing.T) {

	tests := []struct {
		name             string
		mockDb           IntervalActionWriter
		scClient         SchedulerQueueWriter
		intervalAction   contract.IntervalAction
		expectedResult   string
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name:             "Successful database call",
			mockDb:           createAddMockIntervalActionSuccess(),
			scClient:         createAddMockIntervalSCSuccess(),
			intervalAction:   ValidIntervalAction,
			expectedResult:   ValidIntervalAction.ID,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name:             "Error IntervalActionByName",
			mockDb:           createAddMockIntervalActionByNameSameErr(),
			scClient:         createAddMockIntervalSCSuccess(),
			intervalAction:   ValidIntervalAction,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalActionNameInUse(ValidIntervalAction.Name),
		},
		{
			name:             "Error IntervalByName",
			mockDb:           createAddMockIntervalActionByNameErr(),
			scClient:         createAddMockIntervalSCSuccess(),
			intervalAction:   ValidIntervalAction,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalNotFound(Intervals[0].Name),
		},
		{
			name:             "Error No Target",
			mockDb:           createAddMockIntervalActionTargetErr(),
			scClient:         createAddMockIntervalSCSuccess(),
			intervalAction:   InvalidIntervalAction,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalActionTargetNameRequired(InvalidIntervalAction.ID),
		},
		{
			name:             "Error No Interval",
			mockDb:           createAddMockIntervalActionNoIntervalErr(),
			scClient:         createAddMockIntervalSCSuccess(),
			intervalAction:   IntervalActionNoInterval,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalNotFound(IntervalActionNoInterval.ID),
		},
		{
			name:             "Error QueryIntervalActionByName",
			mockDb:           createAddMockIntervalActionSuccess(),
			scClient:         createAddMockIntervalSCQueryError(),
			intervalAction:   ValidIntervalAction,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalActionNameInUse(ValidIntervalAction.Name),
		},
		{
			name:             "Error AddIntervalActionToQueue",
			mockDb:           createAddMockIntervalActionSuccess(),
			scClient:         createAddMockIntervalSCAddError(),
			intervalAction:   ValidIntervalAction,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			name:             "Error AddIntervalAction",
			mockDb:           createAddMockIntervalActionAddErr(),
			scClient:         createAddMockIntervalSCSuccess(),
			intervalAction:   ValidIntervalAction,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewAddExecutor(test.mockDb, test.scClient, test.intervalAction)
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

			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Errorf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
				}
			}
		})
	}
}

func createAddMockIntervalActionSuccess() IntervalActionWriter {
	dbMock := mocks.IntervalActionWriter{}
	dbMock.On("IntervalActionByName", ValidIntervalAction.Name).Return(OtherValidIntervalAction, nil)
	dbMock.On("IntervalByName", Intervals[0].Name).Return(Intervals[0], nil)
	dbMock.On("AddIntervalAction", ValidIntervalAction).Return(ValidIntervalAction.ID, nil)
	return &dbMock
}

func createAddMockIntervalActionAddErr() IntervalActionWriter {
	dbMock := mocks.IntervalActionWriter{}
	dbMock.On("IntervalActionByName", ValidIntervalAction.Name).Return(OtherValidIntervalAction, nil)
	dbMock.On("IntervalByName", Intervals[0].Name).Return(Intervals[0], nil)
	dbMock.On("AddIntervalAction", ValidIntervalAction).Return(ValidIntervalAction.ID, Error)
	return &dbMock
}

func createAddMockIntervalActionByNameSameErr() IntervalActionWriter {
	dbMock := mocks.IntervalActionWriter{}
	dbMock.On("IntervalActionByName", ValidIntervalAction.Name).Return(ValidIntervalAction, nil)
	dbMock.On("IntervalByName", Intervals[0].Name).Return(Intervals[0], nil)
	dbMock.On("AddIntervalAction", ValidIntervalAction).Return(ValidIntervalAction.ID, nil)
	return &dbMock
}

func createAddMockIntervalActionByNameErr() IntervalActionWriter {
	dbMock := mocks.IntervalActionWriter{}
	dbMock.On("IntervalActionByName", ValidIntervalAction.Name).Return(OtherValidIntervalAction, nil)
	dbMock.On("IntervalByName", Intervals[0].Name).Return(Intervals[0], Error)
	dbMock.On("AddIntervalAction", ValidIntervalAction).Return(ValidIntervalAction.ID, nil)
	return &dbMock
}

func createAddMockIntervalActionTargetErr() IntervalActionWriter {
	dbMock := mocks.IntervalActionWriter{}
	dbMock.On("IntervalActionByName", InvalidIntervalAction.Name).Return(OtherValidIntervalAction, nil)
	dbMock.On("IntervalByName", Intervals[0].Name).Return(Intervals[0], nil)
	dbMock.On("AddIntervalAction", InvalidIntervalAction).Return(InvalidIntervalAction.ID, nil)
	return &dbMock
}

func createAddMockIntervalActionNoIntervalErr() IntervalActionWriter {
	dbMock := mocks.IntervalActionWriter{}
	dbMock.On("IntervalActionByName", IntervalActionNoInterval.Name).Return(OtherValidIntervalAction, nil)
	dbMock.On("IntervalByName", Intervals[0].Name).Return(Intervals[0], nil)
	dbMock.On("AddIntervalAction", IntervalActionNoInterval).Return(IntervalActionNoInterval.ID, nil)
	return &dbMock
}

func createAddMockIntervalSCSuccess() SchedulerQueueWriter {
	dbMock := mocks.SchedulerQueueWriter{}
	dbMock.On("QueryIntervalActionByName", ValidIntervalAction.Name).Return(OtherValidIntervalAction, nil)
	dbMock.On("AddIntervalActionToQueue", ValidIntervalAction).Return(nil)
	return &dbMock
}

func createAddMockIntervalSCQueryError() SchedulerQueueWriter {
	dbMock := mocks.SchedulerQueueWriter{}
	dbMock.On("QueryIntervalActionByName", ValidIntervalAction.Name).Return(ValidIntervalAction, nil)
	dbMock.On("AddIntervalActionToQueue", ValidIntervalAction).Return(Error)
	return &dbMock
}

func createAddMockIntervalSCAddError() SchedulerQueueWriter {
	dbMock := mocks.SchedulerQueueWriter{}
	dbMock.On("QueryIntervalActionByName", ValidIntervalAction.Name).Return(OtherValidIntervalAction, nil)
	dbMock.On("AddIntervalActionToQueue", ValidIntervalAction).Return(Error)
	return &dbMock
}
