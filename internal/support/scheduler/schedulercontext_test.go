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

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Test common const
const (
	TestUnexpectedMsg                     = "unexpected result"
	TestUnexpectedMsgFormatStr            = "unexpected result, active: '%s' but expected: '%s'"
	TestUnexpectedMsgFormatStrForIntVal   = "unexpected result, active: '%d' but expected: '%d'"
	TestUnexpectedMsgFormatStrForFloatVal = "unexpected result, active: '%f' but expected: '%f'"
	TestUnexpectedMsgFormatStrForFloatExp = "unexpected result, active: '%s' but expected: '%f'"
	TestUnexpectedMsgFormatStrForBoolVal  = "unexpected result, active: '%t' but expected: '%t'"
	TestUnexpectedMsgFormatStrForInt64Val = "unexpected result, active: '%d' but expected: '%d'"
	TestUnexpectedMsgFormatStrForInt64Exp = "unexpected result, active: '%s' but expected: '%d'"
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

	lc := logger.NewMockClient()

	// init reset
	testIntervalContext := IntervalContext{}
	testIntervalContext.Reset(testInterval, lc)

	if testInterval.Name != testIntervalContext.Interval.Name {
		t.Fatalf(TestUnexpectedMsgFormatStr, testIntervalContext.Interval.Name, testInterval.Name)
	}

	if testIntervalContext.MaxIterations != 1 {
		t.Fatalf(TestUnexpectedMsgFormatStrForIntVal, testIntervalContext.MaxIterations, 1)
	}

	// run times, current and max iteration
	testInterval.RunOnce = false
	testIntervalContext.Reset(testInterval, lc)

	if testIntervalContext.MaxIterations != 0 {
		t.Fatalf(TestUnexpectedMsgFormatStrForIntVal, testIntervalContext.MaxIterations, 0)
	}

	if testIntervalContext.CurrentIterations != 0 {
		t.Fatalf(TestUnexpectedMsgFormatStrForIntVal, testIntervalContext.CurrentIterations, 0)
	}

	// start time
	if testIntervalContext.StartTime == (time.Time{}) {
		t.Fatalf(TestUnexpectedMsg)
	}

	testInterval.Start = "20180101T010101"
	testIntervalContext.Reset(testInterval, lc)

	if testIntervalContext.StartTime.Year() != 2018 {
		t.Fatalf(TestUnexpectedMsgFormatStrForIntVal, testIntervalContext.StartTime.Year(), 2018)
	}

	// end time
	if testIntervalContext.EndTime == (time.Time{}) {
		t.Error(TestUnexpectedMsg)
	}

	testInterval.End = "20170101T010101"
	testIntervalContext.Reset(testInterval, lc)

	if testIntervalContext.EndTime.Year() != 2017 {
		t.Fatalf(TestUnexpectedMsgFormatStrForIntVal, testIntervalContext.EndTime.Year(), 2017)
	}

	// frequency
	if testIntervalContext.Frequency.Hours() != 24 {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatVal, testIntervalContext.Frequency.Hours(), 24.0)
	}

	testInterval.Frequency = "60s"
	testIntervalContext.Reset(testInterval, lc)
	if testIntervalContext.Frequency.Seconds() != 60 {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatVal, testIntervalContext.Frequency.Seconds(), 60.0)
	}

	// re-init time
	testInterval.Start = ""
	testInterval.End = ""
	testInterval.RunOnce = true

	testIntervalContext.Reset(testInterval, lc)

	// next time
	if testIntervalContext.StartTime != testIntervalContext.NextTime {
		t.Error(TestUnexpectedMsg)
	}

	if testIntervalContext.NextTime.Unix() > time.Now().Unix() {
		t.Error(TestUnexpectedMsg)
	}

	testInterval.RunOnce = false
	testIntervalContext.Reset(testInterval, lc)

	if testIntervalContext.StartTime.Unix() > testIntervalContext.NextTime.Unix() {
		t.Error(TestUnexpectedMsg)
	}

	testInterval.Start = "20180101T010101"
	testInterval.Frequency = "24h"
	testIntervalContext.Reset(testInterval, lc)

	if testIntervalContext.StartTime.Unix() > testIntervalContext.NextTime.Unix() {
		t.Error(TestUnexpectedMsg)
	}

	if testIntervalContext.NextTime.Unix() < time.Now().Unix() {
		t.Fatalf(TestUnexpectedMsg)
	}

}

func TestIsComplete(t *testing.T) {
	testInterval := models.Interval{
		Name:      TestIntervalName,
		Start:     TestIntervalStart,
		End:       TestIntervalEnd,
		Frequency: TestIntervalFrequency,
		Cron:      TestIntervalCron,
		RunOnce:   true,
	}

	lc := logger.NewMockClient()

	// init reset
	testIntervalContext := IntervalContext{}
	testIntervalContext.Reset(testInterval, lc)

	if !testIntervalContext.IsComplete() {
		t.Fatalf(TestUnexpectedMsgFormatStrForBoolVal, testIntervalContext.IsComplete(), true)
	}

	testInterval.Start = "20180101T010101"
	testInterval.Frequency = "24h"
	testInterval.RunOnce = false
	testIntervalContext.Reset(testInterval, lc)

	if testIntervalContext.IsComplete() {
		t.Fatalf(TestUnexpectedMsgFormatStrForBoolVal, testIntervalContext.IsComplete(), false)
	}
}

func TestParseNanoSecondFrequency(t *testing.T) {

	durationStr := "50ns"
	duration, err := parseFrequency(durationStr)
	if err != nil {
		t.Fatalf(TestUnexpectedMsgFormatStrForInt64Exp, durationStr, 50)
	}
	if duration.Nanoseconds() != int64(50) {
		t.Fatalf(TestUnexpectedMsgFormatStrForInt64Val, duration.Nanoseconds(), 50)
	}
}

// Note Time.Duration does not support milliseconds, or microseconds directly.
func TestParseMicrosecondsFrequency(t *testing.T) {
	durationStr := "1us"
	duration, err := parseFrequency(durationStr)
	if err != nil {
		t.Fatalf(TestUnexpectedMsgFormatStrForInt64Exp, durationStr, 1000)
	}
	if duration.Nanoseconds() != int64(1000) {
		t.Fatalf(TestUnexpectedMsgFormatStrForInt64Val, duration.Nanoseconds(), 1000)
	}
}

// Note Time.Duration does not support milliseconds, or microseconds directly.
func TestParseMillisecondFrequency(t *testing.T) {

	durationStr := "500ms"
	duration, err := parseFrequency(durationStr)
	if err != nil {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatExp, durationStr, .5)
	}

	if duration.Seconds() != .5 {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatVal, duration.Seconds(), .5)
	}
}

func TestParseFrequency(t *testing.T) {
	durationStr := "24h"
	duration, err := parseFrequency(durationStr)
	if err != nil {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatExp, durationStr, 24.0)
	}
	if duration.Hours() != 24 {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatVal, duration.Hours(), 24.0)
	}

	durationStr = "50s"
	duration, err = parseFrequency(durationStr)

	if err != nil {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatExp, durationStr, 24.0)
	}

	if duration.Seconds() != 50 {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatVal, duration.Seconds(), 50.0)
	}
}
