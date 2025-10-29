//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"embed"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

const defaultDBName = "edgex_db"
const defaultMaxConns = int32(4)
const defaultMaxConnIdleTime = time.Minute * 30
const defaultMaxConnLifetime = time.Hour

var once sync.Once
var dc *Client

// Client represents a Postgres client
type Client struct {
	ConnPool *pgxpool.Pool

	loggingClient logger.LoggingClient
}

// NewClient returns a pointer to the Postgres client
func NewClient(ctx context.Context, config db.Configuration, lc logger.LoggingClient, schemaName, serviceKey, serviceVersion string, sqlFiles embed.FS) (*Client, errors.EdgeX) {
	var edgeXerr errors.EdgeX
	once.Do(func() {
		connPoolConfig, err := connPoolConConfig(config, lc)
		if err != nil {
			edgeXerr = WrapDBError("fail to parse pg conn pool config", err)
			return
		}

		dbPool, err := pgxpool.NewWithConfig(ctx, connPoolConfig)
		if err != nil {
			edgeXerr = WrapDBError("fail to create pg connection pool", err)
		}

		dc = &Client{
			ConnPool:      dbPool,
			loggingClient: lc,
		}
	})
	if edgeXerr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to create a Postgres client", edgeXerr)
	}

	// invoke Ping to test the connectivity to the DB
	if err := dc.ConnPool.Ping(ctx); err != nil {
		return nil, WrapDBError("failed to acquire a connection from database connection pool", err)
	}

	// create a new TableManager instance
	tableManager, err := NewTableManager(dc.ConnPool, lc, schemaName, serviceKey, serviceVersion, sqlFiles)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to create a new TableManager instance", err)
	}

	err = tableManager.RunScripts(ctx)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "TableManager failed to run SQL scripts", err)
	}

	return dc, nil
}

func connPoolConConfig(config db.Configuration, lc logger.LoggingClient) (*pgxpool.Config, error) {
	// Get the database name from the environment variable
	databaseName := os.Getenv("EDGEX_DBNAME")
	if databaseName == "" {
		databaseName = defaultDBName
	}

	// use url encode to prevent special characters in the connection string
	connectionStr := "postgres://" + fmt.Sprintf("%s:%s@%s:%d/%s", url.PathEscape(config.Username), url.PathEscape(config.Password), url.PathEscape(config.Host), config.Port, url.PathEscape(databaseName))
	connPoolConfig, err := pgxpool.ParseConfig(connectionStr)
	if err != nil {
		return nil, WrapDBError("fail to parse pg connection string", err)
	}
	connPoolConfig.MaxConns = config.MaxConns
	if connPoolConfig.MaxConns <= 0 {
		lc.Errorf("The MaxConns too small, use default '%d'", defaultMaxConns)
		connPoolConfig.MaxConns = defaultMaxConns
	}
	if connPoolConfig.MaxConnIdleTime, err = time.ParseDuration(config.MaxConnIdleTime); err != nil {
		connPoolConfig.MaxConnIdleTime = defaultMaxConnIdleTime
		lc.Errorf("Fail to parse pg conn pool MaxConnIdleTime config, use default '%v', err: %v", defaultMaxConnIdleTime, err)
	}
	if connPoolConfig.MaxConnLifetime, err = time.ParseDuration(config.MaxConnLifetime); err != nil {
		connPoolConfig.MaxConnLifetime = defaultMaxConnLifetime
		lc.Errorf("Fail to parse pg conn pool MaxConnLifetime config, use default '%v', err: %v", defaultMaxConnLifetime, err)
	}
	return connPoolConfig, nil
}

// CloseSession closes the connections to postgres
func (c *Client) CloseSession() {
	c.ConnPool.Close()
}
