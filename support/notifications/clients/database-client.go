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

package clients

import (
	"errors"
	"fmt"

	"github.com/edgexfoundry/edgex-go/support/notifications/models"
	"gopkg.in/mgo.v2/bson"
)

type DatabaseType int8 // Database type enum
const (
	MONGO DatabaseType = iota
)

type DBClient interface {
	Notifications() ([]models.Notification, error)

	NotificationById(id string) (models.Notification, error)

	NotificationBySlug(slug string) (models.Notification, error)

	NotificationBySender(sender string, limit int) ([]models.Notification, error)

	NotificationsByLabels(labels []string, limit int) ([]models.Notification, error)

	NotificationsByStartEnd(start int64, end int64, limit int) ([]models.Notification, error)

	NotificationsByStart(start int64, limit int) ([]models.Notification, error)

	NotificationsByEnd(end int64, limit int) ([]models.Notification, error)

	NotificationsNew(limit int) ([]models.Notification, error)

	NotificationsNewNormal(limit int) ([]models.Notification, error)

	AddNotification(n *models.Notification) (bson.ObjectId, error)

	UpdateNotification(n models.Notification) error

	MarkNotificationProcessed(n models.Notification) error

	DeleteNotificationById(id string) error

	DeleteNotificationBySlug(id string) error

	DeleteNotificationsOld(age int64) error

	SubscriptionById(id string) (models.Subscription, error)

	SubscriptionBySlug(slug string) (models.Subscription, error)

	SubscriptionByReceiver(receiver string) ([]models.Subscription, error)

	SubscriptionByCategories(categories []string) ([]models.Subscription, error)

	SubscriptionByLabels(labels []string) ([]models.Subscription, error)

	SubscriptionByCategoriesLabels(categories []string, labels []string) ([]models.Subscription, error)

	AddSubscription(s *models.Subscription) (bson.ObjectId, error)

	DeleteSubscriptionBySlug(id string) error

	AddTransmission(t *models.Transmission) (bson.ObjectId, error)

	UpdateTransmission(t models.Transmission) error

	DeleteTransmission(age int64, status models.TransmissionStatus) error

	TransmissionsByNotificationSlug(slug string, resendLimit int) ([]models.Transmission, error)

	TransmissionsByStartEnd(start int64, end int64, resendLimit int) ([]models.Transmission, error)

	TransmissionsByStart(start int64, resendLimit int) ([]models.Transmission, error)

	TransmissionsByEnd(end int64, resendLimit int) ([]models.Transmission, error)

	TransmissionsByStatus(resendLimit int, status models.TransmissionStatus) ([]models.Transmission, error)

	Cleanup() error

	CleanupOld(age int64) error
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
	CleanupDefaultAge int64
}

var ErrNotFound error = errors.New("Item not found")
var ErrUnsupportedDatabase error = errors.New("Unsuppored database type")
var ErrInvalidObjectId error = errors.New("Invalid object ID")
var ErrNotUnique error = errors.New("Resource already exists")
var ErrSlugEmpty error = errors.New("Slug is nil or empty")

// Return the dbClient interface
func NewDBClient(config DBConfiguration) (DBClient, error) {
	switch config.DbType {
	case MONGO:
		// Create the mongo client
		mc, err := newMongoClient(config)
		if err != nil {
			fmt.Println("Error creating the mongo client: " + err.Error())
			return nil, err
		}
		return mc, nil
	default:
		return nil, ErrUnsupportedDatabase
	}
}
