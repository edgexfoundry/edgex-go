//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

type DBClient interface {
	CloseSession()

	AddScheduleJob(ctx context.Context, scheduleJob model.ScheduleJob) (model.ScheduleJob, errors.EdgeX)
	AllScheduleJobs(ctx context.Context, offset, limit int) ([]model.ScheduleJob, errors.EdgeX)
	UpdateScheduleJob(ctx context.Context, scheduleJob model.ScheduleJob) errors.EdgeX
	DeleteScheduleJobByName(ctx context.Context, name string) errors.EdgeX
	ScheduleJobById(ctx context.Context, id string) (model.ScheduleJob, errors.EdgeX)
	ScheduleJobByName(ctx context.Context, name string) (model.ScheduleJob, errors.EdgeX)
	ScheduleJobTotalCount(ctx context.Context) (uint32, errors.EdgeX)

	AddScheduleActionRecord(scheduleActionRecord model.ScheduleActionRecord) (model.ScheduleActionRecord, errors.EdgeX)
	AllScheduleActionRecords(start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	LatestScheduleActionRecords(offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordsByStatus(status string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordsByJobName(jobName string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordsByJobNameAndStatus(jobName, status string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordTotalCount() (uint32, errors.EdgeX)
	ScheduleActionRecordCountByStatus(status string) (uint32, errors.EdgeX)
	ScheduleActionRecordCountByJobName(jobName string) (uint32, errors.EdgeX)
	ScheduleActionRecordCountByJobNameAndStatus(jobName, status string) (uint32, errors.EdgeX)
	DeleteScheduleActionRecordByAge(age int64) errors.EdgeX
}
