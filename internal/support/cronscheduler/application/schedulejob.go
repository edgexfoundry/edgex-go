//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// TODO: Implement all the functions in this file

func AddScheduleJob(ctx context.Context, job models.ScheduleJob, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	return id, nil
}

func TriggerScheduleJobByName(ctx context.Context, name string, dic *di.Container) errors.EdgeX {
	return nil
}

func ScheduleJobByName(ctx context.Context, name string, dic *di.Container) (dto dtos.ScheduleJob, edgeXerr errors.EdgeX) {
	return dto, nil
}

func AllScheduleJobs(offset int, limit int, dic *di.Container) (scheduleJobDTOs []dtos.ScheduleJob, totalCount uint32, err errors.EdgeX) {
	return scheduleJobDTOs, totalCount, nil
}

func PatchScheduleJob(ctx context.Context, dto dtos.UpdateScheduleJob, dic *di.Container) errors.EdgeX {
	return nil
}

func DeleteScheduleJobByName(ctx context.Context, name string, dic *di.Container) errors.EdgeX {
	return nil
}
