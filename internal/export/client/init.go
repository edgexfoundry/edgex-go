//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/memory"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

const (
	PingApiPath = "/api/v1/ping"
)

// Global variables
var dbClient export.DBClient
var LoggingClient logger.LoggingClient
var Configuration *ConfigurationStruct

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
				LoggingClient = logger.NewClient(internal.ExportClientServiceKey, Configuration.EnableRemoteLogging, logTarget)
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

func Init() bool {
	if Configuration == nil || dbClient == nil {
		return false
	}
	return true
}

func Destroy() {
	if dbClient != nil {
		dbClient.CloseSession()
		dbClient = nil
	}
}

func connectToDatabase() error {
	// Create a database client
	var err error
	dbConfig := db.Configuration{
		DbType:       Configuration.DBType,
		Host:         Configuration.MongoURL,
		Port:         Configuration.MongoPort,
		Timeout:      Configuration.MongoConnectTimeout,
		DatabaseName: Configuration.MongoDatabaseName,
		Username:     Configuration.MongoUsername,
		Password:     Configuration.MongoPassword,
	}
	dbClient, err = newDBClient(Configuration.DBType, dbConfig)
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
		err := connectToConsul(conf)
		if err != nil {
			return nil, err
		}
	}
	return conf, nil
}

func connectToConsul(conf *ConfigurationStruct) error {
	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    internal.ExportClientServiceKey,
		ServicePort:    conf.Port,
		ServiceAddress: conf.Hostname,
		CheckAddress:   "http://" + conf.Hostname + ":" + strconv.Itoa(conf.Port) + PingApiPath,
		CheckInterval:  conf.CheckInterval,
		ConsulAddress:  conf.ConsulHost,
		ConsulPort:     conf.ConsulPort,
	})

	if err != nil {
		return fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	} else {
		// Update configuration data from Consul
		if err := consulclient.CheckKeyValuePairs(conf, internal.ExportClientServiceKey, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func setLoggingTarget() string {
	logTarget := Configuration.LoggingRemoteURL
	if !Configuration.EnableRemoteLogging {
		return Configuration.LogFile
	}
	return logTarget
}
