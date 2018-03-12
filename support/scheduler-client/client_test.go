//
// Copyright (c) 2017 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"encoding/json"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test common const
const (
	TestUnexpectedMsgFormatStr           = "unexpected result, active: '%s' but expected: '%s'"
	TestUnexpectedMsgFormatStrForBoolVal = "unexpected result, active: '%t' but expected: '%t'"
)

// Test comm api path
const (
	TestExpectedScheduleApiPath      = "/api/v1/schedule"
	TestExpectedScheduleEventApiPath = "/api/v1/scheduleevent"
)

// Test Schedule model const fields
const (
	TestScheduleName      = "midnight-1"
	TestScheduleStart     = "20000101T000000"
	TestScheduleEnd       = ""
	TestScheduleFrequency = "P1D"
	TestScheduleCron      = "This is a description"
	TestScheduleRunOnce   = true
)

// Test ScheduleEvent model const fields
const (
	TestScheduleEventName                = "pushed events"
	TestScheduleEventParameters          = ""
	TestScheduleEventService             = "notifications"
	TestScheduleEventSchedule            = "testSchedule"
	TestScheduleEventAddressableName     = "MQTT"
	TestScheduleEventAddressableProtocol = "MQTT"
)

// Test method : SendSchedule
func TestSchedulerClient_SendSchedule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != "POST" {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, "POST")
		}
		if r.URL.EscapedPath() != TestExpectedScheduleApiPath {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), TestExpectedScheduleApiPath)
		}

		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()

		var receivedSchedule models.Schedule
		json.Unmarshal([]byte(result), &receivedSchedule)

		if receivedSchedule.Name != TestScheduleName {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedSchedule.Name, TestScheduleName)
		}

		if receivedSchedule.Start != TestScheduleStart {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedSchedule.Start, TestScheduleStart)
		}

		if receivedSchedule.End != TestScheduleEnd {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedSchedule.End, TestScheduleEnd)
		}

		if receivedSchedule.Frequency != TestScheduleFrequency {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedSchedule.Frequency, TestScheduleFrequency)
		}

		if receivedSchedule.Cron != TestScheduleCron {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedSchedule.Cron, TestScheduleCron)
		}

		if receivedSchedule.RunOnce != TestScheduleRunOnce {
			t.Errorf(TestUnexpectedMsgFormatStrForBoolVal, receivedSchedule.RunOnce, TestScheduleRunOnce)
		}

	}))

	defer ts.Close()
	uriPrefix := ts.URL

	scheduleClient := SchedulerClient{
		RemoteScheduleUrl: uriPrefix + "/api/v1/schedule",
		OwningService:     "notifications",
	}

	schedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	err := scheduleClient.SendSchedule(schedule)
	println(err)
}

// Test method : SendScheduleEvent
func TestSchedulerClient_SendScheduleEvent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != "POST" {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, "POST")
		}
		if r.URL.EscapedPath() != TestExpectedScheduleEventApiPath {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), TestExpectedScheduleEventApiPath)
		}

		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()

		var receivedScheduleEvent models.ScheduleEvent
		json.Unmarshal([]byte(result), &receivedScheduleEvent)

		if receivedScheduleEvent.Name != TestScheduleEventName {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedScheduleEvent.Name, TestScheduleEventName)
		}

		if receivedScheduleEvent.Parameters != TestScheduleEventParameters {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedScheduleEvent.Parameters, TestScheduleEventParameters)
		}

		if receivedScheduleEvent.Service != TestScheduleEventService {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedScheduleEvent.Service, TestScheduleEventService)
		}

		if receivedScheduleEvent.Addressable == (models.Addressable{}) {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedScheduleEvent.Addressable, "nil")
		}

		if receivedScheduleEvent.Addressable.Name != TestScheduleEventAddressableName {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedScheduleEvent.Addressable.Name, TestScheduleEventAddressableName)
		}

		if receivedScheduleEvent.Addressable.Protocol != TestScheduleEventAddressableProtocol {
			t.Errorf(TestUnexpectedMsgFormatStr, receivedScheduleEvent.Addressable.Protocol, TestScheduleEventAddressableProtocol)
		}

	}))

	defer ts.Close()
	uriPrefix := ts.URL

	scheduleClient := SchedulerClient{
		RemoteScheduleEventUrl: uriPrefix + "/api/v1/scheduleevent",
		OwningService:          "notifications",
	}

	scheduleEvent := models.ScheduleEvent{
		Name:       TestScheduleEventName,
		Parameters: TestScheduleEventParameters,
		Service:    TestScheduleEventService,
		Schedule:   TestScheduleEventSchedule,
		Addressable: models.Addressable{
			Name:     TestScheduleEventAddressableName,
			Protocol: TestScheduleEventAddressableProtocol,
		},
	}

	err := scheduleClient.SendScheduleEvent(scheduleEvent)
	println(err)
}
