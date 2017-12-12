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
 * @microservice: core-command-go service
 * @author: Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	consulclient "github.com/edgexfoundry/consul-client-go"
	logger "github.com/edgexfoundry/support-logging-client-go"
)

var loggingClient = logger.NewClient(SERVICENAME, "")

func main() {
	start := time.Now()

	// Load configuration data
	readConfigurationFile(CONFIG)

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
	// Update configuration data from Consul
	consulclient.CheckKeyValuePairs(&configuration)

	// Setup Logging
	loggingClient.RemoteUrl = configuration.LoggingRemoteURL
	loggingClient.LogFilePath = configuration.LogFile

	if err != nil {
		loggingClient.Error("Connection to Consul could not be made: "+err.Error(), "")
		return
	}

	// Start heartbeat
	go heartbeat()

	if strings.Compare(configuration.URLProtocol, REST_HTTP) == 0 {
		r := loadRestRoutes()
		http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration.ServiceTimeout), "Request timed out")
		loggingClient.Info(configuration.AppOpenMessage, "")
		loggingClient.Info("Listening on port: "+strconv.Itoa(configuration.ServerPort), "")

		// Time it took to start service
		loggingClient.Info("Service started in: "+time.Since(start).String(), "")

		loggingClient.Error(http.ListenAndServe(":"+strconv.Itoa(configuration.ServerPort), r).Error())
	}

}

func heartbeat() {
	for true {
		loggingClient.Info(configuration.HeartBeatMessage, "")
		time.Sleep(time.Millisecond * time.Duration(configuration.HeartBeatTime))
	}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

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
