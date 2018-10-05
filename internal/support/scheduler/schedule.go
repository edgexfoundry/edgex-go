//
// Copyright (c) 2018 Tencent
//
// Copyright (c) 2018 Dell Inc.
//
// SPDX-License-Identifier: Apache-2.0
package scheduler

import (
	"errors"
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	queueV1 "gopkg.in/eapache/queue.v1"

	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	ScheduleInterval = 500
)

var (
	mutex                          sync.Mutex
	scheduleQueue                  = queueV1.New()                     // global schedule queue
	scheduleIdToContextMap         = make(map[string]*ScheduleContext) // map : schedule id -> schedule context
	scheduleNameToContextMap       = make(map[string]*ScheduleContext) // map : schedule name -> schedule context
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
	return scheduleQueue.Length()
}

func clearQueue() {
	mutex.Lock()
	defer mutex.Unlock()

	for scheduleQueue.Length() > 0 {
		scheduleQueue.Remove()
	}
}

//endregion

func addScheduleOperation(scheduleId models.Schedule, context *ScheduleContext) {
	scheduleIdToContextMap[scheduleId.Id.Hex()] = context
	scheduleNameToContextMap[scheduleId.Name] = context
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

func querySchedule(scheduleId string) (models.Schedule, error) {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		LoggingClient.Warn("can not find a schedule context with schedule id : " + scheduleId)
		return models.Schedule{}, nil
	}

	LoggingClient.Debug("querying found the schedule with id : " + scheduleId)

	return scheduleContext.Schedule, nil
}

func queryScheduleByName(scheduleName string) (models.Schedule, error) {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleContext, exists := scheduleNameToContextMap[scheduleName]
	if !exists {
		LoggingClient.Warn("can not find a schedule context with schedule name : " + scheduleName)
		return models.Schedule{}, nil
	}

	LoggingClient.Debug("querying found the schedule with name : " + scheduleName)

	return scheduleContext.Schedule, nil
}

func addSchedule(schedule models.Schedule) error {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleId := schedule.Id.Hex()
	LoggingClient.Debug("adding the schedule with id : " + scheduleId + " at time " + schedule.Start)

	if _, exists := scheduleIdToContextMap[scheduleId]; exists {
		LoggingClient.Warn("the schedule context with id " + scheduleId + " already exist ")
		return nil
	}

	context := ScheduleContext{
		ScheduleEventsMap: make(map[string]models.ScheduleEvent),
		MarkedDeleted:     false,
	}

	LoggingClient.Debug("resetting the schedule with id : " + scheduleId)
	context.Reset(schedule)

	addScheduleOperation(schedule, &context)

	LoggingClient.Debug("added the schedule with id : " + scheduleId)

	return nil
}

func updateSchedule(schedule models.Schedule) error {
	mutex.Lock()
	defer mutex.Unlock()

	LoggingClient.Debug("updating the schedule with id : " + schedule.Id.Hex())

	scheduleId := schedule.Id.Hex()
	context, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		LoggingClient.Error("the schedule context with id " + scheduleId + " does not exist ")
		return errors.New("the schedule context with id " + scheduleId + " does not exist ")
	}

	LoggingClient.Debug("resetting the schedule with id " + scheduleId)
	context.Reset(schedule)

	LoggingClient.Debug("updated the schedule with id : " + scheduleId)

	return nil
}

func removeSchedule(scheduleId string) error {
	mutex.Lock()
	defer mutex.Unlock()

	LoggingClient.Debug("removing the schedule with id : " + scheduleId)

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		LoggingClient.Error("can not find schedule context with schedule id : " + scheduleId)
		return errors.New("can not find schedule context with schedule id : " + scheduleId)
	}

	LoggingClient.Debug("removing all the mappings of schedule event id to schedule id : " + scheduleId)
	for eventId := range scheduleContext.ScheduleEventsMap {
		delete(scheduleEventIdToScheduleIdMap, eventId)
	}

	deleteScheduleOperation(scheduleId, scheduleContext)

	LoggingClient.Debug("removed the schedule with id : " + scheduleId)

	return nil
}

func queryScheduleEvent(scheduleEventId string) (models.ScheduleEvent, error) {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleId, exists := scheduleEventIdToScheduleIdMap[scheduleEventId]
	if !exists {
		LoggingClient.Error("can not find schedule id with schedule event id : " + scheduleEventId)
		return models.ScheduleEvent{}, errors.New("Can not find schedule id with schedule event id : " + scheduleEventId)
	}

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		LoggingClient.Warn("can not find a schedule context with schedule id : " + scheduleId)
		return models.ScheduleEvent{}, nil
	}

	scheduleEvent, exists := scheduleContext.ScheduleEventsMap[scheduleEventId]
	if !exists {
		LoggingClient.Error("can not find schedule event with schedule event id : " + scheduleEventId)
		return models.ScheduleEvent{}, errors.New("can not find schedule event with schedule event id : " + scheduleEventId)
	}

	return scheduleEvent, nil
}

func addScheduleEvent(scheduleEvent models.ScheduleEvent) error {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleEventId := scheduleEvent.Id.Hex()
	scheduleName := scheduleEvent.Schedule

	LoggingClient.Debug("adding the schedule event with id " + scheduleEventId + " to schedule : " + scheduleName)

	scheduleContext := scheduleNameToContextMap[scheduleName]

	schedule := scheduleContext.Schedule

	scheduleId := schedule.Id.Hex()
	LoggingClient.Debug("check the schedule with id : " + scheduleId + " exists.")

	if _, exists := scheduleIdToContextMap[scheduleId]; !exists {
		context := ScheduleContext{
			ScheduleEventsMap: make(map[string]models.ScheduleEvent),
			MarkedDeleted:     false,
		}

		context.Reset(schedule)

		addScheduleOperation(schedule, &context)
	}

	addScheduleEventOperation(scheduleId, scheduleEventId, scheduleEvent)

	LoggingClient.Debug("added the schedule event with id : " + scheduleEventId + " to schedule : " + scheduleName)

	return nil
}

func updateScheduleEvent(scheduleEvent models.ScheduleEvent) error {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleEventId := scheduleEvent.Id.Hex()

	LoggingClient.Debug("updating the schedule event with id : " + scheduleEventId)

	oldScheduleId, exists := scheduleEventIdToScheduleIdMap[scheduleEventId]
	if !exists {
		LoggingClient.Error("there is no mapping from schedule event id : " + scheduleEventId + " to schedule.")
		return errors.New("there is no mapping from schedule event id : " + scheduleEventId + " to schedule.")
	}

	scheduleContext, exists := scheduleNameToContextMap[scheduleEvent.Schedule]
	if !exists {
		LoggingClient.Error("query the schedule with name : " + scheduleEvent.Schedule + " occurs a error ")
		return errors.New("there is no mapping from scheduleEvent schedule name: " + scheduleEvent.Schedule + " to schedule")
	}

	//if the schedule event switched schedule
	schedule := scheduleContext.Schedule

	newScheduleId := schedule.Id.Hex()

	if newScheduleId != oldScheduleId {
		LoggingClient.Debug("the schedule event switched schedule from " + oldScheduleId + " to " + newScheduleId)

		//remove Schedule Event
		LoggingClient.Debug("remove the schedule event with id : " + scheduleEventId + " from schedule with id : " + oldScheduleId)
		delete(scheduleContext.ScheduleEventsMap, scheduleEventId)

		//if there are no more events for the schedule, remove the schedule context
		// TODO: Not sure we want to just remove the schedule from the schedule context
		if len(scheduleContext.ScheduleEventsMap) == 0 {
			LoggingClient.Debug("there are no more events for the schedule : " + oldScheduleId + ", remove it.")
			deleteScheduleOperation(oldScheduleId, scheduleContext)
		}

		//add Schedule Event
		LoggingClient.Debug("add the schedule event with id : " + scheduleEventId + " to schedule with id : " + newScheduleId)

		if _, exists := scheduleIdToContextMap[newScheduleId]; !exists {
			context := ScheduleContext{
				ScheduleEventsMap: make(map[string]models.ScheduleEvent),
				MarkedDeleted:     false,
			}
			context.Reset(schedule)

			addScheduleOperation(schedule, &context)
		}

		addScheduleEventOperation(newScheduleId, scheduleEventId, scheduleEvent)
	} else { // if not, just update the schedule event in place
		scheduleContext.ScheduleEventsMap[scheduleEventId] = scheduleEvent
	}

	LoggingClient.Debug("updated the schedule event with id " + scheduleEvent.Id.Hex() + " to schedule id : " + schedule.Id.Hex())

	return nil
}

func removeScheduleEvent(scheduleEventId string) error {
	mutex.Lock()
	defer mutex.Unlock()

	LoggingClient.Debug("removing the schedule event with id " + scheduleEventId)

	scheduleId, exists := scheduleEventIdToScheduleIdMap[scheduleEventId]
	if !exists {
		LoggingClient.Error("can not find schedule id with schedule event id : " + scheduleEventId)
		return errors.New("can not find schedule id with schedule event id : " + scheduleEventId)
	}

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		LoggingClient.Error("can not find schedule context with schedule id : " + scheduleId)
		return errors.New("can not find schedule context with schedule id : " + scheduleId)
	}

	delete(scheduleContext.ScheduleEventsMap, scheduleEventId)

	LoggingClient.Debug("removed the schedule event with id " + scheduleEventId)

	return nil
}

func triggerSchedule() {
	nowEpoch := time.Now().Unix()

	defer func() {
		if err := recover(); err != nil {
			LoggingClient.Error("trigger schedule error : " + err.(string))
		}
	}()

	if scheduleQueue.Length() == 0 {
		return
	}

	var wg sync.WaitGroup

	for i := 0; i < scheduleQueue.Length(); i++ {
		if scheduleQueue.Peek().(*ScheduleContext) != nil {
			scheduleContext := scheduleQueue.Remove().(*ScheduleContext)
			scheduleId := scheduleContext.Schedule.Id.Hex()
			if scheduleContext.MarkedDeleted {
				LoggingClient.Debug("the schedule with id : " + scheduleId + " be marked as deleted, removing it.")
				continue //really delete from the queue
			} else {
				if scheduleContext.NextTime.Unix() <= nowEpoch {
					LoggingClient.Debug("executing schedule, detail : {" + scheduleContext.GetInfo() + "} , at : " + scheduleContext.NextTime.String())

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

func execute(context *ScheduleContext, wg *sync.WaitGroup) error {
	scheduleEventsMap := context.ScheduleEventsMap

	defer wg.Done()

	defer func() {
		if err := recover(); err != nil {
			LoggingClient.Error("schedule execution error : " + err.(string))
		}
	}()

	LoggingClient.Debug(fmt.Sprintf("%d schedule event need to be executed.", len(scheduleEventsMap)))

	//execute schedule event one by one
	for eventId := range scheduleEventsMap {
		LoggingClient.Debug("the event with id : " + eventId + " belongs to schedule : " + context.Schedule.Id.Hex() + " will be executing!")
		scheduleEvent, _ := scheduleEventsMap[eventId]

		executingUrl := getUrlStr(scheduleEvent.Addressable)
		LoggingClient.Debug("the event with id : " + eventId + " will request url : " + executingUrl)

		//TODO: change the method type based on the event

		httpMethod := scheduleEvent.Addressable.HTTPMethod
		if !validMethod(httpMethod) {
			LoggingClient.Error("net/http: invalid method %q", httpMethod)
			return nil
		}

		req, err := http.NewRequest(httpMethod, executingUrl, nil)
		req.Header.Set(ContentTypeKey, ContentTypeJsonValue)

		params := strings.TrimSpace(scheduleEvent.Parameters)

		if len(params) > 0 {
			req.Header.Set(ContentLengthKey, string(len(params)))
		}

		if err != nil {
			LoggingClient.Error("create new request occurs error : " + err.Error())
		}

		client := &http.Client{
			Timeout: time.Duration(Configuration.Service.Timeout) * time.Millisecond,
		}
		responseBytes, statusCode, err := sendRequestAndGetResponse(client, req)
		responseStr := string(responseBytes)

		LoggingClient.Debug(fmt.Sprintf("execution returns status code : %d", statusCode))
		LoggingClient.Debug("execution returns response content : " + responseStr)
	}

	context.UpdateNextTime()
	context.UpdateIterations()

	if context.IsComplete() {
		LoggingClient.Debug("completed schedule, detail : " + context.GetInfo())
	} else {
		LoggingClient.Debug("requeue schedule, detail : " + context.GetInfo())
		scheduleQueue.Add(context)
	}
	return nil
}

func getUrlStr(addressable models.Addressable) string {
	return addressable.GetBaseURL() + addressable.Path
}

func sendRequestAndGetResponse(client *http.Client, req *http.Request) ([]byte, int, error) {
	resp, err := client.Do(req)

	if err != nil {
		println(err.Error())
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

func validMethod(method string) bool {
	/*
	     Method         = "OPTIONS"                ; Section 9.2
	                    | "GET"                    ; Section 9.3
	                    | "HEAD"                   ; Section 9.4
	                    | "POST"                   ; Section 9.5
	                    | "PUT"                    ; Section 9.6
	                    | "DELETE"                 ; Section 9.7
	                    | "TRACE"                  ; Section 9.8
	                    | "CONNECT"                ; Section 9.9
	                    | extension-method
	   extension-method = token
	     token          = 1*<any CHAR except CTLs or separators>
	*/
	a := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "CONNECT"}
	method = strings.ToUpper(method)
	return contains(a, method)
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// Utility function for adding configured locally schedulers and scheduled events
func AddSchedulers() error {

	LoggingClient.Info(fmt.Sprintf("loading default schedules and schedule events..."))

	schedules := Configuration.Schedules

	for i := range schedules {

		schedule := models.Schedule{
			BaseObject: models.BaseObject{},
			Id:         bson.NewObjectId(),
			Name:       schedules[i].Name,
			Start:      schedules[i].Start,
			End:        schedules[i].End,
			Frequency:  schedules[i].Frequency,
			Cron:       schedules[i].Cron,
			RunOnce:    schedules[i].RunOnce,
		}
		err := addSchedule(schedule)
		if err != nil {
			return LoggingClient.Error("AddDefaultSchedulers() - failed to load schedule %s", err.Error())
		} else {
			LoggingClient.Info(fmt.Sprintf("added default schedule %s", schedule.Name))
		}
	}

	scheduleEvents := Configuration.ScheduleEvents

	for e := range scheduleEvents {

		addressable := models.Addressable{
			// TODO: find a better way to initialize perhaps core-metadata
			Id:         bson.NewObjectId(),
			Name:       fmt.Sprintf("Schedule-%s", scheduleEvents[e].Name),
			Path:       scheduleEvents[e].Path,
			Port:       scheduleEvents[e].Port,
			Protocol:   scheduleEvents[e].Protocol,
			HTTPMethod: scheduleEvents[e].Method,
			Address:    scheduleEvents[e].Host,
		}

		scheduleEvent := models.ScheduleEvent{
			Id:          bson.NewObjectId(),
			Name:        scheduleEvents[e].Name,
			Schedule:    scheduleEvents[e].Schedule,
			Parameters:  scheduleEvents[e].Parameters,
			Service:     scheduleEvents[e].Service,
			Addressable: addressable,
		}

		err := addScheduleEvent(scheduleEvent)
		if err != nil {
			return LoggingClient.Error("AddDefaultSchedulers() - failed to load schedule event %s", err.Error())
		} else {
			LoggingClient.Info(fmt.Sprintf("added default schedule event %s", scheduleEvent.Name))
		}
	}

	LoggingClient.Info(fmt.Sprintf("completed loading default schedules and schedule events"))
	return nil
}

//endregion
