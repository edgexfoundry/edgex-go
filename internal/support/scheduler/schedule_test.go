//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/edgexfoundry/edgex-go/support/scheduler-client"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

// Test common const
const (
	TestUnexpectedMsg                     = "unexpected result"
	TestUnexpectedMsgFormatStr            = "unexpected result, active: '%s' but expected: '%s'"
	TestUnexpectedMsgFormatStrForIntVal   = "unexpected result, active: '%d' but expected: '%d'"
	TestUnexpectedMsgFormatStrForFloatVal = "unexpected result, active: '%f' but expected: '%f'"
	TestUnexpectedMsgFormatStrForBoolVal  = "unexpected result, active: '%t' but expected: '%t'"
	ScheduleApiPath                       = "/api/v1/schedule"
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

func setup(t *testing.T) {
	clearQueue()
}

func mockInit(host string, port int) {
	var loggingClient = logger.NewClient(configuration.ApplicationName, configuration.EnableRemoteLogging, "")
	Init(ConfigurationStruct{
		ScheduleInterval: 500,
	}, loggingClient, scheduler.SchedulerClient{
		SchedulerServiceHost: host,
		SchedulerServicePort: port,
	})
	StartTicker()
}

func TestScheduleLifeCycle(t *testing.T) {
	setup(t)
	mockInit("", -1)

	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	loggingClient.Info("----------- test schedule lifecycle start ------------")

	testSchedule.Id = bson.NewObjectId()

	err := addSchedule(testSchedule)
	if err != nil {
		t.Error(err.Error())
	}

	//check exists
	receivedSchedule, err := querySchedule(testSchedule.Id.Hex())
	if err != nil {
		t.Error(err.Error())
	}

	if receivedSchedule == (models.Schedule{}) {
		t.Error("the expected schedule is not exists.")
	}

	if receivedSchedule.Name != TestScheduleName {
		t.Errorf(TestUnexpectedMsgFormatStr, receivedSchedule.Name, TestScheduleName)
	}

	//check queue
	len := queryQueueLen()
	if len != 1 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, len, 1)
	}

	//update
	testSchedule.Name = TestScheduleUpdatingName
	err = updateSchedule(testSchedule)
	if err != nil {
		t.Error(err.Error())
	}

	//check updated
	receivedSchedule, err = querySchedule(testSchedule.Id.Hex())
	if err != nil {
		t.Error(err.Error())
	}

	if receivedSchedule == (models.Schedule{}) {
		t.Error("the expected schedule is not exists.")
	}

	if receivedSchedule.Name != TestScheduleUpdatingName {
		t.Errorf(TestUnexpectedMsgFormatStr, receivedSchedule.Name, TestScheduleUpdatingName)
	}

	//check queue, note that the update combined with remove and add, the remove just mark it with a tag
	len = queryQueueLen()
	if len != 1 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, len, 1)
	}

	//remove
	err = removeSchedule(receivedSchedule.Id.Hex())

	if err != nil {
		t.Error(err.Error())
	}

	//check not exists
	receivedSchedule, err = querySchedule(testSchedule.Id.Hex())
	if err != nil {
		t.Error(err.Error())
	}

	if receivedSchedule != (models.Schedule{}) {
		t.Error("the expected schedule shoud be not exists but is exists.")
	}

	//check queue
	len = queryQueueLen()
	if len != 1 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, len, 1)
	}

	loggingClient.Info("----------- test schedule lifecycle end ------------")
}

func TestScheduleEventLifeCycle(t *testing.T) {
	setup(t)

	//parent schedule model
	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	loggingClient.Info("----------- test schedule event lifecycle start ------------")

	testSchedule.Id = bson.NewObjectId()

	err := addSchedule(testSchedule)
	if err != nil {
		t.Error(err)
	}

	//check queue len
	len := queryQueueLen()
	if len != 1 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, len, 1)
	}

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

		w.WriteHeader(http.StatusOK)

		jsonBytes, err := testSchedule.MarshalJSON()
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

	mockInit(h[0], intPort)

	//add schedule event to schedule
	testScheduleEvent := models.ScheduleEvent{
		Id:         TestScheduleEventId,
		Name:       TestScheduleEventName,
		Parameters: TestScheduleEventParameters,
		Service:    TestScheduleEventService,
		Schedule:   TestScheduleEventSchedule,
		Addressable: models.Addressable{
			Name:     TestScheduleEventAddressableName,
			Protocol: TestScheduleEventAddressableProtocol,
		},
	}

	err = addScheduleEvent(testScheduleEvent)
	if err != nil {
		t.Error(err)
	}

	//check schedule event exists
	receivedScheduleEvent, err := queryScheduleEvent(testScheduleEvent.Id.Hex())
	if err != nil {
		t.Error(err.Error())
	}

	if receivedScheduleEvent.Name != TestScheduleEventName {
		t.Errorf(TestUnexpectedMsgFormatStr, receivedScheduleEvent.Name, TestScheduleEventName)
	}

	//update schedule event
	receivedScheduleEvent.Name = TestScheduleEventUpdatingName
	err = updateScheduleEvent(receivedScheduleEvent)
	if err != nil {
		t.Error(err.Error())
	}

	//check schedule event updated
	receivedScheduleEvent, err = queryScheduleEvent(testScheduleEvent.Id.Hex())
	if err != nil {
		t.Error(err.Error())
	}

	if receivedScheduleEvent.Name != TestScheduleEventUpdatingName {
		t.Errorf(TestUnexpectedMsgFormatStr, receivedScheduleEvent.Name, TestScheduleEventUpdatingName)
	}

	//remove schedule event
	err = removeScheduleEvent(testScheduleEvent.Id.Hex())
	if err != nil {
		t.Error(err.Error())
	}

	//check queue length
	len = queryQueueLen()
	if len != 1 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, len, 1)
	}

	loggingClient.Info("----------- test schedule event lifecycle end ------------")

}
