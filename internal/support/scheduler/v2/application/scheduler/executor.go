//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

const (
	TIMELAYOUT = "20060102T150405"
)

type Executor struct {
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

// Initialize initialize the Executor with interval. This function should be invoked after adding or updating the interval.
func (executor *Executor) Initialize(interval models.Interval, lc logger.LoggingClient) {
	executor.Interval = interval

	// run times, current and max iteration
	if executor.Interval.RunOnce {
		executor.MaxIterations = 1
	} else {
		executor.MaxIterations = 0
	}
	executor.CurrentIterations = 0

	// start and end time
	if executor.Interval.Start == "" {
		executor.StartTime = time.Now()
	} else {
		t, err := time.Parse(TIMELAYOUT, executor.Interval.Start)
		if err != nil {
			lc.Error("parse time error, the original time string is : " + executor.Interval.Start)
		}

		executor.StartTime = t
	}

	if executor.Interval.End == "" {
		// use max time
		executor.EndTime = time.Unix(1<<63-62135596801, 999999999)
	} else {
		t, err := time.Parse(TIMELAYOUT, executor.Interval.End)
		if err != nil {
			lc.Error("parse time error, the original time string is : " + executor.Interval.End)
		}

		executor.EndTime = t
	}

	// frequency and next time
	nowBenchmark := time.Now().Unix()
	if !executor.Interval.RunOnce {
		frequency, err := time.ParseDuration(executor.Interval.Frequency)
		if err != nil {
			lc.Error("interval parse frequency error  %v", err.Error())
		}
		executor.Frequency = frequency
	}

	executor.NextTime = executor.StartTime
	if executor.StartTime.Unix() <= nowBenchmark && !executor.Interval.RunOnce {
		for executor.NextTime.Unix() <= nowBenchmark {
			executor.NextTime = executor.NextTime.Add(executor.Frequency)
		}
	}
}

func (executor *Executor) IsComplete() bool {
	return executor.isComplete(time.Now())
}

func (executor *Executor) UpdateIterations() {
	if !executor.IsComplete() {
		executor.CurrentIterations += 1
	}
}

func (executor *Executor) UpdateNextTime() {
	if !executor.IsComplete() {
		executor.NextTime = executor.NextTime.Add(executor.Frequency)
	}
}

func (executor *Executor) isComplete(time time.Time) bool {
	complete := (executor.StartTime.Unix() < time.Unix() && executor.Interval.RunOnce) ||
		(executor.NextTime.Unix() > executor.EndTime.Unix()) ||
		((executor.MaxIterations != 0) && (executor.CurrentIterations >= executor.MaxIterations))
	return complete
}
