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
	"fmt"
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

func ConnectToConsul(conf ConfigurationStruct) error {
	var err error

	// Initialize service on Consul
	err = consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    conf.Servicename,
		ServicePort:    conf.Serverport,
		ServiceAddress: conf.Serviceaddress,
		CheckAddress:   conf.Consulcheckaddress,
		CheckInterval:  conf.Checkinterval,
		ConsulAddress:  conf.Consulhost,
		ConsulPort:     conf.Consulport,
	})

	if err != nil {
		return fmt.Errorf("Connection to Consul could not be made: %v", err.Error())
	} else {
		// Update configuration data from Consul
		if err := consulclient.CheckKeyValuePairs(&configuration, configuration.Servicename, strings.Split(configuration.Consulprofilesactive, ";")); err != nil {
			return fmt.Errorf("Error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func Init(conf ConfigurationStruct) error {

	var err error
	
	// Create a database client
	dbc, err = clients.NewDBClient(clients.DBConfiguration{
		DbType:       clients.MONGO,
		Host:         conf.Datamongodbhost,
		Port:         conf.Datamongodbport,
		Timeout:      conf.DatamongodbsocketTimeout,
		DatabaseName: conf.Datamongodbdatabase,
		Username:     conf.Datamongodbusername,
		Password:     conf.Datamongodbpassword,
	})
	if err != nil {
		return fmt.Errorf("Couldn't connect to database:  %v", err.Error())
	}

	// Create metadata clients
	mdc = metadataclients.NewDeviceClient(conf.Metadbdeviceurl)
	msc = metadataclients.NewServiceClient(conf.Metadbdeviceserviceurl)

	// Create the event publisher
	ep = messaging.NewZeroMQPublisher(messaging.ZeroMQConfiguration{
		AddressPort: conf.Zeromqaddressport,
	})

	// Start heartbeat
	go heartbeat()
	return nil
}
