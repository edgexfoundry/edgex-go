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

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Test common const
const (
	TestUnexpectedMsg                     = "unexpected result"
	TestUnexpectedMsgFormatStr            = "unexpected result, active: '%s' but expected: '%s'"
	TestUnexpectedMsgFormatStrForIntVal   = "unexpected result, active: '%d' but expected: '%d'"
	TestUnexpectedMsgFormatStrForFloatVal = "unexpected result, active: '%f' but expected: '%f'"
	TestUnexpectedMsgFormatStrForBoolVal  = "unexpected result, active: '%t' but expected: '%t'"
	TestUnexpectedMsgFormatStrForInt64Val = "unexpected result, active: '%d' but expected: '%d'"
)

// Test Schedule model const fields
const (
	TestIntervalName         = "midnight-1"
	TestIntervalStart        = "20000101T000000"
	TestIntervalEnd          = ""
	TestIntervalFrequency    = "24h"
	TestIntervalCron         = "This is a description"
	TestIntervalRunOnce      = true
	TestIntervalUpdatingName = "midnight-2"
	TestBadFrequency         = "423"
)

// Test IntervalAction model const fields
const (
	TestIntervalActionEventId             = "testIntervalActionEventId"
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

	testInterval.Frequency = "60s"
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
	testInterval.Frequency = "24h"
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
	testInterval.Frequency = "24h"
	testInterval.RunOnce = false
	testIntervalContext.Reset(testInterval)

	if testIntervalContext.IsComplete() {
		t.Errorf(TestUnexpectedMsgFormatStrForBoolVal, testIntervalContext.IsComplete(), false)
	}
}

func TestParseNanoSecondFrequency(t *testing.T) {

	durationStr := "50ns"
	duration, err := parseFrequency(durationStr)
	if err != nil {
		t.Errorf(TestUnexpectedMsgFormatStrForInt64Val, duration.Nanoseconds(), 50)
	}
	if duration.Nanoseconds() != int64(50) {
		t.Errorf(TestUnexpectedMsgFormatStrForInt64Val, duration.Nanoseconds(), 50)
	}
}

// Note Time.Duration does not support milliseconds, or microseconds directly.
func TestParseMicrosecondsFrequency(t *testing.T) {
	durationStr := "1us"
	duration, err := parseFrequency(durationStr)
	if err != nil {
		t.Errorf(TestUnexpectedMsgFormatStrForInt64Val, duration.Nanoseconds(), 1000)
	}
	if duration.Nanoseconds() != int64(1000) {
		t.Errorf(TestUnexpectedMsgFormatStrForInt64Val, duration.Nanoseconds(), 1000)
	}
}

// Note Time.Duration does not support milliseconds, or microseconds directly.
func TestParseMillisecondFrequency(t *testing.T) {

	durationStr := "500ms"
	duration, err := parseFrequency(durationStr)
	if err != nil {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, duration.Seconds(), .5)
	}

	if duration.Seconds() != .5 {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, duration.Seconds(), .5)
	}
}

func TestParseFrequency(t *testing.T) {
	durationStr := "24h"
	duration, err := parseFrequency(durationStr)
	if err != nil {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, duration.Hours(), 24.0)
	}
	if duration.Hours() != 24 {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, duration.Hours(), 24.0)
	}

	durationStr = "50s"
	duration, err = parseFrequency(durationStr)

	if err != nil {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, duration.Hours(), 24.0)
	}

	if duration.Seconds() != 50 {
		t.Errorf(TestUnexpectedMsgFormatStrForFloatVal, duration.Seconds(), 50.0)
	}
}
