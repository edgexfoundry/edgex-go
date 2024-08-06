//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

type SchedulerManager interface {
	AddScheduleJob(job models.ScheduleJob) errors.EdgeX
	UpdateScheduleJob(job models.ScheduleJob) errors.EdgeX
	DeleteScheduleJobByName(name string) errors.EdgeX
	StartScheduleJobByName(name string) errors.EdgeX
	StopScheduleJobByName(name string) errors.EdgeX
	TriggerScheduleJobByName(name string) errors.EdgeX

	Shutdown() errors.EdgeX
}
