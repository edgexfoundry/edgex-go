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
package command

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	registryTypes "github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
)

var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient
var mdc metadata.DeviceClient
var cc metadata.CommandClient
var registryClient registry.Client
var registryErrors chan error        //A channel for "config wait errors" sourced from Registry
var registryUpdates chan interface{} //A channel for "config updates" sourced from Registry

func Retry(params startup.BootParams, wait *sync.WaitGroup, ch chan error) {
	now := time.Now()
	until := now.Add(time.Millisecond * time.Duration(params.BootTimeout))
	for time.Now().Before(until) {
		var err error
		//When looping, only handle configuration if it hasn't already been set.
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
				//Check against boot timeout default
				if Configuration.Service.BootTimeout != params.BootTimeout {
					until = now.Add(time.Millisecond * time.Duration(Configuration.Service.BootTimeout))
				}
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(clients.CoreCommandServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Writable.LogLevel)

				//Initialize service clients
				initializeClients(params.UseRegistry)
			}
		}

		if Configuration != nil {
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

func Init(useRegistry bool) bool {
	if Configuration == nil {
		return false
	}

	if useRegistry {
		registryErrors = make(chan error)
		registryUpdates = make(chan interface{})
		go listenForConfigChanges()
	}

	go telemetry.StartCpuUsageAverage()

	return true
}

func Destruct() {
	if registryErrors != nil {
		close(registryErrors)
	}

	if registryUpdates != nil {
		close(registryUpdates)
	}
}

func initializeConfiguration(useRegistry bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain RegistryHost/Port
	configuration := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		return nil, err
	}

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
		ServiceKey:      clients.CoreCommandServiceKey,
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

	registryClient.WatchForChanges(registryUpdates, registryErrors, &WritableInfo{}, internal.WritableKey)

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-signals:
			// Quietly and gracefully stop when SIGINT/SIGTERM received
			return
		case ex := <-registryErrors:
			LoggingClient.Error(ex.Error())
		case raw, ok := <-registryUpdates:
			if ok {
				actual, ok := raw.(*WritableInfo)
				if !ok {
					LoggingClient.Error("listenForConfigChanges() type check failed")
				}
				Configuration.Writable = *actual
				LoggingClient.Info("Writeable configuration has been updated from the Registry")
				LoggingClient.SetLogLevel(Configuration.Writable.LogLevel)
			} else {
				return
			}
		}
	}
}

func initializeClients(useRegistry bool) {
	// Create metadata clients
	params := types.EndpointParams{
		ServiceKey:  clients.CoreMetaDataServiceKey,
		Path:        clients.ApiDeviceRoute,
		UseRegistry: useRegistry,
		Url:         Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
		Interval:    Configuration.Service.ClientMonitor,
	}

	mdc = metadata.NewDeviceClient(params, startup.Endpoint{RegistryClient: &registryClient})
	params.Path = clients.ApiCommandRoute
	params.Url = Configuration.Clients["Metadata"].Url() + clients.ApiCommandRoute
	cc = metadata.NewCommandClient(params, startup.Endpoint{RegistryClient: &registryClient})
}

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}
