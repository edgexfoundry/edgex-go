/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

package metadata

import (
	"context"
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/uom"
	pkgHandlers "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"

	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

func Main(ctx context.Context, cancel context.CancelFunc, router *mux.Router) {
	startupTimer := startup.NewStartUpTimer(common.CoreMetaDataServiceKey)

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

	wg, deferred, success := bootstrap.RunAndReturnWaitGroup(
		ctx,
		cancel,
		f,
		common.CoreMetaDataServiceKey,
		internal.ConfigStemCore,
		configuration,
		nil,
		startupTimer,
		dic,
		true,
		[]interfaces.BootstrapHandler{
			uom.BootstrapHandler,
			pkgHandlers.NewDatabase(httpServer, configuration, container.DBClientInterfaceName).BootstrapHandler, // add v2 db client bootstrap handler
			MessageBusBootstrapHandler,
			handlers.NewServiceMetrics(common.CoreMetaDataServiceKey).BootstrapHandler, // Must be after Messaging
			NewBootstrap(router, common.CoreMetaDataServiceKey).BootstrapHandler,
			telemetry.BootstrapHandler,
			httpServer.BootstrapHandler,
		})

	if !success {
		return
	}

	// Have to call this handler outside the bootstrapping when configuration is loaded and known
	configuration = container.ConfigurationFrom(dic.Get)
	// Only create Notifications Client if going to be using it
	if configuration.Notifications.PostDeviceChanges {
		if !handlers.NewClientsBootstrap().BootstrapHandler(ctx, wg, startupTimer, dic) {
			return
		}
	}

	// Call this handler outside the bootstrapping, so it is always last.
	handlers.NewStartMessage(common.CoreMetaDataServiceKey, edgex.Version).BootstrapHandler(ctx, wg, startupTimer, dic)

	wg.Wait()

	if deferred != nil {
		deferred()
	}
}

// MessageBusBootstrapHandler sets up the MessageBus connection if MessageBus required is true.
// This is required for backwards compatability with older versions of 2.x configuration
// TODO: Remove in EdgeX 3.0
func MessageBusBootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	configuration := container.ConfigurationFrom(dic.Get)
	if configuration.RequireMessageBus {
		return handlers.MessagingBootstrapHandler(ctx, wg, startupTimer, dic)
	}

	// Not required so do nothing
	return true
}
