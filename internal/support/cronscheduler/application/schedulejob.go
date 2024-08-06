//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/edgex-go/internal/support/cronscheduler/container"
	"github.com/edgexfoundry/edgex-go/internal/support/cronscheduler/infrastructure/interfaces"
)

// AddScheduleJob adds a new schedule job
func AddScheduleJob(ctx context.Context, job models.ScheduleJob, dic *di.Container) (string, errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	addedJob, err := dbClient.AddScheduleJob(ctx, job)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	err = schedulerManager.AddScheduleJob(job)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	err = schedulerManager.StartScheduleJobByName(job.Name)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debug("Successfully created the scheduled job")
	return addedJob.Id, nil
}

// TriggerScheduleJobByName triggers a schedule job by name
func TriggerScheduleJobByName(name string, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}

	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	err := schedulerManager.TriggerScheduleJobByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debug("Successfully triggered the scheduled job")
	return nil
}

// ScheduleJobByName queries the schedule job by name
func ScheduleJobByName(ctx context.Context, name string, dic *di.Container) (dto dtos.ScheduleJob, edgeXerr errors.EdgeX) {
	if name == "" {
		return dto, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	job, err := dbClient.ScheduleJobByName(ctx, name)
	if err != nil {
		return dto, errors.NewCommonEdgeXWrapper(err)
	}
	dto = dtos.FromScheduleJobModelToDTO(job)

	return dto, nil
}

// AllScheduleJobs queries all the schedule jobs with offset and limit
func AllScheduleJobs(ctx context.Context, offset, limit int, dic *di.Container) (scheduleJobDTOs []dtos.ScheduleJob, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	jobs, err := dbClient.AllScheduleJobs(ctx, offset, limit)
	if err == nil {
		totalCount, err = dbClient.ScheduleJobTotalCount(ctx)
	}
	if err != nil {
		return scheduleJobDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}

	scheduleJobDTOs = make([]dtos.ScheduleJob, len(jobs))
	for i, job := range jobs {
		dto := dtos.FromScheduleJobModelToDTO(job)
		scheduleJobDTOs[i] = dto
	}

	return scheduleJobDTOs, totalCount, nil
}

// PatchScheduleJob executes the PATCH operation with the DTO to replace the old data
func PatchScheduleJob(ctx context.Context, dto dtos.UpdateScheduleJob, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	job, err := scheduleJobByDTO(ctx, dbClient, dto)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	requests.ReplaceScheduleJobModelFieldsWithDTO(&job, dto)

	err = dbClient.UpdateScheduleJob(ctx, job)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = schedulerManager.UpdateScheduleJob(job)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("Successfully patched the scheduled job %s", job.Name)
	return nil
}

func scheduleJobByDTO(ctx context.Context, dbClient interfaces.DBClient, dto dtos.UpdateScheduleJob) (job models.ScheduleJob, err errors.EdgeX) {
	// The ID or Name is required by DTO and the DTO also accepts empty string ID if the Name is provided
	if dto.Id != nil && *dto.Id != "" {
		job, err = dbClient.ScheduleJobById(ctx, *dto.Id)
		if err != nil {
			return job, errors.NewCommonEdgeXWrapper(err)
		}
	} else {
		job, err = dbClient.ScheduleJobByName(ctx, *dto.Name)
		if err != nil {
			return job, errors.NewCommonEdgeXWrapper(err)
		}
	}
	if dto.Name != nil && *dto.Name != job.Name {
		return job, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("schedule job name '%s' not match the exsting '%s' ", *dto.Name, job.Name), nil)
	}
	return job, nil
}

// DeleteScheduleJobByName deletes the schedule job by name
func DeleteScheduleJobByName(ctx context.Context, name string, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	err := schedulerManager.DeleteScheduleJobByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	err = dbClient.DeleteScheduleJobByName(ctx, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debug("Successfully deleted the scheduled job")
	return nil
}
