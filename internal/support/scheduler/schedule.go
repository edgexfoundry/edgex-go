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
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	queueV1 "gopkg.in/eapache/queue.v1"
)

const (
	ScheduleInterval = 500
)

//the schedule specific shared variables
var (
	mutex                                 sync.Mutex
	scheduleQueue                         = queueV1.New()                     // global schedule queue
	scheduleIdToContextMap                = make(map[string]*ScheduleContext) // map : schedule id -> schedule context
	scheduleNameToContextMap              = make(map[string]*ScheduleContext) // map : schedule name -> schedule context
	scheduleEventIdToScheduleIdMap        = make(map[string]string)           // map : schedule event id -> schedule id
	scheduleEventNameToScheduleIdMap      = make(map[string]string)           // map : schedule event name -> schedule id
	scheduleEventNameToScheduleEventIdMap = make(map[string]string)           // map : schedule event name -> schedule event id
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

// utility function
func clearQueue() {
	mutex.Lock()
	defer mutex.Unlock()

	for scheduleQueue.Length() > 0 {
		scheduleQueue.Remove()
	}
}

func clearMaps() {
	scheduleIdToContextMap = make(map[string]*ScheduleContext)   // map : schedule id -> schedule context
	scheduleNameToContextMap = make(map[string]*ScheduleContext) // map : schedule name -> schedule context
	scheduleEventIdToScheduleIdMap = make(map[string]string)     // map : schedule event id -> schedule id
	scheduleEventNameToScheduleIdMap = make(map[string]string)   // map : schedule event name -> schedule id
	scheduleEventNameToScheduleEventIdMap = make(map[string]string)
}

func addScheduleOperation(scheduleId models.Schedule, context *ScheduleContext) {
	scheduleIdToContextMap[scheduleId.Id.Hex()] = context
	scheduleNameToContextMap[scheduleId.Name] = context
	scheduleQueue.Add(context)
}

func deleteScheduleOperation(schedule models.Schedule, scheduleContext *ScheduleContext) {
	scheduleContext.MarkedDeleted = true
	scheduleIdToContextMap[schedule.Id.Hex()] = scheduleContext
	scheduleNameToContextMap[schedule.Name] = scheduleContext
	delete(scheduleIdToContextMap, schedule.Id.Hex())
}

func addScheduleEventOperation(schedule models.Schedule, scheduleEvent models.ScheduleEvent) {
	scheduleContext, _ := scheduleIdToContextMap[schedule.Id.Hex()]
	scheduleContext.ScheduleEventsMap[scheduleEvent.Id.Hex()] = scheduleEvent
	scheduleEventIdToScheduleIdMap[scheduleEvent.Id.Hex()] = schedule.Id.Hex()
	scheduleEventNameToScheduleIdMap[scheduleEvent.Name] = schedule.Id.Hex()
	scheduleEventNameToScheduleEventIdMap[scheduleEvent.Name] = scheduleEvent.Id.Hex()
}

func querySchedule(scheduleId string) (models.Schedule, error) {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find a schedule context with schedule id : %s", scheduleId)
		LoggingClient.Info(logMsg)
		return models.Schedule{}, errors.New(logMsg)
	}

	LoggingClient.Debug(fmt.Sprintf("querying found the schedule with id : %s", scheduleId))

	return scheduleContext.Schedule, nil
}

func queryScheduleByName(scheduleName string) (models.Schedule, error) {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleContext, exists := scheduleNameToContextMap[scheduleName]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find schedule id with schedule with name : %s", scheduleName)
		LoggingClient.Info(logMsg)
		return models.Schedule{}, errors.New(logMsg)
	}

	LoggingClient.Debug(fmt.Sprintf("scheduler found the schedule with name : %s", scheduleName))

	return scheduleContext.Schedule, nil
}

func addSchedule(schedule models.Schedule) error {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleId := schedule.Id.Hex()
	LoggingClient.Debug(fmt.Sprintf("adding the schedule with id : %s at time %s", scheduleId, schedule.Start))

	if _, exists := scheduleIdToContextMap[scheduleId]; exists {
		LoggingClient.Debug(fmt.Sprintf("the schedule context with id : %s already exists", scheduleId))
		return nil
	}

	context := ScheduleContext{
		ScheduleEventsMap: make(map[string]models.ScheduleEvent),
		MarkedDeleted:     false,
	}

	LoggingClient.Debug(fmt.Sprintf("resetting the schedule with id : %s", scheduleId))
	context.Reset(schedule)

	addScheduleOperation(schedule, &context)

	LoggingClient.Debug(fmt.Sprintf("added the schedule with id : %s ", scheduleId))

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
		logMsg := fmt.Sprintf("scheduler could not find schedule context with schedule id : %s", scheduleId)
		return errors.New(logMsg)
	}

	LoggingClient.Debug("removing all the mappings of schedule event id to schedule id : " + scheduleId)
	for eventId := range scheduleContext.ScheduleEventsMap {
		delete(scheduleEventIdToScheduleIdMap, eventId)
	}

	deleteScheduleOperation(scheduleContext.Schedule, scheduleContext)

	LoggingClient.Debug("removed the schedule with id : " + scheduleId)

	return nil
}

func queryScheduleEvent(scheduleEventId string) (models.ScheduleEvent, error) {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleId, exists := scheduleEventIdToScheduleIdMap[scheduleEventId]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find schedule id with schedule event id : %s", scheduleEventId)
		return models.ScheduleEvent{}, errors.New(logMsg)
	}

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		LoggingClient.Warn("scheduler could not find a schedule context with schedule id : " + scheduleId)
		return models.ScheduleEvent{}, nil
	}

	scheduleEvent, exists := scheduleContext.ScheduleEventsMap[scheduleEventId]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find schedule event with schedule event id : %s", scheduleEventId)
		return models.ScheduleEvent{}, errors.New(logMsg)
	}

	return scheduleEvent, nil
}

func queryScheduleEventByName(scheduleEventName string) (models.ScheduleEvent, error) {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleId, exists := scheduleEventNameToScheduleIdMap[scheduleEventName]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find schedule id with schedule event name : %s", scheduleEventName)
		LoggingClient.Error(logMsg)
		return models.ScheduleEvent{}, errors.New(logMsg)
	}

	scheduleEventId, exists := scheduleEventNameToScheduleEventIdMap[scheduleEventName]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find schedule event id with schedule event name : %s", scheduleEventName)
		LoggingClient.Error(logMsg)
		return models.ScheduleEvent{}, errors.New(logMsg)
	}

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		logMsg := fmt.Sprintf("could not find a schedule context with schedule id : %s", scheduleId)
		LoggingClient.Error(logMsg)
		return models.ScheduleEvent{}, errors.New(logMsg)
	}

	scheduleEvent, exists := scheduleContext.ScheduleEventsMap[scheduleEventId]
	if !exists {
		logMsg := fmt.Sprintf("could not find schedule event with schedule event id :  %s", scheduleContext.Schedule.Id.Hex())
		LoggingClient.Error(logMsg)
		return models.ScheduleEvent{}, errors.New(logMsg)
	}

	return scheduleEvent, nil
}

func addScheduleEvent(scheduleEvent models.ScheduleEvent) error {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleEventId := scheduleEvent.Id.Hex()
	scheduleName := scheduleEvent.Schedule

	LoggingClient.Debug(fmt.Sprintf("adding the schedule event with id  : %s to schedule : %s ", scheduleEventId, scheduleName))

	scheduleContext := scheduleNameToContextMap[scheduleName]

	schedule := scheduleContext.Schedule

	scheduleId := schedule.Id.Hex()
	LoggingClient.Debug(fmt.Sprintf("check the schedule with id : %s exists.", scheduleId))

	if _, exists := scheduleIdToContextMap[scheduleId]; !exists {
		context := ScheduleContext{
			ScheduleEventsMap: make(map[string]models.ScheduleEvent),
			MarkedDeleted:     false,
		}

		context.Reset(schedule)

		addScheduleOperation(schedule, &context)
	}

	addScheduleEventOperation(schedule, scheduleEvent)

	LoggingClient.Debug(fmt.Sprintf("added the schedule event with id : %s to schedule : %s", scheduleEventId, scheduleName))

	return nil
}

func updateScheduleEvent(scheduleEvent models.ScheduleEvent) error {
	mutex.Lock()
	defer mutex.Unlock()

	scheduleEventId := scheduleEvent.Id.Hex()

	LoggingClient.Debug("updating the schedule event with id : " + scheduleEventId)

	oldScheduleId, exists := scheduleEventIdToScheduleIdMap[scheduleEventId]
	if !exists {
		logMsg := fmt.Sprintf("there is no mapping from schedule event id : %s to schedule.", scheduleEventId)
		LoggingClient.Error(logMsg)
		return errors.New(logMsg)
	}

	scheduleContext, exists := scheduleNameToContextMap[scheduleEvent.Schedule]
	if !exists {
		logMsg := fmt.Sprintf("query the schedule with name : %s  and did not exist.", scheduleEvent.Schedule)
		return errors.New(logMsg)
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
			deleteScheduleOperation(schedule, scheduleContext)
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

		addScheduleEventOperation(schedule, scheduleEvent)
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
		logMsg := fmt.Sprintf("could not find schedule id with schedule event id : %s", scheduleEventId)
		return errors.New(logMsg)
	}

	scheduleContext, exists := scheduleIdToContextMap[scheduleId]
	if !exists {
		logMsg := fmt.Sprintf("can not find schedule context with schedule id : %s", scheduleId)
		return errors.New(logMsg)
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
			LoggingClient.Error(fmt.Sprintf("net/http: invalid method %q", httpMethod))
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
		//println(err.Error())
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
