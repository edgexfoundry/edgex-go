//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
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
