//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/container"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/server/handlers"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	bootstrapHandlers "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/labstack/echo/v4"
)

// Configure is the main entry point for configuring the Postgres database before startup
func Configure(ctx context.Context,
	cancel context.CancelFunc,
	flags flags.Common,
	router *echo.Echo) {
	startupTimer := startup.NewStartUpTimer(common.SecuritySecretStoreSetupServiceKey)

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	httpServer := bootstrapHandlers.NewHttpServer(router, true, common.SecuritySecretStoreSetupServiceKey)

	bootstrap.Run(
		ctx,
		cancel,
		flags,
		common.SecuritySecretStoreSetupServiceKey,
		common.ConfigStemSecurity,
		configuration,
		startupTimer,
		dic,
		false,
		bootstrapConfig.ServiceTypeOther,
		[]interfaces.BootstrapHandler{
			handlers.NewBootstrapServer(router).BootstrapServerHandler,
			httpServer.BootstrapHandler,
			bootstrapHandlers.NewStartMessage(common.SecuritySecretStoreSetupServiceKey, edgex.Version).BootstrapHandler,
		},
	)
}
