//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"embed"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	postgresClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
)

type Client struct {
	*postgresClient.Client
	loggingClient logger.LoggingClient
	dic           *di.Container
}

func NewClient(ctx context.Context, config db.Configuration, lc logger.LoggingClient, dic *di.Container, schemaName, serviceKey, serviceVersion string, sqlFiles embed.FS) (*Client, errors.EdgeX) {
	var err error
	dc := &Client{}
	dc.Client, err = postgresClient.NewClient(ctx, config, lc, schemaName, serviceKey, serviceVersion, sqlFiles)
	dc.loggingClient = lc
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "postgres client creation failed", err)
	}
	dc.dic = dic

	return dc, nil
}
