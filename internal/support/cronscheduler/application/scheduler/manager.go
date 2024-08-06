//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"fmt"
	"sync"

	"github.com/go-co-op/gocron/v2"

	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/edgex-go/internal/support/cronscheduler/application/scheduler/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/cronscheduler/infrastructure/interfaces"
	// TODO: import from internal/support/cronscheduler/config if available
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
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
func NewManager(lc logger.LoggingClient, dic *di.Container, config *config.ConfigurationStruct, secretProvider bootstrapInterfaces.SecretProviderExt) interfaces.SchedulerManager {
	return &manager{
		lc:             lc,
		dic:            dic,
		config:         config,
		schedulers:     make(map[string]gocron.Scheduler),
		secretProvider: secretProvider,
	}
}

// AddScheduleJob adds a new ScheduleJob to the scheduler manager
func (m *manager) AddScheduleJob(job models.ScheduleJob) errors.EdgeX {
	if _, err := m.getSchedulerByJobName(job.Name); err == nil {
		return errors.NewCommonEdgeX(errors.KindStatusConflict,
			fmt.Sprintf("the schedule job with name: %s already exists", job.Name), nil)
	}

	if err := m.addNewJob(job); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	m.lc.Infof("New schedule job %s was added into the scheduler manager", job.Name)
	return nil
}

// UpdateScheduleJob updates a ScheduleJob in the scheduler manager
func (m *manager) UpdateScheduleJob(job models.ScheduleJob) errors.EdgeX {
	_, err := m.getSchedulerByJobName(job.Name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// Remove the old job from gocron
	if err := m.DeleteScheduleJobByName(job.Name); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	m.lc.Debugf("The old schedule job %s was removed from the scheduler manager while updating it", job.Name)

	if err := m.addNewJob(job); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	m.lc.Debugf("Schedule job %s was updated into the scheduler manager", job.Name)
	return nil
}

// DeleteScheduleJobByName deletes all the actions of a ScheduleJob by name from the scheduler manager
func (m *manager) DeleteScheduleJobByName(name string) errors.EdgeX {
	scheduler, err := m.getSchedulerByJobName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	if err := scheduler.Shutdown(); err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to shutdown and delete the scheduler for job: %s", name), err)
	}

	m.lc.Debugf("The schedule job %s was stopped and removed from the scheduler manager", name)
	return nil
}

// StartScheduleJobByName starts all the actions of a ScheduleJob by name in the scheduler manager
func (m *manager) StartScheduleJobByName(name string) errors.EdgeX {
	scheduler, err := m.getSchedulerByJobName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	scheduler.Start()
	m.lc.Debugf("The schedule job %s was started", name)
	return nil
}

// StopScheduleJobByName stops all the actions of a ScheduleJob by name in the scheduler manager
func (m *manager) StopScheduleJobByName(name string) errors.EdgeX {
	scheduler, err := m.getSchedulerByJobName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	if err := scheduler.StopJobs(); err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to stop all the actions for job: %s", name), err)
	}

	m.lc.Debugf("The schedule job %s was stopped in the scheduler manager", name)
	return nil
}

// TriggerScheduleJobByName triggers all the actions of a ScheduleJob by name in the scheduler manager
func (m *manager) TriggerScheduleJobByName(name string) errors.EdgeX {
	scheduler, err := m.getSchedulerByJobName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	for _, job := range scheduler.Jobs() {
		if err := job.RunNow(); err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to trigger scheduler action for job: %s", name), err)
		}
	}

	m.lc.Debugf("The schedule job %s has been triggerred manually", name)
	return nil
}

// Shutdown stops all the schedule jobs and removes them from the scheduler manager
func (m *manager) Shutdown() errors.EdgeX {
	for name := range m.schedulers {
		if err := m.DeleteScheduleJobByName(name); err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	m.lc.Debugf("All schedule jobs were stopped and removed from the scheduler manager")
	return nil
}

func (m *manager) getSchedulerByJobName(name string) (gocron.Scheduler, errors.EdgeX) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	scheduler, exists := m.schedulers[name]
	if !exists {
		return nil, errors.NewCommonEdgeX(errors.KindStatusConflict,
			fmt.Sprintf("the schedule job: %s does not exist", name), nil)
	}
	return scheduler, nil
}

func (m *manager) addNewJob(job models.ScheduleJob) errors.EdgeX {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to initialize a new scheduler for job: %s", job.Name), err)
	}

	definition, edgeXerr := utils.ToGocronJobDef(job.Definition)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	for _, action := range job.Actions {
		task, edgeXerr := utils.ToGocronTask(m.lc, m.dic, m.secretProvider, action)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}

		// A "ScheduleAction" will be treated as a "Job" in gocron scheduler
		_, err := scheduler.NewJob(definition, task)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError,
				fmt.Sprintf("failed to create new schedule aciton for job: %s", job.Name), err)
		}
	}

	m.mu.Lock()
	m.schedulers[job.Name] = scheduler
	m.mu.Unlock()

	return nil
}
