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

type GeneralClients map[string]general.GeneralClient

// Instance contains what were global variables.
type Instance struct {
	Configuration  *ConfigurationStruct
	GenClients     GeneralClients
	LoggingClient  logger.LoggingClient
	RegistryClient registry.Client
	chErrors       chan error       //A channel for "config wait error" sourced from Registry
	chUpdates      chan interface{} //A channel for "config updates" sourced from Registry
}

// NewInstance returns an empty Instance struct that is subsequently populated by a call to the Retry method.
func NewInstance() *Instance {
	return &Instance{}
}

func (instance *Instance) Retry(useRegistry bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		// When looping, only handle configuration if it hasn't already been set.
		// Note, too, that the SMA-managed services are bootstrapped by the SMA.
		// Read in those setting, too, which specifies details for those services
		// (Those setting were _previously_ to be found in a now-defunct TOML manifest file).
		if instance.Configuration == nil {
			instance.Configuration, err = instance.initializeConfiguration(useRegistry, useProfile)
			if err != nil {
				ch <- err
				if !useRegistry {
					//Error occurred when attempting to read from local filesystem. Fail fast.
					close(ch)
					wait.Done()
					return
				}
			} else {
				// Setup Logging
				logTarget := instance.setLoggingTarget()
				instance.LoggingClient = logger.NewClient(clients.SystemManagementAgentServiceKey, instance.Configuration.Logging.EnableRemote, logTarget, instance.Configuration.Writable.LogLevel)

				//Initialize service clients
				instance.initializeClients(useRegistry)
			}
		}

		// Exit the loop if the dependencies have been satisfied.
		if instance.Configuration != nil {
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

func (instance *Instance) Init(useRegistry bool) bool {
	if instance.Configuration == nil {
		return false
	}

	if useRegistry && instance.RegistryClient != nil {
		instance.chErrors = make(chan error)
		instance.chUpdates = make(chan interface{})
		go instance.listenForConfigChanges()
	}

	return true
}

func (instance *Instance) Destruct() {

	if instance.chErrors != nil {
		close(instance.chErrors)
	}

	if instance.chUpdates != nil {
		close(instance.chUpdates)
	}
}

func (instance *Instance) initializeConfiguration(useRegistry bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain Registry Host/Port
	configuration := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		return nil, err
	}
	configuration.Registry = config.OverrideFromEnvironment(configuration.Registry)

	if useRegistry {
		err = instance.connectToRegistry(configuration)
		if err != nil {
			return nil, err
		}

		rawConfig, err := instance.RegistryClient.GetConfiguration(configuration)
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

func (instance *Instance) connectToRegistry(conf *ConfigurationStruct) error {
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

	instance.RegistryClient, err = registry.NewRegistryClient(registryConfig)
	if err != nil {
		return fmt.Errorf("connection to Registry could not be made: %v", err.Error())
	}

	// Check if registry service is running
	if !instance.RegistryClient.IsAlive() {
		return fmt.Errorf("registry is not available")
	}

	// Register the service with Registry
	err = instance.RegistryClient.Register()
	if err != nil {
		return fmt.Errorf("could not register service with Registry: %v", err.Error())
	}

	return nil
}

func (instance *Instance) listenForConfigChanges() {
	if instance.RegistryClient == nil {
		instance.LoggingClient.Error("listenForConfigChanges() registry client not set")
		return
	}

	instance.RegistryClient.WatchForChanges(instance.chUpdates, instance.chErrors, &WritableInfo{}, internal.WritableKey)

	// TODO: Refactor names in separate PR: See comments on PR #1133
	chSignals := make(chan os.Signal)
	signal.Notify(chSignals, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-chSignals:
			// Quietly and gracefully stop when SIGINT/SIGTERM received
			return

		case ex := <-instance.chErrors:
			instance.LoggingClient.Error(ex.Error())

		case raw, ok := <-instance.chUpdates:
			if !ok {
				return
			}

			actual, ok := raw.(*WritableInfo)
			if !ok {
				instance.LoggingClient.Error("listenForConfigChanges() type check failed")
				return
			}

			instance.Configuration.Writable = *actual

			instance.LoggingClient.Info("Writeable configuration has been updated from the Registry")
			instance.LoggingClient.SetLogLevel(instance.Configuration.Writable.LogLevel)
		}
	}
}

func (instance *Instance) initializeClients(useRegistry bool) {
	instance.GenClients = make(GeneralClients)

	var updateGenClients = func(serviceKey string, serviceName string) {
		instance.GenClients[serviceKey] = general.NewGeneralClient(
			types.EndpointParams{
				ServiceKey:  serviceKey,
				Path:        "/",
				UseRegistry: useRegistry,
				Url:         instance.Configuration.Clients[serviceName].Url(),
				Interval:    internal.ClientMonitorDefault,
			},
			startup.Endpoint{RegistryClient: &instance.RegistryClient})
	}

	if useRegistry {
		for serviceKey, serviceName := range config.ListDefaultServices() {
			updateGenClients(serviceKey, serviceName)
		}
		return
	}

	// if the registry is not being used, load clients from configurations; assume configuration key is service name
	for key := range instance.Configuration.Clients {
		updateGenClients(key, key)
	}
}

func (instance *Instance) setLoggingTarget() string {
	if instance.Configuration.Logging.EnableRemote {
		return instance.Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return instance.Configuration.Logging.File
}
