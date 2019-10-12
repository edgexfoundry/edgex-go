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
package metadata

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

// Global variables
var Configuration = &ConfigurationStruct{}
var dbClient interfaces.DBClient
var LoggingClient logger.LoggingClient
var nc notifications.NotificationsClient
var vdc coredata.ValueDescriptorClient

// Global ErrorConcept variables
var httpErrorHandler errorconcept.ErrorHandler

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

func (s ServiceInit) initializeClients(useRegistry bool, registryClient registry.Client) {
	// Create notification client
	nParams := types.EndpointParams{
		ServiceKey:  clients.SupportNotificationsServiceKey,
		Path:        clients.ApiNotificationRoute,
		UseRegistry: useRegistry,
		Url:         Configuration.Clients["Notifications"].Url() + clients.ApiNotificationRoute,
		Interval:    Configuration.Service.ClientMonitor,
	}
	nc = notifications.NewNotificationsClient(nParams, endpoint.Endpoint{RegistryClient: &registryClient})

	vParams := types.EndpointParams{
		ServiceKey:  clients.CoreDataServiceKey,
		Path:        clients.ApiValueDescriptorRoute,
		UseRegistry: useRegistry,
		Url:         Configuration.Clients["CoreData"].Url() + clients.ApiValueDescriptorRoute,
		Interval:    Configuration.Service.ClientMonitor,
	}
	vdc = coredata.NewValueDescriptorClient(vParams, endpoint.Endpoint{RegistryClient: &registryClient})
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
	dic *di.Container) bool {

	// update global variables.
	LoggingClient = container.LoggingClientFrom(dic.Get)
	httpErrorHandler = errorconcept.NewErrorHandler(LoggingClient)

	// initialize clients required by service.
	registryClient := container.RegistryFrom(dic.Get)
	s.initializeClients(registryClient != nil, registryClient)

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

	return true
}
