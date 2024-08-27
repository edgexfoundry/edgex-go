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

const scheduleJobTable = "scheduler.schedule_job"

// AddScheduleJob adds a new schedule job to the database
func (c *Client) AddScheduleJob(ctx context.Context, scheduleJob model.ScheduleJob) (model.ScheduleJob, errors.EdgeX) {
	if len(scheduleJob.Id) == 0 {
		scheduleJob.Id = uuid.New().String()
	}
	return addScheduleJob(ctx, c.ConnPool, scheduleJob)
}

// AllScheduleJobs queries the schedule jobs with the given range, offset, and limit
func (c *Client) AllScheduleJobs(ctx context.Context, labels []string, offset, limit int) (jobs []model.ScheduleJob, err errors.EdgeX) {
	offset, limit = getValidOffsetAndLimit(offset, limit)
	if len(labels) > 0 {
		c.loggingClient.Debugf("Querying schedule jobs by labels: %v", labels)
		labelsJSON, err := json.Marshal(labels)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal labels", err)
		}
		jobs, err = queryScheduleJobs(ctx, c.ConnPool, sqlQueryAllByContentLabelsWithPagination(scheduleJobTable), labelsJSON, offset, limit)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all schedule jobs by labels", err)
		}
	} else {
		jobs, err = queryScheduleJobs(ctx, c.ConnPool, sqlQueryAllWithPagination(scheduleJobTable), offset, limit)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all schedule jobs", err)
		}
	}

	return jobs, nil
}

// UpdateScheduleJob updates the schedule job
func (c *Client) UpdateScheduleJob(ctx context.Context, scheduleJob model.ScheduleJob) errors.EdgeX {
	// Check if the schedule job exists
	exists, err := scheduleJobNameExists(ctx, c.ConnPool, scheduleJob.Name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	} else if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("schedule job '%s' does not exist", scheduleJob.Name), nil)
	}

	err = updateScheduleJob(ctx, c.ConnPool, scheduleJob)
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
func (c *Client) ScheduleJobById(ctx context.Context, id string) (model.ScheduleJob, errors.EdgeX) {
	scheduleJob, err := queryScheduleJob(ctx, c.ConnPool, sqlQueryAllById(scheduleJobTable), id)
	if err != nil {
		return scheduleJob, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query schedule job by id %s", id), err)
	}

	return scheduleJob, nil
}

// ScheduleJobByName queries the schedule job by name
func (c *Client) ScheduleJobByName(ctx context.Context, name string) (model.ScheduleJob, errors.EdgeX) {
	scheduleJob, err := queryScheduleJob(ctx, c.ConnPool, sqlQueryAllByName(scheduleJobTable), name)
	if err != nil {
		return scheduleJob, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query schedule job by name %s", name), err)
	}

	return scheduleJob, nil
}

// ScheduleJobTotalCount returns the total count of schedule jobs
func (c *Client) ScheduleJobTotalCount(ctx context.Context, labels []string) (uint32, errors.EdgeX) {
	if len(labels) > 0 {
		labelsJSON, err := json.Marshal(labels)
		if err != nil {
			return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal labels", err)
		}
		return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountContentLabels(scheduleJobTable), labelsJSON)
	}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCount(scheduleJobTable))
}

func addScheduleJob(ctx context.Context, connPool *pgxpool.Pool, scheduleJob model.ScheduleJob) (model.ScheduleJob, errors.EdgeX) {
	// Check if the schedule job name exists
	if exists, _ := scheduleJobNameExists(ctx, connPool, scheduleJob.Name); exists {
		return scheduleJob, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("schedule job name %s already exists", scheduleJob.Name), nil)
	}

	// Marshal the scheduleJob to store it in the database
	scheduleJobJSONBytes, err := json.Marshal(scheduleJob)
	if err != nil {
		return scheduleJob, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal schedule job for Postgres persistence", err)
	}

	_, err = connPool.Exec(ctx, sqlInsertContent(scheduleJobTable), scheduleJob.Id, scheduleJob.Name, scheduleJobJSONBytes)
	if err != nil {
		return scheduleJob, pgClient.WrapDBError("failed to insert schedule job", err)
	}

	return scheduleJob, nil
}

func updateScheduleJob(ctx context.Context, connPool *pgxpool.Pool, updatedScheduleJob model.ScheduleJob) errors.EdgeX {
	modified := time.Now().UTC()
	updatedScheduleJob.Modified = modified.UnixMilli()

	// Marshal the scheduleJob to store it in the database
	updatedScheduleJobJSONBytes, err := json.Marshal(updatedScheduleJob)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal schedule job for Postgres persistence", err)
	}

	_, err = connPool.Exec(ctx, sqlUpdateContentByName(scheduleJobTable), updatedScheduleJobJSONBytes, modified, updatedScheduleJob.Name)
	if err != nil {
		return pgClient.WrapDBError("failed to update schedule job", err)
	}

	return nil
}

func deleteScheduleJobByName(ctx context.Context, connPool *pgxpool.Pool, name string) errors.EdgeX {
	_, err := connPool.Exec(ctx, sqlDeleteByName(scheduleJobTable), name)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete schedule job by name %s", name), err)
	}
	return nil
}

func scheduleJobNameExists(ctx context.Context, connPool *pgxpool.Pool, name string) (bool, errors.EdgeX) {
	var exists bool
	err := connPool.QueryRow(ctx, sqlCheckExistsByName(scheduleJobTable), name).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError("failed to query schedule job by name", err)
	}
	return exists, nil
}

func queryScheduleJob(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (model.ScheduleJob, errors.EdgeX) {
	var job model.ScheduleJob
	var scheduleJobJSONBytes []byte
	var created, modified time.Time
	err := connPool.QueryRow(ctx, sql, args...).Scan(&job.Id, &job.Name, &scheduleJobJSONBytes, &created, &modified)
	if err != nil {
		return job, pgClient.WrapDBError("failed to query scheduler.schedule_job table", err)
	}

	job.Created = created.UnixMilli()
	job.Modified = modified.UnixMilli()

	job, err = toScheduleJobsModel(job, scheduleJobJSONBytes)
	if err != nil {
		return job, errors.NewCommonEdgeXWrapper(err)
	}

	return job, nil
}

func queryScheduleJobs(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]model.ScheduleJob, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("query failed", err)
	}
	defer rows.Close()

	var scheduleJobs []model.ScheduleJob
	for rows.Next() {
		var job model.ScheduleJob
		var created, modified time.Time
		var scheduleJobJSONBytes []byte
		err := rows.Scan(&job.Id, &job.Name, &scheduleJobJSONBytes, &created, &modified)
		if err != nil {
			return nil, pgClient.WrapDBError("failed to scan schedule job", err)
		}

		job.Created = created.UnixMilli()
		job.Modified = modified.UnixMilli()

		job, err = toScheduleJobsModel(job, scheduleJobJSONBytes)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		scheduleJobs = append(scheduleJobs, job)
	}

	if readErr := rows.Err(); readErr != nil {
		return nil, pgClient.WrapDBError("error occurred while query scheduler.schedule_job table", readErr)
	}
	return scheduleJobs, nil
}

func toScheduleJobsModel(scheduleJobs model.ScheduleJob, scheduleJobJSONBytes []byte) (model.ScheduleJob, errors.EdgeX) {
	var storedJob model.ScheduleJob
	if err := json.Unmarshal(scheduleJobJSONBytes, &storedJob); err != nil {
		return scheduleJobs, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON unmarshal schedule job", err)
	}

	scheduleJobs.Actions = storedJob.Actions
	scheduleJobs.AdminState = storedJob.AdminState
	scheduleJobs.Definition = storedJob.Definition
	scheduleJobs.Labels = storedJob.Labels
	scheduleJobs.Properties = storedJob.Properties
	return scheduleJobs, nil
}
