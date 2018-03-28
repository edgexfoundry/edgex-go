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
	"fmt"
	"strings"

	"github.com/edgexfoundry/edgex-go/core/clients/metadataclients"
	"github.com/edgexfoundry/edgex-go/core/data/clients"
	"github.com/edgexfoundry/edgex-go/core/data/messaging"
	consulclient "github.com/edgexfoundry/edgex-go/support/consul-client"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
)

// Global variables
var dbc clients.DBClient
var loggingClient logger.LoggingClient
var ep *messaging.EventPublisher
var mdc metadataclients.DeviceClient
var msc metadataclients.ServiceClient

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
		if err := consulclient.CheckKeyValuePairs(&conf, conf.ServiceName, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

func Init(conf ConfigurationStruct, l logger.LoggingClient) error {
	loggingClient = l
	configuration = conf
	//TODO: The above two are set due to global scope throughout the package. How can this be eliminated / refactored?

	var err error
	
	// Create a database client
	dbc, err = clients.NewDBClient(clients.DBConfiguration{
		DbType:       clients.MONGO,
		Host:         conf.MongoDBHost,
		Port:         conf.MongoDBPort,
		Timeout:      conf.MongoDBConnectTimeout,
		DatabaseName: conf.MongoDatabaseName,
		Username:     conf.MongoDBUserName,
		Password:     conf.MongoDBPassword,
	})
	if err != nil {
		return fmt.Errorf("couldn't connect to database: %v", err.Error())
	}

	// Create metadata clients
	mdc = metadataclients.NewDeviceClient(conf.MetaDeviceURL)
	msc = metadataclients.NewServiceClient(conf.MetaDeviceServiceURL)

	// Create the event publisher
	ep = messaging.NewZeroMQPublisher(messaging.ZeroMQConfiguration{
		AddressPort: conf.ZeroMQAddressPort,
	})

	return nil
}
