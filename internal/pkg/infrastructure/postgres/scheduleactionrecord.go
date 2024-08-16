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
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
)

const (
	scheduleActionRecordTable = "scheduler.schedule_action_record"
	jobNameCol                = "job_name"
	actionCol                 = "action"
	scheduledAtCol            = "scheduled_at"
)

// AddScheduleActionRecord adds a new schedule action record to the database
// Note: the created field should be set manually before calling this function, and all the records belong to the same job should have the same created time.
// So that the created time can be used to query the latest schedule action records of a job.
func (c *Client) AddScheduleActionRecord(ctx context.Context, scheduleActionRecord model.ScheduleActionRecord) (model.ScheduleActionRecord, errors.EdgeX) {
	if len(scheduleActionRecord.Id) == 0 {
		scheduleActionRecord.Id = uuid.New().String()
	}
	return addScheduleActionRecord(ctx, c.ConnPool, scheduleActionRecord)
}

// AllScheduleActionRecords queries the schedule action records with the given range, offset, and limit
func (c *Client) AllScheduleActionRecords(ctx context.Context, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX) {
	var err errors.EdgeX
	start, end, offset, limit, err = getValidRangeParameters(start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryAllWithPaginationAndTimeRange(scheduleActionRecordTable), time.UnixMilli(start), time.UnixMilli(end), offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query all schedule action records", err)
	}

	return records, nil
}

// LatestScheduleActionRecords queries the latest schedule action records of all schedule jobs with the given offset and limit
func (c *Client) LatestScheduleActionRecords(ctx context.Context, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX) {
	// Get all the job names
	jobNames, err := queryScheduleJobNames(ctx, c.ConnPool)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	sqlQueryLatestScheduleActionRecords := `
	SELECT id, job_name, action, status, scheduled_at, created
	FROM(
	    SELECT *
		FROM (
			SELECT *,
				RANK() OVER (PARTITION BY job_name ORDER BY created DESC) AS rnk
			FROM scheduler.schedule_action_record
			WHERE job_name = ANY($1::text[])
		) subquery
		WHERE rnk = 1
	)
    ORDER BY job_name, created DESC
    OFFSET $2
    LIMIT $3;
    `

	// Pass the offset and limit here
	offset, limit = getValidOffsetAndLimit(offset, limit)
	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryLatestScheduleActionRecords, jobNames, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query latest schedule action records", err)
	}

	return records, nil
}

// ScheduleActionRecordsByStatus queries the schedule action records by status with the given range, offset, and limit
func (c *Client) ScheduleActionRecordsByStatus(ctx context.Context, status string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX) {
	var err errors.EdgeX
	start, end, offset, limit, err = getValidRangeParameters(start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryAllByStatusWithPaginationAndTimeRange(scheduleActionRecordTable), status, time.UnixMilli(start), time.UnixMilli(end), offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to query schedule action records by status %s", status), err)
	}

	return records, nil
}

// ScheduleActionRecordsByJobName queries the schedule action records by job name with the given range, offset, and limit
func (c *Client) ScheduleActionRecordsByJobName(ctx context.Context, jobName string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX) {
	var err errors.EdgeX
	start, end, offset, limit, err = getValidRangeParameters(start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryAllByColWithPaginationAndTimeRange(scheduleActionRecordTable, jobNameCol), jobName, time.UnixMilli(start), time.UnixMilli(end), offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to query schedule action records by job name %s", jobName), err)
	}

	return records, nil
}

// ScheduleActionRecordsByJobNameAndStatus queries the schedule action records by job name and status with the given range, offset, and limit
func (c *Client) ScheduleActionRecordsByJobNameAndStatus(ctx context.Context, jobName, status string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX) {
	var err errors.EdgeX
	start, end, offset, limit, err = getValidRangeParameters(start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryAllByColWithPaginationAndTimeRange(scheduleActionRecordTable, jobNameCol, statusCol), jobName, status, time.UnixMilli(start), time.UnixMilli(end), offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to query schedule action records by job name %s and status %s", jobName, status), err)
	}

	return records, nil
}

// ScheduleActionRecordTotalCount returns the total count of all the schedule action records
func (c *Client) ScheduleActionRecordTotalCount(ctx context.Context) (uint32, errors.EdgeX) {
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCount(scheduleActionRecordTable))
}

// LatestScheduleActionRecordTotalCount returns the total count of all the latest schedule action records
func (c *Client) LatestScheduleActionRecordTotalCount(ctx context.Context) (uint32, errors.EdgeX) {
	sqlQueryLatestScheduleActionRecordCount := `
	SELECT COUNT(*)
	FROM (
	    SELECT *
	    FROM (
	        SELECT *,
            RANK() OVER (PARTITION BY job_name ORDER BY created DESC) AS rnk
        	FROM scheduler.schedule_action_record
	    )
	    WHERE rnk = 1
	)
	`
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryLatestScheduleActionRecordCount)
}

// ScheduleActionRecordCountByStatus returns the total count of the schedule action records by status
func (c *Client) ScheduleActionRecordCountByStatus(ctx context.Context, status string) (uint32, errors.EdgeX) {
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByCol(scheduleActionRecordTable, statusCol), status)
}

// ScheduleActionRecordCountByJobName returns the total count of the schedule action records by job name
func (c *Client) ScheduleActionRecordCountByJobName(ctx context.Context, jobName string) (uint32, errors.EdgeX) {
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByCol(scheduleActionRecordTable, jobNameCol), jobName)
}

// ScheduleActionRecordCountByJobNameAndStatus returns the total count of the schedule action records by job name and status
func (c *Client) ScheduleActionRecordCountByJobNameAndStatus(ctx context.Context, jobName, status string) (uint32, errors.EdgeX) {
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByCol(scheduleActionRecordTable, jobNameCol, statusCol), jobName, status)
}

// DeleteScheduleActionRecordByAge deletes the schedule action records by age
func (c *Client) DeleteScheduleActionRecordByAge(ctx context.Context, age int64) errors.EdgeX {
	return deleteScheduleActionRecord(ctx, c.ConnPool, sqlDeleteByAge(scheduleActionRecordTable), age)
}

func addScheduleActionRecord(ctx context.Context, connPool *pgxpool.Pool, scheduleActionRecord model.ScheduleActionRecord) (model.ScheduleActionRecord, errors.EdgeX) {
	// Remove the payload from the action before storing it in the database to reduce the size of the record
	copiedScheduleAction := scheduleActionRecord.Action.WithEmptyPayload()

	// Marshal the action to store it in the database
	actionJSONBytes, err := json.Marshal(copiedScheduleAction)
	if err != nil {
		return scheduleActionRecord, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal schedule action record for Postgres persistence", err)
	}

	_, err = connPool.Exec(
		ctx,
		sqlInsert(scheduleActionRecordTable, idCol, jobNameCol, actionCol, statusCol, scheduledAtCol, createdCol),
		scheduleActionRecord.Id,
		scheduleActionRecord.JobName,
		actionJSONBytes,
		scheduleActionRecord.Status,
		time.UnixMilli(scheduleActionRecord.ScheduledAt).UTC(),
		time.UnixMilli(scheduleActionRecord.Created).UTC())
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
		var record model.ScheduleActionRecord
		var created, scheduledAt time.Time
		var actionJSONBytes []byte
		err := rows.Scan(&record.Id, &record.JobName, &actionJSONBytes, &record.Status, &scheduledAt, &created)
		if err != nil {
			return nil, pgClient.WrapDBError("failed to scan schedule action record", err)
		}

		var action model.ScheduleAction
		action, err = model.UnmarshalScheduleAction(actionJSONBytes)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON unmarshal schedule action record", err)
		}

		record.Action = action
		record.Created = created.UnixMilli()
		record.ScheduledAt = scheduledAt.UnixMilli()
		scheduleActionRecords = append(scheduleActionRecords, record)
	}

	if readErr := rows.Err(); readErr != nil {
		return nil, pgClient.WrapDBError("error occurred while query scheduler.schedule_action_record table", readErr)
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
