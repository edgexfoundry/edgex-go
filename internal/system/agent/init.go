/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package agent

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	registryTypes "github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

// Global variables
var Configuration *ConfigurationStruct
var generalClients map[string]general.GeneralClient
var LoggingClient logger.LoggingClient
var registryClient registry.Client
var chErrors chan error        //A channel for "config wait error" sourced from Registry
var chUpdates chan interface{} //A channel for "config updates" sourced from Registry

// Note that executorClient is the empty interface so that we may type-cast it
// to whatever operation we need it to do at runtime.
var executorClient interface{}

func Retry(params config.BootParams, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(params.Retry.Timeout))
	attempts := 0
	for time.Now().Before(until) && attempts < params.Retry.Count {
		var err error
		// When looping, only handle configuration if it hasn't already been set.
		// Note, too, that the SMA-managed services are bootstrapped by the SMA.
		// Read in those setting, too, which specifies details for those services
		// (Those setting were _previously_ to be found in a now-defunct TOML manifest file).
		if Configuration == nil {
			Configuration, err = initializeConfiguration(params.UseRegistry, params.UseProfile)
			if err != nil {
				ch <- err
				if !params.UseRegistry {
					//Error occurred when attempting to read from local filesystem. Fail fast.
					close(ch)
					wait.Done()
					return
				}
			} else {
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(clients.SystemManagementAgentServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Writable.LogLevel)

				//Initialize service clients
				initializeClients(params.UseRegistry)
			}
		}

		// Exit the loop if the dependencies have been satisfied.
		if Configuration != nil {
			break
		}
		time.Sleep(time.Second * time.Duration(params.Retry.Wait))
		attempts++
	}
	close(ch)
	wait.Done()

	return
}

func Init(useRegistry bool) bool {
	if Configuration == nil {
		return false
	}
	executorClient = &ExecuteApp{}

	if useRegistry && registryClient != nil {
		chErrors = make(chan error)
		chUpdates = make(chan interface{})
		go listenForConfigChanges()
	}

	return true
}

func Destruct() {

	if chErrors != nil {
		close(chErrors)
	}

	if chUpdates != nil {
		close(chUpdates)
	}
}

func initializeConfiguration(useRegistry bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain Registry Host/Port
	configuration := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		return nil, err
	}
	configuration.Registry = config.OverrideFromEnvironment(configuration.Registry)

	if useRegistry {
		err = connectToRegistry(configuration)
		if err != nil {
			return nil, err
		}

		rawConfig, err := registryClient.GetConfiguration(configuration)
		if err != nil {
			return nil, fmt.Errorf("could not get configuration from Registry: %v", err.Error())
		}

		actual, ok := rawConfig.(*ConfigurationStruct)
		if !ok {
			return nil, fmt.Errorf("configuration from Registry failed type check")
		}

		configuration = actual

		// Check that information was successfully read from Registry
		if configuration.Service.Port == 0 {
			return nil, errors.New("error reading configuration from Registry")
		}
	}

	return configuration, nil
}

func connectToRegistry(conf *ConfigurationStruct) error {
	var err error
	registryConfig := registryTypes.Config{
		Host:            conf.Registry.Host,
		Port:            conf.Registry.Port,
		Type:            conf.Registry.Type,
		ServiceKey:      clients.SystemManagementAgentServiceKey,
		ServiceHost:     conf.Service.Host,
		ServicePort:     conf.Service.Port,
		ServiceProtocol: conf.Service.Protocol,
		CheckInterval:   conf.Service.CheckInterval,
		CheckRoute:      clients.ApiPingRoute,
		Stem:            internal.ConfigRegistryStem,
	}

	registryClient, err = registry.NewRegistryClient(registryConfig)
	if err != nil {
		return fmt.Errorf("connection to Registry could not be made: %v", err.Error())
	}

	// Check if registry service is running
	if !registryClient.IsAlive() {
		return fmt.Errorf("registry is not available")
	}

	// Register the service with Registry
	err = registryClient.Register()
	if err != nil {
		return fmt.Errorf("could not register service with Registry: %v", err.Error())
	}

	return nil
}

func listenForConfigChanges() {
	if registryClient == nil {
		LoggingClient.Error("listenForConfigChanges() registry client not set")
		return
	}

	registryClient.WatchForChanges(chUpdates, chErrors, &WritableInfo{}, internal.WritableKey)

	// TODO: Refactor names in separate PR: See comments on PR #1133
	chSignals := make(chan os.Signal)
	signal.Notify(chSignals, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-chSignals:
			// Quietly and gracefully stop when SIGINT/SIGTERM received
			return

		case ex := <-chErrors:
			LoggingClient.Error(ex.Error())

		case raw, ok := <-chUpdates:
			if !ok {
				return
			}

			actual, ok := raw.(*WritableInfo)
			if !ok {
				LoggingClient.Error("listenForConfigChanges() type check failed")
				return
			}

			Configuration.Writable = *actual

			LoggingClient.Info("Writeable configuration has been updated from the Registry")
			LoggingClient.SetLogLevel(Configuration.Writable.LogLevel)
		}
	}
}

func initializeClients(useRegistry bool) {

	generalClients = make(map[string]general.GeneralClient)
	services := config.ListDefaultServices()

	for serviceKey, serviceName := range services {
		params := types.EndpointParams{
			ServiceKey:  serviceKey,
			Path:        "/",
			UseRegistry: useRegistry,
			Url:         Configuration.Clients[serviceName].Url(),
			Interval:    internal.ClientMonitorDefault,
		}
		generalClients[serviceKey] = general.NewGeneralClient(params, startup.Endpoint{RegistryClient: &registryClient})
	}
}

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}
