/*******************************************************************************
 * Copyright 2020 Dell Inc.
 * Copyright 2022-2023 IOTech Ltd.
 * Copyright 2023 Intel Corporation
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
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/edgex-go/internal/core/command/controller/messaging"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	"github.com/gorilla/mux"
)

func Main(ctx context.Context, cancel context.CancelFunc, router *mux.Router) {
	startupTimer := startup.NewStartUpTimer(common.CoreCommandServiceKey)

	// All common command-line flags have been moved to DefaultCommonFlags. Service specific flags can be added here,
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
		common.ConfigStemCore,
		configuration,
		startupTimer,
		dic,
		true,
		bootstrapConfig.ServiceTypeOther,
		[]interfaces.BootstrapHandler{
			handlers.NewClientsBootstrap().BootstrapHandler,
			MessagingBootstrapHandler,
			handlers.NewServiceMetrics(common.CoreCommandServiceKey).BootstrapHandler, // Must be after Messaging
			NewBootstrap(router, common.CoreCommandServiceKey).BootstrapHandler,
			httpServer.BootstrapHandler,
			handlers.NewStartMessage(common.CoreCommandServiceKey, edgex.Version).BootstrapHandler,
		})

	// code here!
}

// MessagingBootstrapHandler sets up the MessageBus and External MQTT connections as well as subscriptions
func MessagingBootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	requestTimeout, err := time.ParseDuration(configuration.Service.RequestTimeout)
	if err != nil {
		lc.Errorf("Failed to parse Service.RequestTimeout configuration value: %v", err)
		return false
	}

	if configuration.ExternalMQTT.Enabled {
		if !handlers.NewExternalMQTT(messaging.OnConnectHandler(requestTimeout, dic)).BootstrapHandler(ctx, wg, startupTimer, dic) {
			return false
		}
	}

	if !handlers.MessagingBootstrapHandler(ctx, wg, startupTimer, dic) {
		return false
	}
	if err := messaging.SubscribeCommandRequests(ctx, requestTimeout, dic); err != nil {
		lc.Errorf("Failed to subscribe commands request from internal message bus, %v", err)
		return false
	}

	if err := messaging.SubscribeCommandQueryRequests(ctx, dic); err != nil {
		lc.Errorf("Failed to subscribe command query request from internal message bus, %v", err)
		return false
	}

	return true
}
