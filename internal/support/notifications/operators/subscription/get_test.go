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
	"errors"
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/operators/subscription/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var Id = "83cb038b-5a94-4707-985d-13effec62de2"

var Error = errors.New("test error")
var ErrorNotFound = db.ErrNotFound
var SuccessfulDatabaseResult = []contract.Subscription{
	{
		Slug:     "notice-test-123",
		Receiver: "System Admin",
		SubscribedCategories: []contract.NotificationsCategory{
			"SECURITY",
			"HW_HEALTH",
			"SW_HEALTH",
		},
		SubscribedLabels: []string{
			"Dell",
			"IoT",
			"test",
		},
		Channels: []contract.Channel{
			{
				Type: "REST",
				Url:  "http://abc.def/alert",
			},
			{
				Type: "EMAIL",
				MailAddresses: []string{
					"cloud@abc.def",
					"jack@abc.def",
				},
			},
		},
	},
}

func TestIdExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         SubscriptionLoader
		expectedResult contract.Subscription
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockSubscriptionLoaderStringArg("GetSubscriptionById", nil, SuccessfulDatabaseResult[0], Id),
			expectedResult: SuccessfulDatabaseResult[0],
			expectedError:  false,
		},
		{
			name:           "Unsuccessful database call",
			mockDb:         createMockSubscriptionLoaderStringArg("GetSubscriptionById", Error, contract.Subscription{}, Id),
			expectedResult: contract.Subscription{},
			expectedError:  true,
		},
		{
			name:           "Subscription not found",
			mockDb:         createMockSubscriptionLoaderStringArg("GetSubscriptionById", ErrorNotFound, contract.Subscription{}, Id),
			expectedResult: contract.Subscription{},
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

func createMockSubscriptionLoaderStringArg(methodName string, err error, ret interface{}, arg string) SubscriptionLoader {
	dbMock := mocks.SubscriptionLoader{}
	dbMock.On(methodName, arg).Return(ret, err)
	return &dbMock
}
