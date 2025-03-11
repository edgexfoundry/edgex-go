/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2023 Intel Corporation
 * Copyright 2024 IOTech Ltd
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
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/environment"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/registration"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/go-mod-registry/v4/registry"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// Deferred defines the signature of a function returned by RunAndReturnWaitGroup that should be executed via defer.
type Deferred func()

// fatalError logs an error and exits the application.  It's intended to be used only within the bootstrap prior to
// any go routines being spawned.
func fatalError(err error, lc logger.LoggingClient) {
	lc.Error(err.Error())
	os.Exit(1)
}

// translateInterruptToCancel spawns a go routine to translate the receipt of a SIGTERM signal to a call to cancel
// the context used by the bootstrap implementation.
func translateInterruptToCancel(ctx context.Context, wg *sync.WaitGroup, cancel context.CancelFunc) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		signalStream := make(chan os.Signal, 1)
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

// RunAndReturnWaitGroup bootstraps an application.  It loads configuration and calls the provided list of handlers.
// Any long-running process should be spawned as a go routine in a handler.  Handlers are expected to return
// immediately.  Once all of the handlers are called this function will return a sync.WaitGroup reference to the caller.
// It is intended that the caller take whatever additional action makes sense before calling Wait() on the returned
// reference to wait for the application to be signaled to stop (and the corresponding goroutines spawned in the
// various handlers to be stopped cleanly).
func RunAndReturnWaitGroup(
	ctx context.Context,
	cancel context.CancelFunc,
	commonFlags flags.Common,
	serviceKey string,
	configStem string,
	serviceConfig interfaces.Configuration,
	configUpdated config.UpdatedStream,
	startupTimer startup.Timer,
	dic *di.Container,
	useSecretProvider bool, // TODO: remove useSecretProvider and use serviceType in place with its constant
	serviceType string,
	handlers []interfaces.BootstrapHandler) (*sync.WaitGroup, Deferred, bool) {

	var err error
	var wg sync.WaitGroup
	deferred := func() {}

	// Check if service provided an initial Logging Client to use. If not create one and add it to the DIC.
	lc := container.LoggingClientFrom(dic.Get)
	if lc == nil {
		lc = logger.NewClient(serviceKey, models.InfoLog)
		dic.Update(di.ServiceConstructorMap{
			container.LoggingClientInterfaceName: func(get di.Get) interface{} {
				return lc
			},
		})
	}

	utils.AdaptLogrusBasedLogging(lc)
	translateInterruptToCancel(ctx, &wg, cancel)

	envVars := environment.NewVariables(lc)

	var secretProvider interfaces.SecretProviderExt
	if useSecretProvider {
		secretProvider, err = secret.NewSecretProvider(serviceConfig, envVars, ctx, startupTimer, dic, serviceKey)
		if err != nil {
			fatalError(fmt.Errorf("failed to create SecretProvider: %s", err.Error()), lc)
		}
	}

	// The SecretProvider is initialized and placed in the DIS as part of processing the configuration due
	// to the need for it to be used to get Access Token for the Configuration Provider and having to wait to
	// initialize it until after the configuration is loaded from file.
	configProcessor := config.NewProcessor(commonFlags, envVars, startupTimer, ctx, &wg, configUpdated, dic)
	if err := configProcessor.Process(serviceKey, serviceType, configStem, serviceConfig, secretProvider, secret.NewJWTSecretProvider(secretProvider)); err != nil {
		fatalError(err, lc)
	}

	var registryClient registry.Client

	envUseRegistry, wasOverridden := envVars.UseRegistry()
	if envUseRegistry || (commonFlags.UseRegistry() && !wasOverridden) {
		registryClient, err = registration.RegisterWithRegistry(
			ctx,
			startupTimer,
			serviceConfig,
			lc,
			serviceKey,
			dic)
		if err != nil {
			fatalError(err, lc)
		}

		deferred = func() {
			lc.Info("Un-Registering service from the Registry")
			err := registryClient.Unregister()
			if err != nil {
				lc.Error("Unable to Un-Register service from the Registry", "error", err.Error())
			}
		}
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationInterfaceName: func(get di.Get) interface{} {
			return serviceConfig
		},
		container.RegistryClientInterfaceName: func(get di.Get) interface{} {
			return registryClient
		},
		container.CancelFuncName: func(get di.Get) interface{} {
			return cancel
		},
	})

	// call individual bootstrap handlers.
	startedSuccessfully := true
	for i := range handlers {
		if !handlers[i](ctx, &wg, startupTimer, dic) {
			cancel()
			startedSuccessfully = false
			break
		}
	}

	// Service that don't use the Security Provider also will not collect metrics. These are the security services that
	// run during bootstrapping of the secure deployment
	if useSecretProvider && startedSuccessfully {
		// Have to delay registering the general common service metrics until all bootstrap handlers have run so that there is
		// opportunity for the MetricsManager to have been created.
		metricsManager := container.MetricsManagerFrom(dic.Get)
		if metricsManager != nil {
			secretProvider := container.SecretProviderExtFrom(dic.Get)
			if secretProvider != nil {
				metrics := secretProvider.GetMetricsToRegister()
				registerMetrics(metricsManager, metrics, lc)

				// TODO: use this same approach to register future service metric controlled by other components
			}
		} else {
			lc.Warn("MetricsManager not available. General common service metrics will not be reported. ")
		}
	}

	return &wg, deferred, startedSuccessfully
}

// Run bootstraps an application.  It loads configuration and calls the provided list of handlers.  Any long-running
// process should be spawned as a go routine in a handler.  Handlers are expected to return immediately.  Once all of
// the handlers are called this function will wait for any go routines spawned inside the handlers to exit before
// returning to the caller.  It is intended that the caller stop executing on the return of this function.
func Run(
	ctx context.Context,
	cancel context.CancelFunc,
	commonFlags flags.Common,
	serviceKey string,
	configStem string,
	serviceConfig interfaces.Configuration,
	startupTimer startup.Timer,
	dic *di.Container,
	useSecretProvider bool,
	serviceType string,
	handlers []interfaces.BootstrapHandler) {

	wg, deferred, success := RunAndReturnWaitGroup(
		ctx,
		cancel,
		commonFlags,
		serviceKey,
		configStem,
		serviceConfig,
		nil,
		startupTimer,
		dic,
		useSecretProvider,
		serviceType,
		handlers,
	)

	if !success {
		// This only occurs when a bootstrap handler has fail.
		// The handler will have logged an error, so not need to log a message here.
		cancel()
		os.Exit(1)
	}

	defer deferred()

	// wait for go routines to stop executing.
	wg.Wait()
}

func registerMetrics(metricsManager interfaces.MetricsManager, metrics map[string]interface{}, lc logger.LoggingClient) {
	for metricName, metric := range metrics {
		err := metricsManager.Register(metricName, metric, nil)
		if err != nil {
			lc.Warnf("Unable to register %s metric for reporting: %v", metricName, err)
			continue
		}

		lc.Infof("%s metric registered and will be reported (if enabled)", metricName)
	}
}
