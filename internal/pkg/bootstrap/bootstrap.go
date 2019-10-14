/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/configuration"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/logging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

// fatalError logs an error and exits the application.  It's intended to be used only within the bootstrap prior to
// any go routines being spawned.
func fatalError(err error, loggingClient logger.LoggingClient) {
	loggingClient.Error(err.Error())
	os.Exit(1)
}

// translateInterruptToCancel spawns a go routine to translate the receipt of a SIGTERM signal to a call to cancel
// the context used by the bootstrap implementation.
func translateInterruptToCancel(wg *sync.WaitGroup, ctx context.Context, cancel context.CancelFunc) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		signalStream := make(chan os.Signal)
		defer func() {
			signal.Stop(signalStream)
			close(signalStream)
		}()
		signal.Notify(signalStream, os.Interrupt, syscall.SIGTERM)
		select {
		case <-signalStream:
			cancel()
			return
		case <-ctx.Done():
			return
		}
	}()
}

// Run bootstraps an application.  It loads configuration and calls the provided list of handlers.  Any long-running
// process should be spawned as a go routine in a handler.  Handlers are expected to return immediately.  Once all of
// the handlers are called this function will wait for any go routines spawned inside the handlers to exit before
// returning to the caller.  It is intended that the caller stop executing on the return of this function.
func Run(
	configDir, profileDir, configFileName string,
	useRegistry bool,
	serviceKey string,
	config interfaces.Configuration,
	startupTimer startup.Timer,
	dic *di.Container,
	handlers []interfaces.BootstrapHandler) {

	loggingClient := logging.FactoryToStdout(serviceKey)
	var err error
	var registryClient registry.Client
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	translateInterruptToCancel(&wg, ctx, cancel)

	// load configuration from file.
	if err = configuration.LoadFromFile(configDir, profileDir, configFileName, config); err != nil {
		fatalError(err, loggingClient)
	}

	// override file-based configuration with environment variables.
	bootstrapConfig := config.GetBootstrap()
	registryInfo, startupInfo := configuration.OverrideFromEnvironment(bootstrapConfig.Registry, bootstrapConfig.Startup)
	config.SetRegistryInfo(registryInfo)
	config.SetStartupInfo(startupInfo)

	//	Update the startup timer to reflect whatever configuration read, if anything available.
	if startupInfo.Duration > 0 {
		startupTimer.SetDuration(startupInfo.Duration)
	}
	if startupInfo.Interval > 0 {
		startupTimer.SetInterval(startupInfo.Interval)
	}

	// set up registryClient and loggingClient; update configuration from registry if we're using a registry.
	switch useRegistry {
	case true:
		registryClient, err = configuration.UpdateFromRegistry(ctx, startupTimer, config, loggingClient, serviceKey)
		if err != nil {
			fatalError(err, loggingClient)
		}
		loggingClient = logging.FactoryFromConfiguration(serviceKey, config)
		configuration.ListenForChanges(&wg, ctx, config, loggingClient, registryClient)
	case false:
		loggingClient = logging.FactoryFromConfiguration(serviceKey, config)
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationInterfaceName: func(get di.Get) interface{} {
			return config
		},
		container.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return loggingClient
		},
		container.RegistryClientInterfaceName: func(get di.Get) interface{} {
			return registryClient
		},
	})

	// call individual bootstrap handlers.
	for i := range handlers {
		if handlers[i](&wg, ctx, startupTimer, dic) == false {
			cancel()
			break
		}
	}

	// wait for go routines to stop executing.
	wg.Wait()
}
