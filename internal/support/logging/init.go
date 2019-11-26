//
// Copyright (c) 2018 Tencent
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/container"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

var Configuration = &ConfigurationStruct{}
var LoggingClient logger.LoggingClient

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

func getPersistence(credentials config.Credentials) (interfaces.Persistence, error) {
	switch Configuration.Writable.Persistence {
	case interfaces.PersistenceFile:
		return &fileLog{filename: Configuration.Logging.File}, nil
	case interfaces.PersistenceDB:
		// TODO: Integrate db layer with internal/pkg/db/ types so we can support other databases
		ms, err := connectToMongo(credentials)
		if err != nil {
			return nil, err
		}
		return &mongoLog{session: ms}, nil
	default:
		return nil, fmt.Errorf("unrecognized value Configuration.Persistence: %s", Configuration.Writable.Persistence)
	}
}

func (s ServiceInit) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	// update global variables.
	LoggingClient = bootstrapContainer.LoggingClientFrom(dic.Get)

	// get database credentials.
	var credentials config.Credentials
	for startupTimer.HasNotElapsed() {
		var err error
		credentials, err =
			bootstrapContainer.CredentialsProviderFrom(dic.Get).GetDatabaseCredentials(Configuration.Databases["Primary"])
		if err == nil {
			break
		}
		LoggingClient.Warn(fmt.Sprintf("couldn't retrieve database credentials: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	// initialize database.
	var dbClient interfaces.Persistence
	for startupTimer.HasNotElapsed() {
		var err error
		dbClient, err = getPersistence(credentials)
		if err == nil {
			break
		}
		LoggingClient.Warn(fmt.Sprintf("couldn't create database client: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	if dbClient == nil {
		return false
	}

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return newPrivateLogger(dbClient)
		},
		container.PersistenceName: func(get di.Get) interface{} {
			return dbClient
		},
	})

	LoggingClient.Info("Database connected")
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		for {
			// wait for httpServer to stop running (e.g. handling requests) before closing the database connection.
			if s.server.IsRunning() == false {
				LoggingClient.Info("Database disconnecting")
				dbClient.CloseSession()
				break
			}
			time.Sleep(time.Second)
		}
	}()

	return true
}
