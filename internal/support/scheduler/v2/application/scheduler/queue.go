//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"fmt"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	queueV1 "gopkg.in/eapache/queue.v1"
)

// the interval specific shared variables
var (
	mutex                sync.Mutex
	scheduleQueue        = queueV1.New()
	intervalToContextMap = make(map[string]*ScheduleContext)
	actionToIntervalMap  = make(map[string]string)
)

func addInterval(interval models.Interval, sc *ScheduleContext) {
	intervalToContextMap[interval.Name] = sc
	scheduleQueue.Add(sc)
}

func deleteInterval(sc *ScheduleContext) {
	delete(intervalToContextMap, sc.Interval.Name)
	// Mark as Deleted and scheduler will remove it from the queue
	sc.MarkedDeleted = true
}

func addIntervalAction(sc *ScheduleContext, action models.IntervalAction) {
	sc.IntervalActionsMap[action.Name] = action
	actionToIntervalMap[action.Name] = sc.Interval.Name
}

func (c *Client) IntervalByName(intervalName string) (models.Interval, errors.EdgeX) {
	mutex.Lock()
	defer mutex.Unlock()

	sc, exists := intervalToContextMap[intervalName]
	if !exists {
		return models.Interval{},
			errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("the schedule context with name %s does not exist", intervalName), nil)
	}
	return sc.Interval, nil
}

func (c *Client) AddInterval(interval models.Interval) errors.EdgeX {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := intervalToContextMap[interval.Name]; exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the schedule context with name : %s already exists", interval.Name), nil)
	}

	context := ScheduleContext{
		IntervalActionsMap: make(map[string]models.IntervalAction),
		MarkedDeleted:      false,
	}
	context.Reset(interval, c.lc)
	addInterval(interval, &context)

	c.lc.Info(fmt.Sprintf("added the interval with name : %s into the scheduler queue", interval.Name))
	return nil
}

func (c *Client) UpdateInterval(interval models.Interval) errors.EdgeX {
	mutex.Lock()
	defer mutex.Unlock()

	context, exists := intervalToContextMap[interval.Name]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the schedule context with name %s does not exist", interval.Name), nil)
	}
	context.Reset(interval, c.lc)

	c.lc.Info(fmt.Sprintf("updated the interval with name: %s in the scheduler queue", interval.Name))
	return nil
}

func (c *Client) DeleteIntervalByName(intervalName string) errors.EdgeX {
	mutex.Lock()
	defer mutex.Unlock()
	c.lc.Debug(fmt.Sprintf("removing the interval with name: %s ", intervalName))

	sc, exists := intervalToContextMap[intervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the schedule context with name %s does not exist", intervalName), nil)
	}
	deleteInterval(sc)

	c.lc.Info(fmt.Sprintf("removed the interval with id: %s from the scheduler queue", intervalName))
	return nil
}

func (c *Client) IntervalActionByName(actionName string) (action models.IntervalAction, err errors.EdgeX) {
	mutex.Lock()
	defer mutex.Unlock()

	intervalName, exists := actionToIntervalMap[actionName]
	if !exists {
		return action,
			errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("scheduler could not find interval name with interval action name : %s", actionName), nil)
	}

	sc, exists := intervalToContextMap[intervalName]
	if !exists {
		return action,
			errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("the schedule context with name %s does not exist", intervalName), nil)
	}

	action, exists = sc.IntervalActionsMap[actionName]
	if !exists {
		return action,
			errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("scheduler could not find interval action with interval action name : %s", actionName), nil)
	}
	return action, nil
}

func (c *Client) AddIntervalAction(action models.IntervalAction) errors.EdgeX {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := actionToIntervalMap[action.Name]; exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the action with name : %s already exists", action.Name), nil)
	}
	// Ensure we have an existing Interval
	sc, exists := intervalToContextMap[action.IntervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the schedule context with name %s does not exist", action.IntervalName), nil)
	}

	addIntervalAction(sc, action)

	c.lc.Info(fmt.Sprintf("added the intervalAction with name: %s to interal: %s into the queue",
		action.Name,
		action.IntervalName))
	return nil
}

func (c *Client) UpdateIntervalAction(action models.IntervalAction) errors.EdgeX {
	mutex.Lock()
	defer mutex.Unlock()
	c.lc.Debug(fmt.Sprintf("updating the intervalAction with name: %s ", action.Name))

	currentContext, exists := intervalToContextMap[action.IntervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the schedule context with name %s does not exist", action.IntervalName), nil)
	}

	// if the interval action switched interval
	previousIntervalName, exists := actionToIntervalMap[action.Name]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("there is no mapping from interval action name : %s to interval", action.Name), nil)
	}
	interval := currentContext.Interval
	if currentContext.Interval.Name != previousIntervalName {
		c.lc.Debugf("the interval action switched interval from name: %s to new name: %s",
			previousIntervalName, currentContext.Interval.Name)

		// remove the action from the previous interval
		previousContext := intervalToContextMap[previousIntervalName]
		delete(previousContext.IntervalActionsMap, action.Name)

		// add Interval action
		addIntervalAction(currentContext, action)
	} else {
		// if not, just update the interval action in place
		currentContext.IntervalActionsMap[action.Name] = action
	}

	c.lc.Infof("updated the intervalAction with name: %s to interval name:  %s", action.Name, interval.Name)
	return nil
}

func (c *Client) DeleteIntervalActionByName(actionName string) errors.EdgeX {
	mutex.Lock()
	defer mutex.Unlock()
	c.lc.Debug("removing the action with name: %s", actionName)

	intervalName, exists := actionToIntervalMap[actionName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("could not find interval name with action name : %s", actionName), nil)
	}

	sc, exists := intervalToContextMap[intervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the schedule context with name %s does not exist", intervalName), nil)
	}

	delete(sc.IntervalActionsMap, actionName)
	delete(actionToIntervalMap, actionName)

	c.lc.Info(fmt.Sprintf("removed the action with name: %s", actionName))
	return nil
}
