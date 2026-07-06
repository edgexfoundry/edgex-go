//
// Copyright (C) 2024-2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/application/action"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces/mocks"
)

var (
	lastRun = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
)

func TestFindMissedIntervalRuns(t *testing.T) {
	// Take the one-hour interval as an example
	interval := time.Hour

	tests := []struct {
		name        string
		lastRun     time.Time
		currentTime time.Time
		interval    time.Duration
		want        []time.Time
	}{
		{
			"Given current time is 50 minutes after last run time, expect no missed runs",
			lastRun,
			lastRun.Add(time.Minute * 50),
			interval,
			nil,
		},
		{
			"Given current time is 1 hour and ten minutes after last run time, expect 1 missed run",
			lastRun,
			lastRun.Add(time.Hour + time.Minute*10),
			interval,
			[]time.Time{lastRun.Add(time.Hour * 1)}},
		{
			"Given current time is 2 hour and ten minutes after last run time, expect 2 missed runs",
			lastRun,
			lastRun.Add(time.Hour*2 + time.Minute*10),
			interval,
			[]time.Time{
				lastRun.Add(time.Hour * 1),
				lastRun.Add(time.Hour * 2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findMissedIntervalRuns(tt.lastRun, tt.currentTime, tt.interval)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindMissedCronRuns(t *testing.T) {
	// Take the "0 * * * *" as an example, which means the job will run every hour
	cronSchedule, _ := action.ParseCronExpression("0 * * * *")

	tests := []struct {
		name         string
		lastRun      time.Time
		currentTime  time.Time
		cronSchedule cron.Schedule
		want         []time.Time
	}{
		{
			"Given current time is 50 minutes after last run time, expect no missed runs",
			lastRun,
			lastRun.Add(time.Minute * 50),
			cronSchedule,
			nil,
		},
		{
			"Given current time is 1 hour and ten minutes after last run time, expect 1 missed run",
			lastRun,
			lastRun.Add(time.Hour + time.Minute*10),
			cronSchedule,
			[]time.Time{lastRun.Add(time.Hour * 1)}},
		{
			"Given current time is 2 hour and ten minutes after last run time, expect 2 missed runs",
			lastRun,
			lastRun.Add(time.Hour*2 + time.Minute*10),
			cronSchedule,
			[]time.Time{
				lastRun.Add(time.Hour * 1),
				lastRun.Add(time.Hour * 2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findMissedCronRuns(tt.lastRun, tt.currentTime, tt.cronSchedule)
			assert.Equal(t, tt.want, got)
		})
	}
}

// captureAddedRecords runs GenerateMissedScheduleActionRecords with a mock DBClient that records the
// missed-record slice passed to AddScheduleActionRecords, and returns that slice.
func captureAddedRecords(t *testing.T, job models.ScheduleJob, latest []models.ScheduleActionRecord) []models.ScheduleActionRecord {
	t.Helper()
	ctx := context.Background()
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})

	var added []models.ScheduleActionRecord
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AddScheduleActionRecords", ctx, mock.AnythingOfType("[]models.ScheduleActionRecord")).
		Return([]models.ScheduleActionRecord{}, nil).
		Run(func(args mock.Arguments) {
			added = args.Get(1).([]models.ScheduleActionRecord)
		})
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	edgeXErr, _ := GenerateMissedScheduleActionRecords(ctx, dic, job, latest)
	require.Nil(t, edgeXErr)
	return added
}

func TestGenerateMissedScheduleActionRecords_ActiveYearlyTimeWindow(t *testing.T) {
	// A daily interval whose latest run is ~120 days ago, so the missed runs span ~120 daily points
	// crossing several months. This guarantees the set contains both in-window and out-of-window dates
	// for any wall-clock "now", keeping the assertions below independent of time.Now().
	latestTime := time.Now().AddDate(0, 0, -120)
	restAction := models.RESTAction{
		BaseScheduleAction: models.BaseScheduleAction{Id: "action-1", Type: "REST"},
	}
	latest := []models.ScheduleActionRecord{
		{JobName: "job", Action: restAction, ScheduledAt: latestTime.UnixMilli()},
	}
	newJob := func(window *models.ActiveYearlyTimeWindow) models.ScheduleJob {
		return models.ScheduleJob{
			Name: "job",
			Definition: models.IntervalScheduleDef{
				BaseScheduleDef: models.BaseScheduleDef{Type: "INTERVAL", ActiveYearlyTimeWindow: window},
				Interval:        "24h",
			},
		}
	}

	// Baseline: no window → every missed run is recorded.
	baseline := captureAddedRecords(t, newJob(nil), latest)
	require.NotEmpty(t, baseline, "baseline should produce missed runs to make the comparison meaningful")

	// A one-month window (March). Only missed runs whose date falls in March survive the filter.
	window := &models.ActiveYearlyTimeWindow{StartMonth: 3, StartDay: 1, EndMonth: 3, EndDay: 31}
	filtered := captureAddedRecords(t, newJob(window), latest)

	// The window must drop at least some runs (120 daily points cannot all be in March).
	assert.Less(t, len(filtered), len(baseline), "windowed run should record fewer than the unfiltered baseline")

	// Every recorded run must fall inside the window, and its status must be Missed.
	for _, r := range filtered {
		runTime := time.UnixMilli(r.ScheduledAt).In(time.Local)
		assert.True(t, action.InWindow(runTime, *window),
			"recorded missed run %s is outside the window and should have been dropped", runTime)
		assert.Equal(t, models.Missed, string(r.Status))
	}

	// Conversely, none of the baseline runs outside the window should appear in the filtered set.
	inFiltered := make(map[int64]bool, len(filtered))
	for _, r := range filtered {
		inFiltered[r.ScheduledAt] = true
	}
	for _, r := range baseline {
		if !action.InWindow(time.UnixMilli(r.ScheduledAt).In(time.Local), *window) {
			assert.False(t, inFiltered[r.ScheduledAt], "an out-of-window run leaked into the filtered set")
		}
	}
}

func TestPurgeRecord(t *testing.T) {
	ctx := context.Background()
	configuration := &config.ConfigurationStruct{
		Retention: config.RecordRetention{
			Enabled:  true,
			Interval: "1s",
			MaxCap:   5,
			MinCap:   3,
		},
	}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})

	tests := []struct {
		name        string
		recordCount int64
	}{
		{"invoke schedule action record purging", int64(configuration.Retention.MaxCap)},
		{"not invoke schedule action record purging", int64(configuration.Retention.MinCap)},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			dbClientMock := &dbMock.DBClient{}
			record := models.ScheduleActionRecord{}
			dbClientMock.On("LatestScheduleActionRecordsByOffset", ctx, configuration.Retention.MinCap).Return(record, nil)
			dbClientMock.On("ScheduleActionRecordTotalCount", ctx, int64(0), mock.AnythingOfType("int64")).Return(testCase.recordCount, nil)
			dbClientMock.On("DeleteScheduleActionRecordByAge", ctx, mock.AnythingOfType("int64")).Return(nil)
			dic.Update(di.ServiceConstructorMap{
				container.DBClientInterfaceName: func(get di.Get) interface{} {
					return dbClientMock
				},
			})
			err := purgeRecord(ctx, dic)
			require.NoError(t, err)
			if testCase.recordCount >= int64(configuration.Retention.MaxCap) {
				dbClientMock.AssertCalled(t, "DeleteScheduleActionRecordByAge", ctx, mock.AnythingOfType("int64"))
			} else {
				dbClientMock.AssertNotCalled(t, "DeleteScheduleActionRecordByAge", ctx, mock.AnythingOfType("int64"))
			}
		})
	}
}
