//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// ScheduleJobClient defines the interface for interactions with the ScheduleJob endpoint on the EdgeX Foundry support-scheduler service.
type ScheduleJobClient interface {
	// Add adds new schedule jobs.
	Add(ctx context.Context, reqs []requests.AddScheduleJobRequest) ([]common.BaseWithIdResponse, errors.EdgeX)
	// Update updates schedule jobs.
	Update(ctx context.Context, reqs []requests.UpdateScheduleJobRequest) ([]common.BaseResponse, errors.EdgeX)
	// AllScheduleJobs returns all schedule jobs.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	AllScheduleJobs(ctx context.Context, labels []string, offset int, limit int) (responses.MultiScheduleJobsResponse, errors.EdgeX)
	// ScheduleJobByName returns a schedule job by name.
	ScheduleJobByName(ctx context.Context, name string) (responses.ScheduleJobResponse, errors.EdgeX)
	// DeleteScheduleJobByName deletes a schedule job by name.
	DeleteScheduleJobByName(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
	// TriggerScheduleJobByName triggers a schedule job by name.
	TriggerScheduleJobByName(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
}
