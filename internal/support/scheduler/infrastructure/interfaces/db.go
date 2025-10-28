//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type DBClient interface {
	CloseSession()

	AddScheduleJob(ctx context.Context, scheduleJob model.ScheduleJob) (model.ScheduleJob, errors.EdgeX)
	AllScheduleJobs(ctx context.Context, labels []string, offset, limit int) ([]model.ScheduleJob, errors.EdgeX)
	UpdateScheduleJob(ctx context.Context, scheduleJob model.ScheduleJob) errors.EdgeX
	DeleteScheduleJobByName(ctx context.Context, name string) errors.EdgeX
	ScheduleJobById(ctx context.Context, id string) (model.ScheduleJob, errors.EdgeX)
	ScheduleJobByName(ctx context.Context, name string) (model.ScheduleJob, errors.EdgeX)
	ScheduleJobTotalCount(ctx context.Context, labels []string) (int64, errors.EdgeX)

	AddScheduleActionRecord(ctx context.Context, scheduleActionRecord model.ScheduleActionRecord) (model.ScheduleActionRecord, errors.EdgeX)
	AddScheduleActionRecords(ctx context.Context, scheduleActionRecord []model.ScheduleActionRecord) ([]model.ScheduleActionRecord, errors.EdgeX)
	AllScheduleActionRecords(ctx context.Context, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	LatestScheduleActionRecordsByJobName(ctx context.Context, jobName string) ([]model.ScheduleActionRecord, errors.EdgeX)
	LatestScheduleActionRecordsByOffset(ctx context.Context, offset uint32) (model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordsByStatus(ctx context.Context, status string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordsByJobName(ctx context.Context, jobName string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordsByJobNameAndStatus(ctx context.Context, jobName, status string, start, end int64, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordTotalCount(ctx context.Context, start, end int64) (int64, errors.EdgeX)
	ScheduleActionRecordCountByStatus(ctx context.Context, status string, start, end int64) (int64, errors.EdgeX)
	ScheduleActionRecordCountByJobName(ctx context.Context, jobName string, start, end int64) (int64, errors.EdgeX)
	ScheduleActionRecordCountByJobNameAndStatus(ctx context.Context, jobName, status string, start, end int64) (int64, errors.EdgeX)
	DeleteScheduleActionRecordByAge(ctx context.Context, age int64) errors.EdgeX
}
