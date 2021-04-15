//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
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
	TestIntervalName      = "midnight-1"
	TestIntervalStart     = "20000101T000000"
	TestIntervalEnd       = ""
	TestIntervalFrequency = "24h"
	TestIntervalRunOnce   = true
)

func TestReset(t *testing.T) {
	testInterval := models.Interval{
		Name:      TestIntervalName,
		Start:     TestIntervalStart,
		End:       TestIntervalEnd,
		Frequency: TestIntervalFrequency,
		RunOnce:   TestIntervalRunOnce,
	}

	lc := logger.NewMockClient()

	// init reset
	executor := Executor{}
	executor.Initialize(testInterval, lc)

	if testInterval.Name != executor.Interval.Name {
		t.Fatalf(TestUnexpectedMsgFormatStr, executor.Interval.Name, testInterval.Name)
	}

	if executor.MaxIterations != 1 {
		t.Fatalf(TestUnexpectedMsgFormatStrForIntVal, executor.MaxIterations, 1)
	}

	// run times, current and max iteration
	testInterval.RunOnce = false
	executor.Initialize(testInterval, lc)

	if executor.MaxIterations != 0 {
		t.Fatalf(TestUnexpectedMsgFormatStrForIntVal, executor.MaxIterations, 0)
	}

	if executor.CurrentIterations != 0 {
		t.Fatalf(TestUnexpectedMsgFormatStrForIntVal, executor.CurrentIterations, 0)
	}

	// start time
	if executor.StartTime == (time.Time{}) {
		t.Fatalf(TestUnexpectedMsg)
	}

	testInterval.Start = "20180101T010101"
	executor.Initialize(testInterval, lc)

	if executor.StartTime.Year() != 2018 {
		t.Fatalf(TestUnexpectedMsgFormatStrForIntVal, executor.StartTime.Year(), 2018)
	}

	// end time
	if executor.EndTime == (time.Time{}) {
		t.Error(TestUnexpectedMsg)
	}

	testInterval.End = "20170101T010101"
	executor.Initialize(testInterval, lc)

	if executor.EndTime.Year() != 2017 {
		t.Fatalf(TestUnexpectedMsgFormatStrForIntVal, executor.EndTime.Year(), 2017)
	}

	// frequency
	if executor.Frequency.Hours() != 24 {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatVal, executor.Frequency.Hours(), 24.0)
	}

	testInterval.Frequency = "60s"
	executor.Initialize(testInterval, lc)
	if executor.Frequency.Seconds() != 60 {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatVal, executor.Frequency.Seconds(), 60.0)
	}

	// re-init time
	testInterval.Start = ""
	testInterval.End = ""
	testInterval.RunOnce = true

	executor.Initialize(testInterval, lc)

	// next time
	if executor.StartTime != executor.NextTime {
		t.Error(TestUnexpectedMsg)
	}

	if executor.NextTime.Unix() > time.Now().Unix() {
		t.Error(TestUnexpectedMsg)
	}

	testInterval.RunOnce = false
	executor.Initialize(testInterval, lc)

	if executor.StartTime.Unix() > executor.NextTime.Unix() {
		t.Error(TestUnexpectedMsg)
	}

	testInterval.Start = "20180101T010101"
	testInterval.Frequency = "24h"
	executor.Initialize(testInterval, lc)

	if executor.StartTime.Unix() > executor.NextTime.Unix() {
		t.Error(TestUnexpectedMsg)
	}

	if executor.NextTime.Unix() < time.Now().Unix() {
		t.Fatalf(TestUnexpectedMsg)
	}

}

func TestIsComplete(t *testing.T) {
	testInterval := models.Interval{
		Name:      TestIntervalName,
		Start:     TestIntervalStart,
		End:       TestIntervalEnd,
		Frequency: TestIntervalFrequency,
		RunOnce:   true,
	}

	lc := logger.NewMockClient()

	// init reset
	executor := Executor{}
	executor.Initialize(testInterval, lc)

	if !executor.IsComplete() {
		t.Fatalf(TestUnexpectedMsgFormatStrForBoolVal, executor.IsComplete(), true)
	}

	testInterval.Start = "20180101T010101"
	testInterval.Frequency = "24h"
	testInterval.RunOnce = false
	executor.Initialize(testInterval, lc)

	if executor.IsComplete() {
		t.Fatalf(TestUnexpectedMsgFormatStrForBoolVal, executor.IsComplete(), false)
	}
}

func TestParseNanoSecondFrequency(t *testing.T) {
	durationStr := "50ns"
	duration, err := time.ParseDuration(durationStr)
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
	duration, err := time.ParseDuration(durationStr)
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
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatExp, durationStr, .5)
	}

	if duration.Seconds() != .5 {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatVal, duration.Seconds(), .5)
	}
}

func TestParseFrequency(t *testing.T) {
	durationStr := "24h"
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatExp, durationStr, 24.0)
	}
	if duration.Hours() != 24 {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatVal, duration.Hours(), 24.0)
	}

	durationStr = "50s"
	duration, err = time.ParseDuration(durationStr)

	if err != nil {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatExp, durationStr, 24.0)
	}

	if duration.Seconds() != 50 {
		t.Fatalf(TestUnexpectedMsgFormatStrForFloatVal, duration.Seconds(), 50.0)
	}
}
