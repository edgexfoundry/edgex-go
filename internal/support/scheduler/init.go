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

	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
)

// Global variables
var Configuration = &ConfigurationStruct{}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the scheduler service.
func BootstrapHandler(wg *sync.WaitGroup, ctx context.Context, startupTimer startup.Timer, dic *di.Container) bool {
	loggingClient := bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient := bootstrapContainer.DBClientFrom(dic.Get)

	// add dependencies to bootstrapContainer
	scClient := NewSchedulerQueueClient(loggingClient)
	dic.Update(di.ServiceConstructorMap{
		container.QueueName: func(get di.Get) interface{} {
			return scClient
		},
	})

	// Initialize the ticker time
	if err := LoadScheduler(loggingClient, dbClient, scClient); err != nil {
		loggingClient.Error(fmt.Sprintf("Failed to load schedules and events %s", err.Error()))
		return false
	}

	ticker := time.NewTicker(time.Duration(Configuration.Writable.ScheduleIntervalTime) * time.Millisecond)
	StartTicker(ticker, loggingClient)

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		StopTicker(ticker)
	}()

	return true
}
