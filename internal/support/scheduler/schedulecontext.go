//
// Copyright (c) 2018 Tencent
//
// Copyright (c) 2017 Dell Inc
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"github.com/edgexfoundry/edgex-go/pkg/models"

	"regexp"
	"strconv"
	"time"
)

type IntervalContext struct {
	Interval          models.Interval
	IntervalActionsMap map[string]models.IntervalAction
	StartTime         time.Time
	EndTime           time.Time
	NextTime          time.Time
	Frequency         time.Duration
	CurrentIterations int64
	MaxIterations     int64
	MarkedDeleted     bool
}

func (sc *IntervalContext) Reset(interval models.Interval) {
	if sc.Interval != (models.Interval{}) && sc.Interval.Name != interval.Name {
		//if interval name has changed, we should clear the old actions map(here just renew one)
		sc.IntervalActionsMap = make(map[string]models.IntervalAction)
	}

	sc.Interval = interval

	//run times, current and max iteration
	if sc.Interval.RunOnce {
		sc.MaxIterations = 1
	} else {
		sc.MaxIterations = 0
	}
	sc.CurrentIterations = 0

	//start and end time
	if sc.Interval.Start == "" {
		sc.StartTime = time.Now()
	} else {
		t, err := time.Parse(TIMELAYOUT, sc.Interval.Start)
		if err != nil {
			LoggingClient.Error("parse time error, the original time string is : " + sc.Interval.Start)
		}

		sc.StartTime = t
	}

	if sc.Interval.End == "" {
		//use max time
		sc.EndTime = time.Unix(1<<63-62135596801, 999999999)
	} else {
		t, err := time.Parse(TIMELAYOUT, sc.Interval.End)
		if err != nil {
			LoggingClient.Error("parse time error, the original time string is : " + sc.Interval.End)
		}

		sc.EndTime = t
	}

	//frequency and next time
	nowBenchmark := time.Now().Unix()
	sc.Frequency = parseFrequency(sc.Interval.Frequency)

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

func parseFrequency(durationStr string) time.Duration {
	durationRegex := regexp.MustCompile(`P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+S)?`)
	matches := durationRegex.FindStringSubmatch(durationStr)

	years := parseInt64(matches[1])
	months := parseInt64(matches[2])
	days := parseInt64(matches[3])
	hours := parseInt64(matches[4])
	minutes := parseInt64(matches[5])
	seconds := parseInt64(matches[6])

	hour := int64(time.Hour)
	minute := int64(time.Minute)
	second := int64(time.Second)
	return time.Duration(years*24*365*hour + months*30*24*hour + days*24*hour + hours*hour + minutes*minute + seconds*second)
}

func parseInt64(value string) int64 {
	if len(value) == 0 {
		return 0
	}
	parsed, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return 0
	}
	return int64(parsed)
}
