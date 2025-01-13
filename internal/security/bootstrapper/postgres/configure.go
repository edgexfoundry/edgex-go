//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/postgres/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/postgres/container"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/postgres/handlers"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
)

// Configure is the main entry point for configuring the Postgres database before startup
func Configure(ctx context.Context,
	cancel context.CancelFunc,
	flags flags.Common) {
	startupTimer := startup.NewStartUpTimer(common.SecurityBootstrapperPostgresKey)

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	// bootstrap.RunAndReturnWaitGroup is needed for the underlying configuration system.
	// Conveniently, it also creates a pipeline of functions as the list of BootstrapHandler's is
	// executed in order.
	_, _, ok := bootstrap.RunAndReturnWaitGroup(
		ctx,
		cancel,
		flags,
		common.SecurityBootstrapperPostgresKey,
		common.ConfigStemCore,
		configuration,
		nil,
		startupTimer,
		dic,
		true,
		bootstrapConfig.ServiceTypeOther,
		[]interfaces.BootstrapHandler{
			handlers.SetupPasswordFile,
			handlers.SetupDBScriptFiles,
		},
	)

	if !ok {
		// had some issue(s) during bootstrapping Postgres
		os.Exit(1)
	}
}
