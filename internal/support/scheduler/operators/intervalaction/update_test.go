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
	"testing"

	intervalErrors "github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/operators/intervalaction/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestUpdateExecutor(t *testing.T) {

	tests := []struct {
		name             string
		dbMock           IntervalActionUpdater
		scClient         SchedulerQueueUpdater
		intervalAction   contract.IntervalAction
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name:             "Error UpdateIntervalActionQueue",
			dbMock:           createMockIntervalActionUpdaterSuccess(),
			scClient:         createMockIntervalUpdaterSCErr(),
			intervalAction:   ValidIntervalAction,
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalActionNotFound(ValidIntervalAction.Name),
		},
		{
			name:             "IntervalActionsByIntervalName success",
			dbMock:           createMockIntervalActionUpdaterSuccess(),
			scClient:         createMockIntervalUpdaterSCSuccess(),
			intervalAction:   ValidIntervalAction,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name:             "Error IntervalActionById",
			dbMock:           createMockIntervalActionUpdaterByIDByNameErr(),
			scClient:         createMockIntervalUpdaterSCSuccess(),
			intervalAction:   ValidIntervalAction,
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalActionNotFound(ValidIntervalAction.ID),
		},
		{
			name:             "Error IntervalActionByName",
			dbMock:           createMockIntervalActionUpdaterByNameErr(),
			scClient:         createMockIntervalUpdaterSCSuccess(),
			intervalAction:   ValidIntervalAction,
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalNotFound(ValidIntervalAction.Interval),
		},
		{
			name:             "Error IntervalActionByName",
			dbMock:           createMockIntervalActionUpdaterNoNameErr(),
			scClient:         createMockIntervalUpdaterSCSuccess(),
			intervalAction:   IntervalActionNoName,
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalActionTargetNameRequired(IntervalActionNoName.Name),
		},
		{
			name:             "Error Name in Use",
			dbMock:           createMockIntervalActionUpdaterNameInUseErr(),
			scClient:         createMockIntervalUpdaterSCSuccess(),
			intervalAction:   ValidIntervalAction,
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalActionNameInUse(ValidIntervalAction.Name),
		},
		{
			name:             "Error UpdateIntervalActionQueue",
			dbMock:           createMockIntervalActionUpdaterSuccess(),
			scClient:         createMockIntervalUpdaterSCErr(),
			intervalAction:   ValidIntervalAction,
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalActionNotFound(ValidIntervalAction.Name),
		},
		{
			name:             "Error No Target",
			dbMock:           createMockIntervalActionUpdaterNoIntervalErr(),
			scClient:         createMockIntervalUpdaterSCNoTargetErr(),
			intervalAction:   IntervalActionNoTarget,
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalActionTargetNameRequired(IntervalActionNoTarget.ID),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewUpdateExecutor(test.dbMock, test.scClient, test.intervalAction)
			err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
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

func createMockIntervalActionUpdaterSuccess() IntervalActionUpdater {
	dbMock := mocks.IntervalActionUpdater{}
	dbMock.On("IntervalActionById", ValidIntervalAction.ID).Return(ValidIntervalAction, nil)
	dbMock.On("IntervalActionByName", ValidIntervalAction.Name).Return(contract.IntervalAction{}, nil)
	dbMock.On("IntervalByName", ValidIntervalAction.Interval).Return(contract.Interval{}, nil)
	dbMock.On("UpdateIntervalAction", ValidIntervalAction).Return(nil)
	return &dbMock
}

func createMockIntervalActionUpdaterByIDByNameErr() IntervalActionUpdater {
	dbMock := mocks.IntervalActionUpdater{}
	dbMock.On("IntervalActionById", ValidIntervalAction.ID).Return(ValidIntervalAction, Error)
	dbMock.On("IntervalActionByName", ValidIntervalAction.Name).Return(contract.IntervalAction{}, Error)
	dbMock.On("IntervalByName", ValidIntervalAction.Interval).Return(contract.Interval{}, nil)
	dbMock.On("UpdateIntervalAction", ValidIntervalAction).Return(nil)
	return &dbMock
}

func createMockIntervalActionUpdaterByNameErr() IntervalActionUpdater {
	dbMock := mocks.IntervalActionUpdater{}
	dbMock.On("IntervalActionById", ValidIntervalAction.ID).Return(ValidIntervalAction, nil)
	dbMock.On("IntervalActionByName", ValidIntervalAction.Name).Return(contract.IntervalAction{}, nil)
	dbMock.On("IntervalByName", ValidIntervalAction.Interval).Return(contract.Interval{}, Error)
	dbMock.On("UpdateIntervalAction", ValidIntervalAction).Return(nil)
	return &dbMock
}

func createMockIntervalActionUpdaterNoNameErr() IntervalActionUpdater {
	dbMock := mocks.IntervalActionUpdater{}
	dbMock.On("IntervalActionById", IntervalActionNoName.ID).Return(IntervalActionNoName, nil)
	dbMock.On("IntervalActionByName", IntervalActionNoName.Name).Return(contract.IntervalAction{}, nil)
	dbMock.On("IntervalByName", IntervalActionNoName.Interval).Return(contract.Interval{}, nil)
	dbMock.On("UpdateIntervalAction", IntervalActionNoName).Return(nil)
	return &dbMock
}

func createMockIntervalActionUpdaterNameInUseErr() IntervalActionUpdater {
	dbMock := mocks.IntervalActionUpdater{}
	dbMock.On("IntervalActionById", ValidIntervalAction.ID).Return(OtherValidIntervalAction, nil)
	dbMock.On("IntervalActionByName", ValidIntervalAction.Name).Return(ValidIntervalAction, nil)
	dbMock.On("IntervalByName", ValidIntervalAction.Interval).Return(contract.Interval{}, nil)
	dbMock.On("UpdateIntervalAction", ValidIntervalAction).Return(nil)
	return &dbMock
}

func createMockIntervalActionUpdaterNoIntervalErr() IntervalActionUpdater {
	dbMock := mocks.IntervalActionUpdater{}
	dbMock.On("IntervalActionById", IntervalActionNoTarget.ID).Return(IntervalActionNoTarget, nil)
	dbMock.On("IntervalActionByName", IntervalActionNoTarget.Name).Return(IntervalActionNoInterval, nil)
	dbMock.On("IntervalByName", IntervalActionNoTarget.Interval).Return(contract.Interval{}, nil)
	dbMock.On("UpdateIntervalAction", IntervalActionNoTarget).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterSCSuccess() SchedulerQueueUpdater {
	dbMock := mocks.SchedulerQueueUpdater{}
	dbMock.On("UpdateIntervalActionQueue", ValidIntervalAction).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterSCErr() SchedulerQueueUpdater {
	dbMock := mocks.SchedulerQueueUpdater{}
	dbMock.On("UpdateIntervalActionQueue", ValidIntervalAction).Return(Error)
	return &dbMock
}

func createMockIntervalUpdaterSCNoTargetErr() SchedulerQueueUpdater {
	dbMock := mocks.SchedulerQueueUpdater{}
	dbMock.On("UpdateIntervalActionQueue", IntervalActionNoTarget).Return(nil)
	return &dbMock
}
