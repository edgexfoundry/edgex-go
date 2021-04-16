//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
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
func (executor *Executor) Initialize(interval models.Interval, lc logger.LoggingClient) errors.EdgeX {
	executor.Interval = interval
	currentTime := time.Now()

	// run times, current and max iteration
	if executor.Interval.RunOnce {
		executor.MaxIterations = 1
	} else {
		executor.MaxIterations = 0
	}
	executor.CurrentIterations = 0

	// start and end time
	if executor.Interval.Start == "" {
		executor.StartTime = currentTime
	} else {
		t, err := time.Parse(TIMELAYOUT, executor.Interval.Start)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("fail to parse the StartTime string %s", executor.Interval.End), err)
		}
		executor.StartTime = t
	}

	if executor.Interval.End == "" {
		// use max time
		executor.EndTime = time.Unix(1<<63-62135596801, 999999999)
	} else {
		t, err := time.Parse(TIMELAYOUT, executor.Interval.End)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("fail to parse the EndTime string %s", executor.Interval.End), err)
		}
		executor.EndTime = t
	}

	executor.NextTime = executor.StartTime

	// Parse frequency when RunOnce is false because we can use frequency or runOnce but not both
	if !executor.Interval.RunOnce {
		frequency, err := time.ParseDuration(executor.Interval.Frequency)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "interval parse frequency error", err)
		}
		executor.Frequency = frequency

		// Increase the NextTime by interval frequency when NextTime small than the CurrentTime
		nowBenchmark := currentTime.Unix()
		for executor.NextTime.Unix() <= nowBenchmark {
			executor.NextTime = executor.NextTime.Add(executor.Frequency)
		}
	}
	return nil
}

// IsComplete checks whether the Executor is complete
func (executor *Executor) IsComplete() bool {
	return executor.isComplete(time.Now())
}

// UpdateIterations increase the CurrentIterations times if the Executor not complete
func (executor *Executor) UpdateIterations() {
	if !executor.IsComplete() {
		executor.CurrentIterations += 1
	}
}

// UpdateNextTime increase the NextTime by frequency if the Executor not complete
func (executor *Executor) UpdateNextTime() {
	if !executor.IsComplete() {
		executor.NextTime = executor.NextTime.Add(executor.Frequency)
	}
}

func (executor *Executor) isComplete(time time.Time) bool {
	ranOnce := executor.StartTime.Unix() < time.Unix() && executor.Interval.RunOnce
	expired := executor.NextTime.Unix() > executor.EndTime.Unix()
	iterationLimitReached := (executor.MaxIterations != 0) && (executor.CurrentIterations >= executor.MaxIterations)
	return ranOnce || expired || iterationLimitReached
}
