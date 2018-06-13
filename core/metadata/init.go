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
package metadata

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/edgex-go/core/db"
	"github.com/edgexfoundry/edgex-go/core/db/memory"
	"github.com/edgexfoundry/edgex-go/core/db/mongo"
	"github.com/edgexfoundry/edgex-go/core/domain/enums"
	"github.com/edgexfoundry/edgex-go/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal"
	consulclient "github.com/edgexfoundry/edgex-go/support/consul-client"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	notifications "github.com/edgexfoundry/edgex-go/support/notifications-client"
)

var loggingClient logger.LoggingClient

func ConnectToConsul(conf ConfigurationStruct) error {
	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    internal.CoreMetaDataServiceKey,
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
		if err := consulclient.CheckKeyValuePairs(&conf, internal.CoreMetaDataServiceKey, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func getDatabase(dbType string, config db.Configuration) (interfaces.DBClient, error) {
	switch dbType {
	case enums.MongoStr:
		return mongo.NewClient(config), nil
	case enums.MemoryStr:
		return &memory.MemDB{}, nil
	}
	return nil, db.ErrNotFound
}

func Init(conf ConfigurationStruct, l logger.LoggingClient) error {
	loggingClient = l
	configuration = conf
	//TODO: The above two are set due to global scope throughout the package. How can this be eliminated / refactored?

	// Initialize notificationsClient based on configuration
	notifications.SetConfiguration(configuration.SupportNotificationsHost, configuration.SupportNotificationsPort)

	var err error
	dbConfig := db.Configuration{
		Host:         configuration.MongoDBHost,
		Port:         configuration.MongoDBPort,
		Timeout:      0,
		DatabaseName: configuration.MongoDatabaseName,
		Username:     configuration.MongoDBUserName,
		Password:     configuration.MongoDBPassword,
	}
	// Create database client
	dbClient, err = getDatabase(configuration.DBType, dbConfig)
	if err != nil {
		return err
	}

	// Connect to the database
	return dbClient.Connect()
}

func Destruct() {
	if dbClient != nil {
		dbClient.CloseSession()
		dbClient = nil
	}
}
