//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// NotificationClient defines the interface for interactions with the Notification endpoint on the EdgeX Foundry support-notifications service.
type NotificationClient interface {
	// SendNotification sends new notifications.
	SendNotification(ctx context.Context, reqs []requests.AddNotificationRequest) ([]common.BaseWithIdResponse, errors.EdgeX)
	// NotificationById query notification by id.
	NotificationById(ctx context.Context, id string) (responses.NotificationResponse, errors.EdgeX)
	// DeleteNotificationById deletes a notification by id.
	DeleteNotificationById(ctx context.Context, id string) (common.BaseResponse, errors.EdgeX)
	// NotificationsByCategory queries notifications with category, offset, ack and limit
	NotificationsByCategory(ctx context.Context, category string, offset int, limit int, ack string) (responses.MultiNotificationsResponse, errors.EdgeX)
	// NotificationsByLabel queries notifications with label, offset, ack and limit
	NotificationsByLabel(ctx context.Context, label string, offset int, limit int, ack string) (responses.MultiNotificationsResponse, errors.EdgeX)
	// NotificationsByStatus queries notifications with status, offset, ack and limit
	NotificationsByStatus(ctx context.Context, status string, offset int, limit int, ack string) (responses.MultiNotificationsResponse, errors.EdgeX)
	// NotificationsByTimeRange query notifications with time range, offset, ack and limit
	NotificationsByTimeRange(ctx context.Context, start, end int64, offset int, limit int, ack string) (responses.MultiNotificationsResponse, errors.EdgeX)
	// NotificationsBySubscriptionName query notifications with subscriptionName, offset, ack and limit
	NotificationsBySubscriptionName(ctx context.Context, subscriptionName string, offset int, limit int, ack string) (responses.MultiNotificationsResponse, errors.EdgeX)
	// CleanupNotificationsByAge removes notifications that are older than age. And the corresponding transmissions will also be deleted.
	// Age is supposed in milliseconds since modified timestamp
	CleanupNotificationsByAge(ctx context.Context, age int) (common.BaseResponse, errors.EdgeX)
	// CleanupNotifications removes notifications and the corresponding transmissions.
	CleanupNotifications(ctx context.Context) (common.BaseResponse, errors.EdgeX)
	// DeleteProcessedNotificationsByAge removes processed notifications that are older than age. And the corresponding transmissions will also be deleted.
	// Age is supposed in milliseconds since modified timestamp
	// Please notice that this API is only for processed notifications (status = PROCESSED). If the deletion purpose includes each kind of notifications, please refer to cleanup API.
	DeleteProcessedNotificationsByAge(ctx context.Context, age int) (common.BaseResponse, errors.EdgeX)
	// NotificationsByQueryConditions queries notifications with offset, limit, acknowledgement status, category and time range
	NotificationsByQueryConditions(ctx context.Context, offset, limit int, ack string, conditionReq requests.GetNotificationRequest) (responses.MultiNotificationsResponse, errors.EdgeX)
	// DeleteNotificationByIds deletes notifications by ids
	DeleteNotificationByIds(ctx context.Context, ids []string) (common.BaseResponse, errors.EdgeX)
	// UpdateNotificationAckStatusByIds updates existing notification's acknowledgement status
	UpdateNotificationAckStatusByIds(ctx context.Context, ack bool, ids []string) (common.BaseResponse, errors.EdgeX)
}
