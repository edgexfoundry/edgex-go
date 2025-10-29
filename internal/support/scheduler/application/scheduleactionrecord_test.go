//
// Copyright (C) 2024 IOTech Ltd
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
	cronSchedule, _ := parseCronExpression("0 * * * *")

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
