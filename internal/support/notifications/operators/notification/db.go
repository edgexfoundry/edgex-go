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

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

// NotificationLoader provides functionality for obtaining Notifications.
type NotificationLoader interface {
	GetNotificationById(id string) (contract.Notification, error)
	GetNotificationBySlug(slug string) (contract.Notification, error)
	GetNotificationBySender(sender string, limit int) ([]contract.Notification, error)
	GetNotificationsByStartEnd(start int64, end int64, limit int) ([]contract.Notification, error)
	GetNotificationsByStart(start int64, limit int) ([]contract.Notification, error)
	GetNotificationsByEnd(end int64, limit int) ([]contract.Notification, error)
	GetNotificationsByLabels(labels []string, limit int) ([]contract.Notification, error)
	GetNewNotifications(limit int) ([]contract.Notification, error)
}

// NotificationWriter provides functionality for adding Notifications.
type NotificationWriter interface {
	AddNotification(n contract.Notification) (string, error)
}

// NotificationDeleter deletes notifications.
type NotificationDeleter interface {
	DeleteNotificationById(id string) error
	DeleteNotificationBySlug(slug string) error
	DeleteNotificationsOld(age int) error
}
