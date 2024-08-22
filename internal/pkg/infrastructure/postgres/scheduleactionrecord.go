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
	actionIdCol               = "action_id"
	jobNameCol                = "job_name"
	actionCol                 = "action"
	scheduledAtCol            = "scheduled_at"
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
	records := make([]model.ScheduleActionRecord, len(scheduleActionRecords))
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
	start, end, offset, limit, err = getValidRangeParameters(start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryAllWithPaginationAndTimeRange(scheduleActionRecordTable), time.UnixMilli(start), time.UnixMilli(end), offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all schedule action records", err)
	}

	return records, nil
}

// LatestScheduleActionRecordsByJobName queries the latest schedule action records by job name
func (c *Client) LatestScheduleActionRecordsByJobName(ctx context.Context, jobName string) ([]model.ScheduleActionRecord, errors.EdgeX) {
	sqlQueryLatestScheduleActionRecords := `
	SELECT id, action_id, job_name, action, status, scheduled_at, created
	FROM(
	    SELECT *
		FROM (
			SELECT *,
				RANK() OVER (PARTITION BY job_name, action_id ORDER BY created DESC) AS rnk
			FROM scheduler.schedule_action_record
			WHERE job_name = $1
		) subquery
		WHERE rnk = 1
	)
    ORDER BY job_name, created DESC;
    `

	records, err := queryScheduleActionRecords(ctx, c.ConnPool, sqlQueryLatestScheduleActionRecords, jobName)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query latest schedule action records", err)
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
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query schedule action records by status %s", status), err)
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
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query schedule action records by job name %s", jobName), err)
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
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query schedule action records by job name %s and status %s", jobName, status), err)
	}

	return records, nil
}

// ScheduleActionRecordTotalCount returns the total count of all the schedule action records
func (c *Client) ScheduleActionRecordTotalCount(ctx context.Context) (uint32, errors.EdgeX) {
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCount(scheduleActionRecordTable))
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
		sqlInsert(scheduleActionRecordTable, idCol, actionIdCol, jobNameCol, actionCol, statusCol, scheduledAtCol),
		scheduleActionRecord.Id,
		copiedScheduleAction.GetBaseScheduleAction().Id,
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
