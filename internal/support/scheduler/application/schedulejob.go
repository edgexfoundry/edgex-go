//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces"
)

// AddScheduleJob adds a new schedule job
func AddScheduleJob(ctx context.Context, job models.ScheduleJob, dic *di.Container) (string, errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	// Add the ID for each action
	for i, action := range job.Actions {
		job.Actions[i] = action.WithId("")
	}

	err := schedulerManager.AddScheduleJob(job, correlationId)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	addedJob, err := dbClient.AddScheduleJob(ctx, job)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("Successfully created the scheduled job. ScheduleJob ID: %s, Correlation-ID: %s", addedJob.Id, correlationId)
	return addedJob.Id, nil
}

// TriggerScheduleJobByName triggers a schedule job by name
func TriggerScheduleJobByName(ctx context.Context, name string, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}

	correlationId := correlation.FromContext(ctx)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	err := schedulerManager.TriggerScheduleJobByName(name, correlationId)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("Successfully triggered the scheduled job. Correlation-ID: %s", correlationId)
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
func AllScheduleJobs(ctx context.Context, labels []string, offset, limit int, dic *di.Container) (scheduleJobDTOs []dtos.ScheduleJob, totalCount int64, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	totalCount, err = dbClient.ScheduleJobTotalCount(ctx, labels)
	if err != nil {
		return scheduleJobDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.ScheduleJob{}, totalCount, err
	}

	jobs, err := dbClient.AllScheduleJobs(ctx, labels, offset, limit)
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
	correlationId := correlation.FromContext(ctx)

	job, err := scheduleJobByDTO(ctx, dbClient, dto)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	requests.ReplaceScheduleJobModelFieldsWithDTO(&job, dto)

	// Add the ID for each action, the old actions will be replaced by the new actions
	for i, action := range job.Actions {
		job.Actions[i] = action.WithId("")
	}

	err = schedulerManager.UpdateScheduleJob(job, correlationId)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = dbClient.UpdateScheduleJob(ctx, job)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("Successfully patched the scheduled job: %s. ScheduleJob ID: %s, Correlation-ID: %s", job.Name, job.Id, correlationId)
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
		return job, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("scheduled job name '%s' not match the exsting '%s' ", *dto.Name, job.Name), nil)
	}
	return job, nil
}

// DeleteScheduleJobByName deletes the schedule job by name
func DeleteScheduleJobByName(ctx context.Context, name string, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	err := schedulerManager.DeleteScheduleJobByName(name, correlationId)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	err = dbClient.DeleteScheduleJobByName(ctx, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("Successfully deleted the scheduled job: %s. Correlation-ID: %s", name, correlationId)
	return nil
}

// LoadScheduleJobsToSchedulerManager loads all the existing schedule jobs to the scheduler manager, the MaxResultCount config is used to limit the number of jobs that will be loaded
func LoadScheduleJobsToSchedulerManager(ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	ctx, correlationId := correlation.FromContextOrNew(ctx)
	config := container.ConfigurationFrom(dic.Get)

	jobs, err := dbClient.AllScheduleJobs(context.Background(), nil, 0, config.Service.MaxResultCount)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to load all existing scheduled jobs", err)
	}

	for _, job := range jobs {
		err := schedulerManager.AddScheduleJob(job, correlationId)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}

		// If endTimestamp is set and expired, the missed schedule action records should not be generated
		isEndExpired := isEndTimestampExpired(job.Definition.GetBaseScheduleDef().EndTimestamp)
		if isEndExpired {
			lc.Debugf("The endTimestamp is expired for the scheduled job: %s, which will not generate missed schedule action records. Correlation-ID: %s", job.Name, correlationId)
			continue
		}
		// Generate missed schedule action records for the existing scheduled jobs
		err, hasMissedAction := generateMissedRecords(ctx, job, dic)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		if hasMissedAction && job.AutoTriggerMissedRecords {
			lc.Debugf("Auto-triggering the missed schedule actions once for the scheduled job: %s. Correlation-ID: %s", job.Name, correlationId)
			err = schedulerManager.TriggerScheduleJobByName(job.Name, correlationId)
			if err != nil {
				return errors.NewCommonEdgeXWrapper(err)
			}
		}

		if !job.AutoTriggerMissedRecords {
			lc.Debugf("AutoTriggerMissedRecords is disabled, the missed schedule actions for the scheduled job: %s will not be auto-triggered. Correlation-ID: %s", job.Name, correlationId)
		}

		lc.Debugf("Successfully loaded the existing scheduled job: %s. Correlation-ID: %s", job.Name, correlationId)
	}

	return nil
}

func isEndTimestampExpired(endTimestamp int64) bool {
	durationUntilEnd := time.Until(time.UnixMilli(endTimestamp))
	return endTimestamp != 0 && durationUntilEnd < 0
}

// generateMissedRecords generates missed schedule action records
func generateMissedRecords(ctx context.Context, job models.ScheduleJob, dic *di.Container) (err errors.EdgeX, hasMissedAction bool) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	if job.AdminState != models.Unlocked {
		lc.Debugf("The scheduled job: %s is locked, skip generating missed schedule action records. ScheduleJob ID: %s, Correlation-ID: %s", job.Name, job.Id, correlationId)
		return nil, hasMissedAction
	}

	// Get the latest schedule action records by job name and generate missed schedule action records
	latestRecords, err := dbClient.LatestScheduleActionRecordsByJobName(ctx, job.Name)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to load the latest schedule action records of job: %s", job.Name), err), hasMissedAction
	}
	err, hasMissedAction = GenerateMissedScheduleActionRecords(ctx, dic, job, latestRecords)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err), hasMissedAction
	}

	return nil, hasMissedAction
}
