/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright (c) 2019 Intel Corporation
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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	bootstrap "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

// Global variables
var Configuration = &ConfigurationStruct{}
var dbClient interfaces.DBClient
var LoggingClient logger.LoggingClient

// TODO: Refactor names in separate PR: See comments on PR #1133
var chEvents chan interface{} // A channel for "domain events" sourced from event operations
var msgClient messaging.MessageClient
var mdc metadata.DeviceClient
var msc metadata.DeviceServiceClient

type server interface {
	IsRunning() bool
}

type ServiceInit struct {
	server server
}

func NewServiceInit(server server) ServiceInit {
	return ServiceInit{
		server: server,
	}
}

func (s ServiceInit) initializeClients(useRegistry bool, registry registry.Client) {
	// Create metadata clients
	mdc = metadata.NewDeviceClient(
		types.EndpointParams{
			ServiceKey:  clients.CoreMetaDataServiceKey,
			Path:        clients.ApiDeviceRoute,
			UseRegistry: useRegistry,
			Url:         Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
			Interval:    Configuration.Service.ClientMonitor,
		},
		endpoint.Endpoint{RegistryClient: &registry})

	msc = metadata.NewDeviceServiceClient(
		types.EndpointParams{
			ServiceKey:  clients.CoreMetaDataServiceKey,
			Path:        clients.ApiDeviceServiceRoute,
			UseRegistry: useRegistry,
			Url:         Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
			Interval:    Configuration.Service.ClientMonitor,
		},
		endpoint.Endpoint{RegistryClient: &registry})

	// Create the messaging client
	var err error
	msgClient, err = messaging.NewMessageClient(
		msgTypes.MessageBusConfig{
			PublishHost: msgTypes.HostInfo{
				Host:     Configuration.MessageQueue.Host,
				Port:     Configuration.MessageQueue.Port,
				Protocol: Configuration.MessageQueue.Protocol,
			},
			Type: Configuration.MessageQueue.Type,
		})

	if err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to create messaging client: %s", err.Error()))
	}
}

// Return the dbClient interface
func (s ServiceInit) newDBClient(dbType string) (interfaces.DBClient, error) {
	switch dbType {
	case db.MongoDB:
		dbConfig := db.Configuration{
			Host:         Configuration.Databases["Primary"].Host,
			Port:         Configuration.Databases["Primary"].Port,
			Timeout:      Configuration.Databases["Primary"].Timeout,
			DatabaseName: Configuration.Databases["Primary"].Name,
			Username:     Configuration.Databases["Primary"].Username,
			Password:     Configuration.Databases["Primary"].Password,
		}
		return mongo.NewClient(dbConfig)
	case db.RedisDB:
		dbConfig := db.Configuration{
			Host: Configuration.Databases["Primary"].Host,
			Port: Configuration.Databases["Primary"].Port,
		}
		redisClient, err := redis.NewCoreDataClient(dbConfig, LoggingClient) // TODO: Verify this also connects to Redis
		if err != nil {
			return nil, err
		}

		return redisClient, nil
	default:
		return nil, db.ErrUnsupportedDatabase
	}
}

func (s ServiceInit) BootstrapHandler(
	wg *sync.WaitGroup,
	ctx context.Context,
	startupTimer startup.Timer,
	config bootstrap.Configuration,
	logging logger.LoggingClient,
	registry registry.Client) bool {

	// update global variables.
	LoggingClient = logging

	// initialize clients required by service.
	s.initializeClients(registry != nil, registry)

	// initialize database.
	for startupTimer.HasNotElapsed() {
		var err error
		dbClient, err = s.newDBClient(Configuration.Databases["Primary"].Type)
		if err == nil {
			break
		}
		dbClient = nil
		LoggingClient.Warn(fmt.Sprintf("couldn't create database client: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	if dbClient == nil {
		return false
	}

	LoggingClient.Info("Database connected")
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		for {
			// wait for httpServer to stop running (e.g. handling requests) before closing the database connection.
			if s.server.IsRunning() == false {
				dbClient.CloseSession()
				break
			}
			time.Sleep(time.Second)
		}
		LoggingClient.Info("Database disconnected")
	}()

	// initialize event handlers
	chEvents = make(chan interface{}, 100)
	initEventHandlers()

	return true
}
