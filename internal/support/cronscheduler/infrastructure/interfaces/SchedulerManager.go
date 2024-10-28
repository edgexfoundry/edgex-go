//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type SchedulerManager interface {
	AddScheduleJob(job models.ScheduleJob, correlationId string) errors.EdgeX
	UpdateScheduleJob(job models.ScheduleJob, correlationId string) errors.EdgeX
	DeleteScheduleJobByName(name, correlationId string) errors.EdgeX
	StartScheduleJobByName(name, correlationId string) errors.EdgeX
	StopScheduleJobByName(name, correlationId string) errors.EdgeX
	TriggerScheduleJobByName(name, correlationId string) errors.EdgeX
	ValidateUpdatingScheduleJob(job models.ScheduleJob) errors.EdgeX

	Shutdown(correlationId string) errors.EdgeX
}
