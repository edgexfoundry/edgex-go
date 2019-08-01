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
	"errors"
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/operators/notification/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var Id = "83cb038b-5a94-4707-985d-13effec62de2"
var Slug = "test-slug"
var Sender = "System Management"
var Limit = 5

var Error = errors.New("test error")
var ErrorNotFound = db.ErrNotFound
var SuccessfulDatabaseResult = []contract.Notification{
	{
		Slug:     "notice-test-123",
		Sender:   "System Management",
		Category: "SECURITY",
		Severity: "CRITICAL",
		Content:  "Hello, Notification!",
		Labels: []string{
			"test_label",
			"test_label2",
		},
	},
}

func TestIdExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         NotificationLoader
		expectedResult contract.Notification
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockNotificiationLoaderStringArg("GetNotificationById", nil, SuccessfulDatabaseResult[0], Id),
			expectedResult: SuccessfulDatabaseResult[0],
			expectedError:  false,
		},
		{
			name:           "Unsuccessful database call",
			mockDb:         createMockNotificiationLoaderStringArg("GetNotificationById", Error, contract.Notification{}, Id),
			expectedResult: contract.Notification{},
			expectedError:  true,
		},
		{
			name:           "Unsuccessful database call",
			mockDb:         createMockNotificiationLoaderStringArg("GetNotificationById", ErrorNotFound, contract.Notification{}, Id),
			expectedResult: contract.Notification{},
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewIdExecutor(test.mockDb, Id)
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

func TestSlugExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         NotificationLoader
		expectedResult contract.Notification
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockNotificiationLoaderStringArg("GetNotificationBySlug", nil, SuccessfulDatabaseResult[0], Id),
			expectedResult: SuccessfulDatabaseResult[0],
			expectedError:  false,
		},
		{
			name:           "Unsuccessful database call",
			mockDb:         createMockNotificiationLoaderStringArg("GetNotificationBySlug", Error, contract.Notification{}, Id),
			expectedResult: contract.Notification{},
			expectedError:  true,
		},
		{
			name:           "Unsuccessful database call",
			mockDb:         createMockNotificiationLoaderStringArg("GetNotificationBySlug", ErrorNotFound, contract.Notification{}, Id),
			expectedResult: contract.Notification{},
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewSlugExecutor(test.mockDb, Id)
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

func createMockNotificiationLoaderStringArg(methodName string, err error, ret interface{}, arg string) NotificationLoader {
	dbMock := mocks.NotificationLoader{}
	dbMock.On(methodName, arg).Return(ret, err)
	return &dbMock
}

func createMockNotificiationLoaderSenderStringArg(methodName string, err error, ret interface{}, sender string, limit int) NotificationLoader {
	dbMock := mocks.NotificationLoader{}
	dbMock.On(methodName, sender, limit).Return(ret, err)
	return &dbMock
}

func TestSenderExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         NotificationLoader
		expectedResult []contract.Notification
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockNotificiationLoaderSenderStringArg("GetNotificationBySender", nil, SuccessfulDatabaseResult, Sender, Limit),
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Unsuccessful database call",
			mockDb:         createMockNotificiationLoaderSenderStringArg("GetNotificationBySender", Error, []contract.Notification{}, Sender, Limit),
			expectedResult: []contract.Notification{},
			expectedError:  true,
		},
		{
			name:           "Unsuccessful database call",
			mockDb:         createMockNotificiationLoaderSenderStringArg("GetNotificationBySender", ErrorNotFound, []contract.Notification{}, Sender, Limit),
			expectedResult: []contract.Notification{},
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewSenderExecutor(test.mockDb, Sender, Limit)
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
