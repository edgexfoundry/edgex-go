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
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type UpdateExecutor interface {
	Execute() error
}

type subscriptionUpdate struct {
	database     SubscriptionUpdater
	subscription models.Subscription
}

func (op subscriptionUpdate) Execute() error {
	// Check if the subscription exists
	s2, err := op.database.GetSubscriptionBySlug(op.subscription.Slug)
	if err != nil {
		return err
	} else {
		op.subscription.ID = s2.ID
	}

	if err = op.database.UpdateSubscription(op.subscription); err != nil {
		return err
	}
	return nil
}

func NewUpdateExecutor(database SubscriptionUpdater, subscription models.Subscription) UpdateExecutor {
	return subscriptionUpdate{
		database:     database,
		subscription: subscription,
	}
}
