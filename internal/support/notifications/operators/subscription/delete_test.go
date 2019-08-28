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

package subscription

import (
	"reflect"
	"testing"

	notificationErrors "github.com/edgexfoundry/edgex-go/internal/support/notifications/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/operators/subscription/mocks"
)

func TestSubscriptionById(t *testing.T) {
	tests := []struct {
		name              string
		database          SubscriptionDeleter
		expectError       bool
		expectedErrorType error
	}{
		{
			name:              "Successful Delete",
			database:          createMockSubscriptionDeleterStringArg("DeleteSubscriptionById", nil, Id),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Subscription not found",
			database:          createMockSubscriptionDeleterStringArg("DeleteSubscriptionById", ErrorNotFound, Id),
			expectError:       true,
			expectedErrorType: notificationErrors.ErrSubscriptionNotFound{},
		},
		{
			name:              "Delete error",
			database:          createMockSubscriptionDeleterStringArg("DeleteSubscriptionById", Error, Id),
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

func TestSubscriptionBySlug(t *testing.T) {
	tests := []struct {
		name              string
		database          SubscriptionDeleter
		expectError       bool
		expectedErrorType error
	}{
		{
			name:              "Successful Delete",
			database:          createMockSubscriptionDeleterStringArg("DeleteSubscriptionBySlug", nil, Slug),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Subscription not found",
			database:          createMockSubscriptionDeleterStringArg("DeleteSubscriptionBySlug", ErrorNotFound, Slug),
			expectError:       true,
			expectedErrorType: notificationErrors.ErrSubscriptionNotFound{},
		},
		{
			name:              "Delete error",
			database:          createMockSubscriptionDeleterStringArg("DeleteSubscriptionBySlug", Error, Slug),
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

func createMockSubscriptionDeleterStringArg(methodName string, err error, arg string) SubscriptionDeleter {
	dbMock := mocks.SubscriptionDeleter{}
	dbMock.On(methodName, arg).Return(err)
	return &dbMock
}
