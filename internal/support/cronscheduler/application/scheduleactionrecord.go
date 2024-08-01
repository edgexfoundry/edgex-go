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
)

// TODO: Implement all the functions in this file

func AllScheduleActionRecords(ctx context.Context, start, end int64, offset, limit int, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, err errors.EdgeX) {
	return scheduleActionRecordDTOs, totalCount, nil
}

func ScheduleActionRecordsByStatus(ctx context.Context, status string, start, end int64, offset, limit int, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, edgeXerr errors.EdgeX) {
	return scheduleActionRecordDTOs, totalCount, nil
}

func ScheduleActionRecordsByJobName(ctx context.Context, jobName string, start, end int64, offset, limit int, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, edgeXerr errors.EdgeX) {
	return scheduleActionRecordDTOs, totalCount, nil
}

func ScheduleActionRecordsByJobNameAndStatus(ctx context.Context, jobName, status string, start, end int64, offset, limit int, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, edgeXerr errors.EdgeX) {
	return scheduleActionRecordDTOs, totalCount, nil
}

func LatestScheduleActionRecords(ctx context.Context, offset, limit int, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, edgeXerr errors.EdgeX) {
	return scheduleActionRecordDTOs, totalCount, nil
}

func DeleteScheduleActionRecordsByAge(ctx context.Context, age int64, dic *di.Container) errors.EdgeX {
	return nil
}
