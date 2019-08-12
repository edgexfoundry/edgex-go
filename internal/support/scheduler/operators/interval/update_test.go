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
	"testing"

	intervalErrors "github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/operators/interval/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var TestInvalidCron = "invalid"
var TestValidCron = "* * * ? * *"

var IntervalHasInvalidCron = contract.Interval{

	ID:        Id,
	Name:      "hourly",
	Start:     "20160101T000000",
	End:       "",
	Frequency: "PT1H",
	Cron:      TestInvalidCron,
}

var IntervalHasValidCron = contract.Interval{

	ID:         Id,
	Name:       OtherName,
	Start:      "20160101T000000",
	End:        "",
	Frequency:  "PT1H",
	Cron:       TestValidCron,
	Timestamps: contract.Timestamps{Origin: 201601011565351081},
}

func TestUpdateExecutor(t *testing.T) {
	successNoId := SuccessfulDatabaseResult[0]
	successNoId.Name = successNoId.ID
	successNoId.ID = ""

	successNewName := SuccessfulDatabaseResult[0]
	successNewName.Name = "something different"

	tests := []struct {
		name             string
		dbMock           IntervalUpdater
		scClient         SchedulerQueueUpdater
		interval         contract.Interval
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name:             "IntervalActionsByIntervalName success",
			dbMock:           createMockIntervalUpdaterIntvActionSuccess(),
			scClient:         createMockIntervalUpdaterSCSuccess(SuccessfulDatabaseResult[0]),
			interval:         SuccessfulDatabaseResult[0],
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name:             "IntervalActionsByIntervalName In Use",
			dbMock:           createMockIntervalUpdaterIntvActionInUse(),
			scClient:         createMockIntervalUpdaterSCSuccess(SuccessfulDatabaseResult[0]),
			interval:         SuccessfulDatabaseResult[0],
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalStillInUse(IntervalHasValidCron.Name),
		},
		{
			name:             "IntervalActionsByIntervalName Error",
			dbMock:           createMockIntervalUpdaterIntvActionErr(),
			scClient:         createMockIntervalUpdaterSCSuccess(SuccessfulDatabaseResult[0]),
			interval:         SuccessfulDatabaseResult[0],
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			name:             "IntervalByName Error",
			dbMock:           createMockIntervalUpdaterNameErr(),
			scClient:         createMockIntervalUpdaterSCSuccess(SuccessfulDatabaseResult[0]),
			interval:         SuccessfulDatabaseResult[0],
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			name:             "IntervalNameInUseErr",
			dbMock:           createMockIntervalUpdaterNameInUseErr(),
			scClient:         createMockIntervalUpdaterSCSuccess(SuccessfulDatabaseResult[0]),
			interval:         SuccessfulDatabaseResult[0],
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalNameInUse(SuccessfulDatabaseResult[0].Name),
		},
		{
			name:             "Successful database call, search by ID",
			dbMock:           createMockIntervalUpdaterSuccess(),
			scClient:         createMockIntervalUpdaterSCSuccess(SuccessfulDatabaseResult[0]),
			interval:         SuccessfulDatabaseResult[0],
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name:             "Not found by ID, Not found by Name",
			dbMock:           createMockIntervalUpdaterNotFoundErr(),
			scClient:         createMockIntervalUpdaterSCSuccess(SuccessfulDatabaseResult[0]),
			interval:         SuccessfulDatabaseResult[0],
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalNotFound(SuccessfulDatabaseResult[0].ID),
		},
		{
			name:             "Error Cron",
			dbMock:           createMockIntervalUpdaterCronErr(),
			scClient:         createMockIntervalUpdaterSCSuccess(SuccessfulDatabaseResult[0]),
			interval:         IntervalHasInvalidCron,
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrInvalidCronFormat(TestInvalidCron),
		},
		{
			name:             "Cron is valid",
			dbMock:           createMockIntervalUpdaterCronValid(),
			scClient:         createMockIntervalUpdaterSCSuccess(IntervalHasValidCron),
			interval:         IntervalHasValidCron,
			expectedError:    false,
			expectedErrorVal: nil,
		},

		{
			name:             "Unexpected error in UpdateIntervalInQueue",
			dbMock:           createMockIntervalUpdaterSuccess(),
			scClient:         createMockIntervalUpdaterSCErr(SuccessfulDatabaseResult[0]),
			interval:         SuccessfulDatabaseResult[0],
			expectedError:    true,
			expectedErrorVal: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewUpdateExecutor(test.dbMock, test.scClient, test.interval)
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

func createMockIntervalUpdaterSuccess() IntervalUpdater {
	dbMock := mocks.IntervalUpdater{}
	dbMock.On("IntervalById", Id).Return(SuccessfulDatabaseResult[0], nil)
	dbMock.On("IntervalByName", Name).Return(SuccessfulDatabaseResult[0], nil)
	dbMock.On("IntervalActionsByIntervalName", Name).Return([]contract.IntervalAction{}, nil)
	dbMock.On("UpdateInterval", SuccessfulDatabaseResult[0]).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterNameInUseErr() IntervalUpdater {
	dbMock := mocks.IntervalUpdater{}
	dbMock.On("IntervalById", Id).Return(IntervalHasValidCron, nil)
	dbMock.On("IntervalByName", Name).Return(IntervalHasValidCron, nil)
	dbMock.On("IntervalActionsByIntervalName", OtherName).Return([]contract.IntervalAction{}, nil)
	dbMock.On("UpdateInterval", SuccessfulDatabaseResult[0]).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterNameErr() IntervalUpdater {
	dbMock := mocks.IntervalUpdater{}
	dbMock.On("IntervalById", Id).Return(IntervalHasValidCron, nil)
	dbMock.On("IntervalByName", Name).Return(IntervalHasValidCron, Error)
	dbMock.On("IntervalActionsByIntervalName", OtherName).Return([]contract.IntervalAction{}, nil)
	dbMock.On("UpdateInterval", SuccessfulDatabaseResult[0]).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterIntvActionErr() IntervalUpdater {
	dbMock := mocks.IntervalUpdater{}
	dbMock.On("IntervalById", Id).Return(IntervalHasValidCron, nil)
	dbMock.On("IntervalByName", Name).Return(contract.Interval{}, nil)
	dbMock.On("IntervalActionsByIntervalName", OtherName).Return([]contract.IntervalAction{}, Error)
	dbMock.On("UpdateInterval", SuccessfulDatabaseResult[0]).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterIntvActionInUse() IntervalUpdater {
	dbMock := mocks.IntervalUpdater{}
	dbMock.On("IntervalById", Id).Return(IntervalHasValidCron, nil)
	dbMock.On("IntervalByName", Name).Return(contract.Interval{}, nil)
	dbMock.On("IntervalActionsByIntervalName", OtherName).Return(SuccessfulIntervalActionResult, nil)
	dbMock.On("UpdateInterval", SuccessfulDatabaseResult[0]).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterIntvActionSuccess() IntervalUpdater {
	dbMock := mocks.IntervalUpdater{}
	dbMock.On("IntervalById", Id).Return(IntervalHasValidCron, nil)
	dbMock.On("IntervalByName", Name).Return(contract.Interval{}, nil)
	dbMock.On("IntervalActionsByIntervalName", OtherName).Return([]contract.IntervalAction{}, nil)
	dbMock.On("UpdateInterval", SuccessfulDatabaseResult[0]).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterCronErr() IntervalUpdater {
	dbMock := mocks.IntervalUpdater{}
	dbMock.On("IntervalById", Id).Return(IntervalHasInvalidCron, nil)
	dbMock.On("IntervalByName", Name).Return(IntervalHasInvalidCron, nil)
	dbMock.On("IntervalActionsByIntervalName", Name).Return([]contract.IntervalAction{}, nil)
	dbMock.On("UpdateInterval", IntervalHasInvalidCron).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterCronValid() IntervalUpdater {
	dbMock := mocks.IntervalUpdater{}
	dbMock.On("IntervalById", Id).Return(IntervalHasValidCron, nil)
	dbMock.On("IntervalByName", Name).Return(IntervalHasValidCron, nil)
	dbMock.On("IntervalActionsByIntervalName", Name).Return([]contract.IntervalAction{}, nil)
	dbMock.On("UpdateInterval", IntervalHasValidCron).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterNotFoundErr() IntervalUpdater {
	dbMock := mocks.IntervalUpdater{}
	dbMock.On("IntervalById", Id).Return(contract.Interval{}, ErrorNotFound)
	dbMock.On("IntervalByName", Name).Return(contract.Interval{}, ErrorNotFound)
	dbMock.On("IntervalActionsByIntervalName", Name).Return([]contract.IntervalAction{}, nil)
	dbMock.On("UpdateInterval", IntervalHasInvalidCron).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterSCSuccess(interval contract.Interval) SchedulerQueueUpdater {
	dbMock := mocks.SchedulerQueueUpdater{}
	dbMock.On("UpdateIntervalInQueue", interval).Return(nil)
	return &dbMock
}

func createMockIntervalUpdaterSCErr(interval contract.Interval) SchedulerQueueUpdater {
	dbMock := mocks.SchedulerQueueUpdater{}
	dbMock.On("UpdateIntervalInQueue", interval).Return(Error)
	return &dbMock
}
