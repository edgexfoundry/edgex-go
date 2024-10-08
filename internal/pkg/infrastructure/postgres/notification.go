//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
)

// AddNotification adds a new notification to the database
func (c *Client) AddNotification(n models.Notification) (models.Notification, errors.EdgeX) {
	ctx := context.Background()
	if len(n.Id) == 0 {
		n.Id = uuid.New().String()
	} else {
		exists, edgexErr := checkNotificationExists(ctx, c.ConnPool, n.Id)
		if edgexErr != nil {
			return n, errors.NewCommonEdgeXWrapper(edgexErr)
		}
		if exists {
			return n, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("notification id %s already exists", n.Id), nil)
		}
	}

	timestamp := time.Now().UTC().UnixMilli()
	n.Created = timestamp
	n.Modified = timestamp
	dataBytes, err := json.Marshal(n)
	if err != nil {
		return n, errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal Notification model", err)
	}

	_, err = c.ConnPool.Exec(ctx, sqlInsert(notificationTableName, idCol, contentCol), n.Id, dataBytes)
	if err != nil {
		return n, pgClient.WrapDBError("failed to insert row to notification table", err)
	}

	return n, nil
}

// NotificationById queries the notification by id
func (c *Client) NotificationById(id string) (models.Notification, errors.EdgeX) {
	notification, err := queryNotification(context.Background(), c.ConnPool, sqlQueryContentById(notificationTableName), id)
	if err != nil {
		return notification, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query notification by id %s", id), err)
	}

	return notification, nil
}

// NotificationsByCategory queries the notification by category
func (c *Client) NotificationsByCategory(offset, limit int, category string) ([]models.Notification, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{categoryField: category}

	notifications, err := queryNotifications(context.Background(), c.ConnPool, sqlQueryContentByJSONFieldWithPagination(notificationTableName), queryObj, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query all notifications by category %s", category), err)
	}

	return notifications, nil
}

// NotificationsByLabel queries the notification by label
func (c *Client) NotificationsByLabel(offset, limit int, label string) ([]models.Notification, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{labelsField: []string{label}}

	notifications, err := queryNotifications(context.Background(), c.ConnPool, sqlQueryContentByJSONFieldWithPagination(notificationTableName), queryObj, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query all notifications by label %s", label), err)
	}

	return notifications, nil
}

// NotificationsByStatus queries the notification by status
func (c *Client) NotificationsByStatus(offset, limit int, status string) ([]models.Notification, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{statusField: status}

	notifications, err := queryNotifications(context.Background(), c.ConnPool, sqlQueryContentByJSONFieldWithPagination(notificationTableName), queryObj, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query all notifications by status %s", status), err)
	}

	return notifications, nil
}

// NotificationsByTimeRange queries the notification by time range
func (c *Client) NotificationsByTimeRange(start int64, end int64, offset, limit int) ([]models.Notification, errors.EdgeX) {
	validStart, validEnd, offset, validLimit, err := getValidRangeParameters(int64(start), int64(end), offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	notifications, err := queryNotifications(context.Background(), c.ConnPool, sqlQueryContentWithTimeRangeAndPagination(notificationTableName), validStart, validEnd, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all notifications by time range", err)
	}

	return notifications, nil
}

// DeleteNotificationById deletes the notification by id
func (c *Client) DeleteNotificationById(id string) errors.EdgeX {
	_, err := c.ConnPool.Exec(context.Background(), sqlDeleteById(notificationTableName), id)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete notification by id %s", id), err)
	}

	return nil
}

// NotificationsByCategoriesAndLabels queries the notification by categories and labels
func (c *Client) NotificationsByCategoriesAndLabels(offset, limit int, categories []string, labels []string) ([]models.Notification, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{
		categoryField: categories,
		labelsField:   labels,
	}

	notifications, err := queryNotifications(context.Background(), c.ConnPool, sqlQueryContentByJSONFieldWithPagination(notificationTableName), queryObj, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all notifications by categories and labels", err)
	}

	return notifications, nil
}

// UpdateNotification updates the notification
func (c *Client) UpdateNotification(n models.Notification) errors.EdgeX {
	modified := time.Now().UTC().UnixMilli()
	n.Modified = modified

	dataBytes, err := json.Marshal(n)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal Notification model", err)
	}

	_, err = c.ConnPool.Exec(context.Background(), sqlUpdateContentById(notificationTableName), dataBytes, n.Id)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to update row by notification id '%s' from notification table", n.Id), err)
	}

	return nil
}

// CleanupNotificationsByAge deletes the notifications that are older than a specific age
func (c *Client) CleanupNotificationsByAge(age int64) errors.EdgeX {
	_, err := c.ConnPool.Exec(context.Background(), sqlDeleteByContentAge(notificationTableName), age)
	if err != nil {
		return pgClient.WrapDBError("failed to cleanup notifications by age", err)
	}

	return nil
}

// DeleteProcessedNotificationsByAge deletes the processed notifications that are older than a specific age
func (c *Client) DeleteProcessedNotificationsByAge(age int64) errors.EdgeX {
	queryObj := map[string]any{statusField: models.Processed}
	_, err := c.ConnPool.Exec(context.Background(), sqlDeleteByJSONFieldAndAge(notificationTableName), queryObj, age)
	if err != nil {
		return pgClient.WrapDBError("failed to delete processed notifications by age", err)
	}

	return nil
}

// NotificationCountByCategory returns the count of notifications by category
func (c *Client) NotificationCountByCategory(category string) (uint32, errors.EdgeX) {
	queryObj := map[string]any{categoryField: category}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONField(notificationTableName), queryObj)
}

// NotificationCountByLabel returns the count of notifications by label
func (c *Client) NotificationCountByLabel(label string) (uint32, errors.EdgeX) {
	queryObj := map[string]any{labelsField: []string{label}}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONField(notificationTableName), queryObj)
}

// NotificationCountByStatus returns the count of notifications by status
func (c *Client) NotificationCountByStatus(status string) (uint32, errors.EdgeX) {
	queryObj := map[string]any{statusField: status}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONField(notificationTableName), queryObj)
}

// NotificationCountByTimeRange returns the count of notifications by time range
func (c *Client) NotificationCountByTimeRange(start int64, end int64) (uint32, errors.EdgeX) {
	validStart, validEnd, err := getValidStartAndEnd(start, end)
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByTimeRange(notificationTableName), validStart, validEnd)
}

// NotificationCountByCategoriesAndLabels returns the count of notifications by categories and labels
func (c *Client) NotificationCountByCategoriesAndLabels(categories []string, labels []string) (uint32, errors.EdgeX) {
	queryObj := map[string]any{categoryField: categories, labelsField: labels}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONField(notificationTableName), queryObj)
}

// NotificationTotalCount returns the total count of notifications
func (c *Client) NotificationTotalCount() (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCount(notificationTableName))
}

// LatestNotificationByOffset returns the latest notification by offset
func (c *Client) LatestNotificationByOffset(offset uint32) (models.Notification, errors.EdgeX) {
	notification, err := queryNotification(context.Background(), c.ConnPool, sqlQueryContentWithPagination(notificationTableName), offset, 1)
	if err != nil {
		return notification, errors.NewCommonEdgeX(errors.Kind(err), "failed to query latest notification by offset", err)
	}

	return notification, nil
}

func queryNotification(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (models.Notification, errors.EdgeX) {
	var notification models.Notification
	row := connPool.QueryRow(ctx, sql, args...)

	if err := row.Scan(&notification); err != nil {
		return notification, pgClient.WrapDBError("failed to query notification", err)
	}
	return notification, nil
}

func queryNotifications(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]models.Notification, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query rows from notification table", err)
	}

	notifications, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Notification, error) {
		var n models.Notification
		scanErr := row.Scan(&n)
		return n, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to Notification model", err)
	}

	return notifications, nil
}

func checkNotificationExists(ctx context.Context, connPool *pgxpool.Pool, id string) (bool, errors.EdgeX) {
	var exists bool
	err := connPool.QueryRow(ctx, sqlCheckExistsById(notificationTableName), id).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError(fmt.Sprintf("failed to query row by id '%s' from notification table", id), err)
	}
	return exists, nil
}
