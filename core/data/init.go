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
package data

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/edgex-go/core/clients/metadata"
	"github.com/edgexfoundry/edgex-go/core/clients/types"
	"github.com/edgexfoundry/edgex-go/core/data/clients"
	"github.com/edgexfoundry/edgex-go/core/data/messaging"
	"github.com/edgexfoundry/edgex-go/core/db"
	"github.com/edgexfoundry/edgex-go/core/db/influx"
	"github.com/edgexfoundry/edgex-go/core/db/memory"
	"github.com/edgexfoundry/edgex-go/core/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal"
	consulclient "github.com/edgexfoundry/edgex-go/support/consul-client"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
)

// Global variables
var dbc clients.DBClient
var loggingClient logger.LoggingClient
var ep *messaging.EventPublisher
var mdc metadata.DeviceClient
var msc metadata.DeviceServiceClient

func ConnectToConsul(conf ConfigurationStruct) error {

	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    internal.CoreDataServiceKey,
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
		if err := consulclient.CheckKeyValuePairs(&conf, internal.CoreDataServiceKey, strings.Split(conf.ConsulProfilesActive, ";")); err != nil {
			return fmt.Errorf("error getting key/values from Consul: %v", err.Error())
		}
	}
	return nil
}

// Return the dbClient interface
func newDBClient(dbType clients.DatabaseType, config db.Configuration) (clients.DBClient, error) {
	switch dbType {
	case clients.MONGO:
		// Create the mongo client
		mc, err := mongo.NewClient(config)
		if err != nil {
			loggingClient.Error("Error creating the mongo client: " + err.Error())
			return nil, err
		}
		return mc, nil
	case clients.INFLUX:
		// Create the influx client
		ic, err := influx.NewClient(config)
		if err != nil {
			loggingClient.Error("Error creating the influx client: " + err.Error())
			return nil, err
		}
		return ic, nil
	case clients.MEMORY:
		// Create the memory client
		mem := &memory.MemDB{}
		return mem, nil
	default:
		return nil, db.ErrUnsupportedDatabase
	}
}

func Init(conf ConfigurationStruct, l logger.LoggingClient, useConsul bool) error {
	loggingClient = l
	configuration = conf
	//TODO: The above two are set due to global scope throughout the package. How can this be eliminated / refactored?

	var err error

	// Create a database client
	dbc, err = newDBClient(clients.MONGO, db.Configuration{
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
	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        conf.MetaDevicePath,
		UseRegistry: useConsul,
		Url:         conf.MetaDeviceURL}

	mdc = metadata.NewDeviceClient(params, types.Endpoint{})

	params.Path = conf.MetaDeviceServicePath
	msc = metadata.NewDeviceServiceClient(params, types.Endpoint{})

	// Create the event publisher
	ep = messaging.NewZeroMQPublisher(messaging.ZeroMQConfiguration{
		AddressPort: conf.ZeroMQAddressPort,
	})

	return nil
}

func Destruct() {
	dbc.CloseSession()
}
