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
 * @microservice: core-data-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package data

import (
	"strings"
	"time"

	"github.com/tsconn23/edgex-go/core/clients/metadataclients"
	"github.com/tsconn23/edgex-go/core/data/clients"
	"github.com/tsconn23/edgex-go/core/data/messaging"
	consulclient "github.com/tsconn23/edgex-go/support/consul-client"
	"github.com/tsconn23/edgex-go/support/logging-client"
)

// Global variables
var dbc clients.DBClient
var loggingClient logger.LoggingClient
var ep *messaging.EventPublisher
var mdc metadataclients.DeviceClient
var msc metadataclients.ServiceClient

// Heartbeat for the data microservice - send a message to logging service
func heartbeat() {
	// Loop forever
	for true {
		loggingClient.Info(configuration.Heartbeatmsg, "")
		time.Sleep(time.Millisecond * time.Duration(configuration.Heartbeattime)) // Sleep based on configuration
	}
}

func Init(conf ConfigurationStruct, logger logger.LoggingClient) error {
	loggingClient = logger
	configuration = conf

	var err error

	// Initialize service on Consul
	err = consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    configuration.Servicename,
		ServicePort:    configuration.Serverport,
		ServiceAddress: configuration.Serviceaddress,
		CheckAddress:   configuration.Consulcheckaddress,
		CheckInterval:  configuration.Checkinterval,
		ConsulAddress:  configuration.Consulhost,
		ConsulPort:     configuration.Consulport,
	})

	if err != nil {
		loggingClient.Error("Connection to Consul could not be made: "+err.Error(), "")
	} else {
		// Update configuration data from Consul
		if err := consulclient.CheckKeyValuePairs(&configuration, configuration.Servicename, strings.Split(configuration.Consulprofilesactive, ";")); err != nil {
			loggingClient.Error("Error getting key/values from Consul: "+err.Error(), "")
		}
	}
	
	// Create a database client
	dbc, err = clients.NewDBClient(clients.DBConfiguration{
		DbType:       clients.MONGO,
		Host:         configuration.Datamongodbhost,
		Port:         configuration.Datamongodbport,
		Timeout:      configuration.DatamongodbsocketTimeout,
		DatabaseName: configuration.Datamongodbdatabase,
		Username:     configuration.Datamongodbusername,
		Password:     configuration.Datamongodbpassword,
	})
	if err != nil {
		loggingClient.Error("Couldn't connect to database: "+err.Error(), "")
		return err
	}

	// Create metadata clients
	mdc = metadataclients.NewDeviceClient(configuration.Metadbdeviceurl)
	msc = metadataclients.NewServiceClient(configuration.Metadbdeviceserviceurl)

	// Create the event publisher
	ep = messaging.NewZeroMQPublisher(messaging.ZeroMQConfiguration{
		AddressPort: configuration.Zeromqaddressport,
	})

	// Start heartbeat
	go heartbeat()
	return nil
}
