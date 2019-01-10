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
package data

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/data/messaging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/pkg/errors"
)

// Global variables
var Configuration *ConfigurationStruct
var dbClient interfaces.DBClient
var LoggingClient logger.LoggingClient
var chEvents chan interface{} //A channel for "domain events" sourced from event operations
var chConfig chan interface{} //A channel for use by ConsulDecoder in detecting configuration mods.
var ep messaging.EventPublisher
var mdc metadata.DeviceClient
var msc metadata.DeviceServiceClient

func Retry(useConsul bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		//When looping, only handle configuration if it hasn't already been set.
		if Configuration == nil {
			Configuration, err = initializeConfiguration(useConsul, useProfile)
			if err != nil {
				ch <- err
				if !useConsul {
					//Error occurred when attempting to read from local filesystem. Fail fast.
					close(ch)
					wait.Done()
					return
				}
			} else {
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(internal.CoreDataServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Logging.Level)

				//Initialize service clients
				initializeClients(useConsul)
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

func Init(useConsul bool) bool {
	if Configuration == nil || dbClient == nil {
		return false
	}
	chEvents = make(chan interface{}, 100)
	initEventHandlers()

	if useConsul {
		chConfig = make(chan interface{})
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
	if chConfig != nil {
		close(chConfig)
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

func initializeConfiguration(useConsul bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain ConsulHost/Port
	conf := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, conf)
	if err != nil {
		return nil, err
	}

	if useConsul {
		conf, err = connectToConsul(conf)
		if err != nil {
			return nil, err
		}
	}
	return conf, nil
}

func connectToConsul(conf *ConfigurationStruct) (*ConfigurationStruct, error) {
	//Obtain ConsulConfig
	cfg := consulclient.NewConsulConfig(conf.Registry, conf.Service, internal.CoreDataServiceKey)
	// Register the service in Consul
	err := consulclient.ConsulInit(cfg)

	if err != nil {
		return conf, fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	}
	// Update configuration data from Consul
	updateCh := make(chan interface{})
	errCh := make(chan error)
	dec := consulclient.NewConsulDecoder(conf.Registry)
	dec.Target = &ConfigurationStruct{}
	dec.Prefix = internal.ConfigRegistryStem + internal.CoreDataServiceKey
	dec.ErrCh = errCh
	dec.UpdateCh = updateCh

	defer dec.Close()
	defer close(updateCh)
	defer close(errCh)
	go dec.Run()

	select {
	case <-time.After(2 * time.Second):
		err = errors.New("timeout loading config from registry")
	case ex := <-errCh:
		err = errors.New(ex.Error())
	case raw := <-updateCh:
		actual, ok := raw.(*ConfigurationStruct)
		if !ok {
			return conf, errors.New("type check failed")
		}
		conf = actual
		//Check that information was successfully read from Consul
		if conf.Service.Port == 0 {
			return nil, errors.New("error reading from Consul")
		}
	}

	return conf, err
}

func listenForConfigChanges() {
	errCh := make(chan error)
	dec := consulclient.NewConsulDecoder(Configuration.Registry)
	dec.Target = &WritableInfo{}
	dec.Prefix = internal.ConfigRegistryStem + internal.CoreDataServiceKey + internal.WritableKey
	dec.ErrCh = errCh
	dec.UpdateCh = chConfig

	defer dec.Close()
	defer close(errCh)

	go dec.Run()
	for {
		select {
		case ex := <-errCh:
			LoggingClient.Error(ex.Error())
		case raw, ok := <-chConfig:
			if ok {
				actual, ok := raw.(*WritableInfo)
				if !ok {
					LoggingClient.Error("listenForConfigChanges() type check failed")
				}
				Configuration.Writable = *actual
			} else {
				return
			}
		}
	}
}

func initializeClients(useConsul bool) {
	// Create metadata clients
	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiDeviceRoute,
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
		Interval:    Configuration.Service.ClientMonitor,
	}

	mdc = metadata.NewDeviceClient(params, startup.Endpoint{})

	params.Path = clients.ApiDeviceServiceRoute
	msc = metadata.NewDeviceServiceClient(params, startup.Endpoint{})

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
