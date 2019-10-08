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

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

// SubscriptionLoader provides functionality for obtaining Subscriptions.
type SubscriptionLoader interface {
	GetSubscriptions() ([]contract.Subscription, error)
	GetSubscriptionById(id string) (contract.Subscription, error)
	GetSubscriptionBySlug(slug string) (contract.Subscription, error)
	GetSubscriptionByCategories(categories []string) ([]contract.Subscription, error)
}

// SubscriptionWriter adds subscriptions.
type SubscriptionWriter interface {
	AddSubscription(s contract.Subscription) (string, error)
}

// SubscriptionDeleter deletes subscriptions.
type SubscriptionDeleter interface {
	DeleteSubscriptionById(id string) error
	DeleteSubscriptionBySlug(id string) error
}

// SubscriptionUpdater updates subscriptions.
type SubscriptionUpdater interface {
	UpdateSubscription(s contract.Subscription) error
	SubscriptionLoader
}
