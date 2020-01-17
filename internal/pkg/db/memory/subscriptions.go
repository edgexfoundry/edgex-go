/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package memory

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

func (c *Client) GetSubscriptionBySlug(slug string) (contract.Subscription, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetSubscriptionByCategories(categories []string) ([]contract.Subscription, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetSubscriptionByLabels(labels []string) ([]contract.Subscription, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetSubscriptionByCategoriesLabels(categories []string, labels []string) ([]contract.Subscription, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetSubscriptionByReceiver(receiver string) ([]contract.Subscription, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetSubscriptionById(id string) (contract.Subscription, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteSubscriptionById(id string) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) AddSubscription(sub contract.Subscription) (string, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) UpdateSubscription(sub contract.Subscription) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteSubscriptionBySlug(slug string) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetSubscriptions() ([]contract.Subscription, error) {
	panic(UnimplementedMethodPanicMessage)
}
