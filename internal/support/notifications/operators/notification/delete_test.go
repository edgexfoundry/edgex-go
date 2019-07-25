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
	"reflect"
	"testing"

	notificationErrors "github.com/edgexfoundry/edgex-go/internal/support/notifications/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/operators/notification/mocks"
)

var notificationID string = "683cfbbf-758a-4dca-b70e-a7265589fed6"

func TestNotificationById(t *testing.T) {
	tests := []struct {
		name              string
		database          NotificationDeleter
		expectError       bool
		expectedErrorType error
	}{
		{
			name:              "Successful Delete",
			database:          createNotificationDeleter(),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Notification not found",
			database:          createNotificationDeleterNotFound(),
			expectError:       true,
			expectedErrorType: notificationErrors.ErrNotificationNotFound{},
		},
		{
			name:              "Delete error",
			database:          createNotificationDeleteError(),
			expectError:       true,
			expectedErrorType: Error,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewDeleteByIDExecutor(test.database, notificationID)
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
func createNotificationDeleter() NotificationDeleter {
	d := mocks.NotificationDeleter{}
	d.On("DeleteNotificationById", notificationID).Return(nil)

	return &d
}

func createNotificationDeleterNotFound() NotificationDeleter {
	d := mocks.NotificationDeleter{}
	d.On("DeleteNotificationById", notificationID).Return(ErrorNotFound)

	return &d
}
func createNotificationDeleteError() NotificationDeleter {
	d := mocks.NotificationDeleter{}
	d.On("DeleteNotificationById", notificationID).Return(Error)

	return &d
}
