/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

var Configuration = &ConfigurationStruct{}
var LoggingClient logger.LoggingClient
var dbClient interfaces.DBClient
var scClient interfaces.SchedulerQueueClient
var ticker *time.Ticker

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
		return redis.NewClient(dbConfig, LoggingClient) //TODO: Verify this also connects to Redis
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

	scClient = NewSchedulerQueueClient()

	// Initialize the ticker time
	if err := LoadScheduler(); err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to load schedules and events %s", err.Error()))
		return false
	}

	ticker = time.NewTicker(time.Duration(Configuration.Writable.ScheduleIntervalTime) * time.Millisecond)
	StartTicker()

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		StopTicker()
	}()

	return true
}
