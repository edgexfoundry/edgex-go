//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
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

func mockInit() {

	var loggingClient = logger.NewMockClient()

	Init(ConfigurationStruct{
		ScheduleInterval: 500,
	}, loggingClient, false)

	StartTicker()
}


// test the schedule and Scheduler
func TestAddSchedule(t *testing.T){

	setup(t)
	mockInit()

	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	testSchedule.Id = bson.NewObjectId()

	err := addSchedule(testSchedule)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestRemoveSchedule(t *testing.T){

	setup(t)
	mockInit()

	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	testSchedule.Id = bson.NewObjectId()

	// Add the schedule first
	err := addSchedule(testSchedule)
	if err != nil {
		t.Errorf("Calling addSchedule() failed. %s", err.Error())
		return
	}

	// Remove the schedule
	err = removeSchedule(testSchedule.Id.Hex())
	if err != nil{
		t.Errorf("Calling removeSchdule() failed. %s", err.Error())
		return
	}

}

func TestQuerySchedule(t *testing.T){
	setup(t)
	mockInit()

	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	testSchedule.Id = bson.NewObjectId()

	// Add the schedule first
	err := addSchedule(testSchedule)
	if err != nil {
		t.Errorf("Calling addSchedule() failed. %s", err.Error())
		return
	}

	// Query the schedule
	schedule, err := querySchedule(testSchedule.Id.Hex())
	if err != nil {
		t.Errorf("Calling querySchedule() failed. %s", err.Error())
		return
	}

	if len(schedule.Id) == 0{
		t.Errorf("Calling querySchedule() failed to return valid schedule")
		return
	}

}

func TestUpdateSchedule(t *testing.T){
	setup(t)
	mockInit()

	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	testUpdateSchedule := models.Schedule{
		Name:      TestScheduleUpdatingName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	testSchedule.Id = bson.NewObjectId()

	// Add the schedule first
	err := addSchedule(testSchedule)
	if err != nil {
		t.Errorf("Calling addSchedule() failed. %s", err.Error())
		return
	}

	// use the same original scheduleId
	testUpdateSchedule.Id = testSchedule.Id

	// update
	err = updateSchedule(testUpdateSchedule)
	if err != nil {
		t.Errorf("Calling updateScheduler() failed. %s", err.Error())
		return
	}

}

// test the scheduleEvent handling in the Scheduler

func TestAddScheduleEvent(t *testing.T){
	setup(t)
	mockInit()

	//parent schedule model
	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	testSchedule.Id = bson.NewObjectId()

	err := addSchedule(testSchedule)
	if err != nil {
		t.Error(err)
	}

	//check queue len
	len := queryQueueLen()   // INFO: Test functionality only.  Should think about exposing thread safe internal function
	if len != 1 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, len, 1)
	}

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

	// query see if it exists
	scheduleEvent, err := queryScheduleEvent(testScheduleEvent.Id.Hex())
	if err != nil {
		t.Errorf("failed to call queryScheduleEvent: %s",err.Error())
	}

	// quick assertion on name
	if scheduleEvent.Name != testScheduleEvent.Name{
		t.Error("failed assertion test on schedulerEvent.Name")
	}
}

func TestRemoveScheduleEvent(t *testing.T){
	setup(t)
	mockInit()

	//parent schedule model
	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	testSchedule.Id = bson.NewObjectId()

	err := addSchedule(testSchedule)
	if err != nil {
		t.Error(err)
	}

	//check queue len
	len := queryQueueLen()   // INFO: Test functionality only.  Should think about exposing thread safe internal function
	if len != 1 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, len, 1)
	}

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

	err = removeScheduleEvent(testScheduleEvent.Id.Hex())
	if err != nil{
		t.Errorf("failed to removeScheduleEvent() %s", err)
	}
}


func TestUpdateScheduleEvent(t *testing.T){

	setup(t)
	mockInit()

	//parent schedule model
	testSchedule := models.Schedule{
		Name:      TestScheduleName,
		Start:     TestScheduleStart,
		End:       TestScheduleEnd,
		Frequency: TestScheduleFrequency,
		Cron:      TestScheduleCron,
		RunOnce:   TestScheduleRunOnce,
	}

	testSchedule.Id = bson.NewObjectId()

	err := addSchedule(testSchedule)
	if err != nil {
		t.Error(err)
	}

	//check queue len
	len := queryQueueLen()   // INFO: Test functionality only.  Should think about exposing thread safe internal function
	if len != 1 {
		t.Errorf(TestUnexpectedMsgFormatStrForIntVal, len, 1)
	}


	//add schedule event to schedule
	testScheduleEvent := models.ScheduleEvent{
		Name:       TestScheduleEventName,
		Parameters: TestScheduleEventParameters,
		Service:    TestScheduleEventService,
		Schedule:   TestScheduleEventSchedule,
		Addressable: models.Addressable{
			Name:     TestScheduleEventAddressableName,
			Protocol: TestScheduleEventAddressableProtocol,
		},
	}

	testScheduleEvent.Id = bson.NewObjectId()

	err = addScheduleEvent(testScheduleEvent)
	if err != nil {
		t.Error(err)
	}

	testScheduleEvent.Name = "UpdatedTestName"

	err = updateScheduleEvent(testScheduleEvent)
	if err != nil{
		t.Errorf("failed to update scheduler event with id %s : %s",TestScheduleEventId,err.Error())
	}

}
