//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/stretchr/testify/assert"
)

// date builds a time.Time at the given month/day (year and time-of-day are irrelevant to InWindow).
func date(month, day int) time.Time {
	return time.Date(2025, time.Month(month), day, 12, 0, 0, 0, time.Local)
}

func TestInWindow(t *testing.T) {
	sameYear := models.ActiveYearlyTimeWindow{StartMonth: 1, StartDay: 5, EndMonth: 3, EndDay: 12}
	yearCrossing := models.ActiveYearlyTimeWindow{StartMonth: 11, StartDay: 15, EndMonth: 2, EndDay: 10}
	sameMonthCrossing := models.ActiveYearlyTimeWindow{StartMonth: 12, StartDay: 10, EndMonth: 12, EndDay: 1}
	singleDay := models.ActiveYearlyTimeWindow{StartMonth: 6, StartDay: 15, EndMonth: 6, EndDay: 15}

	tests := []struct {
		name     string
		window   models.ActiveYearlyTimeWindow
		t        time.Time
		expected bool
	}{
		// same-year window [1/5, 3/12]
		{"same-year inside", sameYear, date(2, 20), true},
		{"same-year on start", sameYear, date(1, 5), true},
		{"same-year on end (inclusive)", sameYear, date(3, 12), true},
		{"same-year day before start", sameYear, date(1, 4), false},
		{"same-year day after end", sameYear, date(3, 13), false},
		{"same-year far before", sameYear, date(1, 1), false},
		{"same-year far after", sameYear, date(12, 31), false},

		// year-crossing window [11/15, 2/10]
		{"crossing inside late-year", yearCrossing, date(12, 25), true},
		{"crossing inside early-year", yearCrossing, date(1, 20), true},
		{"crossing on start", yearCrossing, date(11, 15), true},
		{"crossing on end (inclusive)", yearCrossing, date(2, 10), true},
		{"crossing day before start", yearCrossing, date(11, 14), false},
		{"crossing day after end", yearCrossing, date(2, 11), false},
		{"crossing between months (summer)", yearCrossing, date(7, 1), false},

		// same-month year-crossing window [12/10, 12/1]: start > end within one month
		{"same-month crossing on start", sameMonthCrossing, date(12, 10), true},
		{"same-month crossing on end (inclusive)", sameMonthCrossing, date(12, 1), true},
		{"same-month crossing late-year inside", sameMonthCrossing, date(12, 20), true},
		{"same-month crossing early-year inside", sameMonthCrossing, date(1, 15), true},
		{"same-month crossing in the gap", sameMonthCrossing, date(12, 5), false},

		// single-day window [6/15, 6/15]
		{"single-day on the day", singleDay, date(6, 15), true},
		{"single-day day before", singleDay, date(6, 14), false},
		{"single-day day after", singleDay, date(6, 16), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, InWindow(tt.t, tt.window))
		})
	}
}
