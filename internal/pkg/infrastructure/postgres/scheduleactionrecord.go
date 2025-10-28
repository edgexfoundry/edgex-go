//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
)

// AddScheduleActionRecord adds a new schedule action record to the database
// Note: the scheduledAt field should be set manually before calling this function.
func (c *Client) AddScheduleActionRecord(ctx context.Context, scheduleActionRecord model.ScheduleActionRecord) (model.ScheduleActionRecord, errors.EdgeX) {
	if len(scheduleActionRecord.Id) == 0 {
		scheduleActionRecord.Id = uuid.New().String()
	}
	return addScheduleActionRecord(ctx, c.ConnPool, scheduleActionRecord)
}

// AddScheduleActionRecords adds multiple schedule action records to the database
func (c *Client) AddScheduleActionRecords(ctx context.Context, scheduleActionRecords []model.ScheduleActionRecord) ([]model.ScheduleActionRecord, errors.EdgeX) {
	records := make([]model.ScheduleActionRecord, 0, len(scheduleActionRecords))
	for _, record := range scheduleActionRecords {
		r, err := c.AddScheduleActionRecord(ctx, record)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		records = append(records, r)
	}
	return records, nil
}

// AllScheduleActionRecords queries the schedule action records with the given range, offset, and limit
func (c *Client) AllScheduleActionRecords(ctx context.Context, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX) {
	var err errors.EdgeX
	startTime, endTime, offset, validLimit, err := getValidTimeRangeParameters(start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryAllWithPaginationAndTimeRangeAsNamedArgs(scheduleActionRecordTableName),
		pgx.NamedArgs{startTimeCondition: startTime, endTimeCondition: endTime, offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all schedule action records", err)
	}

	return records, nil
}

// LatestScheduleActionRecordsByJobName queries the latest schedule action records by job name
func (c *Client) LatestScheduleActionRecordsByJobName(ctx context.Context, jobName string) ([]model.ScheduleActionRecord, errors.EdgeX) {
	sqlQueryLatestScheduleActionRecords := fmt.Sprintf(`
	SELECT id, action_id, job_name, action, status, scheduled_at, created
	FROM(
	    SELECT *
		FROM (
			SELECT *,
				RANK() OVER (PARTITION BY job_name, action_id ORDER BY created DESC) AS rnk
			FROM %s
			WHERE job_name = $1
		) subquery
		WHERE rnk = 1
	)
    ORDER BY job_name, created DESC;
    `, scheduleActionRecordTableName)

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryLatestScheduleActionRecords, jobName)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query latest schedule action records", err)
	}

	return records, nil
}

// LatestScheduleActionRecordsByOffset queries the latest schedule action records by offset
func (c *Client) LatestScheduleActionRecordsByOffset(ctx context.Context, offset uint32) (model.ScheduleActionRecord, errors.EdgeX) {
	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryAllWithPaginationDescByCol(scheduleActionRecordTableName, createdCol), pgx.NamedArgs{offsetCondition: offset, limitCondition: 1})
	if err != nil {
		return model.ScheduleActionRecord{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query all schedule action records", err)
	}

	if len(records) == 0 {
		return model.ScheduleActionRecord{}, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("no schedule action record found with offset '%d'", offset), err)
	}
	return records[0], nil
}

// ScheduleActionRecordsByStatus queries the schedule action records by status with the given range, offset, and limit
func (c *Client) ScheduleActionRecordsByStatus(ctx context.Context, status string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX) {
	var err errors.EdgeX
	startTime, endTime, offset, validLimit, err := getValidTimeRangeParameters(start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryAllByStatusWithPaginationAndTimeRange(scheduleActionRecordTableName), status, startTime, endTime, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query schedule action records by status %s", status), err)
	}

	return records, nil
}

// ScheduleActionRecordsByJobName queries the schedule action records by job name with the given range, offset, and limit
func (c *Client) ScheduleActionRecordsByJobName(ctx context.Context, jobName string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX) {
	var err errors.EdgeX
	startTime, endTime, offset, validLimit, err := getValidTimeRangeParameters(start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryAllByColWithPaginationAndTimeRange(scheduleActionRecordTableName, jobNameCol),
		pgx.NamedArgs{jobNameCol: jobName, startTimeCondition: startTime, endTimeCondition: endTime, offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query schedule action records by job name %s", jobName), err)
	}

	return records, nil
}

// ScheduleActionRecordsByJobNameAndStatus queries the schedule action records by job name and status with the given range, offset, and limit
func (c *Client) ScheduleActionRecordsByJobNameAndStatus(ctx context.Context, jobName, status string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX) {
	var err errors.EdgeX
	startTime, endTime, offset, validLimit, err := getValidTimeRangeParameters(start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryAllByColWithPaginationAndTimeRange(scheduleActionRecordTableName, jobNameCol, statusCol),
		pgx.NamedArgs{jobNameCol: jobName, statusCol: status, startTimeCondition: startTime, endTimeCondition: endTime, offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query schedule action records by job name %s and status %s", jobName, status), err)
	}

	return records, nil
}

// ScheduleActionRecordTotalCount returns the total count of all the schedule action records
func (c *Client) ScheduleActionRecordTotalCount(ctx context.Context, start, end int64) (int64, errors.EdgeX) {
	startTime, endTime := getUTCStartAndEndTime(start, end)
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByTimeRangeCol(scheduleActionRecordTableName, createdCol, nil), pgx.NamedArgs{startTimeCondition: startTime, endTimeCondition: endTime})
}

// ScheduleActionRecordCountByStatus returns the total count of the schedule action records by status
func (c *Client) ScheduleActionRecordCountByStatus(ctx context.Context, status string, start, end int64) (int64, errors.EdgeX) {
	startTime, endTime := getUTCStartAndEndTime(start, end)
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByTimeRangeCol(scheduleActionRecordTableName, createdCol, nil, statusCol), pgx.NamedArgs{statusCol: status, startTimeCondition: startTime, endTimeCondition: endTime})
}

// ScheduleActionRecordCountByJobName returns the total count of the schedule action records by job name
func (c *Client) ScheduleActionRecordCountByJobName(ctx context.Context, jobName string, start, end int64) (int64, errors.EdgeX) {
	startTime, endTime := getUTCStartAndEndTime(start, end)
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByTimeRangeCol(scheduleActionRecordTableName, createdCol, nil, jobNameCol), pgx.NamedArgs{jobNameCol: jobName, startTimeCondition: startTime, endTimeCondition: endTime})
}

// ScheduleActionRecordCountByJobNameAndStatus returns the total count of the schedule action records by job name and status
func (c *Client) ScheduleActionRecordCountByJobNameAndStatus(ctx context.Context, jobName, status string, start, end int64) (int64, errors.EdgeX) {
	startTime, endTime := getUTCStartAndEndTime(start, end)
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByTimeRangeCol(scheduleActionRecordTableName, createdCol, nil, jobNameCol, statusCol),
		pgx.NamedArgs{jobNameCol: jobName, statusCol: status, startTimeCondition: startTime, endTimeCondition: endTime})
}

// DeleteScheduleActionRecordByAge deletes the schedule action records by age
func (c *Client) DeleteScheduleActionRecordByAge(ctx context.Context, age int64) errors.EdgeX {
	return deleteScheduleActionRecord(ctx, c.ConnPool, sqlDeleteByAge(scheduleActionRecordTableName), age)
}

func addScheduleActionRecord(ctx context.Context, connPool *pgxpool.Pool, scheduleActionRecord model.ScheduleActionRecord) (model.ScheduleActionRecord, errors.EdgeX) {
	actionId := scheduleActionRecord.Action.GetBaseScheduleAction().Id
	// Remove the payload from the action before storing it in the database to reduce the size of the record
	copiedScheduleAction := scheduleActionRecord.Action.WithEmptyPayloadAndId()

	// Marshal the action to store it in the database
	actionJSONBytes, err := json.Marshal(copiedScheduleAction)
	if err != nil {
		return scheduleActionRecord, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal schedule action record for Postgres persistence", err)
	}

	_, err = connPool.Exec(
		ctx,
		sqlInsert(scheduleActionRecordTableName, idCol, actionIdCol, jobNameCol, actionCol, statusCol, scheduledAtCol),
		scheduleActionRecord.Id,
		actionId,
		scheduleActionRecord.JobName,
		actionJSONBytes,
		scheduleActionRecord.Status,
		time.UnixMilli(scheduleActionRecord.ScheduledAt).UTC())
	if err != nil {
		return scheduleActionRecord, pgClient.WrapDBError("failed to insert schedule action record", err)
	}

	return scheduleActionRecord, nil
}

func queryScheduleActionRecords(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]model.ScheduleActionRecord, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("query failed", err)
	}
	defer rows.Close()

	var scheduleActionRecords []model.ScheduleActionRecord
	for rows.Next() {
		var actionId string
		var record model.ScheduleActionRecord
		var created, scheduledAt time.Time
		var actionJSONBytes []byte
		err := rows.Scan(&record.Id, &actionId, &record.JobName, &actionJSONBytes, &record.Status, &scheduledAt, &created)
		if err != nil {
			return nil, pgClient.WrapDBError("failed to scan schedule action record", err)
		}

		var action model.ScheduleAction
		action, err = model.UnmarshalScheduleAction(actionJSONBytes)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON unmarshal schedule action record", err)
		}
		// Set the action ID back to models.ScheduleAction
		actionWithId := action.WithId(actionId)

		record.Action = actionWithId
		record.Created = created.UnixMilli()
		record.ScheduledAt = scheduledAt.UnixMilli()
		scheduleActionRecords = append(scheduleActionRecords, record)
	}

	if readErr := rows.Err(); readErr != nil {
		return nil, pgClient.WrapDBError("error occurred while query support_scheduler.record table", readErr)
	}
	return scheduleActionRecords, nil
}

func deleteScheduleActionRecord(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) errors.EdgeX {
	_, err := connPool.Exec(ctx, sql, args...)
	if err != nil {
		return pgClient.WrapDBError("failed to delete schedule action records", err)
	}
	return nil
}
