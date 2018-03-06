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
package command

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	consulclient "github.com/tsconn23/edgex-go/support/consul-client"
	logger "github.com/tsconn23/edgex-go/support/logging-client"
)

var loggingClient logger.LoggingClient

func heartbeat() {
	for true {
		loggingClient.Info(configuration.HeartBeatMessage, "")
		time.Sleep(time.Millisecond * time.Duration(configuration.HeartBeatTime))
	}
}

func Start(conf ConfigurationStruct, l logger.LoggingClient) {
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

	if err == nil { // Update configuration data from Consul
		if err := consulclient.CheckKeyValuePairs(&configuration, configuration.ApplicationName, strings.Split(configuration.ConsulProfilesActive, ";")); err != nil {
			loggingClient.Error("Error getting key/values from Consul: "+err.Error(), "")
		}
	} else {
		loggingClient.Error("Connection to Consul could not be made: "+err.Error(), "")
	}

	// Start heartbeat
	go heartbeat()

	if strings.Compare(configuration.URLProtocol, REST_HTTP) == 0 {
		r := loadRestRoutes()
		http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration.ServiceTimeout), "Request timed out")
		loggingClient.Info(configuration.AppOpenMessage, "")
		loggingClient.Info("Listening on port: "+strconv.Itoa(configuration.ServerPort), "")

		loggingClient.Error(http.ListenAndServe(":"+strconv.Itoa(configuration.ServerPort), r).Error())
	}
}
