//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/core/metadata"
	"regexp"
	"strconv"
	"time"
)

type ScheduleContext struct {
	Schedule          models.Schedule
	ScheduleEventsMap map[string]models.ScheduleEvent
	StartTime         time.Time
	EndTime           time.Time
	NextTime          time.Time
	Frequency         time.Duration
	CurrentIterations int64
	MaxIterations     int64
	MarkedDeleted     bool
}

func (sc *ScheduleContext) Reset(schedule models.Schedule) {
	if sc.Schedule != (models.Schedule{}) && sc.Schedule.Name != schedule.Name {
		//if schedule name has changed, we should clear the old events map(here just renew one)
		sc.ScheduleEventsMap = make(map[string]models.ScheduleEvent)
	}

	sc.Schedule = schedule

	//run times, current and max iteration
	if sc.Schedule.RunOnce {
		sc.MaxIterations = 1
	} else {
		sc.MaxIterations = 0
	}
	sc.CurrentIterations = 0

	//start and end time
	if sc.Schedule.Start == "" {
		sc.StartTime = time.Now()
	} else {
		t, err := time.Parse(metadata.TIMELAYOUT, sc.Schedule.Start)
		if err != nil {
			loggingClient.Error("parse time error, the original time string is : " + sc.Schedule.Start)
		}

		sc.StartTime = t
	}

	if sc.Schedule.End == "" {
		//use max time
		sc.EndTime = time.Unix(1<<63-62135596801, 999999999)
	} else {
		t, err := time.Parse(metadata.TIMELAYOUT, sc.Schedule.End)
		if err != nil {
			loggingClient.Error("parse time error, the original time string is : " + sc.Schedule.End)
		}

		sc.EndTime = t
	}

	//frequency and next time
	nowBenchmark := time.Now().Unix()
	sc.Frequency = parseFrequency(sc.Schedule.Frequency)

	sc.NextTime = sc.StartTime
	if sc.StartTime.Unix() <= nowBenchmark && !sc.Schedule.RunOnce {
		for sc.NextTime.Unix() <= nowBenchmark {
			sc.NextTime = sc.NextTime.Add(sc.Frequency)
		}
	}
}

func (sc *ScheduleContext) IsComplete() bool {
	return sc.isComplete(time.Now())
}

func (sc *ScheduleContext) UpdateIterations() {
	if !sc.IsComplete() {
		sc.CurrentIterations += 1
	}
}

func (sc *ScheduleContext) UpdateNextTime() {
	if !sc.IsComplete() {
		sc.NextTime = sc.NextTime.Add(sc.Frequency)
	}
}

func (sc *ScheduleContext) GetInfo() string {
	return sc.Schedule.String()
}

func (sc *ScheduleContext) isComplete(time time.Time) bool {
	complete := (sc.StartTime.Unix() < time.Unix() && sc.Schedule.RunOnce) ||
		(sc.NextTime.Unix() > sc.EndTime.Unix()) ||
		((sc.MaxIterations != 0) && (sc.CurrentIterations >= sc.MaxIterations))
	return complete
}

//region util methods
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

//endregion
