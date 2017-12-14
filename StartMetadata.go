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
	"strings"
	"time"

	consulclient "github.com/edgexfoundry/consul-client-go"
	notifications "github.com/edgexfoundry/support-notifications-client-go"
	logger "github.com/edgexfoundry/support-logging-client-go"
)

// DS : DataStore to retrieve data from database.
var DS DataStore
var notificationsClient = notifications.NotificationsClient{}
var loggingClient logger.LoggingClient

func main() {
	start := time.Now()

	// Load configuration data
	readConfigurationFile(CONFIG)

	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    configuration.ServiceName,
		ServicePort:    configuration.ServerPort,
		ServiceAddress: configuration.ServiceAddress,
		CheckAddress:   configuration.Consulcheckaddress,
		CheckInterval:  configuration.CheckInterval,
		ConsulAddress:  configuration.Consulhost,
		ConsulPort:     configuration.ConsulPort,
	})
	if err != nil {
		fmt.Print("Connection to Consul could not be make: " + err.Error())
	}
	loggingClient = logger.NewClient(configuration.ApplicationName, "")

	// Update configuration data from Consul
	consulclient.CheckKeyValuePairs(&configuration, configuration.ApplicationName, strings.Split(configuration.ConsulProfilesActive, ";"))
	// Update Service CONSTANTS
	DATABASE = configuration.MongoDBName
	PROTOCOL = configuration.Protocol
	SERVERPORT = string(configuration.ServerPort)
	DOCKERMONGO = configuration.MongoDBHost + ":" + string(configuration.MongoDBPort)
	DBUSER = configuration.MongoDBUserName
	DBPASS = configuration.MongoDBPassword

	// Update logging based on configuration
	loggingClient.RemoteUrl = configuration.LoggingRemoteURL
	loggingClient.LogFilePath = configuration.LoggingFile

	// Update notificationsClient based on configuration
	notificationsClient.RemoteUrl = configuration.SupportNotificationsNotificationURL

	// Connect to the database
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
		loggingClient.Error(err.Error(), "")
		return err
	}

	// Decode the configuration as JSON
	err = json.Unmarshal(contents, &configuration)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		return err
	}

	return nil
}
