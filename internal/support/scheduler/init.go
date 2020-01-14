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
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/delegate"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/common/middleware/debugging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/correlationid"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/router"
	schedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/gorilla/mux"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	muxRouter              *mux.Router
	inDebugMode            bool
	inV2AcceptanceTestMode bool
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.  inDebug determines
// whether or not the debug middleware is installed.  inV2AcceptanceTestMode determines if the service is running in
// the test runner context (in which case, we shouldn't load the APIv1 routes).
func NewBootstrap(muxRouter *mux.Router, inDebugMode, inV2AcceptanceTestMode bool) *Bootstrap {
	return &Bootstrap{
		muxRouter:              muxRouter,
		inDebugMode:            inDebugMode,
		inV2AcceptanceTestMode: inV2AcceptanceTestMode,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the scheduler service.
func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := schedulerContainer.ConfigurationFrom(dic.Get)

	if !b.inV2AcceptanceTestMode {
		loadV1Routes(b.muxRouter, dic)
	}
	b.loadV2Routes(dic, lc)

	// add dependencies to bootstrapContainer
	scClient := NewSchedulerQueueClient(lc)
	dic.Update(di.ServiceConstructorMap{
		schedulerContainer.QueueName: func(get di.Get) interface{} {
			return scClient
		},
	})

	// disable v1 behavior that requires database if we're running in v2 acceptance test mode.
	if !b.inV2AcceptanceTestMode {
		err := LoadScheduler(lc, container.DBClientFrom(dic.Get), scClient, configuration)
		if err != nil {
			lc.Error(fmt.Sprintf("Failed to load schedules and events %s", err.Error()))
			return false
		}

		ticker := time.NewTicker(time.Duration(configuration.Writable.ScheduleIntervalTime) * time.Millisecond)
		StartTicker(ticker, lc, configuration)

		wg.Add(1)
		go func() {
			defer wg.Done()

			<-ctx.Done()
			StopTicker(ticker)
		}()
	}

	return true
}

// loadV2Routes creates a new command-query router and handles the related mux.Router initialization for API V2 routes.
func (b *Bootstrap) loadV2Routes(_ *di.Container, lc logger.LoggingClient) {
	correlationid.WireUp(b.muxRouter)

	handlers := []delegate.Handler{}
	if b.inDebugMode {
		handlers = append(handlers, debugging.New(lc).Handler)
	}

	router.Initialize(
		b.muxRouter,
		handlers,
		common.V2Routes(
			b.inV2AcceptanceTestMode,
			[]router.Controller{},
		),
	)
}
