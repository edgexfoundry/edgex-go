/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
 *
 *******************************************************************************/

package interfaces

import (
	"errors"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

type DatabaseType int8 // Database type enum
const (
	MONGO DatabaseType = iota
)

type DBClient interface {
	CloseSession()

	// Notifications
	GetNotifications() ([]contract.Notification, error)
	GetNotificationById(id string) (contract.Notification, error)
	GetNotificationBySlug(slug string) (contract.Notification, error)
	GetNotificationBySender(sender string, limit int) ([]contract.Notification, error)
	GetNotificationsByLabels(labels []string, limit int) ([]contract.Notification, error)
	GetNotificationsByStartEnd(start int64, end int64, limit int) ([]contract.Notification, error)
	GetNotificationsByStart(start int64, limit int) ([]contract.Notification, error)
	GetNotificationsByEnd(end int64, limit int) ([]contract.Notification, error)
	GetNewNotifications(limit int) ([]contract.Notification, error)
	GetNewNormalNotifications(limit int) ([]contract.Notification, error)
	AddNotification(n contract.Notification) (string, error)
	UpdateNotification(n contract.Notification) error
	MarkNotificationProcessed(n contract.Notification) error
	DeleteNotificationById(id string) error
	DeleteNotificationBySlug(id string) error
	DeleteNotificationsOld(age int) error

	// Subscriptions
	GetSubscriptions() ([]contract.Subscription, error)
	GetSubscriptionById(id string) (contract.Subscription, error)
	GetSubscriptionBySlug(slug string) (contract.Subscription, error)
	GetSubscriptionByReceiver(receiver string) ([]contract.Subscription, error)
	GetSubscriptionByCategories(categories []string) ([]contract.Subscription, error)
	GetSubscriptionByLabels(labels []string) ([]contract.Subscription, error)
	GetSubscriptionByCategoriesLabels(categories []string, labels []string) ([]contract.Subscription, error)
	AddSubscription(s contract.Subscription) (string, error)
	UpdateSubscription(s contract.Subscription) error
	DeleteSubscriptionBySlug(id string) error

	// Transmissions
	GetTransmissionsByNotificationSlug(slug string, resendLimit int) ([]contract.Transmission, error)
	GetTransmissionsByStartEnd(start int64, end int64, resendLimit int) ([]contract.Transmission, error)
	GetTransmissionsByStart(start int64, resendLimit int) ([]contract.Transmission, error)
	GetTransmissionsByEnd(end int64, resendLimit int) ([]contract.Transmission, error)
	GetTransmissionsByStatus(resendLimit int, status contract.TransmissionStatus) ([]contract.Transmission, error)
	AddTransmission(t contract.Transmission) (string, error)
	UpdateTransmission(t contract.Transmission) error
	DeleteTransmission(age int64, status contract.TransmissionStatus) error

	// General Cleanup
	Cleanup() error
	CleanupOld(age int) error
}

type DBConfiguration struct {
	DbType            DatabaseType
	Host              string
	Port              int
	Timeout           int
	DatabaseName      string
	Username          string
	Password          string
	ReadMax           int
	ResendLimit       int
	CleanupDefaultAge int
}

var ErrNotFound error = errors.New("Item not found")
var ErrUnsupportedDatabase error = errors.New("Unsuppored database type")
var ErrInvalidObjectId error = errors.New("Invalid object ID")
var ErrNotUnique error = errors.New("Resource already exists")
var ErrSlugEmpty error = errors.New("Slug is nil or empty")
