//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

func (m *manager) addIntervalAction(e *Executor, action models.IntervalAction) {
	e.IntervalActionsMap[action.Name] = action
	m.actionToIntervalMap[action.Name] = e.Interval.Name
}

// AddInterval adds a new interval executor to the SchedulerManager's job queue
func (m *manager) AddInterval(interval models.Interval) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.intervalToExecutorMap[interval.Name]; exists {
		return errors.NewCommonEdgeX(errors.KindStatusConflict,
			fmt.Sprintf("the executor with interval name : %s already exists", interval.Name), nil)
	}

	executor := Executor{
		IntervalActionsMap: make(map[string]models.IntervalAction),
		MarkedDeleted:      false,
	}
	err := executor.Initialize(interval, m.lc)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	m.intervalToExecutorMap[interval.Name] = &executor
	m.executorQueue.Add(&executor)

	m.lc.Infof("added interval %s executor into the scheduler queue", interval.Name)
	return nil
}

// UpdateInterval updates interval executor to the SchedulerManager's job queue
func (m *manager) UpdateInterval(interval models.Interval) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	executor, exists := m.intervalToExecutorMap[interval.Name]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist,
			fmt.Sprintf("the executor with interval name %s does not exist", interval.Name), nil)
	}
	err := executor.Initialize(interval, m.lc)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	m.lc.Infof("updated the interval %s executor in the scheduler queue", interval.Name)
	return nil
}

// DeleteIntervalByName deletes interval executor by intervalName
func (m *manager) DeleteIntervalByName(intervalName string) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lc.Debugf("removing the interval with name: %s ", intervalName)

	executor, exists := m.intervalToExecutorMap[intervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist,
			fmt.Sprintf("the executor with interval name %s does not exist", intervalName), nil)
	}

	delete(m.intervalToExecutorMap, executor.Interval.Name)
	// Mark as Deleted and scheduler will remove it from the queue
	executor.MarkedDeleted = true

	m.lc.Infof("removed the interval %s executor from the scheduler queue", intervalName)
	return nil
}

// AddIntervalAction adds intervalAction to the specified executor
func (m *manager) AddIntervalAction(action models.IntervalAction) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.actionToIntervalMap[action.Name]; exists {
		return errors.NewCommonEdgeX(errors.KindStatusConflict,
			fmt.Sprintf("the action with name : %s already exists", action.Name), nil)
	}
	// Ensure we have an existing Interval
	executor, exists := m.intervalToExecutorMap[action.IntervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist,
			fmt.Sprintf("the executor with interval name %s does not exist", action.IntervalName), nil)
	}

	m.addIntervalAction(executor, action)

	m.lc.Infof("added the intervalAction %s to interval %s executor", action.Name, action.IntervalName)
	return nil
}

// UpdateIntervalAction updates intervalAction to the specified executor
func (m *manager) UpdateIntervalAction(action models.IntervalAction) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lc.Debugf("updating the intervalAction with name: %s ", action.Name)

	currentExecutor, exists := m.intervalToExecutorMap[action.IntervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist,
			fmt.Sprintf("the executor with interval name %s does not exist", action.IntervalName), nil)
	}

	// if the interval action switched interval
	previousIntervalName, exists := m.actionToIntervalMap[action.Name]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist,
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

	m.lc.Infof("updated the intervalAction %s to interval %s executor", action.Name, interval.Name)
	return nil
}

// DeleteIntervalActionByName deletes the intervalAction by name
func (m *manager) DeleteIntervalActionByName(actionName string) errors.EdgeX {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lc.Debugf("removing the action with name: %s", actionName)

	intervalName, exists := m.actionToIntervalMap[actionName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist,
			fmt.Sprintf("could not find interval name with action name : %s", actionName), nil)
	}

	executor, exists := m.intervalToExecutorMap[intervalName]
	if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist,
			fmt.Sprintf("the executor with interval name %s does not exist", intervalName), nil)
	}

	delete(executor.IntervalActionsMap, actionName)
	delete(m.actionToIntervalMap, actionName)

	m.lc.Infof("removed the action with name: %s", actionName)
	return nil
}
