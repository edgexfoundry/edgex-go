//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
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

	scheduleActionRecordDTOs = dtos.FromScheduleActionRecordModelsToDTOs(records)
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

	scheduleActionRecordDTOs = dtos.FromScheduleActionRecordModelsToDTOs(records)
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

	scheduleActionRecordDTOs = dtos.FromScheduleActionRecordModelsToDTOs(records)
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

	scheduleActionRecordDTOs = dtos.FromScheduleActionRecordModelsToDTOs(records)
	return scheduleActionRecordDTOs, totalCount, nil
}

// LatestScheduleActionRecordsByJobName query the latest schedule action records by job name
func LatestScheduleActionRecordsByJobName(ctx context.Context, jobName string, dic *di.Container) (scheduleActionRecordDTOs []dtos.ScheduleActionRecord, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	if _, err := dbClient.ScheduleJobByName(ctx, jobName); err != nil {
		return scheduleActionRecordDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}

	records, err := dbClient.LatestScheduleActionRecordsByJobName(ctx, jobName)
	if err != nil {
		return scheduleActionRecordDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}

	scheduleActionRecordDTOs = dtos.FromScheduleActionRecordModelsToDTOs(records)
	return scheduleActionRecordDTOs, uint32(len(records)), nil
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

// GenerateMissedScheduleActionRecords generates missed schedule action records
func GenerateMissedScheduleActionRecords(ctx context.Context, dic *di.Container, job models.ScheduleJob, latestRecords []models.ScheduleActionRecord) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	for _, latestRecord := range latestRecords {
		actionId := latestRecord.Action.GetBaseScheduleAction().Id
		lastRecordTimestamp := latestRecord.ScheduledAt

		// Compare the last record timestamp with the job's modified timestamp to get the latest time
		latestTime := time.UnixMilli(lastRecordTimestamp)
		modified := time.UnixMilli(job.Modified)
		if latestTime.Before(modified) {
			latestTime = modified
		}

		// Generate missed runs based on the schedule type
		missedRuns, err := generateMissedRuns(job.Definition, latestTime)
		if err != nil {
			lc.Errorf("Failed to generate missed records of job: %s. Correlation-ID: %s", job.Name, correlationId)
			return errors.NewCommonEdgeXWrapper(err)
		}

		var missedRecords []models.ScheduleActionRecord
		if len(missedRuns) != 0 {
			for _, run := range missedRuns {
				actionRecord := models.ScheduleActionRecord{
					JobName:     job.Name,
					Action:      latestRecord.Action,
					Status:      models.Missed,
					ScheduledAt: run.UnixMilli(),
				}

				missedRecords = append(missedRecords, actionRecord)
				lc.Tracef("Missed schedule action record with action id: %s of job: %s have been generated successfully. Correlation-ID: %s", actionId, job.Name, correlationId)
			}

			if _, err := dbClient.AddScheduleActionRecords(ctx, missedRecords); err != nil {
				lc.Errorf("Failed to add missed schedule action records with action id: %s of job: %s to database. Correlation-ID: %s", actionId, job.Name, correlationId)
				return errors.NewCommonEdgeXWrapper(err)
			}

			lc.Debugf("Missed schedule action records with action id: %s of job: %s have been created successfully. Correlation-ID: %s", actionId, job.Name, correlationId)
		}
	}

	return nil
}

func generateMissedRuns(def models.ScheduleDef, latestTime time.Time) (missedRuns []time.Time, err errors.EdgeX) {
	currentTime := time.Now()

	switch def.GetBaseScheduleDef().Type {
	case common.DefCron:
		cronDef, ok := def.(models.CronScheduleDef)
		if !ok {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast ScheduleDefinition to CronScheduleDef", nil)
		}

		cronSchedule, err := parseCronExpression(cronDef.Crontab)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to parse cron expression", err)
		}

		missedRuns = findMissedCronRuns(latestTime, currentTime, cronSchedule)
	case common.DefInterval:
		def, ok := def.(models.IntervalScheduleDef)
		if !ok {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast ScheduleDefinition to IntervalScheduleDef", nil)
		}

		duration, err := time.ParseDuration(def.Interval)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to parse interval string to a duration time value", err)
		}

		missedRuns = findMissedIntervalRuns(latestTime, currentTime, duration)
	default:
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unsupported schedule definition type: %s", def.GetBaseScheduleDef().Type), nil)
	}

	return missedRuns, nil
}

func findMissedIntervalRuns(lastRun, current time.Time, interval time.Duration) (missedRuns []time.Time) {
	for t := lastRun.Add(interval); t.Before(current); t = t.Add(interval) {
		missedRuns = append(missedRuns, t)
	}
	return missedRuns
}

func findMissedCronRuns(lastRun, current time.Time, schedule cron.Schedule) (missedRuns []time.Time) {
	for t := schedule.Next(lastRun); t.Before(current); t = schedule.Next(t) {
		missedRuns = append(missedRuns, t)
	}
	return missedRuns
}

func parseCronExpression(cronExpr string) (cron.Schedule, error) {
	var withLocation string
	if strings.HasPrefix(cronExpr, "TZ=") || strings.HasPrefix(cronExpr, "CRON_TZ=") {
		withLocation = cronExpr
	} else {
		withLocation = fmt.Sprintf("CRON_TZ=%s %s", time.Local.String(), cronExpr)
	}

	// An optional 6th field is used at the beginning since withSeconds is set to true: `* * * * * *`
	p := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := p.Parse(withLocation)
	if err != nil {
		return nil, err
	}
	return schedule, nil
}
