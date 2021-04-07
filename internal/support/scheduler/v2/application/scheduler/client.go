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
)

type Client struct {
	ticker *time.Ticker
	lc     logger.LoggingClient
	config *config.ConfigurationStruct
}

func NewClient(lc logger.LoggingClient, config *config.ConfigurationStruct) interfaces.SchedulerClient {
	return &Client{
		ticker: time.NewTicker(time.Duration(config.Writable.ScheduleIntervalTime) * time.Millisecond),
		lc:     lc,
		config: config,
	}
}

func (c *Client) StartTicker() {
	go func() {
		for range c.ticker.C {
			c.triggerInterval()
		}
	}()
}

func (c *Client) StopTicker() {
	c.ticker.Stop()
}

func (c *Client) triggerInterval() {
	nowEpoch := time.Now().Unix()

	defer func() {
		if err := recover(); err != nil {
			c.lc.Error("trigger interval error : " + err.(string))
		}
	}()

	if scheduleQueue.Length() == 0 {
		return
	}

	var wg sync.WaitGroup

	for i := 0; i < scheduleQueue.Length(); i++ {
		if scheduleQueue.Peek().(*ScheduleContext) != nil {
			intervalContext := scheduleQueue.Remove().(*ScheduleContext)
			if intervalContext.MarkedDeleted {
				c.lc.Debug("the interval with name : " + intervalContext.Interval.Name + " be marked as deleted, removing it.")
				continue // really delete from the queue
			} else {
				if intervalContext.NextTime.Unix() <= nowEpoch {
					c.lc.Debugf(
						"executing interval %s at : %s", intervalContext.Interval.Name, intervalContext.NextTime.String())

					wg.Add(1)

					// execute it in a individual go routine
					go c.execute(intervalContext, &wg)
				} else {
					scheduleQueue.Add(intervalContext)
				}
			}
		}
	}

	wg.Wait()
}

func (c *Client) execute(
	context *ScheduleContext,
	wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			c.lc.Error("interval execution error : " + err.(string))
		}
	}()

	c.lc.Debug(fmt.Sprintf("%d action need to be executed with interval %s.", len(context.IntervalActionsMap), context.Interval.Name))

	// execute interval action one by one
	for _, action := range context.IntervalActionsMap {
		c.lc.Debugf(
			"the action with name: %s belongs to interval: %s will be executing!", action.Name, context.Interval.Name)

		err := utils.SendRequestWithAddress(c.lc, action.Address)
		if err != nil {
			c.lc.Errorf("fail to execute the interval action, err: %v", err)
		}

		c.lc.Debugf("success to execute the action %s with interval %s", action.Name, context.Interval.Name)
	}

	context.UpdateNextTime()
	context.UpdateIterations()

	if context.IsComplete() {
		c.lc.Debugf("completed interval %s", context.Interval.Name)
	} else {
		c.lc.Debugf("requeue interval %s", context.Interval.Name)
		scheduleQueue.Add(context)
	}

	return
}
