//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"gopkg.in/eapache/queue.v1"
)

type manager struct {
	ticker                *time.Ticker
	lc                    logger.LoggingClient
	config                *config.ConfigurationStruct
	once                  sync.Once
	mutex                 sync.Mutex
	executorQueue         *queue.Queue
	intervalToExecutorMap map[string]*Executor
	actionToIntervalMap   map[string]string
}

// NewManager creates a new scheduler manager for running the interval job
func NewManager(lc logger.LoggingClient, config *config.ConfigurationStruct) interfaces.SchedulerManager {
	return &manager{
		ticker:                time.NewTicker(time.Duration(config.ScheduleIntervalTime) * time.Millisecond),
		lc:                    lc,
		config:                config,
		executorQueue:         queue.New(),
		intervalToExecutorMap: make(map[string]*Executor),
		actionToIntervalMap:   make(map[string]string),
	}
}

// StartTicker starts infinite loop with ticker to trigger the interval job
func (m *manager) StartTicker() {
	m.once.Do(func() {
		go func() {
			for range m.ticker.C {
				m.triggerInterval()
			}
		}()
	})
}

// StopTicker stops to trigger the interval job by stopping the ticker
func (m *manager) StopTicker() {
	m.ticker.Stop()
}

func (m *manager) triggerInterval() {
	if m.executorQueue.Length() == 0 {
		return
	}

	var wg sync.WaitGroup
	nowEpoch := time.Now().Unix()
	for i := 0; i < m.executorQueue.Length(); i++ {
		if m.executorQueue.Peek() != nil {
			executor, ok := m.executorQueue.Remove().(*Executor)
			if !ok {
				m.lc.Error("fail to cast the queue element to Executor")
				continue
			}
			if executor.MarkedDeleted {
				m.lc.Debugf("the interval %s be marked as deleted, removing it.", executor.Interval.Name)
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

	m.lc.Debugf("%d action need to be executed with interval %s.", len(executor.IntervalActionsMap), executor.Interval.Name)

	// execute interval action one by one
	for _, action := range executor.IntervalActionsMap {
		if action.AdminState == models.Locked {
			m.lc.Debugf("interval action %s is locked, skip the job execution", action.Name)
			continue
		}
		edgeXerr := m.executeAction(action)
		if edgeXerr != nil {
			m.lc.Errorf("fail to execute the interval action, err: %v", edgeXerr)
		}
	}

	executor.UpdateNextTime()

	if executor.IsComplete() {
		m.lc.Debugf("completed interval %s", executor.Interval.Name)
	} else {
		m.lc.Debugf("requeue interval %s", executor.Interval.Name)
		m.executorQueue.Add(executor)
	}
}

func (m *manager) executeAction(action models.IntervalAction) errors.EdgeX {
	m.lc.Debugf("the action with name: %s belongs to interval: %s will be executing!", action.Name, action.IntervalName)

	switch action.Address.GetBaseAddress().Type {
	case common.REST:
		restAddress, ok := action.Address.(models.RESTAddress)
		if !ok {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to RESTAddress", nil)
		}
		_, err := utils.SendRequestWithRESTAddress(m.lc, action.Content, action.ContentType, restAddress)
		if err != nil {
			m.lc.Errorf("fail to send request with RESTAddress, err: %v", err)
		}
	default:
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Unsupported address type", nil)
	}

	m.lc.Debugf("success to execute the action %s with interval %s", action.Name, action.IntervalName)
	return nil
}
