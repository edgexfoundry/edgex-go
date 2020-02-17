/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package data

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers/database/v1"
	v2 "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers/database/v2"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"

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
		dataContainer.ConfigurationName: func(get di.Get) interface{} {
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
		clients.CoreDataServiceKey,
		internal.ConfigStemCore+internal.ConfigMajorVersion,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			secret.NewSecret().BootstrapHandler,
			// readyStream is nil in production mode; non-nil when running acceptance tests in test runner context.
			// When it's non-nil (i.e. when running acceptance tests), the database bootstrap shouldn't create an
			// APIv1 connection.
			v1.NewDatabaseForCoreData(httpServer, configuration, readyStream != nil).BootstrapHandler,
			// APIv2 persistence.
			v2.NewPersistence(httpServer).BootstrapHandler,
			// readyStream is nil in production mode; non-nil when running acceptance tests in test runner context.
			// When it's non-nil (i.e. when running acceptance tests), the service's bootstrap handler shouldn't
			// wire up the APIv1 endpoints.
			NewBootstrap(router, debugMode, readyStream != nil).BootstrapHandler,
			telemetry.BootstrapHandler,
			httpServer.BootstrapHandler,
			message.NewBootstrap(clients.CoreDataServiceKey, edgex.Version).BootstrapHandler,
			testing.NewBootstrap(httpServer, readyStream).BootstrapHandler,
		},
	)
}
