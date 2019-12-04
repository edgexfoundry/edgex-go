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

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	types "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/config"
)

var Configuration = &config.ConfigurationStruct{}
var dbClient persistence

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

func getPersistence(credentials types.Credentials) (persistence, error) {
	switch Configuration.Writable.Persistence {
	case PersistenceFile:
		return &fileLog{filename: Configuration.Logging.File}, nil
	case PersistenceDB:
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

	dic.Update(di.ServiceConstructorMap{
		container.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return newPrivateLogger()
		},
	})

	loggingClient := container.LoggingClientFrom(dic.Get)

	// get database credentials.
	var credentials types.Credentials
	for startupTimer.HasNotElapsed() {
		var err error
		credentials, err = container.CredentialsProviderFrom(dic.Get).GetDatabaseCredentials(Configuration.Databases["Primary"])
		if err == nil {
			break
		}
		loggingClient.Warn(fmt.Sprintf("couldn't retrieve database credentials: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	// initialize database.
	for startupTimer.HasNotElapsed() {
		var err error
		dbClient, err = getPersistence(credentials)
		if err == nil {
			break
		}
		loggingClient.Warn(fmt.Sprintf("couldn't create database client: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	if dbClient == nil {
		return false
	}

	loggingClient.Info("Database connected")
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		for {
			// wait for httpServer to stop running (e.g. handling requests) before closing the database connection.
			if s.server.IsRunning() == false {
				loggingClient.Info("Database disconnecting")
				dbClient.closeSession()
				break
			}
			time.Sleep(time.Second)
		}
	}()

	return true
}
