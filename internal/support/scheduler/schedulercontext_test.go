//
// Copyright (c) 2018 Tencent
//
// Copyright (c) 2017 Dell Inc
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Test common const
const (
	TestUnexpectedMsg                     = "unexpected result"
	TestUnexpectedMsgFormatStr            = "unexpected result, active: '%s' but expected: '%s'"
	TestUnexpectedMsgFormatStrForIntVal   = "unexpected result, active: '%d' but expected: '%d'"
	TestUnexpectedMsgFormatStrForFloatVal = "unexpected result, active: '%f' but expected: '%f'"
	TestUnexpectedMsgFormatStrForBoolVal  = "unexpected result, active: '%t' but expected: '%t'"
)

// Test Schedule model const fields
const (
	TestScheduleName         = "midnight-1"
	TestScheduleStart        = "20000101T000000"
	TestScheduleEnd          = ""
	TestScheduleFrequency    = "P1D"
	TestScheduleCron         = "This is a description"
	TestScheduleRunOnce      = true
	TestScheduleUpdatingName = "midnight-2"

	TestBadFrequency = "423"
)

// Test ScheduleEvent model const fields
const (
	TestScheduleEventId                  = "testScheduleEventId"
	TestScheduleEventName                = "pushed events"
	TestScheduleEventParameters          = ""
	TestScheduleEventService             = "notifications"
	TestScheduleEventSchedule            = TestScheduleName
	TestScheduleEventAddressableName     = "MQTT"
	TestScheduleEventAddressableProtocol = "MQTT"
	TestScheduleEventUpdatingName        = "pushed events-1"
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
