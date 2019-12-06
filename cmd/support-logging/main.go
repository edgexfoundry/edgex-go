//
// Copyright (c) 2018
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"flag"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers/httpserver"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers/message"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers/secret"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/support/logging"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

func main() {
	startupTimer := startup.NewStartUpTimer(internal.BootRetrySecondsDefault, internal.BootTimeoutSecondsDefault)

	var useRegistry bool
	var configDir, profileDir string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use Registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use Registry.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallback
	flag.Parse()

	dic := di.NewContainer(di.ServiceConstructorMap{})
	httpServer := httpserver.NewBootstrap(logging.LoadRestRoutes(dic))
	bootstrap.Run(
		configDir,
		profileDir,
		internal.ConfigFileName,
		useRegistry,
		clients.SupportLoggingServiceKey,
		logging.Configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			secret.NewSecret().BootstrapHandler,
			logging.NewServiceInit(&httpServer, clients.SupportLoggingServiceKey).BootstrapHandler,
			telemetry.BootstrapHandler,
			httpServer.BootstrapHandler,
			message.NewBootstrap(clients.SupportLoggingServiceKey, edgex.Version).BootstrapHandler,
		})
}
