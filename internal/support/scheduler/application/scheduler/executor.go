//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

const (
	SchedulerTimeFormat = "20060102T150405"
)

type Executor struct {
	Interval           models.Interval
	IntervalActionsMap map[string]models.IntervalAction
	StartTime          time.Time
	EndTime            time.Time
	NextTime           time.Time
	Frequency          time.Duration
	MarkedDeleted      bool
}

// Initialize initialize the Executor with interval. This function should be invoked after adding or updating the interval.
func (executor *Executor) Initialize(interval models.Interval, lc logger.LoggingClient) errors.EdgeX {
	executor.Interval = interval
	currentTime := time.Now()
	loc := currentTime.Location()

	// start and end time
	if executor.Interval.Start == "" {
		executor.StartTime = currentTime
	} else {
		t, err := time.ParseInLocation(SchedulerTimeFormat, executor.Interval.Start, loc)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("fail to parse the StartTime string %s", executor.Interval.End), err)
		}
		executor.StartTime = t
	}

	if executor.Interval.End == "" {
		// use max time
		executor.EndTime = time.Unix(1<<63-62135596801, 999999999)
	} else {
		t, err := time.ParseInLocation(SchedulerTimeFormat, executor.Interval.End, loc)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("fail to parse the EndTime string %s", executor.Interval.End), err)
		}
		executor.EndTime = t
	}

	frequency, err := time.ParseDuration(executor.Interval.Interval)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "interval parse frequency error", err)
	}
	executor.Frequency = frequency

	executor.NextTime = executor.StartTime
	// Increase the NextTime by interval frequency when NextTime small than the CurrentTime
	nowBenchmark := currentTime.Unix()
	for executor.NextTime.Unix() <= nowBenchmark {
		executor.NextTime = executor.NextTime.Add(executor.Frequency)
	}
	return nil
}

// IsComplete checks whether the Executor is complete
func (executor *Executor) IsComplete() bool {
	expired := executor.NextTime.Unix() > executor.EndTime.Unix()
	return expired
}

// UpdateNextTime increase the NextTime by frequency if the Executor not complete
func (executor *Executor) UpdateNextTime() {
	if !executor.IsComplete() {
		executor.NextTime = executor.NextTime.Add(executor.Frequency)
	}
}
