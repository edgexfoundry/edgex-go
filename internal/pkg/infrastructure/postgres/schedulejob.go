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

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
)

// AddScheduleJob adds a new schedule job to the database
func (c *Client) AddScheduleJob(ctx context.Context, j models.ScheduleJob) (models.ScheduleJob, errors.EdgeX) {
	if len(j.Id) == 0 {
		j.Id = uuid.New().String()
	}

	j, err := addScheduleJob(ctx, c.ConnPool, j)
	if err != nil {
		return j, errors.NewCommonEdgeXWrapper(err)
	}
	return j, nil
}

// AllScheduleJobs queries the schedule jobs with the given range, offset, and limit
func (c *Client) AllScheduleJobs(ctx context.Context, labels []string, offset, limit int) (jobs []models.ScheduleJob, err errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	if len(labels) > 0 {
		c.loggingClient.Debugf("Querying schedule jobs by labels: %v", labels)
		queryObj := map[string]any{labelsField: labels}
		jobs, err = queryScheduleJobs(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(scheduleJobTableName),
			pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all schedule jobs by labels", err)
		}
	} else {
		jobs, err = queryScheduleJobs(ctx, c.ConnPool, sqlQueryContentWithPaginationAsNamedArgs(scheduleJobTableName), pgx.NamedArgs{offsetCondition: offset, limitCondition: validLimit})
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all schedule jobs", err)
		}
	}

	return jobs, nil
}

// UpdateScheduleJob updates the schedule job
func (c *Client) UpdateScheduleJob(ctx context.Context, j models.ScheduleJob) errors.EdgeX {
	err := updateScheduleJob(ctx, c.ConnPool, j)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// DeleteScheduleJobByName deletes the schedule job by name
func (c *Client) DeleteScheduleJobByName(ctx context.Context, name string) errors.EdgeX {
	if err := deleteScheduleJobByName(ctx, c.ConnPool, name); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// ScheduleJobById queries the schedule job by id
func (c *Client) ScheduleJobById(ctx context.Context, id string) (models.ScheduleJob, errors.EdgeX) {
	scheduleJob, err := queryScheduleJob(ctx, c.ConnPool, sqlQueryAllById(scheduleJobTableName), id)
	if err != nil {
		return scheduleJob, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query schedule job by id '%s'", id), err)
	}

	return scheduleJob, nil
}

// ScheduleJobByName queries the schedule job by name
func (c *Client) ScheduleJobByName(ctx context.Context, name string) (models.ScheduleJob, errors.EdgeX) {
	queryObj := map[string]any{nameField: name}
	scheduleJob, err := queryScheduleJob(ctx, c.ConnPool, sqlQueryContentByJSONField(scheduleJobTableName), queryObj)
	if err != nil {
		return scheduleJob, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query schedule job by name '%s'", name), err)
	}

	return scheduleJob, nil
}

// ScheduleJobTotalCount returns the total count of schedule jobs
func (c *Client) ScheduleJobTotalCount(ctx context.Context, labels []string) (int64, errors.EdgeX) {
	if len(labels) > 0 {
		queryObj := map[string]any{labelsField: labels}
		return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(scheduleJobTableName), queryObj)
	}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCount(scheduleJobTableName))
}

func addScheduleJob(ctx context.Context, connPool *pgxpool.Pool, j models.ScheduleJob) (models.ScheduleJob, errors.EdgeX) {
	exists, edgexErr := checkScheduleJobExists(ctx, connPool, j.Name)
	if edgexErr != nil {
		return j, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	if exists {
		return j, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("schedule job name '%s' already exists", j.Name), nil)
	}

	timestamp := time.Now().UTC().UnixMilli()
	j.Created = timestamp
	j.Modified = timestamp
	dataBytes, err := json.Marshal(j)
	if err != nil {
		return j, errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal ScheduleJob model", err)
	}

	_, err = connPool.Exec(context.Background(), sqlInsert(scheduleJobTableName, idCol, contentCol), j.Id, dataBytes)
	if err != nil {
		return j, pgClient.WrapDBError("failed to insert row to scheduler.job table", err)
	}

	return j, nil
}

func updateScheduleJob(ctx context.Context, connPool *pgxpool.Pool, j models.ScheduleJob) errors.EdgeX {
	modified := time.Now().UTC().UnixMilli()
	j.Modified = modified

	dataBytes, err := json.Marshal(j)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal ScheduleJob model", err)
	}

	queryObj := map[string]any{nameField: j.Name}
	_, err = connPool.Exec(ctx, sqlUpdateColsByJSONCondCol(scheduleJobTableName, contentCol), dataBytes, queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to update row by schedule job name '%s' from scheduler.job table", j.Name), err)
	}

	return nil
}

func deleteScheduleJobByName(ctx context.Context, connPool *pgxpool.Pool, name string) errors.EdgeX {
	queryObj := map[string]any{nameField: name}
	_, err := connPool.Exec(ctx, sqlDeleteByJSONField(scheduleJobTableName), queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete schedule job by name %s", name), err)
	}
	return nil
}

func checkScheduleJobExists(ctx context.Context, connPool *pgxpool.Pool, name string) (bool, errors.EdgeX) {
	var exists bool
	queryObj := map[string]any{nameField: name}
	err := connPool.QueryRow(ctx, sqlCheckExistsByJSONField(scheduleJobTableName), queryObj).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError(fmt.Sprintf("failed to query row by name '%s' from scheduler.job table", name), err)
	}
	return exists, nil
}

func queryScheduleJob(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (models.ScheduleJob, errors.EdgeX) {
	var job models.ScheduleJob
	row := connPool.QueryRow(ctx, sql, args...)

	if err := row.Scan(&job); err != nil {
		return job, pgClient.WrapDBError("failed to query schedule job", err)
	}
	return job, nil
}

func queryScheduleJobs(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]models.ScheduleJob, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query rows from scheduler.job table", err)
	}

	jobs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.ScheduleJob, error) {
		var j models.ScheduleJob
		scanErr := row.Scan(&j)
		return j, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to ScheduleJob model", err)
	}

	return jobs, nil
}
