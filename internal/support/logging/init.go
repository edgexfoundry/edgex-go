//
// Copyright (c) 2018 Tencent
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

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
	registryTypes "github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
)

var Configuration *ConfigurationStruct
var dbClient persistence
var LoggingClient logger.LoggingClient
var registryClient registry.Client
var registryErrors chan error        //A channel for "config wait errors" sourced from Registry
var registryUpdates chan interface{} //A channel for "config updates" sourced from Registry

func Retry(useRegistry bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	LoggingClient = newPrivateLogger()
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		//When looping, only handle configuration if it hasn't already been set.
		if Configuration == nil {
			Configuration, err = initializeConfiguration(useRegistry, useProfile)
			if err != nil {
				ch <- err
				if !useRegistry {
					//Error occurred when attempting to read from local filesystem. Fail fast.
					close(ch)
					wait.Done()
					return
				}
			}
		}

		//Only attempt to connect to database if configuration has been populated
		if Configuration != nil {
			err = getPersistence()
			if err != nil {
				ch <- err
			} else {
				break
			}
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

func Init(useRegistry bool) bool {
	if Configuration == nil || dbClient == nil {
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
	if dbClient != nil {
		dbClient.closeSession()
		dbClient = nil
	}

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
		ServiceKey:      clients.SupportLoggingServiceKey,
		ServiceHost:     conf.Service.Host,
		ServicePort:     conf.Service.Port,
		ServiceProtocol: conf.Service.Protocol,
		CheckInterval:   conf.Service.CheckInterval,
		CheckRoute:      clients.ApiPingRoute,
		Stem:            internal.ConfigRegistryStemCore + internal.ConfigMajorVersion,
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

func getPersistence() error {
	switch Configuration.Writable.Persistence {
	case PersistenceFile:
		dbClient = &fileLog{filename: Configuration.Logging.File}
	case PersistenceDB:
		// TODO: Integrate db layer with internal/pkg/db/ types so we can support other databases
		ms, err := connectToMongo()
		if err != nil {
			return err
		} else {
			dbClient = &mongoLog{session: ms}
		}
	default:
		return errors.New(fmt.Sprintf("unrecognized value Configuration.Persistence: %s", Configuration.Writable.Persistence))
	}
	return nil
}
