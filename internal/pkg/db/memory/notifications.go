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

func (c *Client) GetNotifications() ([]contract.Notification, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetNotificationById(id string) (contract.Notification, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetNotificationBySlug(slug string) (contract.Notification, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetNotificationBySender(sender string, limit int) ([]contract.Notification, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetNotificationsByLabels(labels []string, limit int) ([]contract.Notification, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetNotificationsByStartEnd(start int64, end int64, limit int) ([]contract.Notification, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetNotificationsByStart(start int64, limit int) ([]contract.Notification, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetNotificationsByEnd(end int64, limit int) ([]contract.Notification, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetNewNotifications(limit int) ([]contract.Notification, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) GetNewNormalNotifications(limit int) ([]contract.Notification, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) AddNotification(n contract.Notification) (string, error) {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) UpdateNotification(n contract.Notification) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) MarkNotificationProcessed(n contract.Notification) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteNotificationById(id string) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteNotificationBySlug(slug string) error {
	panic(UnimplementedMethodPanicMessage)
}

func (c *Client) DeleteNotificationsOld(age int) error {
	panic(UnimplementedMethodPanicMessage)
}
