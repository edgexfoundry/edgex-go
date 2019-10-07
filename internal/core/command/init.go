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
package command

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-registry/registry"

	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces"
	bootstrap "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
)

var Configuration = &ConfigurationStruct{}
var LoggingClient logger.LoggingClient
var mdc metadata.DeviceClient
var dbClient interfaces.DBClient

type server interface {
	IsRunning() bool
}

// ServiceInit encapsulates information needed for the service to startup.
type ServiceInit struct {
	server server
}

// NewServiceInit creates a new ServiceInit
func NewServiceInit(server server) ServiceInit {
	return ServiceInit{
		server: server,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs and startup actions needed by the service.
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
	s.initializeClients(registry)

	// initialize database.
	for startupTimer.HasNotElapsed() {
		var err error
		dbClient, err = s.newDBClient(Configuration.Databases["Primary"].Type)
		if err == nil {
			break
		}
		dbClient = nil
		LoggingClient.Warn(fmt.Sprintf("couldn't create database client: %s", err.Error()))
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

	return true
}

// initializeClients initializes the clients needed by the service.
func (s ServiceInit) initializeClients(registry registry.Client) {
	// Create metadata clients
	params := types.EndpointParams{
		ServiceKey:  clients.CoreMetaDataServiceKey,
		Path:        clients.ApiDeviceRoute,
		UseRegistry: registry != nil,
		Url:         Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
		Interval:    Configuration.Service.ClientMonitor,
	}

	mdc = metadata.NewDeviceClient(params, endpoint.Endpoint{RegistryClient: &registry})
}

// newDBClient creates a DBClient based on the underlying database.
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
		return redis.NewClient(dbConfig, LoggingClient)
	default:
		return nil, db.ErrUnsupportedDatabase
	}
}
