//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
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
	SubscriptionTotalCount() (int64, errors.EdgeX)
	SubscriptionCountByCategory(category string) (int64, errors.EdgeX)
	SubscriptionCountByLabel(label string) (int64, errors.EdgeX)
	SubscriptionCountByReceiver(receiver string) (int64, errors.EdgeX)

	AddNotification(n models.Notification) (models.Notification, errors.EdgeX)
	NotificationById(id string) (models.Notification, errors.EdgeX)
	NotificationsByCategory(offset, limit int, ack, category string) ([]models.Notification, errors.EdgeX)
	NotificationsByLabel(offset, limit int, ack, label string) ([]models.Notification, errors.EdgeX)
	NotificationsByStatus(offset, limit int, ack, status string) ([]models.Notification, errors.EdgeX)
	NotificationsByTimeRange(start int64, end int64, offset int, limit int, ack string) ([]models.Notification, errors.EdgeX)
	NotificationsByQueryConditions(offset, limit int, condition requests.NotificationQueryCondition, ack string) ([]models.Notification, errors.EdgeX)
	DeleteNotificationById(id string) errors.EdgeX
	DeleteNotificationByIds(ids []string) errors.EdgeX
	NotificationsByCategoriesAndLabels(offset, limit int, categories []string, labels []string, ack string) ([]models.Notification, errors.EdgeX)
	UpdateNotification(s models.Notification) errors.EdgeX
	UpdateNotificationAckStatusByIds(ack bool, ids []string) errors.EdgeX
	CleanupNotificationsByAge(age int64) errors.EdgeX
	DeleteProcessedNotificationsByAge(age int64) errors.EdgeX
	NotificationCountByCategory(category string, ack string) (int64, errors.EdgeX)
	NotificationCountByLabel(label string, ack string) (int64, errors.EdgeX)
	NotificationCountByStatus(status string, ack string) (int64, errors.EdgeX)
	NotificationCountByTimeRange(start int64, end int64, ack string) (int64, errors.EdgeX)
	NotificationCountByCategoriesAndLabels(categories []string, labels []string, ack string) (int64, errors.EdgeX)
	NotificationCountByQueryConditions(condition requests.NotificationQueryCondition, ack string) (int64, errors.EdgeX)
	NotificationTotalCount() (int64, errors.EdgeX)
	LatestNotificationByOffset(offset uint32) (models.Notification, errors.EdgeX)

	AddTransmission(trans models.Transmission) (models.Transmission, errors.EdgeX)
	UpdateTransmission(trans models.Transmission) errors.EdgeX
	TransmissionById(id string) (models.Transmission, errors.EdgeX)
	TransmissionsByTimeRange(start int64, end int64, offset int, limit int) ([]models.Transmission, errors.EdgeX)
	AllTransmissions(offset int, limit int) ([]models.Transmission, errors.EdgeX)
	TransmissionsByStatus(offset, limit int, status string) ([]models.Transmission, errors.EdgeX)
	DeleteProcessedTransmissionsByAge(age int64) errors.EdgeX
	TransmissionsBySubscriptionName(offset, limit int, subscriptionName string) ([]models.Transmission, errors.EdgeX)
	TransmissionTotalCount() (int64, errors.EdgeX)
	TransmissionCountBySubscriptionName(subscriptionName string) (int64, errors.EdgeX)
	TransmissionCountByStatus(status string) (int64, errors.EdgeX)
	TransmissionCountByTimeRange(start int64, end int64) (int64, errors.EdgeX)
	TransmissionsByNotificationId(offset, limit int, id string) ([]models.Transmission, errors.EdgeX)
	TransmissionCountByNotificationId(id string) (int64, errors.EdgeX)
}
