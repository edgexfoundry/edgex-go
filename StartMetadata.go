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
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	consulclient "github.com/edgexfoundry/consul-client-go"
	enums "github.com/edgexfoundry/core-domain-go/enums"
	logger "github.com/edgexfoundry/support-logging-client-go"
	notifications "github.com/edgexfoundry/support-notifications-client-go"
)

// DS : DataStore to retrieve data from database.
var DS DataStore
var notificationsClient = notifications.NotificationsClient{}
var loggingClient logger.LoggingClient

func main() {
	start := time.Now()

	// Load configuration data
	err := readConfigurationFile(CONFIG)
	if err != nil {
		fmt.Printf("Could not read configuration file(%s): %#v\n", CONFIG, err)
		os.Exit(1)
	}

	// Initialize service on Consul
	err = consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    configuration.ServiceName,
		ServicePort:    configuration.ServerPort,
		ServiceAddress: configuration.ServiceAddress,
		CheckAddress:   configuration.ConsulCheckAddress,
		CheckInterval:  configuration.CheckInterval,
		ConsulAddress:  configuration.ConsulHost,
		ConsulPort:     configuration.ConsulPort,
	})
	if err != nil {
		fmt.Print("Connection to Consul could not be make: " + err.Error())
	}
	loggingClient = logger.NewClient(configuration.ApplicationName, "")

	// Update configuration data from Consul
	consulclient.CheckKeyValuePairs(&configuration, configuration.ApplicationName, strings.Split(configuration.ConsulProfilesActive, ";"))
	// Update Service CONSTANTS
	MONGODATABASE = configuration.MongoDatabaseName
	PROTOCOL = configuration.Protocol
	SERVERPORT = strconv.Itoa(configuration.ServerPort)
	DBTYPE = configuration.DBType
	DOCKERMONGO = configuration.MongoDBHost + ":" + strconv.Itoa(configuration.MongoDBPort)
	DBUSER = configuration.MongoDBUserName
	DBPASS = configuration.MongoDBPassword

	// Update logging based on configuration
	loggingClient.RemoteUrl = configuration.LoggingRemoteURL
	loggingClient.LogFilePath = configuration.LoggingFile

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

	// Start heartbeat
	go heartbeat()

	if strings.Compare(PROTOCOL, REST) == 0 {
		r := loadRestRoutes()
		http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration.ServerTimeout), "Request timed out")
		loggingClient.Info(configuration.AppOpenMsg, "")
		fmt.Println("Listening on port: " + SERVERPORT)

		// Time it took to start service
		loggingClient.Info("Service started in: "+time.Since(start).String(), "")

		loggingClient.Error(http.ListenAndServe(":"+SERVERPORT, r).Error())
	}

}

// Heartbeat for the metadata microservice - send a message to logging service
func heartbeat() {
	// Loop forever
	for true {
		loggingClient.Info(configuration.HeartBeatMsg, "")
		time.Sleep(time.Millisecond * time.Duration(configuration.HeartBeatTime)) // Sleep based on configuration
	}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// Read the configuration file and
func readConfigurationFile(path string) error {
	// Read the configuration file
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Decode the configuration as JSON
	err = json.Unmarshal(contents, &configuration)
	if err != nil {
		return err
	}

	return nil
}
