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
	"reflect"
	"testing"

	intervalErrors "github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/operators/interval/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var ErrorStillInUse = intervalErrors.ErrIntervalStillUsedByIntervalActions{}

func createMockIdDBSuccessDel() IntervalDeleter {
	dbMock := mocks.IntervalDeleter{}
	dbMock.On("DeleteIntervalById", Id).Return(nil)
	return &dbMock
}

func createMockIdDBNotFoundErr() IntervalDeleter {
	dbMock := mocks.IntervalDeleter{}
	dbMock.On("DeleteIntervalById", Id).Return(ErrorNotFound)
	return &dbMock
}

func createMockIdDBErr() IntervalDeleter {
	dbMock := mocks.IntervalDeleter{}
	dbMock.On("DeleteIntervalById", Id).Return(Error)
	return &dbMock
}

func createMockIdSCSuccessDel() SchedulerQueueDeleter {
	dbMock := mocks.SchedulerQueueDeleter{}
	dbMock.On("QueryIntervalByID", Id).Return(SuccessfulDatabaseResult[0], nil)
	dbMock.On("RemoveIntervalInQueue", Id).Return(nil)
	return &dbMock
}

func createMockIdSCNotFoundErr() SchedulerQueueDeleter {
	dbMock := mocks.SchedulerQueueDeleter{}
	dbMock.On("QueryIntervalByID", Id).Return(contract.Interval{}, ErrorNotFound)
	dbMock.On("RemoveIntervalInQueue", Id).Return(nil)
	return &dbMock
}

func createMockIdSCNotFoundQueueErr() SchedulerQueueDeleter {
	dbMock := mocks.SchedulerQueueDeleter{}
	dbMock.On("QueryIntervalByID", Id).Return(SuccessfulDatabaseResult[0], nil)
	dbMock.On("RemoveIntervalInQueue", Id).Return(ErrorNotFound)
	return &dbMock
}

func TestIntervalById(t *testing.T) {
	tests := []struct {
		name              string
		database          IntervalDeleter
		scLoader          SchedulerQueueLoader
		scDeleter         SchedulerQueueDeleter
		expectError       bool
		expectedErrorType error
	}{
		{
			name:              "Successful Delete",
			database:          createMockIdDBSuccessDel(),
			scDeleter:         createMockIdSCSuccessDel(),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Interval not found",
			database:          createMockIdDBNotFoundErr(),
			scDeleter:         createMockIdSCSuccessDel(),
			expectError:       true,
			expectedErrorType: intervalErrors.ErrIntervalNotFound{},
		},
		{
			name:              "Delete error",
			database:          createMockIdDBErr(),
			scDeleter:         createMockIdSCSuccessDel(),
			expectError:       true,
			expectedErrorType: Error,
		},
		{
			name:              "Error with getting interval",
			database:          createMockIdDBSuccessDel(),
			scDeleter:         createMockIdSCNotFoundErr(),
			expectError:       true,
			expectedErrorType: intervalErrors.ErrIntervalNotFound{},
		},
		{
			name:              "Error with removing interval in queue",
			database:          createMockIdDBSuccessDel(),
			scDeleter:         createMockIdSCNotFoundQueueErr(),
			expectError:       true,
			expectedErrorType: intervalErrors.ErrIntervalNotFound{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewDeleteByIDExecutor(test.database, test.scDeleter, Id)
			err := op.Execute()

			if test.expectError && err == nil {
				t.Error("We expected an error but did not get one")
			}

			if !test.expectError && err != nil {
				t.Errorf("We do not expected an error but got one. %s", err.Error())
			}

			if test.expectError {
				eet := reflect.TypeOf(test.expectedErrorType)
				aet := reflect.TypeOf(err)
				if !aet.AssignableTo(eet) {
					t.Errorf("Expected error of type %v, but got an error of type %v", eet, aet)
				}
			}
			return
		})
	}
}

func createMockNameDBSuccessDel() IntervalDeleter {
	dbMock := mocks.IntervalDeleter{}
	dbMock.On("DeleteIntervalById", Id).Return(nil)
	dbMock.On("IntervalByName", Name).Return(SuccessfulDatabaseResult[0], nil)
	dbMock.On("IntervalActionsByIntervalName", Name).Return([]contract.IntervalAction{}, nil)
	return &dbMock
}

func createMockNameSCSuccessDel() SchedulerQueueDeleter {
	dbMock := mocks.SchedulerQueueDeleter{}
	dbMock.On("QueryIntervalByName", Name).Return(SuccessfulDatabaseResult[0], nil)
	dbMock.On("RemoveIntervalInQueue", Id).Return(nil)
	return &dbMock
}

func createMockNameDBNotFound() IntervalDeleter {
	dbMock := mocks.IntervalDeleter{}
	dbMock.On("DeleteIntervalById", Id).Return(ErrorNotFound)
	dbMock.On("IntervalByName", Name).Return(contract.Interval{}, ErrorNotFound)
	dbMock.On("IntervalActionsByIntervalName", Name).Return([]contract.IntervalAction{}, nil)
	return &dbMock
}

func createMockNameIntervalByNameErr() IntervalDeleter {
	dbMock := mocks.IntervalDeleter{}
	dbMock.On("DeleteIntervalById", Id).Return(nil)
	dbMock.On("IntervalByName", Name).Return(contract.Interval{}, Error)
	dbMock.On("IntervalActionsByIntervalName", Name).Return([]contract.IntervalAction{}, nil)
	return &dbMock
}

func createMockNameIntervalStillInUseErr() IntervalDeleter {
	dbMock := mocks.IntervalDeleter{}
	dbMock.On("DeleteIntervalById", Id).Return(nil)
	dbMock.On("IntervalByName", Name).Return(SuccessfulDatabaseResult[0], nil)
	dbMock.On("IntervalActionsByIntervalName", Name).Return(SuccessfulIntervalActionResult, nil)
	return &dbMock
}

func createMockNameDeleteErr() IntervalDeleter {
	dbMock := mocks.IntervalDeleter{}
	dbMock.On("DeleteIntervalById", Id).Return(Error)
	dbMock.On("IntervalByName", Name).Return(SuccessfulDatabaseResult[0], nil)
	dbMock.On("IntervalActionsByIntervalName", Name).Return([]contract.IntervalAction{}, nil)
	return &dbMock
}

func createMockNameRemoveIntervalInQueueErr() SchedulerQueueDeleter {
	dbMock := mocks.SchedulerQueueDeleter{}
	dbMock.On("QueryIntervalByName", Name).Return(SuccessfulDatabaseResult[0], nil)
	dbMock.On("RemoveIntervalInQueue", Id).Return(ErrorNotFound)
	return &dbMock
}

func createMockNameIntervalActionsByIntervalNameErr() IntervalDeleter {
	dbMock := mocks.IntervalDeleter{}
	dbMock.On("DeleteIntervalById", Id).Return(nil)
	dbMock.On("IntervalByName", Name).Return(SuccessfulDatabaseResult[0], nil)
	dbMock.On("IntervalActionsByIntervalName", Name).Return([]contract.IntervalAction{}, Error)
	return &dbMock
}

func createMockNameQueryIntervalByNameErr() SchedulerQueueDeleter {
	dbMock := mocks.SchedulerQueueDeleter{}
	dbMock.On("QueryIntervalByName", Name).Return(contract.Interval{}, ErrorNotFound)
	dbMock.On("RemoveIntervalInQueue", Id).Return(nil)
	return &dbMock
}

func TestIntervalByName(t *testing.T) {
	tests := []struct {
		name              string
		database          IntervalDeleter
		scDeleter         SchedulerQueueDeleter
		expectError       bool
		expectedErrorType error
	}{
		{
			name:              "Successful Delete",
			database:          createMockNameDBSuccessDel(),
			scDeleter:         createMockNameSCSuccessDel(),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Interval not found",
			database:          createMockNameDBNotFound(),
			scDeleter:         createMockNameSCSuccessDel(),
			expectError:       true,
			expectedErrorType: intervalErrors.ErrIntervalNotFound{},
		},
		{
			name:              "IntervalByName error",
			database:          createMockNameIntervalByNameErr(),
			scDeleter:         createMockNameSCSuccessDel(),
			expectError:       true,
			expectedErrorType: Error,
		},
		{
			name:              "IntervalStillInUse error",
			database:          createMockNameIntervalStillInUseErr(),
			scDeleter:         createMockNameSCSuccessDel(),
			expectError:       true,
			expectedErrorType: ErrorStillInUse,
		},
		{
			name:              "Delete error",
			database:          createMockNameDeleteErr(),
			scDeleter:         createMockNameSCSuccessDel(),
			expectError:       true,
			expectedErrorType: Error,
		},
		{
			name:              "Delete RemoveIntervalInQueue error",
			database:          createMockNameDBSuccessDel(),
			scDeleter:         createMockNameRemoveIntervalInQueueErr(),
			expectError:       true,
			expectedErrorType: intervalErrors.ErrIntervalNotFound{},
		},
		{
			name:              "Delete IntervalActionsByIntervalName error",
			database:          createMockNameIntervalActionsByIntervalNameErr(),
			scDeleter:         createMockNameSCSuccessDel(),
			expectError:       true,
			expectedErrorType: Error,
		},
		{
			name:              "Error QueryIntervalByName",
			database:          createMockNameDBSuccessDel(),
			scDeleter:         createMockNameQueryIntervalByNameErr(),
			expectError:       true,
			expectedErrorType: intervalErrors.ErrIntervalNotFound{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewDeleteByNameExecutor(test.database, test.scDeleter, Name)
			err := op.Execute()

			if test.expectError && err == nil {
				t.Error("We expected an error but did not get one")
			}

			if !test.expectError && err != nil {
				t.Errorf("We do not expected an error but got one. %s", err.Error())
			}

			if test.expectError {
				eet := reflect.TypeOf(test.expectedErrorType)
				aet := reflect.TypeOf(err)
				if !aet.AssignableTo(eet) {
					t.Errorf("Expected error of type %v, but got an error of type %v", eet, aet)
				}
			}
			return
		})
	}
}
