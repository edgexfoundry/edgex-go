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

var ValidInterval = SuccessfulDatabaseResult[0]
var OtherValidInterval = SuccessfulDatabaseResult[1]
var InvalidStartInterval = SuccessfulDatabaseResult[2]
var InvalidEndInterval = SuccessfulDatabaseResult[3]
var InvalidFreqInterval = SuccessfulDatabaseResult[4]

func TestAddExecutor(t *testing.T) {

	tests := []struct {
		name             string
		mockDb           IntervalWriter
		scClient         SchedulerQueueWriter
		interval         contract.Interval
		expectedResult   string
		expectedError    bool
		expectedErrorVal error
	}{

		{
			name:             "Successful database call",
			mockDb:           createAddMockIntervalSuccess(),
			scClient:         createAddMockIntervalSCSuccess(),
			interval:         ValidInterval,
			expectedResult:   ValidInterval.ID,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name:             "Error Interval In Use",
			mockDb:           createAddMockIntervalInUse(),
			scClient:         createAddMockIntervalSCSuccess(),
			interval:         ValidInterval,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: intervalErrors.NewErrIntervalNameInUse(SuccessfulDatabaseResult[0].Name),
		},
		{
			name:             "Error AddInterval",
			mockDb:           createAddMockIntervalError(),
			scClient:         createAddMockIntervalSCSuccess(),
			interval:         ValidInterval,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			name:             "Error AddToQueue",
			mockDb:           createAddMockIntervalSuccess(),
			scClient:         createAddMockIntervalSCError(),
			interval:         ValidInterval,
			expectedResult:   ValidInterval.ID,
			expectedError:    true,
			expectedErrorVal: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewAddExecutor(test.mockDb, test.scClient, test.interval)
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

func createAddMockIntervalSuccess() IntervalWriter {
	dbMock := mocks.IntervalWriter{}
	dbMock.On("IntervalByName", Name).Return(OtherValidInterval, nil)
	dbMock.On("AddInterval", ValidInterval).Return(Id, nil)
	return &dbMock
}

func createAddMockIntervalInUse() IntervalWriter {
	dbMock := mocks.IntervalWriter{}
	dbMock.On("IntervalByName", Name).Return(ValidInterval, nil)
	dbMock.On("AddInterval", ValidInterval).Return(Id, nil)
	return &dbMock
}

func createAddMockIntervalInvalidFreq() IntervalWriter {
	dbMock := mocks.IntervalWriter{}
	dbMock.On("IntervalByName", InvalidFreqInterval.Name).Return(OtherValidInterval, nil)
	dbMock.On("AddInterval", ValidInterval).Return(Id, nil)
	return &dbMock
}

func createAddMockIntervalError() IntervalWriter {
	dbMock := mocks.IntervalWriter{}
	dbMock.On("IntervalByName", ValidInterval.Name).Return(OtherValidInterval, nil)
	dbMock.On("AddInterval", ValidInterval).Return("", Error)
	return &dbMock
}

func createAddMockInvalidStart() IntervalWriter {
	dbMock := mocks.IntervalWriter{}
	dbMock.On("IntervalByName", InvalidStartInterval.Name).Return(ValidInterval, nil)
	dbMock.On("AddInterval", ValidInterval).Return(Id, nil)
	return &dbMock
}

func createAddMockInvalidEnd() IntervalWriter {
	dbMock := mocks.IntervalWriter{}
	dbMock.On("IntervalByName", InvalidEndInterval.Name).Return(ValidInterval, nil)
	dbMock.On("AddInterval", ValidInterval).Return(Id, nil)
	return &dbMock
}

func createAddMockIntervalSCSuccess() SchedulerQueueWriter {
	dbMock := mocks.SchedulerQueueWriter{}
	dbMock.On("AddIntervalToQueue", ValidInterval).Return(nil)
	return &dbMock
}

func createAddMockIntervalSCError() SchedulerQueueWriter {
	dbMock := mocks.SchedulerQueueWriter{}
	dbMock.On("AddIntervalToQueue", ValidInterval).Return(Error)
	return &dbMock
}
