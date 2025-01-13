//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/application/action"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces"
)

const (
	// validationTag is the tag used to validate the ScheduleJob internally and then remove those jobs from the scheduler by this tag
	validationTag = "::validation::"
)

type manager struct {
	lc             logger.LoggingClient
	dic            *di.Container
	config         *config.ConfigurationStruct
	mu             sync.RWMutex
	schedulers     map[string]gocron.Scheduler
	secretProvider bootstrapInterfaces.SecretProviderExt
}

// NewManager creates a new scheduler manager for running the ScheduleJob
func NewManager(dic *di.Container) interfaces.SchedulerManager {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	secretProvider := bootstrapContainer.SecretProviderExtFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	return &manager{
		lc:             lc,
		dic:            dic,
		config:         configuration,
		schedulers:     make(map[string]gocron.Scheduler),
		secretProvider: secretProvider,
	}
}

// AddScheduleJob adds a new ScheduleJob to the scheduler manager
func (m *manager) AddScheduleJob(job models.ScheduleJob, correlationId string) errors.EdgeX {
	if _, err := m.getSchedulerByJobName(job.Name); err == nil {
		return errors.NewCommonEdgeX(errors.KindStatusConflict,
			fmt.Sprintf("the scheduled job with name: %s already exists", job.Name), nil)
	}

	if err := m.addNewJob(job); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	m.lc.Infof("New scheduled job %s was added into the scheduler manager. ScheduleJob ID: %s, Correlation-ID: %s", job.Name, job.Id, correlationId)
	return nil
}

// UpdateScheduleJob updates a ScheduleJob in the scheduler manager
func (m *manager) UpdateScheduleJob(job models.ScheduleJob, correlationId string) errors.EdgeX {
	// Validate the ScheduleJob before updating it
	if err := m.ValidateUpdatingScheduleJob(job); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to validate the scheduled job", err)
	}

	// Remove the old job from gocron
	if err := m.DeleteScheduleJobByName(job.Name, correlationId); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	m.lc.Debugf("The old scheduled job %s was removed from the scheduler manager while updating it. ScheduleJob ID: %s, Correlation-ID: %s", job.Name, job.Id, correlationId)

	// Create a new job with the updated ScheduleJob
	if err := m.addNewJob(job); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	m.lc.Debugf("Scheduled job %s was updated into the scheduler manager. ScheduleJob ID: %s, Correlation-ID: %s", job.Name, job.Id, correlationId)
	return nil
}

// DeleteScheduleJobByName deletes all the actions of a ScheduleJob by name from the scheduler manager
func (m *manager) DeleteScheduleJobByName(name, correlationId string) errors.EdgeX {
	scheduler, err := m.getSchedulerByJobName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	if err := scheduler.Shutdown(); err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to shutdown and delete the scheduler for job: %s", name), err)
	}

	m.mu.Lock()
	delete(m.schedulers, name)
	m.mu.Unlock()

	m.lc.Debugf("The scheduled job %s was stopped and removed from the scheduler manager. Correlation-ID: %s", name, correlationId)
	return nil
}

// StartScheduleJobByName starts all the actions of a ScheduleJob by name in the scheduler manager
func (m *manager) StartScheduleJobByName(name, correlationId string) errors.EdgeX {
	scheduler, err := m.getSchedulerByJobName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	scheduler.Start()
	m.lc.Debugf("The scheduled job %s was started. Correlation-ID: %s", name, correlationId)
	return nil
}

// StopScheduleJobByName stops all the actions of a ScheduleJob by name in the scheduler manager
func (m *manager) StopScheduleJobByName(name, correlationId string) errors.EdgeX {
	scheduler, err := m.getSchedulerByJobName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	if err := scheduler.StopJobs(); err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to stop all the actions for job: %s", name), err)
	}

	m.lc.Debugf("The scheduled job %s was stopped in the scheduler manager. Correlation-ID: %s", name, correlationId)
	return nil
}

// TriggerScheduleJobByName triggers all the actions of a ScheduleJob by name in the scheduler manager
func (m *manager) TriggerScheduleJobByName(name, correlationId string) errors.EdgeX {
	scheduler, err := m.getSchedulerByJobName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	for _, job := range scheduler.Jobs() {
		if err := job.RunNow(); err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to trigger scheduler action for job: %s", name), err)
		}
	}

	m.lc.Debugf("The scheduled job %s has been triggerred manually. Correlation-ID: %s", name, correlationId)
	return nil
}

// Shutdown stops all the schedule jobs and removes them from the scheduler manager
func (m *manager) Shutdown(correlationId string) errors.EdgeX {
	for name := range m.schedulers {
		if err := m.DeleteScheduleJobByName(name, correlationId); err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	m.mu.Lock()
	m.schedulers = make(map[string]gocron.Scheduler)
	m.mu.Unlock()

	m.lc.Debugf("All scheduled jobs were stopped and removed from the scheduler manager. Correlation-ID: %s", correlationId)
	return nil
}

// ValidateUpdatingScheduleJob validates the ScheduleJob that will be updated, this function mainly checks the definition and actions of the ScheduleJob with gocron
func (m *manager) ValidateUpdatingScheduleJob(job models.ScheduleJob) errors.EdgeX {
	if job.Name == "" && job.Id == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name or ID is required", nil)
	}
	if job.Definition == nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "definition field is required", nil)
	}
	if len(job.Actions) == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "actions field is required", nil)
	}

	scheduler, err := m.getSchedulerByJobName(job.Name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	definition, edgeXerr := action.ToGocronJobDef(job.Definition)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	for _, a := range job.Actions {
		task, edgeXerr := action.ToGocronTask(m.lc, m.dic, m.secretProvider, a)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}

		// A "ScheduleAction" will be treated as a "Job" in gocron scheduler
		// Those "Jobs" will be created with a validation tag, and then removed from the scheduler if there is no error while creating them
		_, err := scheduler.NewJob(definition, task, gocron.WithTags(validationTag))
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError,
				fmt.Sprintf("failed to create scheduled job: %s", job.Name), err)
		}
	}

	// Remove the jobs for validation from the scheduler
	scheduler.RemoveByTags(validationTag)

	return nil
}

func (m *manager) getSchedulerByJobName(name string) (gocron.Scheduler, errors.EdgeX) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	scheduler, exists := m.schedulers[name]
	if !exists {
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist,
			fmt.Sprintf("the scheduled job: %s does not exist", name), nil)
	}
	return scheduler, nil
}

func (m *manager) addNewJob(job models.ScheduleJob) errors.EdgeX {
	ctx, correlationId := correlation.FromContextOrNew(context.Background())

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to initialize a new scheduler for job: %s", job.Name), err)
	}

	definition, edgeXerr := action.ToGocronJobDef(job.Definition)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	var jobOptions []gocron.JobOption

	// Add options for the scheduled job based on the startTimestamp and endTimestamp
	toTrigger, startOption, endOption := m.arrangeScheduleJob(ctx, job)
	if toTrigger {
		if startOption != nil {
			jobOptions = append(jobOptions, startOption)
		}
		if endOption != nil {
			jobOptions = append(jobOptions, endOption)
		}

		// If toTrigger is true, the ScheduleAction will be added to the scheduler and ready to be triggered
		for _, a := range job.Actions {
			copiedAction := a
			task, edgeXerr := action.ToGocronTask(m.lc, m.dic, m.secretProvider, a)
			if edgeXerr != nil {
				return errors.NewCommonEdgeXWrapper(edgeXerr)
			}

			// Add event listeners to the job options for recording the schedule action records
			jobOptions = append(jobOptions, gocron.WithEventListeners(
				gocron.AfterJobRuns(
					func(jobID uuid.UUID, jobName string) {
						gocronJob := getGocronJobByID(scheduler.Jobs(), jobID)
						lastRun, err := gocronJob.LastRun()
						if err != nil {
							m.lc.Errorf("failed to get the last run time for job: %s, Correlation-ID: %s, err: %v", job.Name, correlationId, err)
						}

						record := models.ScheduleActionRecord{
							JobName:     job.Name,
							Action:      copiedAction,
							Status:      models.Succeeded,
							ScheduledAt: lastRun.UnixMilli(),
						}
						m.addScheduleActionRecord(ctx, record, nil)
					}),
				gocron.AfterJobRunsWithError(
					func(jobID uuid.UUID, jobName string, err error) {
						gocronJob := getGocronJobByID(scheduler.Jobs(), jobID)
						lastRun, timeErr := gocronJob.LastRun()
						if timeErr != nil {
							m.lc.Errorf("failed to get the last run time for job: %s, Correlation-ID: %s, err: %v", job.Name, correlationId, timeErr)
						}

						record := models.ScheduleActionRecord{
							JobName:     job.Name,
							Action:      copiedAction,
							Status:      models.Failed,
							ScheduledAt: lastRun.UnixMilli(),
						}
						m.addScheduleActionRecord(ctx, record, err)
					}),
			))

			// A "ScheduleAction" will be treated as a "Job" in gocron scheduler
			_, err := scheduler.NewJob(definition, task, jobOptions...)
			if err != nil {
				return errors.NewCommonEdgeX(errors.KindServerError,
					fmt.Sprintf("failed to create new scheduled aciton for job: %s", job.Name), err)
			}
		}

		scheduler.Start()
		m.lc.Debugf("The scheduled job %s was started. Correlation-ID: %s", job.Name, correlationId)
	}

	// Whether the job is going to be triggered or not, the scheduler will be added to the manager to sync with the database
	m.mu.Lock()
	m.schedulers[job.Name] = scheduler
	m.mu.Unlock()

	return nil
}

func (m *manager) addScheduleActionRecord(ctx context.Context, record models.ScheduleActionRecord, err error) {
	dbClient := container.DBClientFrom(m.dic.Get)
	correlationId := correlation.FromContext(ctx)

	newRecord, dbErr := dbClient.AddScheduleActionRecord(ctx, record)
	if dbErr != nil {
		m.lc.Errorf("failed to add a new schedule action record for job: %s, Correlation-ID: %s, err: %v", record.JobName, correlationId, dbErr)
	} else {
		if err != nil {
			m.lc.Debugf("A new schedule action record with type: %s and status: %s was added for job: %s, record ID: %s, action error: %v, Correlation-ID: %s",
				record.Action.GetBaseScheduleAction().Type, record.Status, record.JobName, newRecord.Id, err, correlationId)
		} else {
			m.lc.Debugf("A new schedule action record with type: %s and status: %s was added for job: %s, record ID: %s, Correlation-ID: %s",
				record.Action.GetBaseScheduleAction().Type, record.Status, record.JobName, newRecord.Id, correlationId)
		}
	}
}

func getGocronJobByID(jobs []gocron.Job, id uuid.UUID) gocron.Job {
	for _, j := range jobs {
		if j.ID() == id {
			return j
		}
	}
	return nil
}

// arrangeScheduleJob arranges the schedule job based on the startTimestamp and endTimestamp and return the corresponding job options for gocron
func (m *manager) arrangeScheduleJob(ctx context.Context, job models.ScheduleJob) (toTrigger bool, startOption, endOption gocron.JobOption) {
	correlationId := correlation.FromContext(ctx)
	toTrigger = false

	if job.AdminState != models.Unlocked {
		m.lc.Debugf("The scheduled job is ready but not started because the admin state is locked. ScheduleJob ID: %s, Correlation-ID: %s", job.Id, correlationId)
		return toTrigger, nil, nil
	}

	startTimestamp := job.Definition.GetBaseScheduleDef().StartTimestamp
	startTime := time.UnixMilli(startTimestamp)
	endTimestamp := job.Definition.GetBaseScheduleDef().EndTimestamp
	endTime := time.UnixMilli(endTimestamp)

	durationUntilStart := time.Until(time.UnixMilli(startTimestamp))
	durationUntilEnd := time.Until(time.UnixMilli(endTimestamp))
	isEndExpired := endTimestamp != 0 && durationUntilEnd < 0

	// If endTimestamp is set and expired, the scheduled job should not be triggered
	if isEndExpired {
		m.lc.Warnf("The endTimestamp is expired for the scheduled job: %s, which will not be started. Correlation-ID: %s", job.Name, correlationId)
		return toTrigger, nil, nil
	}

	// If startTimestamp is expired, the scheduled job should be started immediately
	if durationUntilStart < 0 {
		m.lc.Debugf("The startTimestamp is expired for the scheduled job: %s, which will be started immediately. Correlation-ID: %s", job.Name, correlationId)
		durationUntilStart = 0
	} else if durationUntilStart > 0 {
		m.lc.Debugf("The scheduled job: %s will be started at %v (timestamp: %v). Correlation-ID: %s", job.Name, startTime, startTimestamp, correlationId)
	}

	// Regardless of whether startTimestamp has a value or not, the job should always be started by default if endTimestamp is not expired.
	if durationUntilStart != 0 {
		startOption = gocron.WithStartAt(gocron.WithStartDateTime(startTime))
	}

	// If the endTimestamp is set and the duration until the end is greater than 0, the scheduled job will be stopped at the endTimestamp
	if endTimestamp != 0 && durationUntilEnd > 0 {
		m.lc.Debugf("The scheduled job: %s will be stopped at %v (timestamp: %v). Correlation-ID: %s", job.Name, endTime, endTimestamp, correlationId)
		endOption = gocron.WithStopAt(gocron.WithStopDateTime(endTime))
	}

	toTrigger = true
	return toTrigger, startOption, endOption
}
