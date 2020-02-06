/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	dbInterfaces "github.com/edgexfoundry/edgex-go/internal/pkg/db/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// httpServer defines the contract used to determine whether or not the http httpServer is running.
type httpServer interface {
	IsRunning() bool
}

// Database contains references to dependencies required by the database bootstrap implementation.
type Database struct {
	httpServer httpServer
	database   interfaces.Database
	isCoreData bool
}

// NewDatabase is a factory method that returns an initialized Database receiver struct.
func NewDatabase(httpServer httpServer, database interfaces.Database) Database {
	return Database{
		httpServer: httpServer,
		database:   database,
		isCoreData: false,
	}
}

// NewDatabaseForCoreData is a factory method that returns an initialized Database receiver struct.
func NewDatabaseForCoreData(httpServer httpServer, database interfaces.Database) Database {
	return Database{
		httpServer: httpServer,
		database:   database,
		isCoreData: true,
	}
}

// Return the dbClient interface
func (d Database) newDBClient(
	lc logger.LoggingClient,
	credentials bootstrapConfig.Credentials) (dbInterfaces.DBClient, error) {

	databaseInfo := d.database.GetDatabaseInfo()["Primary"]
	switch databaseInfo.Type {
	case db.MongoDB:
		return mongo.NewClient(
			db.Configuration{
				Host:         databaseInfo.Host,
				Port:         databaseInfo.Port,
				Timeout:      databaseInfo.Timeout,
				DatabaseName: databaseInfo.Name,
				Username:     credentials.Username,
				Password:     credentials.Password,
			})
	case db.RedisDB:
		if d.isCoreData {
			return redis.NewCoreDataClient(
				db.Configuration{
					Host: databaseInfo.Host,
					Port: databaseInfo.Port,
				},
				lc)
		}
		return redis.NewClient(db.Configuration{Host: databaseInfo.Host, Port: databaseInfo.Port}, lc)
	default:
		return nil, db.ErrUnsupportedDatabase
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and initializes the database.
func (d Database) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	// get database credentials.
	var credentials bootstrapConfig.Credentials
	for startupTimer.HasNotElapsed() {
		var err error
		credentials, err = bootstrapContainer.CredentialsProviderFrom(dic.Get).GetDatabaseCredentials(d.database.GetDatabaseInfo()["Primary"])
		if err == nil {
			break
		}
		lc.Warn(fmt.Sprintf("couldn't retrieve database credentials: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	// initialize database.
	var dbClient dbInterfaces.DBClient
	for startupTimer.HasNotElapsed() {
		var err error
		dbClient, err = d.newDBClient(lc, credentials)
		if err == nil {
			break
		}
		dbClient = nil
		lc.Warn(fmt.Sprintf("couldn't create database client: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	if dbClient == nil {
		return false
	}

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClient
		},
	})

	lc.Info("Database connected")
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		for {
			// wait for httpServer to stop running (e.g. handling requests) before closing the database connection.
			if d.httpServer.IsRunning() == false {
				dbClient.CloseSession()
				break
			}
			time.Sleep(time.Second)
		}
		lc.Info("Database disconnected")
	}()

	return true
}
