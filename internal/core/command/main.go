/*******************************************************************************
 * Copyright 2020 Dell Inc.
 * Copyright 2022 IOTech Ltd.
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

package command

import (
	"context"
	"os"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/edgex-go/internal/core/command/controller/messaging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	"github.com/gorilla/mux"
)

func Main(ctx context.Context, cancel context.CancelFunc, router *mux.Router) {
	startupTimer := startup.NewStartUpTimer(common.CoreCommandServiceKey)

	// All common command-line flags have been moved to DefaultCommonFlags. Service specific flags can be add here,
	// by inserting service specific flag prior to call to commonFlags.Parse().
	// Example:
	// 		flags.FlagSet.StringVar(&myvar, "m", "", "Specify a ....")
	//      ....
	//      flags.Parse(os.Args[1:])
	//
	f := flags.New()
	f.Parse(os.Args[1:])

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	httpServer := handlers.NewHttpServer(router, true)

	bootstrap.Run(
		ctx,
		cancel,
		f,
		common.CoreCommandServiceKey,
		internal.ConfigStemCore,
		configuration,
		startupTimer,
		dic,
		true,
		[]interfaces.BootstrapHandler{
			handlers.NewClientsBootstrap().BootstrapHandler,
			MessageBusBootstrapHandler,
			handlers.NewServiceMetrics(common.CoreCommandServiceKey).BootstrapHandler, // Must be after Messaging
			NewBootstrap(router, common.CoreCommandServiceKey).BootstrapHandler,
			telemetry.BootstrapHandler,
			httpServer.BootstrapHandler,
			handlers.NewStartMessage(common.CoreCommandServiceKey, edgex.Version).BootstrapHandler,
		})

	// code here!
}

// MessageBusBootstrapHandler sets up the MessageBus connection if MessageBus required is true.
// This is required for backwards compatability with older versions of 2.x configuration
// TODO: Remove in EdgeX 3.0
func MessageBusBootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	router := messaging.NewMessagingRouter()

	if configuration.RequireMessageBus {
		if configuration.MessageQueue.External.Enabled {
			if !handlers.NewExternalMQTT(messaging.OnConnectHandler(router, dic)).BootstrapHandler(ctx, wg, startupTimer, dic) {
				return false
			}
		}
		if !handlers.MessagingBootstrapHandler(ctx, wg, startupTimer, dic) {
			return false
		}
		if err := messaging.SubscribeCommandRequests(ctx, router, dic); err != nil {
			lc.Errorf("Failed to subscribe commands request from internal message bus, %v", err)
			return false
		}
		if err := messaging.SubscribeCommandResponses(ctx, router, dic); err != nil {
			lc.Errorf("Failed to subscribe commands response from internal message bus, %v", err)
			return false
		}
		if err := messaging.SubscribeCommandQueryRequests(ctx, dic); err != nil {
			lc.Errorf("Failed to subscribe command query request from internal message bus, %v", err)
			return false
		}
	}

	// Not required so do nothing
	return true
}
