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
	TestIntervalName         = "midnight-1"
	TestIntervalStart        = "20000101T000000"
	TestIntervalEnd          = ""
	TestIntervalFrequency    = "P1D"
	TestIntervalCron         = "This is a description"
	TestIntervalRunOnce      = true
	TestIntervalUpdatingName = "midnight-2"

	TestBadFrequency = "423"
)

// Test ScheduleEvent model const fields
const (
	TestIntervalActionEventId                  = "testScheduleEventId"
	TestIntervalActionName                = "pushed events"
	TestIntervalActionParameters          = ""
	TestIntervalActionService             = "notifications"
	TestIntervalActionSchedule            = TestIntervalName
	TestIntervalActionAddressableName     = "MQTT"
	TestIntervalActionAddressableProtocol = "MQTT"
	TestIntervalActionUpdatingName        = "pushed events-1"
)

func TestRet(t *testing.T) {
	testInterval := models.Interval{
		Name:      TestIntervalName,
		Start:     TestIntervalStart,
		End:       TestIntervalEnd,
		Frequency: TestIntervalFrequency,
		Cron:      TestIntervalCron,
		RunOnce:   TestIntervalRunOnce,
	}

	//init reset
	testIntervalContext := IntervalContext{}
	testIntervalContext.Reset(testInterval)

	if testInterval.Name != testIntervalContext.Interval.Name {
		t.Errorf(TestUnexpectedMsgFormatStr, testIntervalContext.Interval.Name, testInterval.Name)
	}

	if testIntervalContext.MaxIterations != 1 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, testIntervalContext.MaxIterations, 1)
	}

	//run times, current and max iteration
	testInterval.RunOnce = false
	testIntervalContext.Reset(testInterval)

	if testIntervalContext.MaxIterations != 0 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, testIntervalContext.MaxIterations, 0)
	}

	if testIntervalContext.CurrentIterations != 0 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, testIntervalContext.CurrentIterations, 0)
	}

	//start time
	if testIntervalContext.StartTime == (time.Time{}) {
		t.Errorf(TestUnexpectedMsg)
	}

	testInterval.Start = "20180101T010101"
	testIntervalContext.Reset(testInterval)

	if testIntervalContext.StartTime.Year() != 2018 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, testIntervalContext.StartTime.Year(), 2018)
	}

	//end time
	if testIntervalContext.EndTime == (time.Time{}) {
		t.Error(TestUnexpectedMsg)
	}

	testInterval.End = "20170101T010101"
	testIntervalContext.Reset(testInterval)

	if testIntervalContext.EndTime.Year() != 2017 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, testIntervalContext.EndTime.Year(), 2017)
	}

	//frequency
	if testIntervalContext.Frequency.Hours() != 24 {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, testIntervalContext.Frequency.Hours(), 24.0)
	}

	testInterval.Frequency = "PT60S"
	testIntervalContext.Reset(testInterval)
	if testIntervalContext.Frequency.Seconds() != 60 {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, testIntervalContext.Frequency.Seconds(), 60.0)
	}

	//re-init time
	testInterval.Start = ""
	testInterval.End = ""
	testInterval.RunOnce = true

	testIntervalContext.Reset(testInterval)

	//next time
	if testIntervalContext.StartTime != testIntervalContext.NextTime {
		t.Error(TestUnexpectedMsg)
	}

	if testIntervalContext.NextTime.Unix() > time.Now().Unix() {
		t.Error(TestUnexpectedMsg)
	}

	testInterval.RunOnce = false
	testIntervalContext.Reset(testInterval)

	if testIntervalContext.StartTime.Unix() > testIntervalContext.NextTime.Unix() {
		t.Error(TestUnexpectedMsg)
	}

	testInterval.Start = "20180101T010101"
	testInterval.Frequency = "P1D"
	testIntervalContext.Reset(testInterval)

	if testIntervalContext.StartTime.Unix() > testIntervalContext.NextTime.Unix() {
		t.Error(TestUnexpectedMsg)
	}

	if testIntervalContext.NextTime.Unix() < time.Now().Unix() {
		t.Errorf(TestUnexpectedMsg)
	}

}

func TestIsComplete(t *testing.T) {
	testInterval := models.Interval{
		Name:      TestIntervalName,
		Start:     TestIntervalStart,
		End:       TestIntervalEnd,
		Frequency: TestIntervalFrequency,
		Cron:      TestIntervalCron,
		RunOnce:   TestIntervalRunOnce,
	}

	//init reset
	testIntervalContext := IntervalContext{}
	testIntervalContext.Reset(testInterval)

	if !testIntervalContext.IsComplete() {
		t.Errorf(TestUnexpectedMsgFormatStrForBoolVal, testIntervalContext.IsComplete(), true)
	}

	testInterval.Start = "20180101T010101"
	testInterval.Frequency = "P1D"
	testInterval.RunOnce = false
	testIntervalContext.Reset(testInterval)

	if testIntervalContext.IsComplete() {
		t.Errorf(TestUnexpectedMsgFormatStrForBoolVal, testIntervalContext.IsComplete(), false)
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
