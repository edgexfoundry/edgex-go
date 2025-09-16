//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"embed"
	"sync"
	"time"

	bootstrapInterfaces "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/postgres"
	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/redis"
	"github.com/edgexfoundry/edgex-go/internal/pkg/interfaces"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

const (
	redisDBType    = "redisdb"
	postgresDBType = "postgres"
)

// httpServer defines the contract used to determine whether the http httpServer is running.
type httpServer interface {
	IsRunning() bool
}

// Database contains references to dependencies required by the database bootstrap implementation.
type Database struct {
	httpServer            httpServer
	database              bootstrapInterfaces.Database
	dBClientInterfaceName string
	schemaName            string
	serviceKey            string
	serviceVersion        string
	sqlFiles              embed.FS
}

// NewDatabase is a factory method that returns an initialized Database receiver struct.
func NewDatabase(httpServer httpServer, database bootstrapInterfaces.Database, dBClientInterfaceName, schemaName, serviceKey, version string, sqlFiles embed.FS) Database {
	return Database{
		httpServer:            httpServer,
		database:              database,
		dBClientInterfaceName: dBClientInterfaceName,
		schemaName:            schemaName,
		serviceKey:            serviceKey,
		serviceVersion:        version,
		sqlFiles:              sqlFiles,
	}
}

// Return the dbClient interface
func (d Database) newDBClient(
	ctx context.Context,
	lc logger.LoggingClient,
	dic *di.Container,
	credentials bootstrapConfig.Credentials) (interfaces.DBClient, error) {
	databaseInfo := d.database.GetDatabaseInfo()

	databaseConfig := db.Configuration{
		Host:            databaseInfo.Host,
		Port:            databaseInfo.Port,
		Password:        credentials.Password,
		Timeout:         databaseInfo.Timeout,
		MaxConns:        databaseInfo.MaxConns,
		MaxConnIdleTime: databaseInfo.MaxConnIdleTime,
		MaxConnLifetime: databaseInfo.MaxConnLifetime,
	}

	switch databaseInfo.Type {
	case redisDBType:
		return redis.NewClient(databaseConfig, lc)
	case postgresDBType:
		databaseConfig.Username = credentials.Username
		return postgres.NewClient(ctx, databaseConfig, lc, dic, d.schemaName, d.serviceKey, d.serviceVersion, d.sqlFiles)
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

	dbInfo := d.database.GetDatabaseInfo()
	if len(dbInfo.Host) == 0 || dbInfo.Port == 0 || len(dbInfo.Type) == 0 || len(dbInfo.Timeout) == 0 {
		lc.Error("Database configuration is empty or incomplete, missing common config? Use -cp or -cc flags for common config")
		return false
	}

	var credentials bootstrapConfig.Credentials
	dbCredsRetrieved := false
	for startupTimer.HasNotElapsed() {
		var err error

		secrets, err := secretProvider.GetSecret(d.database.GetDatabaseInfo().Type)
		if err == nil {
			credentials = bootstrapConfig.Credentials{
				Username: secrets[secret.UsernameKey],
				Password: secrets[secret.PasswordKey],
			}

			dbCredsRetrieved = true
			break
		}

		lc.Warnf("couldn't retrieve database credentials: %v and will retry it again, %s", err.Error(),
			"missing common config? Use -cp or -cc flags for common config")
		startupTimer.SleepForInterval()
	}

	// using this check to avoid the confusion with the case of both Username and Password being set to empty from credentials
	if !dbCredsRetrieved {
		// shouldn't go further if database credentials failed to retrieve
		lc.Error("bootstrap failed: couldn't retrieve database credentials after some retries, missing common config? Use -cp or -cc flags for common config")
		return false
	}

	// initialize database.
	var dbClient interfaces.DBClient

	for startupTimer.HasNotElapsed() {
		var err error
		dbClient, err = d.newDBClient(ctx, lc, dic, credentials)
		if err == nil {
			break
		}
		dbClient = nil
		lc.Warnf("couldn't create database client: %v", err.Error())
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

	lc.Infof("%s database connected", dbInfo.Type)
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
		lc.Infof("%s database disconnected", dbInfo.Type)
	}()

	return true
}
