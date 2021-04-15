//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/infrastructure/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	queueV1 "gopkg.in/eapache/queue.v1"
)

type manager struct {
	ticker                *time.Ticker
	lc                    logger.LoggingClient
	config                *config.ConfigurationStruct
	once                  sync.Once
	mutex                 sync.Mutex
	executorQueue         *queueV1.Queue
	intervalToExecutorMap map[string]*Executor
	actionToIntervalMap   map[string]string
}

func NewManager(lc logger.LoggingClient, config *config.ConfigurationStruct) interfaces.SchedulerManager {
	return &manager{
		ticker:                time.NewTicker(time.Duration(config.Writable.ScheduleIntervalTime) * time.Millisecond),
		lc:                    lc,
		config:                config,
		executorQueue:         queueV1.New(),
		intervalToExecutorMap: make(map[string]*Executor),
		actionToIntervalMap:   make(map[string]string),
	}
}

func (m *manager) StartTicker() {
	m.once.Do(func() {
		go func() {
			for range m.ticker.C {
				m.triggerInterval()
			}
		}()
	})
}

func (m *manager) StopTicker() {
	m.ticker.Stop()
}

func (m *manager) triggerInterval() {
	nowEpoch := time.Now().Unix()

	defer func() {
		if err := recover(); err != nil {
			m.lc.Error("trigger interval error : " + err.(string))
		}
	}()

	if m.executorQueue.Length() == 0 {
		return
	}

	var wg sync.WaitGroup

	for i := 0; i < m.executorQueue.Length(); i++ {
		if m.executorQueue.Peek().(*Executor) != nil {
			executor := m.executorQueue.Remove().(*Executor)
			if executor.MarkedDeleted {
				m.lc.Debug("the interval with name : " + executor.Interval.Name + " be marked as deleted, removing it.")
				continue // really delete from the queue
			} else {
				if executor.NextTime.Unix() <= nowEpoch {
					m.lc.Debugf(
						"executing interval %s at : %s", executor.Interval.Name, executor.NextTime.String())

					wg.Add(1)

					// execute it in a individual go routine
					go m.execute(executor, &wg)
				} else {
					m.executorQueue.Add(executor)
				}
			}
		}
	}

	wg.Wait()
}

func (m *manager) execute(
	executor *Executor,
	wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			m.lc.Error("interval execution error : " + err.(string))
		}
	}()

	m.lc.Debug(fmt.Sprintf("%d action need to be executed with interval %s.", len(executor.IntervalActionsMap), executor.Interval.Name))

	// execute interval action one by one
	for _, action := range executor.IntervalActionsMap {
		m.lc.Debugf(
			"the action with name: %s belongs to interval: %s will be executing!", action.Name, executor.Interval.Name)

		err := utils.SendRequestWithAddress(m.lc, action.Address)
		if err != nil {
			m.lc.Errorf("fail to execute the interval action, err: %v", err)
		}

		m.lc.Debugf("success to execute the action %s with interval %s", action.Name, executor.Interval.Name)
	}

	executor.UpdateNextTime()
	executor.UpdateIterations()

	if executor.IsComplete() {
		m.lc.Debugf("completed interval %s", executor.Interval.Name)
	} else {
		m.lc.Debugf("requeue interval %s", executor.Interval.Name)
		m.executorQueue.Add(executor)
	}

	return
}
