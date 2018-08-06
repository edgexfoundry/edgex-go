/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package notifications

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/consul"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
)

// Global variables
var Configuration *ConfigurationStruct
var dbClient interfaces.DBClient
var LoggingClient logger.LoggingClient
var limitMax int
var resendLimit int
var smtpPort string
var smtpHost string
var smtpSender string
var smtpPassword string
var smtpSubject string

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
				LoggingClient = logger.NewClient(internal.SupportNotificationsServiceKey, Configuration.EnableRemoteLogging, logTarget)
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

func Destruct() {
	if dbClient != nil {
		dbClient.CloseSession()
		dbClient = nil
	}
}

func connectToDatabase() error {
	// Create a database client
	var err error
	dbConfig := db.Configuration{
		Host:         Configuration.MongoDBHost,
		Port:         Configuration.MongoDBPort,
		Timeout:      Configuration.MongoDBConnectTimeout,
		DatabaseName: Configuration.MongoDatabaseName,
		Username:     Configuration.MongoDBUserName,
		Password:     Configuration.MongoDBPassword,
	}
	dbClient, err = newDBClient(Configuration.DBType, dbConfig)
	if err != nil {
		return fmt.Errorf("couldn't create database client: %v", err.Error())
	}

	// Connect to the database
	err = dbClient.Connect()
	if err != nil {
		return fmt.Errorf("couldn't connect to database: %v", err.Error())
	}
	return nil
}

// Return the dbClient interface
func newDBClient(dbType string, config db.Configuration) (interfaces.DBClient, error) {
	switch dbType {
	case db.MongoDB:
		return mongo.NewClient(config), nil
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
		ServiceName:    internal.SupportNotificationsServiceKey,
		ServicePort:    conf.ServicePort,
		ServiceAddress: conf.ServiceAddress,
		CheckAddress:   conf.ConsulCheckAddress,
		CheckInterval:  conf.CheckInterval,
		ConsulAddress:  conf.ConsulHost,
		ConsulPort:     conf.ConsulPort,
	})

	if err != nil {
		return fmt.Errorf("connection to Consul could not be made: %v", err.Error())
	} else {
		// Update configuration data from Consul
		if err := consulclient.CheckKeyValuePairs(conf, internal.SupportNotificationsServiceKey, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func setLoggingTarget() string {
	logTarget := Configuration.LoggingRemoteURL
	if !Configuration.EnableRemoteLogging {
		return Configuration.LoggingFile
	}
	return logTarget
}
