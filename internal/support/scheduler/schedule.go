//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"errors"
	"fmt"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"gopkg.in/eapache/queue.v1"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	ScheduleEventInvokeRequestTimeOut = 5000
	ScheduleInterval                  = 500
)

//the schedule specific shared variables
var (
	mutex                          sync.Mutex
	scheduleQueue                  = queue.New()                       // global schedule queue
	scheduleIdToContextMap         = make(map[string]*ScheduleContext) // map : schedule id -> schedule context
	scheduleEventIdToScheduleIdMap = make(map[string]string)           // map : schedule event id -> schedule id
)

func StartTicker() {
	go func() {
		for range ticker.C {
			triggerSchedule()
		}
	}()
}

func StopTicker() {
	ticker.Stop()
}

//region queue util for testing
func queryQueueLen() int {
	mutex.Lock()
	defer mutex.Unlock()

	len := scheduleQueue.Length()

	return len
}

func clearQueue() {
	mutex.Lock()
	defer mutex.Unlock()

	for scheduleQueue.Length() > 0 {
		scheduleQueue.Remove()
	}
}

//endregion

//region util methods access shared variable, Warning : these methods should be called in sychronization block
func addScheduleOperation(scheduleId string, context *ScheduleContext) {
	scheduleIdToContextMap[scheduleId] = context
	scheduleQueue.Add(context)
}

func deleteScheduleOperation(scheduleId string, scheduleContext *ScheduleContext) {
	scheduleContext.MarkedDeleted = true
	scheduleIdToContextMap[scheduleId] = scheduleContext
	delete(scheduleIdToContextMap, scheduleId)
}

func addScheduleEventOperation(scheduleId string, scheduleEventId string, scheduleEvent models.ScheduleEvent) {
	scheduleContext, _ := scheduleIdToContextMap[scheduleId]
	scheduleContext.ScheduleEventsMap[scheduleEventId] = scheduleEvent
	scheduleEventIdToScheduleIdMap[scheduleEventId] = scheduleId
}

//endregion

//region schedule CRUD
func querySchedule(scheduleId string) (models.Schedule, error) {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		loggingClient.Warn("can not find a schedule context with schedule id : " + scheduleId)
		return models.Schedule{}, nil
	}

	return scheduleContext.Schedule, nil
}

func addSchedule(schedule models.Schedule) error {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleId := schedule.Id.Hex()
	loggingClient.Info("adding the schedule with id : " + scheduleId)

	if _, exists := scheduleIdToContextMap[scheduleId]; exists {
		loggingClient.Warn("the schedule context with id " + scheduleId + " already exist ")
		return nil
	}

	context := ScheduleContext{
		ScheduleEventsMap: make(map[string]models.ScheduleEvent),
		MarkedDeleted:     false,
	}

	loggingClient.Debug("resetting the schedule with id : " + scheduleId)
	context.Reset(schedule)

	addScheduleOperation(scheduleId, &context)

	loggingClient.Info("added the schedule with id : " + scheduleId)

	return nil
}

func updateSchedule(schedule models.Schedule) error {
	mutex.Lock()
	defer mutex.Unlock()

	loggingClient.Info("updating the schedule with id : " + schedule.Id.Hex())

	scheduleId := schedule.Id.Hex()
	context, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		loggingClient.Error("the schedule context with id " + scheduleId + " dose not exist ")

		return errors.New("the schedule context with id " + scheduleId + " dose not exist ")
	}

	loggingClient.Debug("resetting the schedule whit id " + scheduleId)
	context.Reset(schedule)

	loggingClient.Info("updated the schedule with id : " + scheduleId)

	return nil
}

func removeSchedule(scheduleId string) error {
	mutex.Lock()
	defer mutex.Unlock()

	loggingClient.Info("removing the schedule with id : " + scheduleId)

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		loggingClient.Error("can not find schedule context with schedule id : " + scheduleId)
		return errors.New("can not find schedule context with schedule id : " + scheduleId)
	}

	loggingClient.Debug("removing all the mappings of schedule event id to schedule id : " + scheduleId)
	for eventId := range scheduleContext.ScheduleEventsMap {
		delete(scheduleEventIdToScheduleIdMap, eventId)
	}

	deleteScheduleOperation(scheduleId, scheduleContext)

	loggingClient.Info("removed the schedule with id : " + scheduleId)

	return nil
}

//endregion

//region schedule event CRUD
func queryScheduleEvent(scheduleEventId string) (models.ScheduleEvent, error) {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleId, exists := scheduleEventIdToScheduleIdMap[scheduleEventId]
	if !exists {
		loggingClient.Error("can not find schedule id with schedule event id : " + scheduleEventId)
		return models.ScheduleEvent{}, errors.New("Can not find schedule id with schedule event id : " + scheduleEventId)
	}

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		loggingClient.Warn("can not find a schedule context with schedule id : " + scheduleId)
		return models.ScheduleEvent{}, nil
	}

	scheduleEvent, exists := scheduleContext.ScheduleEventsMap[scheduleEventId]
	if !exists {
		loggingClient.Error("can not find schedule event with schedule event id : " + scheduleEventId)
		return models.ScheduleEvent{}, errors.New("can not find schedule event with schedule event id : " + scheduleEventId)
	}

	return scheduleEvent, nil
}

func addScheduleEvent(scheduleEvent models.ScheduleEvent) error {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleEventId := scheduleEvent.Id.Hex()
	scheduleName := scheduleEvent.Schedule

	loggingClient.Info("adding the schedule event with id " + scheduleEventId + " to schedule : " + scheduleName)

	schedule, err := schedulerClient.QueryScheduleWithName(scheduleName)
	if err != nil {
		loggingClient.Error("query the schedule with name : " + scheduleName + " occurs a error : " + err.Error())
		return err
	}

	scheduleId := schedule.Id.Hex()
	loggingClient.Info("check the schedule with id : " + scheduleId + " exists.")

	if _, exists := scheduleIdToContextMap[scheduleId]; !exists {
		context := ScheduleContext{
			ScheduleEventsMap: make(map[string]models.ScheduleEvent),
			MarkedDeleted:     false,
		}

		context.Reset(schedule)

		addScheduleOperation(scheduleId, &context)
	}

	addScheduleEventOperation(scheduleId, scheduleEventId, scheduleEvent)

	loggingClient.Info("added the schedule event with id : " + scheduleEventId + " to schedule : " + scheduleName)

	return nil
}

func updateScheduleEvent(scheduleEvent models.ScheduleEvent) error {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleEventId := scheduleEvent.Id.Hex()

	loggingClient.Info("updating the schedule event with id : " + scheduleEventId)

	oldScheduleId, exists := scheduleEventIdToScheduleIdMap[scheduleEventId]
	if !exists {
		loggingClient.Error("there is no mapping from schedule event id : " + scheduleEventId + " to schedule.")
		return errors.New("there is no mapping from schedule event id : " + scheduleEventId + " to schedule.")
	}

	schedule, err := schedulerClient.QueryScheduleWithName(scheduleEvent.Schedule)
	if err != nil {
		loggingClient.Error("query the schedule with name : " + scheduleEvent.Schedule + " occurs a error : " + err.Error())
		return err
	}

	scheduleContext, exists := scheduleIdToContextMap[oldScheduleId]
	if !exists {
		loggingClient.Error("can not find the mapping from the old schedule id : " + oldScheduleId + " to schedule context")
		return errors.New("can not find the mapping from the old schedule id : " + oldScheduleId + " to schedule context")
	}

	//if the schedule event switched schedule
	newScheduleId := schedule.Id.Hex()
	if newScheduleId != oldScheduleId {
		loggingClient.Debug("the schedule event switched schedule from " + oldScheduleId + " to " + newScheduleId)

		//remove Schedule Event
		loggingClient.Debug("remove the schedule event with id : " + scheduleEventId + " from schedule with id : " + oldScheduleId)
		delete(scheduleContext.ScheduleEventsMap, scheduleEventId)

		//if there are no more events for the schedule, remove the schedule context
		if len(scheduleContext.ScheduleEventsMap) == 0 {
			loggingClient.Debug("there are no more events for the schedule : " + oldScheduleId + ", remove it.")

			deleteScheduleOperation(oldScheduleId, scheduleContext)
		}

		//add Schedule Event
		loggingClient.Debug("add the schedule event with id : " + scheduleEventId + " to schedule with id : " + newScheduleId)

		if _, exists := scheduleIdToContextMap[newScheduleId]; !exists {
			context := ScheduleContext{
				ScheduleEventsMap: make(map[string]models.ScheduleEvent),
				MarkedDeleted:     false,
			}
			context.Reset(schedule)

			addScheduleOperation(newScheduleId, &context)
		}

		addScheduleEventOperation(newScheduleId, scheduleEventId, scheduleEvent)
	} else { // if not, just update the schedule event in place
		scheduleContext.ScheduleEventsMap[scheduleEventId] = scheduleEvent
	}

	loggingClient.Info("updated the schedule event with id " + scheduleEvent.Id.Hex() + " to schedule id : " + schedule.Id.Hex())

	return nil
}

func removeScheduleEvent(scheduleEventId string) error {
	mutex.Lock()
	defer mutex.Unlock()

	loggingClient.Info("removing the schedule event with id " + scheduleEventId)

	scheduleId, exists := scheduleEventIdToScheduleIdMap[scheduleEventId]
	if !exists {
		loggingClient.Error("can not find schedule id with schedule event id : " + scheduleEventId)
		return errors.New("can not find schedule id with schedule event id : " + scheduleEventId)
	}

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		loggingClient.Error("can not find schedule context with schedule id : " + scheduleId)
		return errors.New("can not find schedule context with schedule id : " + scheduleId)
	}

	delete(scheduleContext.ScheduleEventsMap, scheduleEventId)

	//if there are no more events for the schedule, remove the schedule context
	if len(scheduleContext.ScheduleEventsMap) == 0 {
		loggingClient.Info("there are no more events for the schedule : " + scheduleId + ", remove it.")
		deleteScheduleOperation(scheduleId, scheduleContext)
	}

	loggingClient.Info("removed the schedule event with id " + scheduleEventId)

	return nil
}

//endregion

//region schedule event execution
func triggerSchedule() {
	nowEpoch := time.Now().Unix()

	defer func() {
		if err := recover(); err != nil {
			loggingClient.Error("trigger schedule error : " + err.(string))
		}
	}()

	if scheduleQueue.Length() == 0 {
		return
	}

	loggingClient.Debug(fmt.Sprintf("%d item in schedule queue.", scheduleQueue.Length()))

	var wg sync.WaitGroup

	for i := 0; i < scheduleQueue.Length(); i++ {
		if scheduleQueue.Peek().(*ScheduleContext) != nil {
			scheduleContext := scheduleQueue.Remove().(*ScheduleContext)
			scheduleId := scheduleContext.Schedule.Id.Hex()
			loggingClient.Debug("check schedule with id : " + scheduleId)
			if scheduleContext.MarkedDeleted {
				loggingClient.Debug("the schedule with id : " + scheduleId + " be marked as deleted, removing it.")
				continue //really delete from the queue
			} else {
				loggingClient.Debug("schedule with id : " + scheduleId + " next schedule time is : " + scheduleContext.NextTime.String())
				if scheduleContext.NextTime.Unix() <= nowEpoch {
					loggingClient.Info("executing schedule, detail : {" + scheduleContext.GetInfo() + "} , at : " + scheduleContext.NextTime.String())

					wg.Add(1)

					//execute it in a individual go routine
					go execute(scheduleContext, &wg)
				} else {
					scheduleQueue.Add(scheduleContext)
				}
			}
		}
	}

	wg.Wait()
}

func execute(context *ScheduleContext, wg *sync.WaitGroup) {
	scheduleEventsMap := context.ScheduleEventsMap

	defer wg.Done()

	defer func() {
		if err := recover(); err != nil {
			loggingClient.Error("schedule execution error : " + err.(string))
		}
	}()

	loggingClient.Debug(fmt.Sprintf("%d schedule event need to be executed.", len(scheduleEventsMap)))

	//execute schedule event one by one
	for eventId := range scheduleEventsMap {
		loggingClient.Info("the event with id : " + eventId + " belongs schedule : " + context.Schedule.Id.Hex() + " is be executing!")
		scheduleEvent, _ := scheduleEventsMap[eventId]

		executingUrl := getUrlStr(scheduleEvent.Addressable)
		loggingClient.Debug("the event with id : " + eventId + " will request url : " + executingUrl)

		req, err := http.NewRequest(http.MethodPost, executingUrl, nil)
		req.Header.Set(ContentTypeKey, ContentTypeJsonValue)
		req.Header.Set(ContentLengthKey, string(len(scheduleEvent.Parameters)))

		if err != nil {
			loggingClient.Error("create new request occurs error : " + err.Error())
		}

		client := &http.Client{
			Timeout: ScheduleEventInvokeRequestTimeOut,
		}
		responseBytes, statusCode, err := sendRequestAndGetResponse(client, req)
		responseStr := string(responseBytes)

		loggingClient.Debug(fmt.Sprintf("execution returns status code : %d", statusCode))
		loggingClient.Debug("execution returns response content : " + responseStr)
	}

	context.UpdateNextTime()
	context.UpdateIterations()

	if context.IsComplete() {
		loggingClient.Info("completed schedule, detail : " + context.GetInfo())
	} else {
		loggingClient.Info("requeue schedule, detail : " + context.GetInfo())
		scheduleQueue.Add(context)
	}
}

func getUrlStr(addressable models.Addressable) string {
	return addressable.GetBaseURL() + addressable.Path
}

func sendRequestAndGetResponse(client *http.Client, req *http.Request) ([]byte, int, error) {
	resp, err := client.Do(req)

	if err != nil {
		return []byte{}, 500, err
	}

	defer resp.Body.Close()
	resp.Close = true

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, 500, err
	}

	return bodyBytes, resp.StatusCode, nil
}

//endregion
