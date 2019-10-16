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

func TestUpdateExecutor(t *testing.T) {

	tests := []struct {
		name             string
		mockDb           SubscriptionUpdater
		subscription     contract.Subscription
		expectedResult   string
		expectedError    bool
		expectedErrorVal error
	}{

		{
			name:             "Successful database call",
			mockDb:           createUpdateMockSubscriptionSuccess(),
			subscription:     ValidSubscription,
			expectedResult:   ValidSubscription.ID,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name:             "GetSubscriptionBySlug Error",
			mockDb:           createUpdateMockSubscriptionGetErr(),
			subscription:     ValidSubscription,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			name:             "Error",
			mockDb:           createUpdateMockSubscriptionErr(),
			subscription:     ValidSubscription,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewUpdateExecutor(test.mockDb, test.subscription)
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

func createUpdateMockSubscriptionSuccess() SubscriptionUpdater {
	dbMock := mocks.SubscriptionUpdater{}
	dbMock.On("GetSubscriptionBySlug", ValidSubscription.Slug).Return(ValidSubscription, nil)
	dbMock.On("UpdateSubscription", ValidSubscription).Return(nil)
	return &dbMock
}

func createUpdateMockSubscriptionGetErr() SubscriptionUpdater {
	dbMock := mocks.SubscriptionUpdater{}
	dbMock.On("GetSubscriptionBySlug", ValidSubscription.Slug).Return(ValidSubscription, Error)
	dbMock.On("UpdateSubscription", ValidSubscription).Return(Error)
	return &dbMock
}

func createUpdateMockSubscriptionErr() SubscriptionUpdater {
	dbMock := mocks.SubscriptionUpdater{}
	dbMock.On("GetSubscriptionBySlug", ValidSubscription.Slug).Return(ValidSubscription, nil)
	dbMock.On("UpdateSubscription", ValidSubscription).Return(Error)
	return &dbMock
}
