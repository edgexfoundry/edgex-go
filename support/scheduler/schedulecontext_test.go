//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"testing"
	"time"
)

func TestRet(t *testing.T) {
	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	//init reset
	testScheduleContext := ScheduleContext{}
	testScheduleContext.Reset(testSchedule)

	if testSchedule.Name != testScheduleContext.Schedule.Name {
		t.Errorf(TestUnexpectedMsgFormatStr, testScheduleContext.Schedule.Name, testSchedule.Name)
	}

	if testScheduleContext.MaxIterations != 1 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, testScheduleContext.MaxIterations, 1)
	}

	//run times, current and max iteration
	testSchedule.RunOnce = false
	testScheduleContext.Reset(testSchedule)

	if testScheduleContext.MaxIterations != 0 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, testScheduleContext.MaxIterations, 0)
	}

	if testScheduleContext.CurrentIterations != 0 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, testScheduleContext.CurrentIterations, 0)
	}

	//start time
	if testScheduleContext.StartTime == (time.Time{}) {
		t.Errorf(TestUnexpectedMsg)
	}

	testSchedule.Start = "20180101T010101"
	testScheduleContext.Reset(testSchedule)

	if testScheduleContext.StartTime.Year() != 2018 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, testScheduleContext.StartTime.Year(), 2018)
	}

	//end time
	if testScheduleContext.EndTime == (time.Time{}) {
		t.Error(TestUnexpectedMsg)
	}

	testSchedule.End = "20170101T010101"
	testScheduleContext.Reset(testSchedule)

	if testScheduleContext.EndTime.Year() != 2017 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, testScheduleContext.EndTime.Year(), 2017)
	}

	//frequency
	if testScheduleContext.Frequency.Hours() != 24 {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, testScheduleContext.Frequency.Hours(), 24.0)
	}

	testSchedule.Frequency = "PT60S"
	testScheduleContext.Reset(testSchedule)
	if testScheduleContext.Frequency.Seconds() != 60 {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, testScheduleContext.Frequency.Seconds(), 60.0)
	}

	//re-init time
	testSchedule.Start = ""
	testSchedule.End = ""
	testSchedule.RunOnce = true

	testScheduleContext.Reset(testSchedule)

	//next time
	if testScheduleContext.StartTime != testScheduleContext.NextTime {
		t.Error(TestUnexpectedMsg)
	}

	if testScheduleContext.NextTime.Unix() > time.Now().Unix() {
		t.Error(TestUnexpectedMsg)
	}

	testSchedule.RunOnce = false
	testScheduleContext.Reset(testSchedule)

	if testScheduleContext.StartTime.Unix() > testScheduleContext.NextTime.Unix() {
		t.Error(TestUnexpectedMsg)
	}

	testSchedule.Start = "20180101T010101"
	testSchedule.Frequency = "P1D"
	testScheduleContext.Reset(testSchedule)

	if testScheduleContext.StartTime.Unix() > testScheduleContext.NextTime.Unix() {
		t.Error(TestUnexpectedMsg)
	}

	if testScheduleContext.NextTime.Unix() < time.Now().Unix() {
		t.Errorf(TestUnexpectedMsg)
	}

}

func TestIsComplete(t *testing.T) {
	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	//init reset
	testScheduleContext := ScheduleContext{}
	testScheduleContext.Reset(testSchedule)

	if !testScheduleContext.IsComplete() {
		t.Errorf(TestUnexpectedMsgFormatStrForBoolVal, testScheduleContext.IsComplete(), true)
	}

	testSchedule.Start = "20180101T010101"
	testSchedule.Frequency = "P1D"
	testSchedule.RunOnce = false
	testScheduleContext.Reset(testSchedule)

	if testScheduleContext.IsComplete() {
		t.Errorf(TestUnexpectedMsgFormatStrForBoolVal, testScheduleContext.IsComplete(), false)
	}
}

func TestParseFrequency(t *testing.T) {
	durationStr := "P1D"
	duration := parseFrequency(durationStr)

	if duration.Hours() != 24 {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, duration.Hours(), 24.0)
	}

	durationStr = "PT50S"
	duration = parseFrequency(durationStr)

	if duration.Seconds() != 50 {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, duration.Seconds(), 50.0)
	}

	//exception case
	durationStr = "TP1234"
	duration = parseFrequency(durationStr)

	if duration.Seconds() != 0 {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, duration.Seconds(), 0.0)
	}
}
