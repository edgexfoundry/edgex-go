//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

type DBClient interface {
	CloseSession()

	AddScheduleJob(scheduleJob model.ScheduleJob) (model.ScheduleJob, errors.EdgeX)
	AllScheduleJobs(start, end, offset, limit int) ([]model.ScheduleJob, errors.EdgeX)
	UpdateScheduleJob(scheduleJob model.ScheduleJob) errors.EdgeX
	DeleteScheduleJobByName(name string) errors.EdgeX
	ScheduleJobById(id string) (model.ScheduleJob, errors.EdgeX)
	ScheduleJobByName(name string) (model.ScheduleJob, errors.EdgeX)
	ScheduleJobTotalCount() (uint32, errors.EdgeX)
	ScheduleJobCountById(id string) (uint32, errors.EdgeX)
	ScheduleJobCountByName(name string) (uint32, errors.EdgeX)

	AddScheduleActionRecord(scheduleActionRecord model.ScheduleActionRecord) (model.ScheduleActionRecord, errors.EdgeX)
	AllScheduleActionRecords(start, end, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	LatestScheduleActionRecords(offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordsByStatus(status string, start, end, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordsByJobName(jobName string, start, end, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordsByJobNameAndStatus(jobName, status string, start, end, offset, limit int) ([]model.ScheduleActionRecord, errors.EdgeX)
	ScheduleActionRecordTotalCount() (uint32, errors.EdgeX)
	ScheduleActionRecordCountByStatus(status string) (uint32, errors.EdgeX)
	ScheduleActionRecordCountByJobName(jobName string) (uint32, errors.EdgeX)
	ScheduleActionRecordCountByJobNameAndStatus(jobName, status string) (uint32, errors.EdgeX)
}
