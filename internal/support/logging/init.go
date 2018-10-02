//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/memory"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/pkg/errors"
)

var Configuration *ConfigurationStruct
var dbClient DBClient
var LoggingClient logger.LoggingClient

func Retry(useConsul bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	LoggingClient = newPrivateLogger()
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

func Init() bool {
	if Configuration == nil || dbClient == nil {
		return false
	}
	return true
}

func Destruct() {
	if dbClient != nil {
		dbClient.CloseSession()
		dbClient = nil
	}
}

func initializeConfiguration(useConsul bool, useProfile string) (*ConfigurationStruct, error) {
	//We currently have to load configuration from filesystem first in order to obtain ConsulHost/Port
	conf := &ConfigurationStruct{}
	err := config.LoadFromFileV2(useProfile, conf)
	if err != nil {
		return nil, err
	}

	if useConsul {
		err := connectToConsul(conf)
		if err != nil {
			return nil, err
		}
	}
	return conf, nil
}

func connectToConsul(conf *ConfigurationStruct) error {
	//Obtain ConsulConfig
	cfg := consulclient.NewConsulConfig(conf.Registry, conf.Service, internal.SupportLoggingServiceKey)
	// Register the service in Consul
	err := consulclient.ConsulInit(cfg)

	if err != nil {
		return fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	} else {
		// Update configuration data from Consul
		updateCh := make(chan interface{})
		errCh := make(chan error)
		dec := consulclient.NewConsulDecoder(conf.Registry)
		dec.Target = &ConfigurationStruct{}
		dec.Prefix = internal.ConfigV2Stem + internal.SupportLoggingServiceKey
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
				return errors.New("type check failed")
			}
			Configuration = actual
		}
	}
	return err
}

// Return the dbClient interface
func newDBClient(dbType string, config db.Configuration) (DBClient, error) {
	switch dbType {
	case db.MongoDB:
		return mongo.NewClient(config), nil
	case db.MemoryDB:
		return &memory.MemDB{}, nil
	default:
		return nil, db.ErrUnsupportedDatabase
	}
}

func getPersistence() error {
	if Configuration.Persistence == PersistenceFile {
		dbClient = &fileLog{filename: Configuration.Logging.File}
	} else if Configuration.Persistence == PersistenceDB {
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
	}
	return nil
}
