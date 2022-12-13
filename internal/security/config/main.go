//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0'
//

package config

import (
	"context"
	"os"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	"github.com/edgexfoundry/edgex-go/internal/security/config/command"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/container"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
)

const securitySecretsConfigServiceKey = "secrets-config"

// Main function called from cmd/secrets-config
func Main(ctx context.Context, cancel context.CancelFunc) int {

	startupTimer := startup.NewStartUpTimer(securitySecretsConfigServiceKey)

	// Common Command-line flags have been moved to command.CommonFlags, but this service doesn't use all
	// the common flags so we are using our own implementation of the CommonFlags interface
	f := command.NewCommonFlags()
	f.Parse(os.Args[1:])

	lc := logger.NewClient(securitySecretsConfigServiceKey, models.ErrorLog)
	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
	})

	serviceHandler := NewBootstrap()

	bootstrap.RunAndReturnWaitGroup(
		ctx,
		cancel,
		f,
		securitySecretsConfigServiceKey,
		common.ConfigStemSecurity,
		configuration,
		nil,
		startupTimer,
		dic,
		false,
		[]interfaces.BootstrapHandler{
			serviceHandler.BootstrapHandler,
		},
	)

	return serviceHandler.ExitStatusCode()
}
