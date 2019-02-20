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
package scheduler

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"
	registry "github.com/edgexfoundry/go-mod-registry"
	"github.com/edgexfoundry/go-mod-registry/pkg/factory"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

var Configuration *ConfigurationStruct
var LoggingClient logger.LoggingClient
var dbClient interfaces.DBClient
var scClient interfaces.SchedulerQueueClient
var registryClient registry.Client
var errChan chan error          //A channel for "config wait error" sourced from Registry
var updateChan chan interface{} //A channel for "config updates" sourced from Registry

var chConfig chan interface{} //A channel for use by RegistryDecoder in detecting configuration mods.
var ticker *time.Ticker

func Retry(useRegistry bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	now := time.Now()
	until := now.Add(time.Millisecond * time.Duration(timeout))
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
				//Check against boot timeout default
				if Configuration.Service.BootTimeout != timeout {
					until = now.Add(time.Millisecond * time.Duration(Configuration.Service.BootTimeout))
				}
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(internal.SupportSchedulerServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Writable.LogLevel)

				// Initialize the ticker time
				ticker = time.NewTicker(time.Duration(Configuration.Writable.ScheduleIntervalTime) * time.Millisecond)
			}
		}

		if Configuration != nil {
			err := connectToDatabase()
			if err != nil {
				ch <- err
			}
			err = connectToSchedulerQueue()
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
		errChan = make(chan error)
		updateChan = make(chan interface{})
		go listenForConfigChanges()
	}
	return true
}

func Destruct() {
	if errChan != nil {
		close(errChan)
	}

	if updateChan != nil {
		close(updateChan)
	}

	if ticker != nil {
		StopTicker()
	}

	if dbClient != nil {
		dbClient.CloseSession()
		dbClient = nil
	}

	if scClient != nil {
		scClient = nil
	}
}

func initializeConfiguration(useRegistry bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain RegistryHost/Port
	conf := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, conf)
	if err != nil {
		return nil, err
	}

	if useRegistry {
		err = connectToRegistry(conf)
		if err != nil {
			return nil, err
		}

		rawConfig, err := registryClient.GetConfiguration(conf)
		if err != nil {
			return nil, fmt.Errorf("could not get configuration from Registry: %v", err.Error())
		}

		actual, ok := rawConfig.(*ConfigurationStruct)
		if !ok {
			return nil, fmt.Errorf("configuration from Registry failed type check")
		}

		conf = actual
	}
	return conf, nil
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

	registryClient, err = factory.NewRegistryClient(registryConfig, internal.SupportSchedulerServiceKey)
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

	registryClient.WatchForChanges(updateChan, errChan, &WritableInfo{}, internal.WritableKey)

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-signalChan:
			// Quietly and gracefully stop when SIGINT/SIGTERM received
			return
		case raw, ok := <-updateChan:
			if ok {
				actual, ok := raw.(*WritableInfo)
				if !ok {
					LoggingClient.Error("listenForConfigChanges() type check failed")
				}
				Configuration.Writable = *actual
				LoggingClient.Info("Writeable configuration has been updated. Setting log level to " + Configuration.Writable.LogLevel)
				LoggingClient.SetLogLevel(Configuration.Writable.LogLevel)
			} else {
				return
			}
		}
	}
}

func connectToDatabase() error {
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
	return nil
}

func connectToSchedulerQueue() error {
	var err error
	scClient, err = newScheduleQueueClient()
	if err != nil {
		scClient = nil
		return fmt.Errorf("couldn't create scheduler queue client: %v", err.Error())
	}
	return nil
}
func newScheduleQueueClient() (interfaces.SchedulerQueueClient, error) {
	return NewSchedulerQueueClient(), nil
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

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}
