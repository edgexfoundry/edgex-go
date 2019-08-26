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

package notification

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/notifications/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/operators/notification/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var ValidNotification = SuccessfulDatabaseResult[0]
var OtherValidNotification = SuccessfulDatabaseResult[1]

func TestAddExecutor(t *testing.T) {

	tests := []struct {
		name             string
		mockDb           NotificationWriter
		notification     contract.Notification
		expectedResult   string
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name:             "Successful adding",
			mockDb:           createAddMockNotificationSuccess(),
			notification:     ValidNotification,
			expectedResult:   Id,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name:             "Error adding already in use",
			mockDb:           createAddMockNotificationInUseError(),
			notification:     ValidNotification,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: errors.NewErrNotificationInUse(ValidNotification.Slug),
		},
		{
			name:             "Error adding",
			mockDb:           createAddMockNotificationError(),
			notification:     ValidNotification,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewAddExecutor(test.mockDb, test.notification)
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

func createAddMockNotificationSuccess() NotificationWriter {
	dbMock := mocks.NotificationWriter{}
	dbMock.On("AddNotification", ValidNotification).Return(Id, nil)
	return &dbMock
}

func createAddMockNotificationInUseError() NotificationWriter {
	dbMock := mocks.NotificationWriter{}
	dbMock.On("AddNotification", ValidNotification).Return("", db.ErrNotUnique)
	return &dbMock
}

func createAddMockNotificationError() NotificationWriter {
	dbMock := mocks.NotificationWriter{}
	dbMock.On("AddNotification", ValidNotification).Return("", Error)
	return &dbMock
}
