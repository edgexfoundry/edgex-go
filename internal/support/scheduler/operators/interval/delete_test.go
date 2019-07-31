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

//TODO: TestFailDeleteOnExistingIntervalActions
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
			database:          createMockIntervalDeleterStringArg("DeleteIntervalById", nil, Id),
			scLoader:          createMockSCLoaderStringArg("QueryIntervalByID", nil, SuccessfulDatabaseResult[0], Id),
			scDeleter:         createMockSCDeleterStringArg("RemoveIntervalInQueue", nil, Id),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Interval not found",
			database:          createMockIntervalDeleterStringArg("DeleteIntervalById", ErrorNotFound, Id),
			scLoader:          createMockSCLoaderStringArg("QueryIntervalByID", nil, SuccessfulDatabaseResult[0], Id),
			scDeleter:         createMockSCDeleterStringArg("RemoveIntervalInQueue", nil, Id),
			expectError:       true,
			expectedErrorType: intervalErrors.ErrIntervalNotFound{},
		},
		{
			name:              "Delete error",
			database:          createMockIntervalDeleterStringArg("DeleteIntervalById", Error, Id),
			scLoader:          createMockSCLoaderStringArg("QueryIntervalByID", nil, SuccessfulDatabaseResult[0], Id),
			scDeleter:         createMockSCDeleterStringArg("RemoveIntervalInQueue", nil, Id),
			expectError:       true,
			expectedErrorType: Error,
		},
		{
			name:              "Error with getting interval",
			database:          createMockIntervalDeleterStringArg("DeleteIntervalById", nil, Id),
			scLoader:          createMockSCLoaderStringArg("QueryIntervalByID", ErrorNotFound, contract.Interval{}, Id),
			scDeleter:         createMockSCDeleterStringArg("RemoveIntervalInQueue", nil, Id),
			expectError:       true,
			expectedErrorType: intervalErrors.ErrIntervalNotFound{},
		},
		{
			name:              "Error with removing interval in queue",
			database:          createMockIntervalDeleterStringArg("DeleteIntervalById", nil, Id),
			scLoader:          createMockSCLoaderStringArg("QueryIntervalByID", nil, SuccessfulDatabaseResult[0], Id),
			scDeleter:         createMockSCDeleterStringArg("RemoveIntervalInQueue", ErrorNotFound, Id),
			expectError:       true,
			expectedErrorType: intervalErrors.ErrIntervalNotFound{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewDeleteByIDExecutor(test.database, test.scLoader, test.scDeleter, Id)
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

func createMockIntervalDeleterStringArg(methodName string, err error, arg string) IntervalDeleter {
	dbMock := mocks.IntervalDeleter{}
	dbMock.On(methodName, arg).Return(err)
	return &dbMock
}
func createMockSCLoaderStringArg(methodName string, err error, ret interface{}, arg string) SchedulerQueueLoader {
	dbMock := mocks.SchedulerQueueLoader{}
	dbMock.On(methodName, arg).Return(ret, err)
	return &dbMock
}
func createMockSCDeleterStringArg(methodName string, err error, arg string) SchedulerQueueDeleter {
	dbMock := mocks.SchedulerQueueDeleter{}
	dbMock.On(methodName, arg).Return(err)
	return &dbMock
}
