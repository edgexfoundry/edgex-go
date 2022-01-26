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
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/application"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/application/scheduler"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/gorilla/mux"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	router      *mux.Router
	serviceName string
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(router *mux.Router, serviceName string) *Bootstrap {
	return &Bootstrap{
		router:      router,
		serviceName: serviceName,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the scheduler service.
func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	LoadRestRoutes(b.router, dic, b.serviceName)

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	// V2 Scheduler
	schedulerManager := scheduler.NewManager(lc, configuration)
	dic.Update(di.ServiceConstructorMap{
		container.SchedulerManagerName: func(get di.Get) interface{} {
			return schedulerManager
		},
	})

	err := application.LoadIntervalToSchedulerManager(dic)
	if err != nil {
		lc.Errorf("Failed to load interval to scheduler, %v", err)
		return false
	}

	err = application.LoadIntervalActionToSchedulerManager(dic)
	if err != nil {
		lc.Errorf("Failed to load intervalAction to scheduler, %v", err)
		return false
	}

	schedulerManager.StartTicker()

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		schedulerManager.StopTicker()
	}()

	return true
}
