//
// Copyright (C) 2021-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/google/uuid"
)

var asyncPurgeNotificationOnce sync.Once

// The AddNotification function accepts the new Notification model from the controller function
// and then invokes AddNotification function of infrastructure layer to add new Notification
func AddNotification(n models.Notification, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	addedNotification, edgeXerr := dbClient.AddNotification(n)
	if edgeXerr != nil {
		return "", errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	lc.Debugf("Notification created on DB successfully. Notification ID: %s, Correlation-ID: %s ",
		addedNotification.Id,
		correlation.FromContext(ctx))

	go distribute(dic, addedNotification) // nolint:errcheck

	return addedNotification.Id, nil
}

// NotificationsByCategory queries notifications with offset, limit, ack, and category
func NotificationsByCategory(offset, limit int, ack string, category string, dic *di.Container) (notifications []dtos.Notification, totalCount int64, err errors.EdgeX) {
	if category == "" {
		return notifications, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "category is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	totalCount, err = dbClient.NotificationCountByCategory(category, ack)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.Notification{}, totalCount, err
	}

	notificationModels, err := dbClient.NotificationsByCategory(offset, limit, ack, category)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return dtos.FromNotificationModelsToDTOs(notificationModels), totalCount, nil
}

// NotificationsByLabel queries notifications with offset, limit, ack and label
func NotificationsByLabel(offset, limit int, ack string, label string, dic *di.Container) (notifications []dtos.Notification, totalCount int64, err errors.EdgeX) {
	if label == "" {
		return notifications, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "label is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	totalCount, err = dbClient.NotificationCountByLabel(label, ack)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.Notification{}, totalCount, err
	}

	notificationModels, err := dbClient.NotificationsByLabel(offset, limit, ack, label)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return dtos.FromNotificationModelsToDTOs(notificationModels), totalCount, nil
}

// NotificationById queries notification by ID
func NotificationById(id string, dic *di.Container) (notification dtos.Notification, edgeXerr errors.EdgeX) {
	if id == "" {
		return notification, errors.NewCommonEdgeX(errors.KindContractInvalid, "ID is empty", nil)
	}
	if _, err := uuid.Parse(id); err != nil {
		return notification, errors.NewCommonEdgeX(errors.KindContractInvalid, "ID is not a valid UUID", err)
	}

	dbClient := container.DBClientFrom(dic.Get)
	notificationModel, edgeXerr := dbClient.NotificationById(id)
	if edgeXerr != nil {
		return notification, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	notification = dtos.FromNotificationModelToDTO(notificationModel)
	return notification, nil
}

// NotificationsByStatus queries notifications with offset, limit, ack and status
func NotificationsByStatus(offset, limit int, status, ack string, dic *di.Container) (notifications []dtos.Notification, totalCount int64, err errors.EdgeX) {
	if status == "" {
		return notifications, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "status is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	totalCount, err = dbClient.NotificationCountByStatus(status, ack)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.Notification{}, totalCount, err
	}

	notificationModels, err := dbClient.NotificationsByStatus(offset, limit, ack, status)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return dtos.FromNotificationModelsToDTOs(notificationModels), totalCount, nil
}

// NotificationsByTimeRange query notifications with offset, limit and time range
func NotificationsByTimeRange(start int64, end int64, offset int, limit int, ack string, dic *di.Container) (notifications []dtos.Notification, totalCount int64, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	totalCount, err = dbClient.NotificationCountByTimeRange(start, end, ack)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.Notification{}, totalCount, err
	}

	notificationModels, err := dbClient.NotificationsByTimeRange(start, end, offset, limit, ack)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return dtos.FromNotificationModelsToDTOs(notificationModels), totalCount, nil
}

// DeleteNotificationById deletes the notification by id and all of its associated transmissions
func DeleteNotificationById(id string, dic *di.Container) errors.EdgeX {
	if id == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "id is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	_, err := dbClient.NotificationById(id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = dbClient.DeleteNotificationById(id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// DeleteNotificationByIds deletes the notifications by ids and all of their associated transmissions
func DeleteNotificationByIds(ids []string, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	err := dbClient.DeleteNotificationByIds(ids)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// NotificationsBySubscriptionName queries notifications by offset, limit and subscriptionName
func NotificationsBySubscriptionName(offset, limit int, subscriptionName, ack string, dic *di.Container) (notifications []dtos.Notification, totalCount int64, err errors.EdgeX) {
	if subscriptionName == "" {
		return notifications, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "subscriptionName is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	subscription, err := dbClient.SubscriptionByName(subscriptionName)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	totalCount, err = dbClient.NotificationCountByCategoriesAndLabels(subscription.Categories, subscription.Labels, ack)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.Notification{}, totalCount, err
	}

	notificationModels, err := dbClient.NotificationsByCategoriesAndLabels(offset, limit, subscription.Categories, subscription.Labels, ack)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return dtos.FromNotificationModelsToDTOs(notificationModels), totalCount, nil
}

// CleanupNotificationsByAge invokes the infrastructure layer function to remove notifications that are older than age. And the corresponding transmissions will also be deleted
// Age is supposed in milliseconds since modified timestamp.
func CleanupNotificationsByAge(age int64, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)

	err := dbClient.CleanupNotificationsByAge(age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// DeleteProcessedNotificationsByAge invokes the infrastructure layer function to remove processed notifications that are older than age. And the corresponding transmissions will also be deleted
// Age is supposed in milliseconds since modified timestamp.
func DeleteProcessedNotificationsByAge(age int64, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)

	err := dbClient.DeleteProcessedNotificationsByAge(age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// NotificationByQueryConditions queries notifications with offset, limit, ack, categories, and time range
func NotificationByQueryConditions(offset, limit int, ack string, conditions requests.NotificationQueryCondition,
	dic *di.Container) (notifications []dtos.Notification, totalCount int64, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	totalCount, err = dbClient.NotificationCountByQueryConditions(conditions, ack)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.Notification{}, totalCount, err
	}

	notificationModels, err := dbClient.NotificationsByQueryConditions(offset, limit, conditions, ack)
	if err != nil {
		return notifications, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return dtos.FromNotificationModelsToDTOs(notificationModels), totalCount, nil
}

func UpdateNotificationAckStatus(ack bool, ids []string, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	err := dbClient.UpdateNotificationAckStatusByIds(ack, ids)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// AsyncPurgeNotification purge notifications and related transmissions according to the retention capability.
func AsyncPurgeNotification(interval time.Duration, ctx context.Context, dic *di.Container) {
	asyncPurgeNotificationOnce.Do(func() {
		go func() {
			lc := bootstrapContainer.LoggingClientFrom(dic.Get)
			timer := time.NewTimer(interval)
			for {
				timer.Reset(interval)
				select {
				case <-ctx.Done():
					lc.Info("Exiting notification retention")
					return
				case <-timer.C:
					err := purgeNotification(dic)
					if err != nil {
						lc.Errorf("failed to purge notifications and transmissions, %v", err)
					}
				}
			}
		}()
	})
}

func purgeNotification(dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	total, err := dbClient.NotificationTotalCount()
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "failed to query notification total count, %v", err)
	}
	if total >= int64(config.Retention.MaxCap) {
		lc.Debugf("Purging the notification amount %d to the minimum capacity %d", total, config.Retention.MinCap)
		// Query the latest notification and clean notifications by modified date.
		notification, err := dbClient.LatestNotificationByOffset(config.Retention.MinCap)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query notification with offset '%d'", config.Retention.MinCap), err)
		}
		now := time.Now().UnixMilli()
		age := now - notification.Modified
		err = dbClient.CleanupNotificationsByAge(age)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to delete notifications and transmissions by age '%d'", age), err)
		}
	}
	return nil
}
