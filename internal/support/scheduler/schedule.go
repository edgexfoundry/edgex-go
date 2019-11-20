//
// Copyright (c) 2018 Tencent
//
// Copyright (c) 2018 Dell Inc.
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	queueV1 "gopkg.in/eapache/queue.v1"
)

// the interval specific shared variables
var (
	mutex                                   sync.Mutex
	intervalQueue                           = queueV1.New()
	intervalIdToContextMap                  = make(map[string]*IntervalContext)
	intervalNameToContextMap                = make(map[string]*IntervalContext)
	intervalNameToIdMap                     = make(map[string]string)
	intervalActionIdToIntervalMap           = make(map[string]string)
	intervalActionNameToIntervalMap         = make(map[string]string)
	intervalActionNameToIntervalActionIdMap = make(map[string]string)
)

func StartTicker(loggingClient logger.LoggingClient) {
	go func() {
		for range ticker.C {
			triggerInterval(loggingClient)
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

	for intervalQueue.Length() > 0 {
		intervalQueue.Remove()
	}
}

func clearMaps() {
	intervalIdToContextMap = make(map[string]*IntervalContext)        // map : interval id -> interval context
	intervalNameToContextMap = make(map[string]*IntervalContext)      // map : interval name -> interval context
	intervalNameToIdMap = make(map[string]string)                     // map : interval name -> interval id
	intervalActionIdToIntervalMap = make(map[string]string)           // map : interval action id -> interval id
	intervalActionNameToIntervalMap = make(map[string]string)         // map : interval action name -> interval id
	intervalActionNameToIntervalActionIdMap = make(map[string]string) // map : interval action name -> interval actionId

}

func addIntervalOperation(interval contract.Interval, context *IntervalContext) {
	intervalIdToContextMap[interval.ID] = context
	intervalNameToContextMap[interval.Name] = context
	intervalNameToIdMap[interval.Name] = interval.ID
	intervalQueue.Add(context)
}

func deleteIntervalOperation(interval contract.Interval, intervalContext *IntervalContext) {
	intervalContext.MarkedDeleted = true
	intervalIdToContextMap[interval.ID] = intervalContext
	intervalNameToContextMap[interval.Name] = intervalContext
	delete(intervalIdToContextMap, interval.ID)
	delete(intervalNameToContextMap, interval.Name)
}

func addIntervalActionOperation(interval contract.Interval, intervalAction contract.IntervalAction) {
	intervalContext, _ := intervalIdToContextMap[interval.ID]
	intervalContext.IntervalActionsMap[intervalAction.ID] = intervalAction
	intervalActionIdToIntervalMap[intervalAction.ID] = interval.ID
	intervalActionNameToIntervalMap[intervalAction.Name] = interval.ID
	intervalActionNameToIntervalActionIdMap[intervalAction.Name] = intervalAction.ID
}

func (qc *QueueClient) Connect() (string, error) {
	return "alive..", nil
}
func (qc *QueueClient) QueryIntervalByID(intervalId string) (contract.Interval, error) {

	mutex.Lock()
	defer mutex.Unlock()

	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		return contract.Interval{},
		errors.New(fmt.Sprintf("scheduler could not find a interval context with interval id : %s", intervalId))
	}

	qc.loggingClient.Debug(fmt.Sprintf("querying found the interval with id : %s", intervalId))

	return intervalContext.Interval, nil
}

func (qc *QueueClient) QueryIntervalByName(intervalName string) (contract.Interval, error) {

	mutex.Lock()
	defer mutex.Unlock()

	intervalContext, exists := intervalNameToContextMap[intervalName]
	if !exists {
		return contract.Interval{},
		errors.New(fmt.Sprintf("scheduler could not find interval with interval with name : %s", intervalName))
	}

	qc.loggingClient.Debug(fmt.Sprintf("scheduler found the interval with name : %s", intervalName))

	return intervalContext.Interval, nil
}

func (qc *QueueClient) AddIntervalToQueue(interval contract.Interval) error {
	mutex.Lock()
	defer mutex.Unlock()

	intervalId := interval.ID
	qc.loggingClient.Debug(fmt.Sprintf("adding the interval with id : %s at time %s", intervalId, interval.Start))

	if _, exists := intervalIdToContextMap[intervalId]; exists {
		qc.loggingClient.Debug(fmt.Sprintf("the interval context with id : %s already exists", intervalId))
		return nil
	}

	context := IntervalContext{
		IntervalActionsMap: make(map[string]contract.IntervalAction),
		MarkedDeleted:      false,
	}

	qc.loggingClient.Debug(fmt.Sprintf("resetting the interval with id : %s", intervalId))
	context.Reset(interval, qc.loggingClient)

	addIntervalOperation(interval, &context)

	qc.loggingClient.Info(fmt.Sprintf("added the interval with id : %s into the scheduler queue", intervalId))

	return nil
}

func (qc *QueueClient) UpdateIntervalInQueue(interval contract.Interval) error {
	mutex.Lock()
	defer mutex.Unlock()

	intervalId := interval.ID
	context, exists := intervalIdToContextMap[intervalId]
	if !exists {
		return errors.New("the interval context with id " + intervalId + " does not exist ")
	}

	// remove the old map entry and create new one
	_, exists = intervalNameToIdMap[context.Interval.Name]
	if exists {
		delete(intervalNameToIdMap, context.Interval.Name)
	}

	// add new map entry
	intervalNameToIdMap[interval.Name] = interval.ID

	qc.loggingClient.Debug(fmt.Sprintf("resting the interval context with id: %s in the scheduler queue", intervalId))
	context.Reset(interval, qc.loggingClient)

	qc.loggingClient.Info(fmt.Sprintf("updated the interval with id: %s in the scheduler queue", intervalId))

	return nil
}

func (qc *QueueClient) RemoveIntervalInQueue(intervalId string) error {
	mutex.Lock()
	defer mutex.Unlock()

	qc.loggingClient.Debug(fmt.Sprintf("removing the interval with id: %s ", intervalId))

	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		return errors.New(fmt.Sprintf("scheduler could not find interval context with interval id : %s", intervalId))
	}

	qc.loggingClient.Debug(fmt.Sprintf("removing all the mappings of interval action id to interval id: %s ", intervalId))
	for eventId := range intervalContext.IntervalActionsMap {
		delete(intervalActionIdToIntervalMap, eventId)
	}

	deleteIntervalOperation(intervalContext.Interval, intervalContext)

	qc.loggingClient.Info(fmt.Sprintf("removed the interval with id: %s from the scheduler queue", intervalId))

	return nil
}

func (qc *QueueClient) QueryIntervalActionByID(intervalActionId string) (contract.IntervalAction, error) {

	mutex.Lock()
	defer mutex.Unlock()

	intervalId, exists := intervalActionIdToIntervalMap[intervalActionId]
	if !exists {
		return contract.IntervalAction{},
		errors.New(fmt.Sprintf("scheduler could not find interval id with interval action id : %s", intervalActionId))
	}

	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		qc.loggingClient.Warn("scheduler could not find a interval context with interval id : " + intervalId)
		return contract.IntervalAction{}, nil
	}

	intervalAction, exists := intervalContext.IntervalActionsMap[intervalActionId]
	if !exists {
		return contract.IntervalAction{},
		errors.New(fmt.Sprintf("scheduler could not find interval action with interval action id : %s", intervalActionId))
	}

	return intervalAction, nil
}

func (qc *QueueClient) QueryIntervalActionByName(intervalActionName string) (contract.IntervalAction, error) {

	mutex.Lock()
	defer mutex.Unlock()

	intervalId, exists := intervalActionNameToIntervalMap[intervalActionName]
	if !exists {
		return contract.IntervalAction{},
		errors.New(fmt.Sprintf("scheduler could not find interval id with intervalAction name : %s", intervalActionName))
	}

	intervalActionId, exists := intervalActionNameToIntervalActionIdMap[intervalActionName]
	if !exists {
		return contract.IntervalAction{},
		errors.New(fmt.Sprintf(
			"scheduler could not find intervalAction id with intervalAction name : %s",
			intervalActionName))
	}

	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		return contract.IntervalAction{},
		errors.New(fmt.Sprintf(
			"scheduler could not find a interval context with interval id : %s",
			intervalId))
	}

	intervalAction, exists := intervalContext.IntervalActionsMap[intervalActionId]
	if !exists {
		return contract.IntervalAction{},
		errors.New(fmt.Sprintf(
			"scheduler could not find intervalAction with intervalAction id :  %s",
			intervalContext.Interval.ID))
	}

	return intervalAction, nil
}

func (qc *QueueClient) AddIntervalActionToQueue(intervalAction contract.IntervalAction) error {

	mutex.Lock()
	defer mutex.Unlock()

	intervalActionId := intervalAction.ID
	intervalName := intervalAction.Interval
	intervalActionName := intervalAction.Name

	qc.loggingClient.Debug(fmt.Sprintf(
		"adding the intervalAction with id  : %s to interval : %s into the queue",
		intervalActionId,
		intervalName))

	if _, exists := intervalActionNameToIntervalMap[intervalActionName]; exists {
		return errors.New(fmt.Sprintf("scheduler found existing intervalAction with same name: %s", intervalName))
	}

	// Ensure we have an existing Interval
	intervalId, exists := intervalNameToIdMap[intervalName]
	if !exists {
		return errors.New(fmt.Sprintf("scheduler could not find a interval with interval name : %s", intervalName))
	}

	// Get the Schedule Context
	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		return errors.New(fmt.Sprintf("scheduler could not find a interval with interval name : %s", intervalName))
	}

	interval := intervalContext.Interval

	addIntervalActionOperation(interval, intervalAction)

	qc.loggingClient.Info(fmt.Sprintf(
		"added the intervalAction with id: %s to interal: %s into the queue",
		intervalActionId,
		intervalName))

	return nil
}

func (qc *QueueClient) UpdateIntervalActionQueue(intervalAction contract.IntervalAction) error {

	mutex.Lock()
	defer mutex.Unlock()

	intervalActionId := intervalAction.ID

	qc.loggingClient.Debug(fmt.Sprintf("updating the intervalAction with id: %s ", intervalActionId))

	oldIntervalId, exists := intervalActionIdToIntervalMap[intervalActionId]
	if !exists {
		return errors.New(fmt.Sprintf(
			"there is no mapping from interval action id : %s to interval.",
			intervalActionId))
	}

	intervalContext, exists := intervalNameToContextMap[intervalAction.Interval]
	if !exists {
		return errors.New(fmt.Sprintf(
			"query the interval with name : %s  and did not exist.",
			intervalAction.Interval))
	}

	// if the interval action switched interval
	interval := intervalContext.Interval

	newIntervalId := interval.ID

	if newIntervalId != oldIntervalId {
		qc.loggingClient.Debug(fmt.Sprintf(
			"the interval action switched interval from ID: %s to new ID: %s",
			oldIntervalId,
			newIntervalId))

		// remove the old interval action entry
		qc.loggingClient.Debug(fmt.Sprintf("remove the intervalAction with ID: %s from interval with ID: %s ",
			intervalActionId,
			oldIntervalId))
		delete(intervalContext.IntervalActionsMap, intervalActionId)

		// if there are no more events for the interval, remove the interval context
		// TODO: Not sure we want to just remove the interval from the interval context
		if len(intervalContext.IntervalActionsMap) == 0 {
			qc.loggingClient.Debug("there are no more events for the interval : " + oldIntervalId + ", remove it.")
			deleteIntervalOperation(interval, intervalContext)
		}

		// add Interval Event
		qc.loggingClient.Info(fmt.Sprintf(
			"add the intervalAction with id: %s to interval with id: %s",
			intervalActionId,
			newIntervalId))

		if _, exists := intervalIdToContextMap[newIntervalId]; !exists {
			context := IntervalContext{
				IntervalActionsMap: make(map[string]contract.IntervalAction),
				MarkedDeleted:      false,
			}
			context.Reset(interval, qc.loggingClient)

			addIntervalOperation(interval, &context)
		}
		addIntervalActionOperation(interval, intervalAction)
	} else { // if not, just update the interval action in place
		intervalContext.IntervalActionsMap[intervalActionId] = intervalAction
	}

	qc.loggingClient.Info(fmt.Sprintf(
		"updated the intervalAction with id: %s to interval id:  %s",
		intervalAction.ID,
		interval.ID))

	return nil
}

func (qc *QueueClient) RemoveIntervalActionQueue(intervalActionId string) error {
	mutex.Lock()
	defer mutex.Unlock()

	qc.loggingClient.Debug(fmt.Sprintf("removing the intervalAction with id: %s", intervalActionId))

	intervalId, exists := intervalActionIdToIntervalMap[intervalActionId]
	if !exists {
		return errors.New(fmt.Sprintf("could not find interval id with interval action id : %s", intervalActionId))
	}

	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		return errors.New(fmt.Sprintf("can not find interval context with interval id : %s", intervalId))
	}

	action, exists := intervalContext.IntervalActionsMap[intervalActionId]
	if exists {
		delete(intervalActionNameToIntervalMap, action.Name)
	}

	delete(intervalContext.IntervalActionsMap, intervalActionId)

	qc.loggingClient.Info(fmt.Sprintf("removed the intervalAction with id: %s", intervalActionId))

	return nil
}

func triggerInterval(loggingClient logger.LoggingClient) {
	nowEpoch := time.Now().Unix()

	defer func() {
		if err := recover(); err != nil {
			loggingClient.Error("trigger interval error : " + err.(string))
		}
	}()

	if intervalQueue.Length() == 0 {
		return
	}

	var wg sync.WaitGroup

	for i := 0; i < intervalQueue.Length(); i++ {
		if intervalQueue.Peek().(*IntervalContext) != nil {
			intervalContext := intervalQueue.Remove().(*IntervalContext)
			intervalId := intervalContext.Interval.ID
			if intervalContext.MarkedDeleted {
				loggingClient.Debug("the interval with id : " + intervalId + " be marked as deleted, removing it.")
				continue // really delete from the queue
			} else {
				if intervalContext.NextTime.Unix() <= nowEpoch {
					loggingClient.Debug(
						"executing interval, detail : {" + intervalContext.GetInfo() + "} ," +
							" at : " + intervalContext.NextTime.String())

					wg.Add(1)

					// execute it in a individual go routine
					go execute(intervalContext, &wg, loggingClient)
				} else {
					intervalQueue.Add(intervalContext)
				}
			}
		}
	}

	wg.Wait()
}

func execute(context *IntervalContext, wg *sync.WaitGroup, loggingClient logger.LoggingClient) {
	intervalActionMap := context.IntervalActionsMap

	defer wg.Done()

	defer func() {
		if err := recover(); err != nil {
			loggingClient.Error("interval execution error : " + err.(string))
		}
	}()

	loggingClient.Debug(fmt.Sprintf("%d interval action need to be executed.", len(intervalActionMap)))

	// execute interval action one by one
	for eventId := range intervalActionMap {
		loggingClient.Debug(
			"the event with id : " + eventId +
				" belongs to interval : " + context.Interval.ID + " will be executing!")
		intervalAction, _ := intervalActionMap[eventId]

		executingUrl := getUrlStr(intervalAction)
		loggingClient.Debug("the event with id : " + eventId + " will request url : " + executingUrl)

		httpMethod := intervalAction.HTTPMethod
		if !validMethod(httpMethod) {
			loggingClient.Error(fmt.Sprintf("net/http: invalid method %q", httpMethod))
			return
		}

		req, err := getHttpRequest(httpMethod, executingUrl, intervalAction, loggingClient)

		if err != nil {
			loggingClient.Error("create new request occurs error : " + err.Error())
		}

		client := &http.Client{
			Timeout: time.Duration(Configuration.Service.Timeout) * time.Millisecond,
		}
		responseBytes, statusCode, err := sendRequestAndGetResponse(client, req)
		responseStr := string(responseBytes)

		loggingClient.Debug(fmt.Sprintf("execution returns status code : %d", statusCode))
		loggingClient.Debug("execution returns response content : " + responseStr)
	}

	context.UpdateNextTime()
	context.UpdateIterations()

	if context.IsComplete() {
		loggingClient.Debug("completed interval, detail : " + context.GetInfo())
	} else {
		loggingClient.Debug("requeue interval, detail : " + context.GetInfo())
		intervalQueue.Add(context)
	}

	return
}

// TODO xmlviking We may need to modify this for authorization type in the future
func getHttpRequest(
	httpMethod string,
	executingUrl string,
	intervalAction contract.IntervalAction,
	loggingClient logger.LoggingClient) (*http.Request, error) {
	var body []byte

	params := strings.TrimSpace(intervalAction.Parameters)

	if len(params) > 0 {
		body = []byte(params)
	} else {
		body = nil
	}

	req, err := http.NewRequest(httpMethod, executingUrl, bytes.NewBuffer(body))
	if err != nil {
		loggingClient.Error("create new request occurs error : " + err.Error())
		return nil, err
	}

	req.Header.Set(ContentTypeKey, ContentTypeJsonValue)

	if len(params) > 0 {
		req.Header.Set(ContentLengthKey, string(len(params)))
	}

	return req, err
}

func getUrlStr(intervalAction contract.IntervalAction) string {
	return intervalAction.GetBaseURL() + intervalAction.Path
}

func sendRequestAndGetResponse(client *http.Client, req *http.Request) ([]byte, int, error) {
	resp, err := client.Do(req)

	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}

	defer resp.Body.Close()
	resp.Close = true

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
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
	var methods = map[string]struct{}{
		"GET": {}, "HEAD": {}, "POST": {}, "PUT": {}, "DELETE": {}, "TRACE": {}, "CONNECT": {},
	}

	_, contains := methods[strings.ToUpper(method)]
	return contains
}
