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
	"net/http"
	"strconv"
	"strings"
	"time"

	enums "github.com/edgexfoundry/edgex-go/core/domain/enums"
	consulclient "github.com/edgexfoundry/edgex-go/support/consul-client"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	notifications "github.com/edgexfoundry/edgex-go/support/notifications-client"
)

// DS : DataStore to retrieve data from database.
var DS DataStore
var loggingClient logger.LoggingClient
var notificationsClient = notifications.NotificationsClient{}


func Init(conf ConfigurationStruct, l logger.LoggingClient) {
	loggingClient = l
	configuration = conf

	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    configuration.ServiceName,
		ServicePort:    configuration.ServerPort,
		ServiceAddress: configuration.ServiceAddress,
		CheckAddress:   configuration.ConsulCheckAddress,
		CheckInterval:  configuration.CheckInterval,
		ConsulAddress:  configuration.ConsulHost,
		ConsulPort:     configuration.ConsulPort,
	})
	if err != nil {
		loggingClient.Error("Connection to Consul could not be make: " + err.Error())
	} else {
		// Update configuration data from Consul
		consulclient.CheckKeyValuePairs(&configuration, configuration.ApplicationName, strings.Split(configuration.ConsulProfilesActive, ";"))
	}

	// Update Service CONSTANTS
	MONGODATABASE = configuration.MongoDatabaseName
	PROTOCOL = configuration.Protocol
	SERVERPORT = strconv.Itoa(configuration.ServerPort)
	DBTYPE = configuration.DBType
	DOCKERMONGO = configuration.MongoDBHost + ":" + strconv.Itoa(configuration.MongoDBPort)
	DBUSER = configuration.MongoDBUserName
	DBPASS = configuration.MongoDBPassword

	// Update notificationsClient based on configuration
	notificationsClient.RemoteUrl = configuration.SupportNotificationsNotificationURL

	// Connect to the database
	DATABASE, err = enums.GetDatabaseType(DBTYPE)
	if err != nil {
		loggingClient.Error(err.Error())
		return
	}
	if !dbConnect() {
		loggingClient.Error("Error connecting to Database")
		return
	}

	if strings.Compare(PROTOCOL, REST) == 0 {
		r := loadRestRoutes()
		http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration.ServerTimeout), "Request timed out")
		loggingClient.Info(configuration.AppOpenMsg, "")
		fmt.Println("Listening on port: " + SERVERPORT)

		loggingClient.Error(http.ListenAndServe(":"+SERVERPORT, r).Error())
	}

}
