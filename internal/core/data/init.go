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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/edgexfoundry-holding/go-mod-core-security/pkg"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"
	registryTypes "github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
)

// Global variables
var Configuration *ConfigurationStruct
var dbClient interfaces.DBClient
var LoggingClient logger.LoggingClient
var registryClient registry.Client
var secretsClient pkg.SecretClient

// TODO: Refactor names in separate PR: See comments on PR #1133
var chEvents chan interface{}  // A channel for "domain events" sourced from event operations
var chErrors chan error        // A channel for "config wait error" sourced from Registry
var chUpdates chan interface{} // A channel for "config updates" sourced from Registry

var msgClient messaging.MessageClient
var mdc metadata.DeviceClient
var msc metadata.DeviceServiceClient

func Retry(params startup.BootParams, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(params.BootTimeout))
	for time.Now().Before(until) {
		var err error
		// When looping, only handle configuration if it hasn't already been set.
		if Configuration == nil {
			Configuration, err = initializeConfiguration(params)
			if err != nil {
				ch <- err
				if !params.UseRegistry {
					// Error occurred when attempting to read from local filesystem. Fail fast.
					close(ch)
					wait.Done()
					return
				}
			} else {
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(clients.CoreDataServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Writable.LogLevel)

				// Initialize service clients
				initializeClients(params.UseRegistry)
			}
		}

		if Configuration != nil {
			// Attempt to connect to secrets service. Fall back to local config on failure.
			if !params.UseLocalSecrets {
				err = connectAndPollSecrets()

				// Error occurred trying to read remote secrets. Fail fast.
				if err != nil {
					ch <- err
					ch <- fmt.Errorf("could not fetch remote secrets, shutting down")
					close(ch)
					wait.Done()
					return
				}
			}

			// Only attempt to connect to database if configuration has been populated
			err := connectToDatabase()
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
	chEvents = make(chan interface{}, 100)
	initEventHandlers()

	if useRegistry && registryClient != nil {
		chErrors = make(chan error)
		chUpdates = make(chan interface{})
		go listenForConfigChanges()
	}

	go telemetry.StartCpuUsageAverage()

	return true
}

func Destruct() {
	if dbClient != nil {
		dbClient.CloseSession()
		dbClient = nil
	}
	if chEvents != nil {
		close(chEvents)
	}

	if chErrors != nil {
		close(chErrors)
	}

	if chUpdates != nil {
		close(chUpdates)
	}
}

func connectToDatabase() error {
	// Create a database client
	var err error

	dbClient, err = newDBClient(Configuration.Databases["Primary"].Type)
	if err != nil {
		dbClient = nil
		return fmt.Errorf("couldn't create database client: %v", err.Error())
	}

	return nil
}

// Return the dbClient interface
func newDBClient(dbType string) (interfaces.DBClient, error) {
	switch dbType {
	case db.MongoDB:
		dbConfig := db.Configuration{
			Host:         Configuration.Databases["Primary"].Host,
			Port:         Configuration.Databases["Primary"].Port,
			Timeout:      Configuration.Databases["Primary"].Timeout,
			DatabaseName: Configuration.Databases["Primary"].Name,
			Username:     Configuration.Databases["Primary"].Username,
			Password:     Configuration.Databases["Primary"].Password,
		}
		return mongo.NewClient(dbConfig)
	case db.RedisDB:
		dbConfig := db.Configuration{
			Host: Configuration.Databases["Primary"].Host,
			Port: Configuration.Databases["Primary"].Port,
		}
		return redis.NewClient(dbConfig) // TODO: Verify this also connects to Redis
	default:
		return nil, db.ErrUnsupportedDatabase
	}
}

func initializeConfiguration(params startup.BootParams) (*ConfigurationStruct, error) {
	// We currently have to load configuration from filesystem first in order to obtain Registry Host/Port
	configuration := &ConfigurationStruct{}
	err := config.LoadFromFile(params.UseProfile, configuration)
	if err != nil {
		return nil, err
	}

	if params.UseRegistry {
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

func connectAndPollSecrets() error {
	var err error
	secretsClient, err = pkg.NewSecretClient(Configuration.Secrets)
	if err != nil {
		return err
	}

	username, err := secretsClient.GetValue("coredata")
	if err != nil {
		return err
	}

	password, err := secretsClient.GetValue("coredatapasswd")
	if err != nil {
		return err
	}

	Configuration.Databases["Primary"] = config.DatabaseInfo{
		Type:     Configuration.Databases["Primary"].Type,
		Timeout:  Configuration.Databases["Primary"].Timeout,
		Host:     Configuration.Databases["Primary"].Host,
		Port:     Configuration.Databases["Primary"].Port,
		Username: username,
		Password: password,
		Name:     Configuration.Databases["Primary"].Name,
	}

	return nil
}

func connectToRegistry(conf *ConfigurationStruct) error {
	var err error
	registryConfig := registryTypes.Config{
		Host:            conf.Registry.Host,
		Port:            conf.Registry.Port,
		Type:            conf.Registry.Type,
		ServiceKey:      clients.CoreDataServiceKey,
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
	// Create metadata clients
	params := types.EndpointParams{
		ServiceKey:  clients.CoreMetaDataServiceKey,
		Path:        clients.ApiDeviceRoute,
		UseRegistry: useRegistry,
		Url:         Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
		Interval:    Configuration.Service.ClientMonitor,
	}

	mdc = metadata.NewDeviceClient(params, startup.Endpoint{RegistryClient: &registryClient})

	params.Path = clients.ApiDeviceServiceRoute
	msc = metadata.NewDeviceServiceClient(params, startup.Endpoint{RegistryClient: &registryClient})

	// Create the messaging client
	var err error
	msgClient, err = messaging.NewMessageClient(msgTypes.MessageBusConfig{
		PublishHost: msgTypes.HostInfo{
			Host:     Configuration.MessageQueue.Host,
			Port:     Configuration.MessageQueue.Port,
			Protocol: Configuration.MessageQueue.Protocol,
		},
		Type: Configuration.MessageQueue.Type,
	})

	if err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to create messaging client: %s", err.Error()))
	}
}

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}
