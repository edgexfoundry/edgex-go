//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// ScheduleActionRecordClient defines the interface for interactions with the ScheduleActionRecord endpoint on the EdgeX Foundry support-scheduler service.
type ScheduleActionRecordClient interface {
	// AllScheduleActionRecords query schedule action records with start, end, offset, and limit
	AllScheduleActionRecords(ctx context.Context, start, end int64, offset, limit int) (responses.MultiScheduleActionRecordsResponse, errors.EdgeX)
	// LatestScheduleActionRecordsByJobName query the latest schedule action records by job name
	LatestScheduleActionRecordsByJobName(ctx context.Context, jobName string) (responses.MultiScheduleActionRecordsResponse, errors.EdgeX)
	// ScheduleActionRecordsByStatus queries schedule action records with status, start, end, offset, and limit
	ScheduleActionRecordsByStatus(ctx context.Context, status string, start, end int64, offset, limit int) (responses.MultiScheduleActionRecordsResponse, errors.EdgeX)
	// ScheduleActionRecordsByJobName query schedule action records with jobName, start, end, offset, and limit
	ScheduleActionRecordsByJobName(ctx context.Context, jobName string, start, end int64, offset, limit int) (responses.MultiScheduleActionRecordsResponse, errors.EdgeX)
	// ScheduleActionRecordsByJobNameAndStatus query schedule action records with jobName, status, start, end, offset, and limit
	ScheduleActionRecordsByJobNameAndStatus(ctx context.Context, jobName, status string, start, end int64, offset, limit int) (responses.MultiScheduleActionRecordsResponse, errors.EdgeX)
}
