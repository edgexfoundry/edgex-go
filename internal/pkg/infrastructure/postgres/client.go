//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"embed"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	postgresClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/postgres/cache"
)

type Client struct {
	*postgresClient.Client
	loggingClient     logger.LoggingClient
	deviceInfoIdCache cache.DeviceInfoIdCache
}

func NewClient(ctx context.Context, config db.Configuration, lc logger.LoggingClient, schemaName, serviceKey, serviceVersion string, sqlFiles embed.FS) (*Client, errors.EdgeX) {
	var err error
	dc := &Client{}
	dc.Client, err = postgresClient.NewClient(ctx, config, lc, schemaName, serviceKey, serviceVersion, sqlFiles)
	dc.loggingClient = lc
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "postgres client creation failed", err)
	}
	dc.deviceInfoIdCache = cache.NewDeviceInfoIdCache(lc)

	return dc, nil
}
