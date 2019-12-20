//
// Copyright (c) 2018 Tencent
//
// Copyright (c) 2017 Dell Inc
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type IntervalContext struct {
	Interval           models.Interval
	IntervalActionsMap map[string]models.IntervalAction
	StartTime          time.Time
	EndTime            time.Time
	NextTime           time.Time
	Frequency          time.Duration
	CurrentIterations  int64
	MaxIterations      int64
	MarkedDeleted      bool
}

func (sc *IntervalContext) Reset(interval models.Interval, lc logger.LoggingClient) {
	if sc.Interval != (models.Interval{}) && sc.Interval.Name != interval.Name {
		// if interval name has changed, we should clear the old actions map(here just renew one)
		sc.IntervalActionsMap = make(map[string]models.IntervalAction)
	}

	sc.Interval = interval

	// run times, current and max iteration
	if sc.Interval.RunOnce {
		sc.MaxIterations = 1
	} else {
		sc.MaxIterations = 0
	}
	sc.CurrentIterations = 0

	// start and end time
	if sc.Interval.Start == "" {
		sc.StartTime = time.Now()
	} else {
		t, err := time.Parse(TIMELAYOUT, sc.Interval.Start)
		if err != nil {
			lc.Error("parse time error, the original time string is : " + sc.Interval.Start)
		}

		sc.StartTime = t
	}

	if sc.Interval.End == "" {
		// use max time
		sc.EndTime = time.Unix(1<<63-62135596801, 999999999)
	} else {
		t, err := time.Parse(TIMELAYOUT, sc.Interval.End)
		if err != nil {
			lc.Error("parse time error, the original time string is : " + sc.Interval.End)
		}

		sc.EndTime = t
	}

	// frequency and next time
	nowBenchmark := time.Now().Unix()
	if !sc.Interval.RunOnce {
		frequency, err := parseFrequency(sc.Interval.Frequency)
		if err != nil {
			lc.Error("interval parse frequency error  %v", err.Error())
		}
		sc.Frequency = frequency
	}

	sc.NextTime = sc.StartTime
	if sc.StartTime.Unix() <= nowBenchmark && !sc.Interval.RunOnce {
		for sc.NextTime.Unix() <= nowBenchmark {
			sc.NextTime = sc.NextTime.Add(sc.Frequency)
		}
	}
}

func (sc *IntervalContext) IsComplete() bool {
	return sc.isComplete(time.Now())
}

func (sc *IntervalContext) UpdateIterations() {
	if !sc.IsComplete() {
		sc.CurrentIterations += 1
	}
}

func (sc *IntervalContext) UpdateNextTime() {
	if !sc.IsComplete() {
		sc.NextTime = sc.NextTime.Add(sc.Frequency)
	}
}

func (sc *IntervalContext) GetInfo() string {
	return sc.Interval.String()
}

func (sc *IntervalContext) isComplete(time time.Time) bool {
	complete := (sc.StartTime.Unix() < time.Unix() && sc.Interval.RunOnce) ||
		(sc.NextTime.Unix() > sc.EndTime.Unix()) ||
		((sc.MaxIterations != 0) && (sc.CurrentIterations >= sc.MaxIterations))
	return complete
}
