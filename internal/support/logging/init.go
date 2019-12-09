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
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/logging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	types "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/config"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/container"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/logger/file"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/logger/mongo"
)

const (
	PersistenceDB   = "database"
	PersistenceFile = "file"
)

type server interface {
	IsRunning() bool
}

type ServiceInit struct {
	server     server
	serviceKey string
}

func NewServiceInit(server server, serviceKey string) ServiceInit {
	return ServiceInit{
		server:     server,
		serviceKey: serviceKey,
	}
}

func getPersistence(
	credentials *types.Credentials,
	configuration *config.ConfigurationStruct) (interfaces.Logger, error) {

	switch configuration.Writable.Persistence {
	case PersistenceFile:
		return file.NewLogger(configuration.Logging.File), nil
	case PersistenceDB:
		return mongo.NewLogger(credentials, configuration)
	default:
		return nil, fmt.Errorf("unrecognized value Configuration.Logger: %s", configuration.Writable.Persistence)
	}
}

func (s ServiceInit) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	loggingClient := logging.FactoryToStdout(s.serviceKey)
	configuration := container.ConfigurationFrom(dic.Get)

	// get database credentials.
	credentialsProvider := bootstrapContainer.CredentialsProviderFrom(dic.Get)
	var credentials types.Credentials
	for startupTimer.HasNotElapsed() {
		var err error
		credentials, err = credentialsProvider.GetDatabaseCredentials(configuration.Databases["Primary"])
		if err == nil {
			break
		}
		loggingClient.Warn(fmt.Sprintf("couldn't retrieve database credentials: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	// initialize database.
	var dbClient interfaces.Logger
	var err error
	for startupTimer.HasNotElapsed() {
		dbClient, err = getPersistence(&credentials, configuration)
		if err == nil {
			break
		}
		loggingClient.Warn(fmt.Sprintf("couldn't create database client: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	if err != nil {
		return false
	}

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return loggingClient
		},
		container.LoggerInterfaceName: func(get di.Get) interface{} {
			return dbClient
		},
	})

	loggingClient.Info("Database connected")
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		for {
			// wait for httpServer to stop running (e.g. handling requests) before closing the database connection.
			if s.server.IsRunning() == false {
				loggingClient.Info("Database disconnecting")
				dbClient.CloseSession()
				break
			}
			time.Sleep(time.Second)
		}
	}()

	return true
}
