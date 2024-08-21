//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

const defaultDBName = "edgex_db"

var once sync.Once
var dc *Client

// Client represents a Postgres client
type Client struct {
	ConnPool *pgxpool.Pool

	loggingClient logger.LoggingClient
}

// NewClient returns a pointer to the Postgres client
func NewClient(ctx context.Context, config db.Configuration, baseScriptPath, extScriptPath string, lc logger.LoggingClient) (*Client, errors.EdgeX) {
	// Get the database name from the environment variable
	databaseName := os.Getenv("EDGEX_DBNAME")
	if databaseName == "" {
		databaseName = defaultDBName
	}

	var edgeXerr errors.EdgeX
	once.Do(func() {
		// use url encode to prevent special characters in the connection string
		connectionStr := "postgres://" + fmt.Sprintf("%s:%s@%s:%d/%s", url.PathEscape(config.Username), url.PathEscape(config.Password), url.PathEscape(config.Host), config.Port, url.PathEscape(databaseName))
		dbPool, err := pgxpool.New(ctx, connectionStr)
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

	// execute base DB scripts
	if edgeXerr = executeDBScripts(ctx, dc.ConnPool, baseScriptPath); edgeXerr != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(edgeXerr), "failed to execute Postgres base DB scripts", edgeXerr)
	}
	lc.Info("successfully execute Postgres base DB scripts")

	// execute extension DB scripts
	if edgeXerr = executeDBScripts(ctx, dc.ConnPool, extScriptPath); edgeXerr != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(edgeXerr), "failed to execute Postgres extension DB scripts", edgeXerr)
	}

	return dc, nil
}

// CloseSession closes the connections to postgres
func (c *Client) CloseSession() {
	c.ConnPool.Close()
}
