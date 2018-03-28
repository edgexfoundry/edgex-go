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
 *
 * @microservice: core-metadata-go service
 * @author: Spencer Bull & Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package metadata

import (
	"fmt"
	"strconv"
	"strings"

	enums "github.com/edgexfoundry/edgex-go/core/domain/enums"
	consulclient "github.com/edgexfoundry/edgex-go/support/consul-client"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	notifications "github.com/edgexfoundry/edgex-go/support/notifications-client"
)

// DS : DataStore to retrieve data from database.
var DS DataStore
var loggingClient logger.LoggingClient
var notificationsClient = notifications.NotificationsClient{}

func ConnectToConsul(conf ConfigurationStruct) error {
	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    conf.ServiceName,
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
		if err := consulclient.CheckKeyValuePairs(&conf, conf.ApplicationName, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func Init(conf ConfigurationStruct, l logger.LoggingClient) error {
	loggingClient = l
	configuration = conf
	//TODO: The above two are set due to global scope throughout the package. How can this be eliminated / refactored?

	// Update Service CONSTANTS
	MONGODATABASE = configuration.MongoDatabaseName
	PROTOCOL = configuration.Protocol
	SERVERPORT = strconv.Itoa(configuration.ServicePort)
	DBTYPE = configuration.DBType
	DOCKERMONGO = configuration.MongoDBHost + ":" + strconv.Itoa(configuration.MongoDBPort)
	DBUSER = configuration.MongoDBUserName
	DBPASS = configuration.MongoDBPassword

	// Update notificationsClient based on configuration
	notificationsClient.RemoteUrl = configuration.SupportNotificationsNotificationURL

	var err error
	// Connect to the database
	DATABASE, err = enums.GetDatabaseType(DBTYPE)
	if err != nil {
		return err
	}
	if !dbConnect() {
		return err
	}

	return nil
}
