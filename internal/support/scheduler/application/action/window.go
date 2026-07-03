//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// point is a (month, day) pair used to compare calendar positions within a year, ignoring the year.
type point struct{ month, day int }

// less reports whether a comes strictly before b in the calendar, comparing month first then day.
func less(a, b point) bool {
	if a.month != b.month {
		return a.month < b.month
	}
	return a.day < b.day
}

// InWindow reports whether now falls inside the recurring month/day window. EndDay is inclusive; a start
// after end (e.g. Nov 15 -> Feb 10) is a year-crossing window.
func InWindow(now time.Time, window models.ActiveYearlyTimeWindow) bool {
	cur := point{int(now.Month()), now.Day()}
	start := point{window.StartMonth, window.StartDay}
	end := point{window.EndMonth, window.EndDay}

	if !less(end, start) { // start <= end -> same-year window
		return !less(cur, start) && !less(end, cur) // [start, end], end inclusive
	}
	return !less(cur, start) || !less(end, cur) // year-crossing window
}
