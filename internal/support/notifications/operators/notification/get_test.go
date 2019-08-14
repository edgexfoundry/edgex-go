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
var Start int64 = 1564758450
var End int64 = 1564758650
var Labels = []string{
	"test_label",
	"test_label2",
}
var Error = errors.New("test error")
var ErrorNotFound = db.ErrNotFound
var SuccessfulDatabaseResult = []contract.Notification{
	{
		Slug:     "notice-test-123",
		Sender:   "System Management",
		Category: "SECURITY",
		Severity: "CRITICAL",
		Content:  "Hello, Notification!",
		Labels:   Labels,
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
			name:           "Notification not found",
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
			name:           "Notification not found",
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

func createMockNotificiationLoaderStartStringArg(methodName string, err error, ret interface{}, start int64, limit int) NotificationLoader {
	dbMock := mocks.NotificationLoader{}
	dbMock.On(methodName, start, limit).Return(ret, err)
	return &dbMock
}

func createMockNotificiationLoaderStartEndStringArg(methodName string, err error, ret interface{}, start int64, end int64, limit int) NotificationLoader {
	dbMock := mocks.NotificationLoader{}
	dbMock.On(methodName, start, end, limit).Return(ret, err)
	return &dbMock
}

func createMockNotificiationLoaderLabelsStringArg(methodName string, err error, ret interface{}, labels []string, limit int) NotificationLoader {
	dbMock := mocks.NotificationLoader{}
	dbMock.On(methodName, labels, limit).Return(ret, err)
	return &dbMock
}

func TestSenderExecutor(t *testing.T) {
	tests := []struct {
		name            string
		mockDb          NotificationLoader
		expectedResult  []contract.Notification
		expectedError   bool
		expectedErrType error
	}{
		{
			name:            "Successful database call",
			mockDb:          createMockNotificiationLoaderSenderStringArg("GetNotificationBySender", nil, SuccessfulDatabaseResult, Sender, Limit),
			expectedResult:  SuccessfulDatabaseResult,
			expectedError:   false,
			expectedErrType: nil,
		},
		{
			name:            "Unsuccessful database call",
			mockDb:          createMockNotificiationLoaderSenderStringArg("GetNotificationBySender", Error, []contract.Notification{}, Sender, Limit),
			expectedResult:  []contract.Notification{},
			expectedError:   true,
			expectedErrType: Error,
		},
		{
			name:            "Notification not found",
			mockDb:          createMockNotificiationLoaderSenderStringArg("GetNotificationBySender", nil, []contract.Notification{}, Sender, Limit),
			expectedResult:  []contract.Notification{},
			expectedError:   true,
			expectedErrType: ErrorNotFound,
		},
		{
			name:            "Unknown Error",
			mockDb:          createMockNotificiationLoaderSenderStringArg("GetNotificationBySender", Error, SuccessfulDatabaseResult, Sender, Limit),
			expectedResult:  SuccessfulDatabaseResult,
			expectedError:   true,
			expectedErrType: Error,
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

			if !reflect.DeepEqual(test.expectedErrType, err) {
				t.Errorf("Expected error result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedErrType, err)
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestStartExecutor(t *testing.T) {
	tests := []struct {
		name            string
		mockDb          NotificationLoader
		expectedResult  []contract.Notification
		expectedError   bool
		expectedErrType error
	}{
		{
			name:            "Successful database call",
			mockDb:          createMockNotificiationLoaderStartStringArg("GetNotificationsByStart", nil, SuccessfulDatabaseResult, Start, Limit),
			expectedResult:  SuccessfulDatabaseResult,
			expectedError:   false,
			expectedErrType: nil,
		},
		{
			name:            "Unsuccessful database call",
			mockDb:          createMockNotificiationLoaderStartStringArg("GetNotificationsByStart", Error, []contract.Notification{}, Start, Limit),
			expectedResult:  []contract.Notification{},
			expectedError:   true,
			expectedErrType: Error,
		},
		{
			name:            "Notification not found",
			mockDb:          createMockNotificiationLoaderStartStringArg("GetNotificationsByStart", nil, []contract.Notification{}, Start, Limit),
			expectedResult:  []contract.Notification{},
			expectedError:   true,
			expectedErrType: ErrorNotFound,
		},
		{
			name:            "Unknown Error",
			mockDb:          createMockNotificiationLoaderStartStringArg("GetNotificationsByStart", Error, SuccessfulDatabaseResult, Start, Limit),
			expectedResult:  SuccessfulDatabaseResult,
			expectedError:   true,
			expectedErrType: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewStartExecutor(test.mockDb, Start, Limit)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedErrType, err) {
				t.Errorf("Expected error result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedErrType, err)
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestEndExecutor(t *testing.T) {
	tests := []struct {
		name            string
		mockDb          NotificationLoader
		expectedResult  []contract.Notification
		expectedError   bool
		expectedErrType error
	}{
		{
			name:            "Successful database call",
			mockDb:          createMockNotificiationLoaderStartStringArg("GetNotificationsByEnd", nil, SuccessfulDatabaseResult, End, Limit),
			expectedResult:  SuccessfulDatabaseResult,
			expectedError:   false,
			expectedErrType: nil,
		},
		{
			name:            "Unsuccessful database call",
			mockDb:          createMockNotificiationLoaderStartStringArg("GetNotificationsByEnd", Error, []contract.Notification{}, End, Limit),
			expectedResult:  []contract.Notification{},
			expectedError:   true,
			expectedErrType: Error,
		},
		{
			name:            "Notification not found",
			mockDb:          createMockNotificiationLoaderStartStringArg("GetNotificationsByEnd", nil, []contract.Notification{}, End, Limit),
			expectedResult:  []contract.Notification{},
			expectedError:   true,
			expectedErrType: ErrorNotFound,
		},
		{
			name:            "Unknown Error",
			mockDb:          createMockNotificiationLoaderStartStringArg("GetNotificationsByEnd", Error, SuccessfulDatabaseResult, End, Limit),
			expectedResult:  SuccessfulDatabaseResult,
			expectedError:   true,
			expectedErrType: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewEndExecutor(test.mockDb, End, Limit)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedErrType, err) {
				t.Errorf("Expected error result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedErrType, err)
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestStartEndExecutor(t *testing.T) {
	tests := []struct {
		name            string
		mockDb          NotificationLoader
		expectedResult  []contract.Notification
		expectedError   bool
		expectedErrType error
	}{
		{
			name:            "Successful database call",
			mockDb:          createMockNotificiationLoaderStartEndStringArg("GetNotificationsByStartEnd", nil, SuccessfulDatabaseResult, Start, End, Limit),
			expectedResult:  SuccessfulDatabaseResult,
			expectedError:   false,
			expectedErrType: nil,
		},
		{
			name:            "Unsuccessful database call",
			mockDb:          createMockNotificiationLoaderStartEndStringArg("GetNotificationsByStartEnd", Error, []contract.Notification{}, Start, End, Limit),
			expectedResult:  []contract.Notification{},
			expectedError:   true,
			expectedErrType: Error,
		},
		{
			name:            "Notification not found",
			mockDb:          createMockNotificiationLoaderStartEndStringArg("GetNotificationsByStartEnd", nil, []contract.Notification{}, Start, End, Limit),
			expectedResult:  []contract.Notification{},
			expectedError:   true,
			expectedErrType: ErrorNotFound,
		},
		{
			name:            "Unknown Error",
			mockDb:          createMockNotificiationLoaderStartEndStringArg("GetNotificationsByStartEnd", Error, SuccessfulDatabaseResult, Start, End, Limit),
			expectedResult:  SuccessfulDatabaseResult,
			expectedError:   true,
			expectedErrType: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewStartEndExecutor(test.mockDb, Start, End, Limit)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedErrType, err) {
				t.Errorf("Expected error result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedErrType, err)
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestLabelsExecutor(t *testing.T) {
	tests := []struct {
		name            string
		mockDb          NotificationLoader
		expectedResult  []contract.Notification
		expectedError   bool
		expectedErrType error
	}{
		{
			name:            "Successful database call",
			mockDb:          createMockNotificiationLoaderLabelsStringArg("GetNotificationsByLabels", nil, SuccessfulDatabaseResult, Labels, Limit),
			expectedResult:  SuccessfulDatabaseResult,
			expectedError:   false,
			expectedErrType: nil,
		},
		{
			name:            "Unsuccessful database call",
			mockDb:          createMockNotificiationLoaderLabelsStringArg("GetNotificationsByLabels", Error, []contract.Notification{}, Labels, Limit),
			expectedResult:  []contract.Notification{},
			expectedError:   true,
			expectedErrType: Error,
		},
		{
			name:            "Notification not found",
			mockDb:          createMockNotificiationLoaderLabelsStringArg("GetNotificationsByLabels", nil, []contract.Notification{}, Labels, Limit),
			expectedResult:  []contract.Notification{},
			expectedError:   true,
			expectedErrType: ErrorNotFound,
		},
		{
			name:            "Unknown Error",
			mockDb:          createMockNotificiationLoaderLabelsStringArg("GetNotificationsByLabels", Error, SuccessfulDatabaseResult, Labels, Limit),
			expectedResult:  SuccessfulDatabaseResult,
			expectedError:   true,
			expectedErrType: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewLabelsExecutor(test.mockDb, Labels, Limit)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedErrType, err) {
				t.Errorf("Expected error result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedErrType, err)
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}
