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
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/notifications/operators/subscription/mocks"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var ValidSubscription = SuccessfulDatabaseResult[0]

func TestAddExecutor(t *testing.T) {

	tests := []struct {
		name             string
		mockDb           SubscriptionWriter
		subscription     contract.Subscription
		expectedResult   string
		expectedError    bool
		expectedErrorVal error
	}{

		{
			name:             "Successful database call",
			mockDb:           createAddMockSubscriptionSuccess(),
			subscription:     ValidSubscription,
			expectedResult:   ValidSubscription.ID,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name:             "Error",
			mockDb:           createAddMockSubscriptionErr(),
			subscription:     ValidSubscription,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewAddExecutor(test.mockDb, test.subscription)
			err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
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

func createAddMockSubscriptionSuccess() SubscriptionWriter {
	dbMock := mocks.SubscriptionWriter{}
	dbMock.On("AddSubscription", ValidSubscription).Return(Id, nil)
	return &dbMock
}

func createAddMockSubscriptionErr() SubscriptionWriter {
	dbMock := mocks.SubscriptionWriter{}
	dbMock.On("AddSubscription", ValidSubscription).Return(Id, Error)
	return &dbMock
}
