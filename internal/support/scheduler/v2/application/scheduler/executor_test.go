//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	lc := logger.NewMockClient()

	tests := []struct {
		name              string
		intervalName      string
		startTime         string
		endTime           string
		interval          string
		runOnce           bool
		expectedErrorKind errors.ErrKind
	}{
		{"run once", "midnight", "20000101T000000", "", "24h", true, ""},
		{"run with interval", "midnight", "20000101T000000", "22000101T000000", "24h", false, ""},
		{"run without startTime ", "midnight", "", "22000101T000000", "24h", false, ""},
		{"wrong startTime string format", "midnight", "20000101T", "", "24h", false, errors.KindContractInvalid},
		{"wrong endTime string format", "midnight", "", "20000101T", "24h", false, errors.KindContractInvalid},
		{"wrong frequency string format", "midnight", "", "", "24", false, errors.KindContractInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			current := time.Now()
			interval := models.Interval{
				Name:  testCase.intervalName,
				Start: testCase.startTime, End: testCase.endTime,
				Interval: testCase.interval, RunOnce: testCase.runOnce,
			}
			executor := Executor{}

			err := executor.Initialize(interval, lc)
			if testCase.expectedErrorKind != "" {
				require.Equal(t, errors.Kind(err), testCase.expectedErrorKind)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, interval.Name, executor.Interval.Name)
			if interval.Start != "" {
				assert.Equal(t, interval.Start, executor.StartTime.Format("20060102T000000"))
			}
			if interval.End != "" {
				assert.Equal(t, interval.End, executor.EndTime.Format("20060102T000000"))
			}
			if interval.RunOnce {
				assert.LessOrEqual(t, executor.NextTime.Unix(), current.Unix())
				assert.EqualValues(t, executor.MaxIterations, 1)
				assert.EqualValues(t, executor.Frequency.Seconds(), 0)
				if interval.Start == "" {
					assert.Equal(t, executor.StartTime, executor.NextTime)
					assert.LessOrEqual(t, executor.NextTime.Unix(), current.Unix())
				}
				assert.True(t, executor.IsComplete())
			} else {
				assert.GreaterOrEqual(t, executor.NextTime.Unix(), current.Unix())
				assert.Greater(t, executor.Frequency.Seconds(), float64(0))
				assert.False(t, executor.IsComplete())
			}
		})
	}
}
