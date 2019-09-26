//
// Copyright (c) 2018 Tencent
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	bootstrap "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

var Configuration = &ConfigurationStruct{}
var dbClient persistence
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

func getPersistence() (persistence, error) {
	switch Configuration.Writable.Persistence {
	case PersistenceFile:
		return &fileLog{filename: Configuration.Logging.File}, nil
	case PersistenceDB:
		// TODO: Integrate db layer with internal/pkg/db/ types so we can support other databases
		ms, err := connectToMongo()
		if err != nil {
			return nil, err
		}
		return &mongoLog{session: ms}, nil
	default:
		return nil, errors.New(fmt.Sprintf("unrecognized value Configuration.Persistence: %s", Configuration.Writable.Persistence))
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
	LoggingClient = newPrivateLogger()

	// initialize database.
	for startupTimer.HasNotElapsed() {
		var err error
		dbClient, err = getPersistence()
		if err == nil {
			break
		}
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
				LoggingClient.Info("Database disconnecting")
				dbClient.closeSession()
				break
			}
			time.Sleep(time.Second)
		}
	}()

	return true
}
