/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright 2018 Dell Technologies Inc.
 * Copyright (C) 2020-2021 IOTech Ltd
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
 *
 * @microservice: support-notifications
 * @author: Jim White, Dell Technologies
 * @version: 0.5.0
 *******************************************************************************/

// main is the central entry point for the application and calls all the startup logic.
package notifications

import (
	"context"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"os"

	"github.com/edgexfoundry/edgex-go"
	pkgHandlers "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers"
	notificationsConfig "github.com/edgexfoundry/edgex-go/internal/support/notifications/config"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/gorilla/mux"
)

func Main(ctx context.Context, cancel context.CancelFunc, router *mux.Router) {
	startupTimer := startup.NewStartUpTimer(common.SupportNotificationsServiceKey)

	// All common command-line flags have been moved to DefaultCommonFlags. Service specific flags can be add here,
	// by inserting service specific flag prior to call to commonFlags.Parse().
	// Example:
	// 		flags.FlagSet.StringVar(&myvar, "m", "", "Specify a ....")
	//      ....
	//      flags.Parse(os.Args[1:])
	//
	f := flags.New()
	f.Parse(os.Args[1:])

	configuration := &notificationsConfig.ConfigurationStruct{}
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
		common.SupportNotificationsServiceKey,
		common.ConfigStemCore,
		configuration,
		startupTimer,
		dic,
		true,
		config.ServiceTypeOther,
		[]interfaces.BootstrapHandler{
			pkgHandlers.NewDatabase(httpServer, configuration, container.DBClientInterfaceName).BootstrapHandler, // add v2 db client bootstrap handler
			handlers.MessagingBootstrapHandler,
			handlers.NewServiceMetrics(common.SupportNotificationsServiceKey).BootstrapHandler, // Must be after Messaging
			handlers.NewClientsBootstrap().BootstrapHandler,
			NewBootstrap(router, common.SupportNotificationsServiceKey).BootstrapHandler,
			httpServer.BootstrapHandler,
			handlers.NewStartMessage(common.SupportNotificationsServiceKey, edgex.Version).BootstrapHandler,
		})
}
