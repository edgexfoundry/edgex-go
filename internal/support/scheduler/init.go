/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright (c) 2019 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// Global variables
var Configuration = &ConfigurationStruct{}
var LoggingClient logger.LoggingClient
var dbClient interfaces.DBClient
var scClient interfaces.SchedulerQueueClient
var ticker *time.Ticker

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the scheduler service.
func BootstrapHandler(wg *sync.WaitGroup, ctx context.Context, startupTimer startup.Timer, dic *di.Container) bool {
	// update global variables.
	LoggingClient = container.LoggingClientFrom(dic.Get)
	dbClient = container.DBClientFrom(dic.Get)

	scClient = NewSchedulerQueueClient()

	// Initialize the ticker time
	if err := LoadScheduler(); err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to load schedules and events %s", err.Error()))
		return false
	}

	ticker = time.NewTicker(time.Duration(Configuration.Writable.ScheduleIntervalTime) * time.Millisecond)
	StartTicker()

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		StopTicker()
	}()

	return true
}
