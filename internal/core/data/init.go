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
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/data/messaging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	registry "github.com/edgexfoundry/go-mod-registry"
	"github.com/edgexfoundry/go-mod-registry/pkg/factory"
)

// Global variables
var Configuration *ConfigurationStruct
var dbClient interfaces.DBClient
var LoggingClient logger.LoggingClient
var Registry registry.Client
var chEvents chan interface{}      //A channel for "domain events" sourced from event operations
var errChannel chan error          //A channel for "config wait error" sourced from Registry
var updateChannel chan interface{} //A channel for "config updates" sourced from Registry

var ep messaging.EventPublisher
var mdc metadata.DeviceClient
var msc metadata.DeviceServiceClient

func Retry(useRegistry bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
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
			} else {
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(internal.CoreDataServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Writable.LogLevel)

				//Initialize service clients
				initializeClients(useRegistry)
			}
		}

		//Only attempt to connect to database if configuration has been populated
		if Configuration != nil {
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

	if useRegistry && Registry != nil {
		errChannel = make(chan error)
		updateChannel = make(chan interface{})
		go listenForConfigChanges()
	}
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

	if errChannel != nil {
		close(errChannel)
	}

	if updateChannel != nil {
		close(updateChannel)
	}
}

func connectToDatabase() error {
	// Create a database client
	var err error
	dbConfig := db.Configuration{
		Host:         Configuration.Databases["Primary"].Host,
		Port:         Configuration.Databases["Primary"].Port,
		Timeout:      Configuration.Databases["Primary"].Timeout,
		DatabaseName: Configuration.Databases["Primary"].Name,
		Username:     Configuration.Databases["Primary"].Username,
		Password:     Configuration.Databases["Primary"].Password,
	}
	dbClient, err = newDBClient(Configuration.Databases["Primary"].Type, dbConfig)
	if err != nil {
		dbClient = nil
		return fmt.Errorf("couldn't create database client: %v", err.Error())
	}

	return err
}

// Return the dbClient interface
func newDBClient(dbType string, config db.Configuration) (interfaces.DBClient, error) {
	switch dbType {
	case db.MongoDB:
		return mongo.NewClient(config)
	default:
		return nil, db.ErrUnsupportedDatabase
	}
}

func initializeConfiguration(useRegistry bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain Registry Host/Port
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

		rawConfig, err := Registry.GetConfiguration(configuration)
		if err != nil {
			return configuration, fmt.Errorf("could not get configuration from Registry: %v", err.Error())
		}

		actual, ok := rawConfig.(*ConfigurationStruct)
		if !ok {
			return configuration, fmt.Errorf("configuration from Registry failed type check")
		}

		configuration = actual
	}

	return configuration, nil
}

func connectToRegistry(conf *ConfigurationStruct) error {
	var err error
	registryConfig := registry.Config{
		Host:            conf.Registry.Host,
		Port:            conf.Registry.Port,
		Type:            conf.Registry.Type,
		ServiceHost:     conf.Service.Host,
		ServicePort:     conf.Service.Port,
		ServiceProtocol: conf.Service.Protocol,
		CheckInterval:   conf.Service.CheckInterval,
		CheckRoute:      clients.ApiPingRoute,
		Stem:            internal.ConfigRegistryStem,
	}

	Registry, err = factory.NewRegistryClient(registryConfig, internal.CoreDataServiceKey)
	if err != nil {
		return fmt.Errorf("connection to Registry could not be made: %v", err.Error())
	}

	// Check if registry service is running
	if !Registry.IsRegistryRunning() {
		return fmt.Errorf("registry is not available")
	}

	// Register the service with Registry
	err = Registry.Register()
	if err != nil {
		return fmt.Errorf("could not register service with Registry: %v", err.Error())
	}

	return nil
}

func listenForConfigChanges() {
	if Registry == nil {
		LoggingClient.Error("listenForConfigChanges() registry client not set")
		return
	}

	Registry.WatchForChanges(updateChannel, errChannel, &WritableInfo{}, internal.WritableKey)

	for {
		select {
		case ex := <-errChannel:
			LoggingClient.Error(ex.Error())

		case raw, ok := <-updateChannel:
			if !ok {
				return
			}

			actual, ok := raw.(*WritableInfo)
			if !ok {
				LoggingClient.Error("listenForConfigChanges() type check failed")
				return
			}

			Configuration.Writable = *actual

			LoggingClient.Info("Writeable configuration has been updated. Setting log level to " + Configuration.Writable.LogLevel)
			LoggingClient.SetLogLevel(Configuration.Writable.LogLevel)
		}
	}
}

func initializeClients(useRegistry bool) {
	// Create metadata clients
	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiDeviceRoute,
		UseRegistry: useRegistry,
		Url:         Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
		Interval:    Configuration.Service.ClientMonitor,
	}

	mdc = metadata.NewDeviceClient(params, startup.Endpoint{RegistryClient: &Registry})

	params.Path = clients.ApiDeviceServiceRoute
	msc = metadata.NewDeviceServiceClient(params, startup.Endpoint{RegistryClient: &Registry})

	// Create the event publisher
	ep = messaging.NewEventPublisher(messaging.PubSubConfiguration{
		AddressPort: Configuration.MessageQueue.Uri(),
	})
}

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}
