//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	postgresClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
)

type Client struct {
	*postgresClient.Client
	loggingClient logger.LoggingClient
}

func NewClient(ctx context.Context, config db.Configuration, baseScriptPath, extScriptPath string, lc logger.LoggingClient) (*Client, errors.EdgeX) {
	var err error
	dc := &Client{}
	dc.Client, err = postgresClient.NewClient(ctx, config, baseScriptPath, extScriptPath, lc)
	dc.loggingClient = lc
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "postgres client creation failed", err)
	}

	return dc, nil
}
