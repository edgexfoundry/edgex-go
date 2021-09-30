//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"fmt"
	"sync"
	"time"

	bootstrapInterfaces "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/redis"
	"github.com/edgexfoundry/edgex-go/internal/pkg/interfaces"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

// httpServer defines the contract used to determine whether or not the http httpServer is running.
type httpServer interface {
	IsRunning() bool
}

// Database contains references to dependencies required by the database bootstrap implementation.
type Database struct {
	httpServer            httpServer
	database              bootstrapInterfaces.Database
	dBClientInterfaceName string
}

// NewDatabase is a factory method that returns an initialized Database receiver struct.
func NewDatabase(httpServer httpServer, database bootstrapInterfaces.Database, dBClientInterfaceName string) Database {
	return Database{
		httpServer:            httpServer,
		database:              database,
		dBClientInterfaceName: dBClientInterfaceName,
	}
}

// Return the dbClient interface
func (d Database) newDBClient(
	lc logger.LoggingClient,
	credentials bootstrapConfig.Credentials) (interfaces.DBClient, error) {
	databaseInfo := d.database.GetDatabaseInfo()[common.Primary]
	switch databaseInfo.Type {
	case "redisdb":
		return redis.NewClient(
			db.Configuration{
				Host:     databaseInfo.Host,
				Port:     databaseInfo.Port,
				Password: credentials.Password,
			},
			lc)
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
	secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)

	// get database credentials.
	var credentials bootstrapConfig.Credentials
	for startupTimer.HasNotElapsed() {
		var err error

		secrets, err := secretProvider.GetSecret(d.database.GetDatabaseInfo()[common.Primary].Type)
		if err == nil {
			credentials = bootstrapConfig.Credentials{
				Username: secrets[secret.UsernameKey],
				Password: secrets[secret.PasswordKey],
			}

			break
		}

		lc.Warn(fmt.Sprintf("couldn't retrieve database credentials: %v", err.Error()))
		startupTimer.SleepForInterval()
	}

	// initialize database.
	var dbClient interfaces.DBClient

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
		d.dBClientInterfaceName: func(get di.Get) interface{} {
			return dbClient
		},
	})

	lc.Info("Database for V2 API connected")
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		for {
			// wait for httpServer to stop running (e.g. handling requests) before closing the database connection.
			if !d.httpServer.IsRunning() {
				dbClient.CloseSession()
				break
			}
			time.Sleep(time.Second)
		}
		lc.Info("Database for V2 API disconnected")
	}()

	return true
}
