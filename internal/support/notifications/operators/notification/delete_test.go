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

var TestAge int = 1564594093

func TestNotificationById(t *testing.T) {
	tests := []struct {
		name              string
		database          NotificationDeleter
		expectError       bool
		expectedErrorType error
	}{
		{
			name:              "Successful Delete",
			database:          createMockNotificiationDeleterStringArg("DeleteNotificationById", nil, Id),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Notification not found",
			database:          createMockNotificiationDeleterStringArg("DeleteNotificationById", ErrorNotFound, Id),
			expectError:       true,
			expectedErrorType: notificationErrors.ErrNotificationNotFound{},
		},
		{
			name:              "Delete error",
			database:          createMockNotificiationDeleterStringArg("DeleteNotificationById", Error, Id),
			expectError:       true,
			expectedErrorType: Error,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewDeleteByIDExecutor(test.database, Id)
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

func TestNotificationBySlug(t *testing.T) {
	tests := []struct {
		name              string
		database          NotificationDeleter
		expectError       bool
		expectedErrorType error
	}{
		{
			name:              "Successful Delete",
			database:          createMockNotificiationDeleterStringArg("DeleteNotificationBySlug", nil, Slug),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Notification not found",
			database:          createMockNotificiationDeleterStringArg("DeleteNotificationBySlug", ErrorNotFound, Slug),
			expectError:       true,
			expectedErrorType: notificationErrors.ErrNotificationNotFound{},
		},
		{
			name:              "Delete error",
			database:          createMockNotificiationDeleterStringArg("DeleteNotificationBySlug", Error, Slug),
			expectError:       true,
			expectedErrorType: Error,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewDeleteBySlugExecutor(test.database, Slug)
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

func createMockNotificiationDeleterStringArg(methodName string, err error, arg string) NotificationDeleter {
	dbMock := mocks.NotificationDeleter{}
	dbMock.On(methodName, arg).Return(err)
	return &dbMock
}

func createMockNotificiationDeleterIntArg(methodName string, err error, arg int) NotificationDeleter {
	dbMock := mocks.NotificationDeleter{}
	dbMock.On(methodName, arg).Return(err)
	return &dbMock
}

func TestNotificationsByAge(t *testing.T) {
	tests := []struct {
		name              string
		database          NotificationDeleter
		expectError       bool
		expectedErrorType error
	}{
		{
			name:              "Successful Delete",
			database:          createMockNotificiationDeleterIntArg("DeleteNotificationsOld", nil, TestAge),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Delete error",
			database:          createMockNotificiationDeleterIntArg("DeleteNotificationsOld", Error, TestAge),
			expectError:       true,
			expectedErrorType: Error,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewDeleteByAgeExecutor(test.database, TestAge)
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
