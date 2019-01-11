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
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	queueV1 "gopkg.in/eapache/queue.v1"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)


//the interval specific shared variables
var (
	mutex                                   sync.Mutex
	intervalQueue                           = queueV1.New()                     // global interval queue
	intervalIdToContextMap                  = make(map[string]*IntervalContext) // map : interval id -> interval context
	intervalNameToContextMap                = make(map[string]*IntervalContext) // map : interval name -> interval context
	intervalNameToIdMap	                    = make(map[string]string)           // map : interval name -> interval id
	intervalActionIdToIntervalMap           = make(map[string]string)           // map : interval action id -> interval id
	intervalActionNameToIntervalMap         = make(map[string]string)           // map : interval action name -> interval id
	intervalActionNameToIntervalActionIdMap = make(map[string]string)           // map : interval action name -> interval action id
)

func StartTicker() {
	go func() {
		for range ticker.C {
			triggerInterval()
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
	intervalNameToIdMap           = make(map[string]string)           // map : interval name -> interval id
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
}

func addIntervalActionOperation(interval contract.Interval, intervalAction contract.IntervalAction) {
	intervalContext, _ := intervalIdToContextMap[interval.ID]
	intervalContext.IntervalActionsMap[intervalAction.ID] = intervalAction
	intervalActionIdToIntervalMap[intervalAction.ID] = interval.ID
	intervalActionNameToIntervalMap[intervalAction.Name] = interval.ID
	intervalActionNameToIntervalActionIdMap[intervalAction.Name] = intervalAction.ID
}

func (qc *QueueClient) Connect()(string, error){
	return "alive..",nil
}
func (qc *QueueClient) QueryIntervalByID(intervalId string) (contract.Interval, error) {
	mutex.Lock()
	defer mutex.Unlock()

	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find a interval context with interval id : %s", intervalId)
		LoggingClient.Info(logMsg)
		return contract.Interval{}, errors.New(logMsg)
	}

	LoggingClient.Debug(fmt.Sprintf("querying found the interval with id : %s", intervalId))

	return intervalContext.Interval, nil
}

func (qc *QueueClient) QueryIntervalByName(intervalName string) (contract.Interval, error) {
	mutex.Lock()
	defer mutex.Unlock()

	intervalContext, exists := intervalNameToContextMap[intervalName]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find interval with interval with name : %s", intervalName)
		LoggingClient.Info(logMsg)
		return contract.Interval{}, errors.New(logMsg)
	}

	LoggingClient.Debug(fmt.Sprintf("scheduler found the interval with name : %s", intervalName))

	return intervalContext.Interval, nil
}

func (qc *QueueClient) AddIntervalToQueue(interval contract.Interval) error {
	mutex.Lock()
	defer mutex.Unlock()

	intervalId := interval.ID
	LoggingClient.Debug(fmt.Sprintf("adding the interval with id : %s at time %s", intervalId, interval.Start))

	if _, exists := intervalIdToContextMap[intervalId]; exists {
		LoggingClient.Debug(fmt.Sprintf("the interval context with id : %s already exists", intervalId))
		return nil
	}

	context := IntervalContext{
		IntervalActionsMap: make(map[string]contract.IntervalAction),
		MarkedDeleted:      false,
	}

	LoggingClient.Debug(fmt.Sprintf("resetting the interval with id : %s", intervalId))
	context.Reset(interval)

	addIntervalOperation(interval, &context)

	LoggingClient.Info(fmt.Sprintf("added the interval with id : %s into the scheduler queue", intervalId))

	return nil
}

func (qc *QueueClient) UpdateIntervalInQueue(interval contract.Interval) error {
	mutex.Lock()
	defer mutex.Unlock()

	intervalId := interval.ID
	context, exists := intervalIdToContextMap[intervalId]
	if !exists {
		LoggingClient.Error("the interval context with id " + intervalId + " does not exist ")
		return errors.New("the interval context with id " + intervalId + " does not exist ")
	}

	// remove the old map entry and create new one
	_, exists = intervalNameToIdMap[context.Interval.Name]
	if exists{
		delete(intervalNameToIdMap,context.Interval.Name)
	}

	// add new map entry
	intervalNameToIdMap[interval.Name] = interval.ID

	LoggingClient.Debug(fmt.Sprintf("resting the interval context with id: %s in the scheduler queue",intervalId))
	context.Reset(interval)

	LoggingClient.Info(fmt.Sprintf("updated the interval with id: %s in the scheduler queue",intervalId))

	return nil
}

func (qc *QueueClient) RemoveIntervalInQueue(intervalId string) error {
	mutex.Lock()
	defer mutex.Unlock()

	LoggingClient.Debug(fmt.Sprintf("removing the interval with id: %s ", intervalId))

	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find interval context with interval id : %s", intervalId)
		return errors.New(logMsg)
	}

	LoggingClient.Debug(fmt.Sprintf("removing all the mappings of interval action id to interval id: %s ", intervalId))
	for eventId := range intervalContext.IntervalActionsMap {
		delete(intervalActionIdToIntervalMap, eventId)
	}

	deleteIntervalOperation(intervalContext.Interval, intervalContext)

	LoggingClient.Info(fmt.Sprintf("removed the interval with id: %s from the scheduler queue",intervalId))

	return nil
}

func (qc *QueueClient) QueryIntervalActionByID(intervalActionId string) (contract.IntervalAction, error) {
	mutex.Lock()
	defer mutex.Unlock()

	intervalId, exists := intervalActionIdToIntervalMap[intervalActionId]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find interval id with interval action id : %s", intervalActionId)
		return contract.IntervalAction{}, errors.New(logMsg)
	}

	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		LoggingClient.Warn("scheduler could not find a interval context with interval id : " + intervalId)
		return contract.IntervalAction{}, nil
	}

	intervalAction, exists := intervalContext.IntervalActionsMap[intervalActionId]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find interval action with interval action id : %s", intervalActionId)
		return contract.IntervalAction{}, errors.New(logMsg)
	}

	return intervalAction, nil
}

func (qc *QueueClient) QueryIntervalActionByName(intervalActionName string) (contract.IntervalAction, error) {
	mutex.Lock()
	defer mutex.Unlock()

	intervalId, exists := intervalActionNameToIntervalMap[intervalActionName]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find interval id with intervalAction name : %s", intervalActionName)
		LoggingClient.Warn(logMsg)
		return contract.IntervalAction{}, errors.New(logMsg)
	}

	intervalActionId, exists := intervalActionNameToIntervalActionIdMap[intervalActionName]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find intervalAction id with intervalAction name : %s", intervalActionName)
		LoggingClient.Warn(logMsg)
		return contract.IntervalAction{}, errors.New(logMsg)
	}

	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find a interval context with interval id : %s", intervalId)
		LoggingClient.Warn(logMsg)
		return contract.IntervalAction{}, errors.New(logMsg)
	}

	intervalAction, exists := intervalContext.IntervalActionsMap[intervalActionId]
	if !exists {
		logMsg := fmt.Sprintf("scheduler could not find intervalAction with intervalAction id :  %s", intervalContext.Interval.ID)
		LoggingClient.Warn(logMsg)
		return contract.IntervalAction{}, errors.New(logMsg)
	}

	return intervalAction, nil
}

func (qc *QueueClient) AddIntervalActionToQueue(intervalAction contract.IntervalAction) error {
	mutex.Lock()
	defer mutex.Unlock()

	intervalActionId := intervalAction.ID
	intervalName := intervalAction.Interval
	intervalActionName := intervalAction.Name

	LoggingClient.Debug(fmt.Sprintf("adding the intervalAction with id  : %s to interval : %s into the queue", intervalActionId, intervalName))

	if _, exists := intervalActionNameToIntervalMap[intervalActionName]; exists {
		logMsg := fmt.Sprintf("scheduler found existing intervalAction with same name: %s", intervalName)
		LoggingClient.Warn(logMsg)
		return errors.New(logMsg)
	}

	// Ensure we have an existing Interval
	intervalId, exists := intervalNameToIdMap[intervalName]
	if !exists{
		logMsg := fmt.Sprintf("scheduler could not find a interval with interval name : %s", intervalName)
		LoggingClient.Warn(logMsg)
		return errors.New(logMsg)
	}

	// Get the Schedule Context
	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists{
		logMsg := fmt.Sprintf("scheduler could not find a interval with interval name : %s", intervalName)
		LoggingClient.Warn(logMsg)
		return errors.New(logMsg)
	}

	interval := intervalContext.Interval

	addIntervalActionOperation(interval, intervalAction)

	LoggingClient.Info(fmt.Sprintf("added the intervalAction with id: %s to interal: %s into the queue", intervalActionId, intervalName))

	return nil
}

func (qc *QueueClient) UpdateIntervalActionQueue(intervalAction contract.IntervalAction) error {
	mutex.Lock()
	defer mutex.Unlock()

	intervalActionId := intervalAction.ID

	LoggingClient.Debug(fmt.Sprintf("updating the intervalAction with id: %s ", intervalActionId))

	oldIntervalId, exists := intervalActionIdToIntervalMap[intervalActionId]
	if !exists {
		logMsg := fmt.Sprintf("there is no mapping from interval action id : %s to interval.", intervalActionId)
		LoggingClient.Error(logMsg)
		return errors.New(logMsg)
	}

// Refactor START

	intervalContext, exists := intervalNameToContextMap[intervalAction.Interval]
	if !exists {
		logMsg := fmt.Sprintf("query the interval with name : %s  and did not exist.", intervalAction.Interval)
		return errors.New(logMsg)
	}

	//if the interval action switched interval
	interval := intervalContext.Interval

	newIntervalId := interval.ID

	if newIntervalId != oldIntervalId {
		LoggingClient.Debug(fmt.Sprintf("the interval action switched interval from ID: %s to new ID: %s" ,oldIntervalId, newIntervalId))

		//remove the old interval action entry
		LoggingClient.Debug(fmt.Sprintf("remove the intervalAction with ID: %s from interval with ID: %s ",intervalActionId ,oldIntervalId))
		delete(intervalContext.IntervalActionsMap, intervalActionId)

		//if there are no more events for the interval, remove the interval context
		// TODO: Not sure we want to just remove the interval from the interval context
		if len(intervalContext.IntervalActionsMap) == 0 {
			LoggingClient.Debug("there are no more events for the interval : " + oldIntervalId + ", remove it.")
			deleteIntervalOperation(interval, intervalContext)
		}

		//add Interval Event
		LoggingClient.Info(fmt.Sprintf("add the intervalAction with id: %s to interval with id: %s",intervalActionId, newIntervalId))

		if _, exists := intervalIdToContextMap[newIntervalId]; !exists {
			context := IntervalContext{
				IntervalActionsMap: make(map[string]contract.IntervalAction),
				MarkedDeleted:      false,
			}
			context.Reset(interval)

			addIntervalOperation(interval, &context)
		}
		addIntervalActionOperation(interval, intervalAction)
	} else { // if not, just update the interval action in place
		intervalContext.IntervalActionsMap[intervalActionId] = intervalAction
	}

	LoggingClient.Info(fmt.Sprintf("updated the intervalAction with id: %s to interval id:  %s", intervalAction.ID,interval.ID))

	return nil
}

func(qc *QueueClient)  RemoveIntervalActionQueue(intervalActionId string) error {
	mutex.Lock()
	defer mutex.Unlock()

	LoggingClient.Debug(fmt.Sprintf("removing the intervalAction with id: %s", intervalActionId))

	intervalId, exists := intervalActionIdToIntervalMap[intervalActionId]
	if !exists {
		logMsg := fmt.Sprintf("could not find interval id with interval action id : %s", intervalActionId)
		return errors.New(logMsg)
	}

	intervalContext, exists := intervalIdToContextMap[intervalId]
	if !exists {
		logMsg := fmt.Sprintf("can not find interval context with interval id : %s", intervalId)
		return errors.New(logMsg)
	}

	delete(intervalContext.IntervalActionsMap, intervalActionId)

	LoggingClient.Info(fmt.Sprintf("removed the intervalAction with id: %s", intervalActionId))

	return nil
}

func triggerInterval() {
	nowEpoch := time.Now().Unix()

	defer func() {
		if err := recover(); err != nil {
			LoggingClient.Error("trigger interval error : " + err.(string))
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
				LoggingClient.Debug("the interval with id : " + intervalId + " be marked as deleted, removing it.")
				continue //really delete from the queue
			} else {
				if intervalContext.NextTime.Unix() <= nowEpoch {
					LoggingClient.Debug("executing interval, detail : {" + intervalContext.GetInfo() + "} , at : " + intervalContext.NextTime.String())

					wg.Add(1)

					//execute it in a individual go routine
					go execute(intervalContext, &wg)
				} else {
					intervalQueue.Add(intervalContext)
				}
			}
		}
	}

	wg.Wait()
}

func execute(context *IntervalContext, wg *sync.WaitGroup) error {
	intervalActionMap := context.IntervalActionsMap

	defer wg.Done()

	defer func() {
		if err := recover(); err != nil {
			LoggingClient.Error("interval execution error : " + err.(string))
		}
	}()

	LoggingClient.Debug(fmt.Sprintf("%d interval action need to be executed.", len(intervalActionMap)))

	//execute interval action one by one
	for eventId := range intervalActionMap {
		LoggingClient.Debug("the event with id : " + eventId + " belongs to interval : " + context.Interval.ID + " will be executing!")
		intervalAction, _ := intervalActionMap[eventId]

		executingUrl := getUrlStr(intervalAction)
		LoggingClient.Debug("the event with id : " + eventId + " will request url : " + executingUrl)

		//TODO: change the method type based on the event

		httpMethod := intervalAction.HTTPMethod
		if !validMethod(httpMethod) {
			LoggingClient.Error(fmt.Sprintf("net/http: invalid method %q", httpMethod))
			return nil
		}

		req, err := http.NewRequest(httpMethod, executingUrl, nil)
		req.Header.Set(ContentTypeKey, ContentTypeJsonValue)

		params := strings.TrimSpace(intervalAction.Parameters)

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
		LoggingClient.Debug("completed interval, detail : " + context.GetInfo())
	} else {
		LoggingClient.Debug("requeue interval, detail : " + context.GetInfo())
		intervalQueue.Add(context)
	}
	return nil
}

func getUrlStr(intervalAction contract.IntervalAction) string {
	return intervalAction.GetBaseURL() + intervalAction.Path
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
