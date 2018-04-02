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
	"net/url"
	"strconv"
	"strings"
	"testing"
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
		if r.URL.EscapedPath() != ScheduleApiPath {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), ScheduleApiPath)
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

	scheduleClient := SchedulerClient{
		SchedulerServiceHost: h[0],
		SchedulerServicePort: intPort,
		OwningService:        "notifications",
	}

	schedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	error := scheduleClient.AddSchedule(schedule)
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

		urlWithIdPath := ScheduleApiPath + "/" + TestScheduleId

		if r.URL.EscapedPath() != urlWithIdPath {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), urlWithIdPath)
		}

		id := strings.TrimPrefix(r.URL.EscapedPath(), ScheduleApiPath+"/")

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

	scheduleClient := SchedulerClient{
		SchedulerServiceHost: h[0],
		SchedulerServicePort: intPort,
		OwningService:        "notifications",
	}

	receivedSchedule, err := scheduleClient.QuerySchedule(TestScheduleId)
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

		urlWithNamePart := ScheduleApiPath + "/name/" + TestScheduleName

		if r.URL.EscapedPath() != urlWithNamePart {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), urlWithNamePart)
		}

		name := strings.TrimPrefix(r.URL.EscapedPath(), ScheduleApiPath+"/name/")

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

	scheduleClient := SchedulerClient{
		SchedulerServiceHost: h[0],
		SchedulerServicePort: intPort,
		OwningService:        "notifications",
	}

	receivedSchedule, err := scheduleClient.QueryScheduleWithName(TestScheduleName)
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
		if r.URL.EscapedPath() != ScheduleApiPath {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), ScheduleApiPath)
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

	scheduleClient := SchedulerClient{
		SchedulerServiceHost: h[0],
		SchedulerServicePort: intPort,
		OwningService:        "notifications",
	}

	schedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	error := scheduleClient.UpdateSchedule(schedule)
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

		if !strings.HasPrefix(r.URL.EscapedPath(), ScheduleApiPath) {
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

	scheduleClient := SchedulerClient{
		SchedulerServiceHost: h[0],
		SchedulerServicePort: intPort,
		OwningService:        "notifications",
	}

	error := scheduleClient.RemoveSchedule(TestScheduleIdForTest)
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
		if r.URL.EscapedPath() != ScheduleEventApiPath {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), ScheduleEventApiPath)
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

	scheduleClient := SchedulerClient{
		SchedulerServiceHost: h[0],
		SchedulerServicePort: intPort,
		OwningService:        "notifications",
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

	error := scheduleClient.AddScheduleEvent(scheduleEvent)
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

		urlWithIdPath := ScheduleEventApiPath + "/" + TestScheduleEventId

		if r.URL.EscapedPath() != urlWithIdPath {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), urlWithIdPath)
		}

		id := strings.TrimPrefix(r.URL.EscapedPath(), ScheduleEventApiPath+"/")
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

	scheduleClient := SchedulerClient{
		SchedulerServiceHost: h[0],
		SchedulerServicePort: intPort,
		OwningService:        "notifications",
	}

	receivedScheduleEvent, error := scheduleClient.QueryScheduleEvent(TestScheduleEventId)
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
		if r.URL.EscapedPath() != ScheduleEventApiPath {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), ScheduleEventApiPath)
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

	scheduleClient := SchedulerClient{
		SchedulerServiceHost: h[0],
		SchedulerServicePort: intPort,
		OwningService:        "notifications",
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

	error := scheduleClient.UpdateScheduleEvent(scheduleEvent)
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

		if !strings.HasPrefix(r.URL.EscapedPath(), ScheduleEventApiPath) {
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

	scheduleClient := SchedulerClient{
		SchedulerServiceHost: h[0],
		SchedulerServicePort: intPort,
		OwningService:        "notifications",
	}

	error := scheduleClient.RemoveScheduleEvent(TestScheduleEventId)
	if error != nil {
		t.Error(error)
	}
}
