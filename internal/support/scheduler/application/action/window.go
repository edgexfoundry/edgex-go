//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"time"

	"github.com/robfig/cron/v3"

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

// WindowLocation returns the timezone in which a job's window should be evaluated. For a CRON definition it
// is the schedule's own location (a TZ=/CRON_TZ= prefix, else time.Local), so the window matches the
// schedule's calendar day. Any non-CRON definition uses time.Local. An error is returned only when the
// crontab fails to parse, so the caller can reject the job rather than evaluate the window in the wrong zone.
func WindowLocation(def models.ScheduleDef) (*time.Location, error) {
	cronDef, ok := def.(models.CronScheduleDef)
	if !ok {
		return time.Local, nil
	}

	schedule, err := ParseCronExpression(cronDef.Crontab)
	if err != nil {
		return nil, err
	}
	if spec, ok := schedule.(*cron.SpecSchedule); ok && spec.Location != nil {
		return spec.Location, nil
	}
	return time.Local, nil
}
