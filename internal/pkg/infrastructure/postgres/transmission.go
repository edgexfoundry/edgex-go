//
// Copyright (C) 2024-2025 IOTech Ltd
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

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
)

// AddTransmission adds a new transmission to the database
func (c *Client) AddTransmission(t models.Transmission) (models.Transmission, errors.EdgeX) {
	ctx := context.Background()

	if len(t.Id) == 0 {
		t.Id = uuid.New().String()
	} else {
		exists, edgexErr := checkTransmissionExists(ctx, c.ConnPool, t.Id)
		if edgexErr != nil {
			return t, errors.NewCommonEdgeXWrapper(edgexErr)
		}
		if exists {
			return t, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("transmission id %s already exists", t.Id), nil)
		}
	}

	timestamp := time.Now().UTC().UnixMilli()
	t.Created = timestamp
	dataBytes, err := json.Marshal(t)
	if err != nil {
		return t, errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal Transmission model", err)
	}

	_, err = c.ConnPool.Exec(ctx, sqlInsert(transmissionTableName, idCol, notificationIdCol, contentCol), t.Id, t.NotificationId, dataBytes)
	if err != nil {
		return t, pgClient.WrapDBError("failed to insert row to transmission table", err)
	}
	return t, nil
}

// UpdateTransmission updates a transmission in the database
func (c *Client) UpdateTransmission(t models.Transmission) errors.EdgeX {
	dataBytes, err := json.Marshal(t)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal Transmission model", err)
	}

	_, err = c.ConnPool.Exec(context.Background(), sqlUpdateContentById(transmissionTableName), dataBytes, t.Id)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to update row by transmission id '%s' from transmission table", t.Id), err)
	}

	return nil
}

// TransmissionById queries the transmission by id
func (c *Client) TransmissionById(id string) (models.Transmission, errors.EdgeX) {
	transmission, err := queryTransmission(context.Background(), c.ConnPool, sqlQueryContentById(transmissionTableName), id)
	if err != nil {
		return transmission, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query transmission by id %s", id), err)
	}

	return transmission, nil
}

// TransmissionsByTimeRange queries the transmissions by time range
func (c *Client) TransmissionsByTimeRange(start int64, end int64, offset, limit int) ([]models.Transmission, errors.EdgeX) {
	validStart, validEnd, offset, validLimit, err := getValidRangeParameters(start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	transmission, err := queryTransmissions(context.Background(), c.ConnPool, sqlQueryContentWithTimeRangeAndPaginationAsNamedArgs(transmissionTableName),
		pgx.NamedArgs{startTimeCondition: validStart, endTimeCondition: validEnd, jsonContentCondition: map[string]any{}, offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all transmission by time range", err)
	}

	return transmission, nil
}

// AllTransmissions queries transmission with the given offset, and limit
func (c *Client) AllTransmissions(offset, limit int) ([]models.Transmission, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	transmissions, err := queryTransmissions(context.Background(), c.ConnPool, sqlQueryContentWithPaginationAsNamedArgs(transmissionTableName), pgx.NamedArgs{offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all transmissions", err)
	}

	return transmissions, nil
}

// TransmissionsByStatus queries the transmissions by status
func (c *Client) TransmissionsByStatus(offset, limit int, status string) ([]models.Transmission, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{statusField: status}

	transmissions, err := queryTransmissions(context.Background(), c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(transmissionTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query all transmissions by status %s", status), err)
	}

	return transmissions, nil
}

// DeleteProcessedTransmissionsByAge deletes the processed transmissions that are older than a specific age
func (c *Client) DeleteProcessedTransmissionsByAge(age int64) errors.EdgeX {
	status := []string{models.Sent, models.Acknowledged, models.Escalated}
	conditions := fmt.Sprintf("(content -> '%s') ?| $2", statusField)
	_, err := c.ConnPool.Exec(context.Background(), sqlDeleteByContentAgeWithConds(transmissionTableName, conditions), age, status)
	if err != nil {
		return pgClient.WrapDBError("failed to delete processed transmissions by age", err)
	}

	return nil
}

// TransmissionsBySubscriptionName queries the transmissions by subscription name
func (c *Client) TransmissionsBySubscriptionName(offset, limit int, subscriptionName string) ([]models.Transmission, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{subscriptionNameField: subscriptionName}

	transmissions, err := queryTransmissions(context.Background(), c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(transmissionTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query all transmissions by subscription name %s", subscriptionName), err)
	}

	return transmissions, nil
}

// TransmissionTotalCount returns the total count of transmissions
func (c *Client) TransmissionTotalCount() (int64, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCount(transmissionTableName))
}

// TransmissionCountBySubscriptionName returns the count of transmissions by subscription name
func (c *Client) TransmissionCountBySubscriptionName(subscriptionName string) (int64, errors.EdgeX) {
	queryObj := map[string]any{subscriptionNameField: subscriptionName}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONField(transmissionTableName), queryObj)
}

// TransmissionCountByStatus returns the count of transmissions by status
func (c *Client) TransmissionCountByStatus(status string) (int64, errors.EdgeX) {
	queryObj := map[string]any{statusField: status}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONField(transmissionTableName), queryObj)
}

// TransmissionCountByTimeRange returns the count of transmissions by time range
func (c *Client) TransmissionCountByTimeRange(start int64, end int64) (int64, errors.EdgeX) {
	validStart, validEnd, err := getValidStartAndEnd(start, end)
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByTimeRange(transmissionTableName), validStart, validEnd)
}

// TransmissionsByNotificationId queries the transmissions by notification id
func (c *Client) TransmissionsByNotificationId(offset, limit int, id string) ([]models.Transmission, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{notificationIdField: id}

	transmissions, err := queryTransmissions(context.Background(), c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(transmissionTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query all transmissions by notification id %s", id), err)
	}

	return transmissions, nil
}

// TransmissionCountByNotificationId returns the count of transmissions by notification id
func (c *Client) TransmissionCountByNotificationId(id string) (int64, errors.EdgeX) {
	queryObj := map[string]any{notificationIdField: id}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONField(transmissionTableName), queryObj)
}

func queryTransmission(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (models.Transmission, errors.EdgeX) {
	var transmission models.Transmission
	row := connPool.QueryRow(ctx, sql, args...)

	if err := row.Scan(&transmission); err != nil {
		return transmission, pgClient.WrapDBError("failed to query transmission", err)
	}
	return transmission, nil
}

func queryTransmissions(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]models.Transmission, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query rows from transmission table", err)
	}

	transmissions, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Transmission, error) {
		var t models.Transmission
		scanErr := row.Scan(&t)
		return t, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to Transmission model", err)
	}

	return transmissions, nil
}

func checkTransmissionExists(ctx context.Context, connPool *pgxpool.Pool, id string) (bool, errors.EdgeX) {
	var exists bool
	err := connPool.QueryRow(ctx, sqlCheckExistsById(transmissionTableName), id).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError(fmt.Sprintf("failed to query row by id '%s' from transmission table", id), err)
	}
	return exists, nil
}
