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

package configuration

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	registryTypes "github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

// createRegistryClient creates and returns a registry.Client instance.
func createRegistryClient(serviceKey string, config interfaces.Configuration) (registry.Client, error) {
	bootstrapConfig := config.GetBootstrap()
	return registry.NewRegistryClient(
		registryTypes.Config{
			Host:            bootstrapConfig.Registry.Host,
			Port:            bootstrapConfig.Registry.Port,
			Type:            bootstrapConfig.Registry.Type,
			ServiceKey:      serviceKey,
			ServiceHost:     bootstrapConfig.Service.Host,
			ServicePort:     bootstrapConfig.Service.Port,
			ServiceProtocol: bootstrapConfig.Service.Protocol,
			CheckInterval:   bootstrapConfig.Service.CheckInterval,
			CheckRoute:      clients.ApiPingRoute,
			Stem:            internal.ConfigRegistryStemCore + internal.ConfigMajorVersion,
		})
}

// UpdateFromRegistry connects to the registry, registers the service, gets configuration, and updates the service's
// configuration struct.
func UpdateFromRegistry(
	ctx context.Context,
	startupTimer startup.Timer,
	config interfaces.Configuration,
	loggingClient logger.LoggingClient,
	serviceKey string) (registry.Client, error) {

	var updateFromRegistry = func(registryClient registry.Client) error {
		if !registryClient.IsAlive() {
			return errors.New("registry is not available")
		}

		if err := registryClient.Register(); err != nil {
			return errors.New(fmt.Sprintf("could not register service with Registry: %v", err.Error()))
		}

		rawConfig, err := registryClient.GetConfiguration(config)
		if err != nil {
			return errors.New(fmt.Sprintf("could not get configuration from Registry: %v", err.Error()))
		}

		if !config.UpdateFromRaw(rawConfig) {
			return errors.New("configuration from Registry failed type check")
		}

		return nil
	}

	registryClient, err := createRegistryClient(serviceKey, config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("createRegistryClient failed: %v", err.Error()))
	}

	for startupTimer.HasNotElapsed() {
		if err := updateFromRegistry(registryClient); err != nil {
			loggingClient.Warn(err.Error())
			select {
			case <-ctx.Done():
				return nil, errors.New("aborted UpdateFromRegistry()")
			default:
				startupTimer.SleepForInterval()
				continue
			}
		}
		return registryClient, nil
	}
	return nil, errors.New("unable to update configuration from registry in allotted time")
}

// ListenForChanges leverages the registry client's WatchForChanges() method to receive changes to and update the
// service's configuration struct's writable substruct.  It's assumed the log level is universally part of the
// writable struct and this function explicitly updates the loggingClient's log level when new configuration changes
// are received.
func ListenForChanges(
	wg *sync.WaitGroup,
	ctx context.Context,
	config interfaces.Configuration,
	loggingClient logger.LoggingClient,
	registryClient registry.Client) {

	wg.Add(1)
	go func() {
		defer wg.Done()

		errorStream := make(chan error)
		defer close(errorStream)

		updateStream := make(chan interface{})
		defer close(updateStream)

		registryClient.WatchForChanges(updateStream, errorStream, config.EmptyWritablePtr(), internal.WritableKey)

		for {
			select {
			case <-ctx.Done():
				return

			case ex := <-errorStream:
				loggingClient.Error(ex.Error())

			case raw, ok := <-updateStream:
				if !ok {
					return
				}

				if !config.UpdateWritableFromRaw(raw) {
					loggingClient.Error("ListenForChanges() type check failed")
					return
				}

				loggingClient.Info("Writeable configuration has been updated from the Registry")
				loggingClient.SetLogLevel(config.GetLogLevel())
			}
		}
	}()
}
