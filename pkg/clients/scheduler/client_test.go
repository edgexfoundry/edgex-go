//
// Copyright (c) 2017 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Test common const
const (
	TestUnexpectedMsg                    = "unexpected result"
	TestUnexpectedMsgFormatStr           = "unexpected result, active: '%s' but expected: '%s'"
	TestUnexpectedMsgFormatStrForBoolVal = "unexpected result, active: '%t' but expected: '%t'"
)

// Test Schedule model const fields
const (
	TestScheduleId        = "testScheduleId"
	TestScheduleName      = "midnight-1"
	TestScheduleStart     = "20000101T000000"
	TestScheduleEnd       = ""
	TestScheduleFrequency = "P1D"
	TestScheduleCron      = "This is a description"
	TestScheduleRunOnce   = true
	TestScheduleIdForTest = "testScheduleId"
)

// Test ScheduleEvent model const fields
const (
	TestScheduleEventName                = "pushed events"
	TestScheduleEventParameters          = ""
	TestScheduleEventService             = "notifications"
	TestScheduleEventSchedule            = "testSchedule"
	TestScheduleEventAddressableName     = "MQTT"
	TestScheduleEventAddressableProtocol = "MQTT"
	TestScheduleEventId                  = "testScheduleEventId"
)

// Test method : AddSchedule
func TestAddSchedule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != http.MethodPost {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodPost)
		}
		if r.URL.EscapedPath() != clients.ApiScheduleRoute {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), clients.ApiScheduleRoute)
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

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err.Error())
	}

	h := strings.Split(u.Host, ":")

	intPort, e := strconv.Atoi(h[1])
	if e != nil {
		t.Error(e)
	}

	SetConfiguration(h[0], intPort)

	schedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	error := GetSchedulerClient().AddSchedule(schedule)
	if error != nil {
		t.Error(error)
	}
}

// Test method : QuerySchedule
func TestQuerySchedule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodGet)
		}

		urlWithIdPath := clients.ApiScheduleRoute + "/" + TestScheduleId

		if r.URL.EscapedPath() != urlWithIdPath {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), urlWithIdPath)
		}

		id := strings.TrimPrefix(r.URL.EscapedPath(), clients.ApiScheduleRoute+"/")

		if id != TestScheduleId {
			t.Errorf(TestUnexpectedMsgFormatStr, id, TestScheduleId)
		}

		schedule := models.Schedule{
			Name:      TestScheduleName,
			Start:     TestScheduleStart,
			End:       TestScheduleEnd,
			Frequency: TestScheduleFrequency,
			Cron:      TestScheduleCron,
			RunOnce:   TestScheduleRunOnce,
		}

		w.WriteHeader(http.StatusOK)

		jsonBytes, err := schedule.MarshalJSON()
		if err != nil {
			t.Error(err.Error())
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	}))

	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err.Error())
	}

	h := strings.Split(u.Host, ":")

	intPort, e := strconv.Atoi(h[1])
	if e != nil {
		t.Error(e)
	}

	SetConfiguration(h[0], intPort)

	receivedSchedule, err := GetSchedulerClient().QuerySchedule(TestScheduleId)
	if err != nil {
		t.Error(err.Error())
	}

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
}

// Test method : QueryScheduleWithName
func TestQueryScheduleWithName(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodGet)
		}

		urlWithNamePart := clients.ApiScheduleRoute + "/name/" + TestScheduleName

		if r.URL.EscapedPath() != urlWithNamePart {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), urlWithNamePart)
		}

		name := strings.TrimPrefix(r.URL.EscapedPath(), clients.ApiScheduleRoute+"/name/")

		if name != TestScheduleName {
			t.Errorf(TestUnexpectedMsgFormatStr, name, TestScheduleName)
		}

		schedule := models.Schedule{
			Name:      TestScheduleName,
			Start:     TestScheduleStart,
			End:       TestScheduleEnd,
			Frequency: TestScheduleFrequency,
			Cron:      TestScheduleCron,
			RunOnce:   TestScheduleRunOnce,
		}

		w.WriteHeader(http.StatusOK)

		jsonBytes, err := schedule.MarshalJSON()
		if err != nil {
			t.Error(err.Error())
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	}))

	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err.Error())
	}

	h := strings.Split(u.Host, ":")

	intPort, e := strconv.Atoi(h[1])
	if e != nil {
		t.Error(e)
	}

	SetConfiguration(h[0], intPort)

	receivedSchedule, err := GetSchedulerClient().QueryScheduleWithName(TestScheduleName)
	if err != nil {
		t.Error(err.Error())
	}

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
}

// Test method : UpdateSchedule
func TestUpdateSchedule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != http.MethodPut {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodPut)
		}
		if r.URL.EscapedPath() != clients.ApiScheduleRoute {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), clients.ApiScheduleRoute)
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

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err.Error())
	}

	h := strings.Split(u.Host, ":")

	intPort, e := strconv.Atoi(h[1])
	if e != nil {
		t.Error(e)
	}

	SetConfiguration(h[0], intPort)

	schedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	error := GetSchedulerClient().UpdateSchedule(schedule)
	if error != nil {
		t.Error(error)
	}
}

// Test method : RemoveSchedule
func TestRemoveSchedule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != http.MethodDelete {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodDelete)
		}

		if !strings.HasPrefix(r.URL.EscapedPath(), clients.ApiScheduleRoute) {
			t.Errorf(TestUnexpectedMsg)
		}

		if !strings.HasSuffix(r.URL.EscapedPath(), TestScheduleIdForTest) {
			t.Errorf(TestUnexpectedMsg)
		}
	}))

	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err.Error())
	}

	h := strings.Split(u.Host, ":")

	intPort, e := strconv.Atoi(h[1])
	if e != nil {
		t.Error(e)
	}

	SetConfiguration(h[0], intPort)

	error := GetSchedulerClient().RemoveSchedule(TestScheduleIdForTest)
	if error != nil {
		t.Error(error)
	}
}

// Test method : AddScheduleEvent
func TestAddScheduleEvent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != http.MethodPost {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodPost)
		}
		if r.URL.EscapedPath() != clients.ApiScheduleEventRoute {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), clients.ApiScheduleEventRoute)
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

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err.Error())
	}

	h := strings.Split(u.Host, ":")

	intPort, e := strconv.Atoi(h[1])
	if e != nil {
		t.Error(e)
	}

	SetConfiguration(h[0], intPort)

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

	error := GetSchedulerClient().AddScheduleEvent(scheduleEvent)
	if error != nil {
		t.Error(error)
	}
}

// Test method : QueryScheduleEvent
func TestQueryScheduleEvent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodGet)
		}

		urlWithIdPath := clients.ApiScheduleEventRoute + "/" + TestScheduleEventId

		if r.URL.EscapedPath() != urlWithIdPath {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), urlWithIdPath)
		}

		id := strings.TrimPrefix(r.URL.EscapedPath(), clients.ApiScheduleEventRoute+"/")
		if id != TestScheduleEventId {
			t.Errorf(TestUnexpectedMsgFormatStr, id, TestScheduleEventId)
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

		jsonBytes, err := scheduleEvent.MarshalJSON()
		if err != nil {
			t.Error(err.Error())
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	}))

	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err.Error())
	}

	h := strings.Split(u.Host, ":")

	intPort, e := strconv.Atoi(h[1])
	if e != nil {
		t.Error(e)
	}

	SetConfiguration(h[0], intPort)

	receivedScheduleEvent, error := GetSchedulerClient().QueryScheduleEvent(TestScheduleEventId)
	if error != nil {
		t.Error(error)
	}

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

}

// Test method : UpdateScheduleEvent
func TestUpdateScheduleEvent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != http.MethodPut {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodPut)
		}
		if r.URL.EscapedPath() != clients.ApiScheduleEventRoute {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), clients.ApiScheduleEventRoute)
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

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err.Error())
	}

	h := strings.Split(u.Host, ":")

	intPort, e := strconv.Atoi(h[1])
	if e != nil {
		t.Error(e)
	}

	SetConfiguration(h[0], intPort)

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

	error := GetSchedulerClient().UpdateScheduleEvent(scheduleEvent)
	if error != nil {
		t.Error(error)
	}
}

// Test method : RemoveScheduleEvent
func TestRemoveScheduleEvent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != http.MethodDelete {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodDelete)
		}

		if !strings.HasPrefix(r.URL.EscapedPath(), clients.ApiScheduleEventRoute) {
			t.Errorf(TestUnexpectedMsg)
		}

		if !strings.HasSuffix(r.URL.EscapedPath(), TestScheduleEventId) {
			t.Errorf(TestUnexpectedMsg)
		}
	}))

	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err.Error())
	}

	h := strings.Split(u.Host, ":")

	intPort, e := strconv.Atoi(h[1])
	if e != nil {
		t.Error(e)
	}

	SetConfiguration(h[0], intPort)

	error := GetSchedulerClient().RemoveScheduleEvent(TestScheduleEventId)
	if error != nil {
		t.Error(error)
	}
}
