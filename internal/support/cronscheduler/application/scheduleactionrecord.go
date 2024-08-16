//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/support/cronscheduler/container"
)

// AllScheduleActionRecords query the schedule action records with the specified offset, limit, and time range
func AllScheduleActionRecords(ctx context.Context, start, end int64, offset, limit int, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	records, err := dbClient.AllScheduleActionRecords(ctx, start, end, offset, limit)
	if err == nil {
		totalCount, err = dbClient.ScheduleActionRecordTotalCount(ctx)
	}
	if err != nil {
		return scheduleActionRecordDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}

	scheduleActionRecordDTOs = fromScheduleActionRecordModelsToDTOs(records)
	return scheduleActionRecordDTOs, totalCount, nil
}

// ScheduleActionRecordsByStatus query the schedule action records with the specified status, offset, limit, and time range
func ScheduleActionRecordsByStatus(ctx context.Context, status string, start, end int64, offset, limit int, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	records, err := dbClient.ScheduleActionRecordsByStatus(ctx, status, start, end, offset, limit)
	if err == nil {
		totalCount, err = dbClient.ScheduleActionRecordCountByStatus(ctx, status)
	}
	if err != nil {
		return scheduleActionRecordDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}

	scheduleActionRecordDTOs = fromScheduleActionRecordModelsToDTOs(records)
	return scheduleActionRecordDTOs, totalCount, nil
}

// ScheduleActionRecordsByJobName query the schedule action records with the specified job name, offset, limit, and time range
func ScheduleActionRecordsByJobName(ctx context.Context, jobName string, start, end int64, offset, limit int, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	records, err := dbClient.ScheduleActionRecordsByJobName(ctx, jobName, start, end, offset, limit)
	if err == nil {
		totalCount, err = dbClient.ScheduleActionRecordCountByJobName(ctx, jobName)
	}
	if err != nil {
		return scheduleActionRecordDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}

	scheduleActionRecordDTOs = fromScheduleActionRecordModelsToDTOs(records)
	return scheduleActionRecordDTOs, totalCount, nil
}

// ScheduleActionRecordsByJobNameAndStatus query the schedule action records with the specified job name, status, offset, limit, and time range
func ScheduleActionRecordsByJobNameAndStatus(ctx context.Context, jobName, status string, start, end int64, offset, limit int, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	records, err := dbClient.ScheduleActionRecordsByJobNameAndStatus(ctx, jobName, status, start, end, offset, limit)
	if err == nil {
		totalCount, err = dbClient.ScheduleActionRecordCountByJobNameAndStatus(ctx, jobName, status)
	}
	if err != nil {
		return scheduleActionRecordDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}

	scheduleActionRecordDTOs = fromScheduleActionRecordModelsToDTOs(records)
	return scheduleActionRecordDTOs, totalCount, nil
}

// LatestScheduleActionRecords query the latest schedule action records with the specified offset and limit
func LatestScheduleActionRecords(ctx context.Context, offset, limit int, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	records, err := dbClient.LatestScheduleActionRecords(ctx, offset, limit)
	if err == nil {
		totalCount, err = dbClient.LatestScheduleActionRecordTotalCount(ctx)
	}
	if err != nil {
		return scheduleActionRecordDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}

	scheduleActionRecordDTOs = fromScheduleActionRecordModelsToDTOs(records)
	return scheduleActionRecordDTOs, totalCount, nil
}

// DeleteScheduleActionRecordsByAge deletes the schedule action records by age
func DeleteScheduleActionRecordsByAge(ctx context.Context, age int64, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	err := dbClient.DeleteScheduleActionRecordByAge(ctx, age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("Successfully deleted the scheduled action record by age: %v. Correlation-ID: %s", age, correlationId)

	return nil
}

func fromScheduleActionRecordModelsToDTOs(records []models.ScheduleActionRecord) []dtos.ScheduleActionRecord {
	scheduleActionRecordDTOs := make([]dtos.ScheduleActionRecord, len(records))
	for i, record := range records {
		scheduleActionRecordDTOs[i] = dtos.FromScheduleActionRecordModelToDTO(record)
	}
	return scheduleActionRecordDTOs
}
