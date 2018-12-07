//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/memory"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/export/distro"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/pkg/errors"
)

// Global variables
var chConfig chan interface{} //A channel for use by ConsulDecoder in detecting configuration mods.
var dbClient export.DBClient
var LoggingClient logger.LoggingClient
var Configuration *ConfigurationStruct
var dc distro.DistroClient

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
				LoggingClient = logger.NewClient(internal.ExportClientServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Logging.Level)

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

	if chConfig != nil {
		close(chConfig)
		chConfig = nil
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

	// Connect to the database
	err = dbClient.Connect()
	if err != nil {
		dbClient = nil
		return fmt.Errorf("couldn't connect to database: %v", err.Error())
	}
	return nil
}

// Return the dbClient interface
func newDBClient(dbType string, config db.Configuration) (export.DBClient, error) {
	switch dbType {
	case db.MongoDB:
		return mongo.NewClient(config), nil
	case db.MemoryDB:
		return &memory.MemDB{}, nil
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

func initializeClients(useConsul bool) {
	// Create export-distro client
	params := types.EndpointParams{
		ServiceKey:  internal.ExportDistroServiceKey,
		UseRegistry: useConsul,
		Url:         Configuration.Clients["Distro"].Url(),
		Interval:    Configuration.Service.ClientMonitor,
	}

	dc = distro.NewDistroClient(params, startup.Endpoint{})
}

func connectToConsul(conf *ConfigurationStruct) (*ConfigurationStruct, error) {
	//Obtain ConsulConfig
	cfg := consulclient.NewConsulConfig(conf.Registry, conf.Service, internal.ExportClientServiceKey)
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
	dec.Prefix = internal.ConfigV2Stem + internal.ExportClientServiceKey
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
	dec.Target = &ConfigurationStruct{}
	dec.Prefix = internal.ConfigV2Stem + internal.ExportClientServiceKey
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
				actual, ok := raw.(*ConfigurationStruct)
				if !ok {
					LoggingClient.Error("listenForConfigChanges() type check failed")
				}
				Configuration = actual //Mutex needed?
			} else {
				return
			}
		}
	}
}

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}
