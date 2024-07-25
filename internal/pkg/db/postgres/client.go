//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

const defaultDBName = "postgres"

var once sync.Once
var dc *Client

// Client represents a Postgres client
type Client struct {
	ConnPool *pgxpool.Pool

	loggingClient logger.LoggingClient
}

// NewClient returns a pointer to the Postgres client
func NewClient(ctx context.Context, config db.Configuration, baseScriptPath, extScriptPath string, lc logger.LoggingClient) (*Client, errors.EdgeX) {
	// TODO: Should set the database's name in the configuration file as well
	databaseName := config.DatabaseName
	if databaseName == "" {
		databaseName = defaultDBName
	}

	var edgeXerr errors.EdgeX
	once.Do(func() {
		connectionStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", config.Username, config.Password, config.Host, config.Port, databaseName)
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
		return nil, errors.NewCommonEdgeX(errors.Kind(edgeXerr), "failed to execute base DB scripts", edgeXerr)
	}
	lc.Info("successfully execute base DB scripts")

	// execute extension DB scripts
	if edgeXerr = executeDBScripts(ctx, dc.ConnPool, extScriptPath); edgeXerr != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(edgeXerr), "failed to execute extension DB scripts", edgeXerr)
	}

	return dc, nil
}

// CloseSession closes the connections to postgres
func (c *Client) CloseSession() {
	c.ConnPool.Close()
}
