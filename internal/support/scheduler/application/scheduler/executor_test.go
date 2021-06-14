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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

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
		expectedErrorKind errors.ErrKind
	}{
		{"run with interval", "midnight", "20000101T000000", "22000101T000000", "24h", ""},
		{"run without startTime ", "midnight", "", "22000101T000000", "24h", ""},
		{"wrong startTime string format", "midnight", "20000101T", "", "24h", errors.KindContractInvalid},
		{"wrong endTime string format", "midnight", "", "20000101T", "24h", errors.KindContractInvalid},
		{"wrong frequency string format", "midnight", "", "", "24", errors.KindContractInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			current := time.Now()
			interval := models.Interval{
				Name:  testCase.intervalName,
				Start: testCase.startTime, End: testCase.endTime,
				Interval: testCase.interval,
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
			assert.GreaterOrEqual(t, executor.NextTime.Unix(), current.Unix())
			assert.Greater(t, executor.Frequency.Seconds(), float64(0))
			assert.False(t, executor.IsComplete())
		})
	}
}
