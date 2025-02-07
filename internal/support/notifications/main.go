/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright 2018 Dell Technologies Inc.
 * Copyright (C) 2020-2025 IOTech Ltd
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

	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"

	"github.com/edgexfoundry/edgex-go"
	pkgHandlers "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers"
	notificationsConfig "github.com/edgexfoundry/edgex-go/internal/support/notifications/config"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/embed"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/labstack/echo/v4"
)

func Main(ctx context.Context, cancel context.CancelFunc, router *echo.Echo, args []string) {
	startupTimer := startup.NewStartUpTimer(common.SupportNotificationsServiceKey)

	// All common command-line flags have been moved to DefaultCommonFlags. Service specific flags can be added here,
	// by inserting service specific flag prior to call to commonFlags.Parse().
	// Example:
	// 		flags.FlagSet.StringVar(&myvar, "m", "", "Specify a ....")
	//      ....
	//      flags.Parse(args)
	//
	f := flags.New()
	f.Parse(args)

	configuration := &notificationsConfig.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	httpServer := handlers.NewHttpServer(router, true, common.SupportNotificationsServiceKey)
	dbHandler := pkgHandlers.NewDatabase(httpServer, configuration, container.DBClientInterfaceName, embed.SchemaName,
		common.SupportNotificationsServiceKey, edgex.Version, embed.SQLFiles)

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
			handlers.NewClientsBootstrap().BootstrapHandler,
			dbHandler.BootstrapHandler, // add db client bootstrap handler
			handlers.MessagingBootstrapHandler,
			handlers.NewServiceMetrics(common.SupportNotificationsServiceKey).BootstrapHandler, // Must be after Messaging
			NewBootstrap(router, common.SupportNotificationsServiceKey).BootstrapHandler,
			httpServer.BootstrapHandler,
			handlers.NewStartMessage(common.SupportNotificationsServiceKey, edgex.Version).BootstrapHandler,
		})
}
