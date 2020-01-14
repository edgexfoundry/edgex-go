//
// Copyright (c) 2018
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/config"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/container"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers/httpserver"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers/message"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers/testing"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"

	"github.com/gorilla/mux"
)

// Main is the service's execution entry point.  It takes a ctx, corresponding cancel function, a mux router, and
// a boolean readyStream; these facilitate acceptance testing.  The production service has its own main function
// that supplies default values for these; acceptance testing calls this function directly with its own values for
// the parameters specific to testing.  readyStream is nil for production environments; non-nil when run in the
// test runner context for acceptance testing.
func Main(ctx context.Context, cancel context.CancelFunc, router *mux.Router, readyStream chan<- bool) {
	startupTimer := startup.NewStartUpTimer(internal.BootRetrySecondsDefault, internal.BootTimeoutSecondsDefault)

	// All common command-line flags have been moved to DefaultCommonFlags. Service specific flags can be add here,
	// by inserting service specific flag prior to call to commonFlags.Parse().
	// Example:
	// 		flags.FlagSet.StringVar(&myvar, "m", "", "Specify a ....")
	//      ....
	//      flags.Parse(os.Args[1:])
	//
	f := flags.New()

	var debugMode bool
	f.FlagSet.BoolVar(&debugMode, "debug", false, "Turns on request/response debug logging.")

	f.Parse(os.Args[1:])

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	// readyStream is nil in production mode; non-nil when running acceptance tests in test runner context.  When
	// it's non-nil (i.e. when running acceptance tests), the httpServer bootstrap shouldn't bind and listen on a
	// specific port.
	httpServer := httpserver.NewBootstrap(router, readyStream == nil)

	bootstrap.Run(
		ctx,
		cancel,
		f,
		clients.SupportLoggingServiceKey,
		internal.ConfigStemCore+internal.ConfigMajorVersion,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			secret.NewSecret().BootstrapHandler,
			NewBootstrap(
				router,
				httpServer,
				clients.SupportLoggingServiceKey,
				debugMode,
				// readyStream is nil in production mode; non-nil when running acceptance tests in test runner context.
				// When it's non-nil (i.e. when running acceptance tests), the service's bootstrap handler shouldn't
				// wire up the APIv1 endpoints.
				readyStream != nil).BootstrapHandler,
			telemetry.BootstrapHandler,
			httpServer.BootstrapHandler,
			message.NewBootstrap(clients.SupportLoggingServiceKey, edgex.Version).BootstrapHandler,
			testing.NewBootstrap(httpServer, readyStream).BootstrapHandler,
		})
}
