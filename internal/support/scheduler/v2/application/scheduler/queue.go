//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

func (m *manager) addInterval(interval models.Interval, sc *Executor) {
	m.intervalToExecutorMap[interval.Name] = sc
	m.executorQueue.Add(sc)
}

func (m *manager) deleteInterval(sc *Executor) {
	delete(m.intervalToExecutorMap, sc.Interval.Name)
	// Mark as Deleted and scheduler will remove it from the queue
	sc.MarkedDeleted = true
}

func (m *manager) addIntervalAction(sc *Executor, action models.IntervalAction) {
	sc.IntervalActionsMap[action.Name] = action
	m.actionToIntervalMap[action.Name] = sc.Interval.Name
}

func (m *manager) IntervalByName(intervalName string) (models.Interval, errors.EdgeX) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sc, exists := m.intervalToExecutorMap[intervalName]
	if !exists {
		return models.Interval{},
			errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("the executor with name %s does not exist", intervalName), nil)
	}
	return sc.Interval, nil
}

func (m *manager) AddInterval(interval models.Interval) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.intervalToExecutorMap[interval.Name]; exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the executor with name : %s already exists", interval.Name), nil)
	}

	executor := Executor{
		IntervalActionsMap: make(map[string]models.IntervalAction),
		MarkedDeleted:      false,
	}
	executor.Reset(interval, m.lc)
	m.addInterval(interval, &executor)

	m.lc.Info(fmt.Sprintf("added the interval with name : %s into the scheduler queue", interval.Name))
	return nil
}

func (m *manager) UpdateInterval(interval models.Interval) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	executor, exists := m.intervalToExecutorMap[interval.Name]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the executor with name %s does not exist", interval.Name), nil)
	}
	executor.Reset(interval, m.lc)

	m.lc.Info(fmt.Sprintf("updated the interval with name: %s in the scheduler queue", interval.Name))
	return nil
}

func (m *manager) DeleteIntervalByName(intervalName string) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lc.Debug(fmt.Sprintf("removing the interval with name: %s ", intervalName))

	sc, exists := m.intervalToExecutorMap[intervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the executor with name %s does not exist", intervalName), nil)
	}
	m.deleteInterval(sc)

	m.lc.Info(fmt.Sprintf("removed the interval with id: %s from the scheduler queue", intervalName))
	return nil
}

func (m *manager) IntervalActionByName(actionName string) (action models.IntervalAction, err errors.EdgeX) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	intervalName, exists := m.actionToIntervalMap[actionName]
	if !exists {
		return action,
			errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("scheduler could not find interval name with interval action name : %s", actionName), nil)
	}

	sc, exists := m.intervalToExecutorMap[intervalName]
	if !exists {
		return action,
			errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("the executor with name %s does not exist", intervalName), nil)
	}

	action, exists = sc.IntervalActionsMap[actionName]
	if !exists {
		return action,
			errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("scheduler could not find interval action with interval action name : %s", actionName), nil)
	}
	return action, nil
}

func (m *manager) AddIntervalAction(action models.IntervalAction) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.actionToIntervalMap[action.Name]; exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the action with name : %s already exists", action.Name), nil)
	}
	// Ensure we have an existing Interval
	sc, exists := m.intervalToExecutorMap[action.IntervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the executor with name %s does not exist", action.IntervalName), nil)
	}

	m.addIntervalAction(sc, action)

	m.lc.Info(fmt.Sprintf("added the intervalAction with name: %s to interal: %s into the queue",
		action.Name,
		action.IntervalName))
	return nil
}

func (m *manager) UpdateIntervalAction(action models.IntervalAction) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lc.Debug(fmt.Sprintf("updating the intervalAction with name: %s ", action.Name))

	currentExecutor, exists := m.intervalToExecutorMap[action.IntervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the executor with name %s does not exist", action.IntervalName), nil)
	}

	// if the interval action switched interval
	previousIntervalName, exists := m.actionToIntervalMap[action.Name]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("there is no mapping from interval action name : %s to interval", action.Name), nil)
	}
	interval := currentExecutor.Interval
	if currentExecutor.Interval.Name != previousIntervalName {
		m.lc.Debugf("the interval action switched interval from name: %s to new name: %s",
			previousIntervalName, currentExecutor.Interval.Name)

		// remove the action from the previous interval
		previousExecutor := m.intervalToExecutorMap[previousIntervalName]
		delete(previousExecutor.IntervalActionsMap, action.Name)

		// add Interval action
		m.addIntervalAction(currentExecutor, action)
	} else {
		// if not, just update the interval action in place
		currentExecutor.IntervalActionsMap[action.Name] = action
	}

	m.lc.Infof("updated the intervalAction with name: %s to interval name:  %s", action.Name, interval.Name)
	return nil
}

func (m *manager) DeleteIntervalActionByName(actionName string) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lc.Debug("removing the action with name: %s", actionName)

	intervalName, exists := m.actionToIntervalMap[actionName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("could not find interval name with action name : %s", actionName), nil)
	}

	sc, exists := m.intervalToExecutorMap[intervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("the executor with name %s does not exist", intervalName), nil)
	}

	delete(sc.IntervalActionsMap, actionName)
	delete(m.actionToIntervalMap, actionName)

	m.lc.Info(fmt.Sprintf("removed the action with name: %s", actionName))
	return nil
}
