//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type DBClient interface {
	CloseSession()

	AddSubscription(e models.Subscription) (models.Subscription, errors.EdgeX)
	SubscriptionById(id string) (models.Subscription, errors.EdgeX)
	AllSubscriptions(offset int, limit int) ([]models.Subscription, errors.EdgeX)
	SubscriptionByName(name string) (models.Subscription, errors.EdgeX)
	SubscriptionsByCategory(offset, limit int, category string) ([]models.Subscription, errors.EdgeX)
	SubscriptionsByLabel(offset, limit int, label string) ([]models.Subscription, errors.EdgeX)
	SubscriptionsByReceiver(offset, limit int, receiver string) ([]models.Subscription, errors.EdgeX)
	DeleteSubscriptionByName(name string) errors.EdgeX
	UpdateSubscription(s models.Subscription) errors.EdgeX
	SubscriptionsByCategoriesAndLabels(offset, limit int, categories []string, labels []string) ([]models.Subscription, errors.EdgeX)

	AddNotification(n models.Notification) (models.Notification, errors.EdgeX)
	NotificationById(id string) (models.Notification, errors.EdgeX)
	NotificationsByCategory(offset, limit int, category string) ([]models.Notification, errors.EdgeX)
	NotificationsByLabel(offset, limit int, label string) ([]models.Notification, errors.EdgeX)
	NotificationsByStatus(offset, limit int, status string) ([]models.Notification, errors.EdgeX)
	NotificationsByTimeRange(start int, end int, offset int, limit int) ([]models.Notification, errors.EdgeX)
	DeleteNotificationById(id string) errors.EdgeX
	NotificationsByCategoriesAndLabels(offset, limit int, categories []string, labels []string) ([]models.Notification, errors.EdgeX)
	UpdateNotification(s models.Notification) errors.EdgeX
	CleanupNotificationsByAge(age int64) errors.EdgeX
	DeleteProcessedNotificationsByAge(age int64) errors.EdgeX

	AddTransmission(trans models.Transmission) (models.Transmission, errors.EdgeX)
	UpdateTransmission(trans models.Transmission) errors.EdgeX
	TransmissionById(id string) (models.Transmission, errors.EdgeX)
	TransmissionsByTimeRange(start int, end int, offset int, limit int) ([]models.Transmission, errors.EdgeX)
	AllTransmissions(offset int, limit int) ([]models.Transmission, errors.EdgeX)
	TransmissionsByStatus(offset, limit int, status string) ([]models.Transmission, errors.EdgeX)
	DeleteProcessedTransmissionsByAge(age int64) errors.EdgeX
	TransmissionsBySubscriptionName(offset, limit int, subscriptionName string) ([]models.Transmission, errors.EdgeX)
}
