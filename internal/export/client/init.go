//
// Copyright (c) 2018 Tencent
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/export"
	bootstrap "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/export/distro"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

// Global variables
var dbClient export.DBClient
var LoggingClient logger.LoggingClient
var Configuration = &ConfigurationStruct{}
var dc distro.DistroClient

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
	// Create export-distro client
	params := types.EndpointParams{
		ServiceKey:  clients.ExportDistroServiceKey,
		UseRegistry: useRegistry,
		Url:         Configuration.Clients["Distro"].Url(),
		Interval:    Configuration.Service.ClientMonitor,
	}

	dc = distro.NewDistroClient(params, endpoint.Endpoint{RegistryClient: &registry})
}

// Return the dbClient interface
func (s ServiceInit) newDBClient(dbType string) (export.DBClient, error) {
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

	return true
}
