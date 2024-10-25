//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	goErrors "errors"
	"fmt"
	"time"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (c *Client) AddRegistration(r models.Registration) (models.Registration, errors.EdgeX) {
	ctx := context.Background()
	exists, edgexErr := checkRegistrationExists(c.ConnPool, ctx, r.ServiceId)
	if edgexErr != nil {
		return models.Registration{}, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	if exists {
		return models.Registration{}, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("registry with service id '%s' already exists", r.ServiceId), nil)
	}

	timestamp := time.Now().UTC().UnixMilli()
	r.Created = timestamp
	r.Modified = timestamp
	dataBytes, err := json.Marshal(r)
	if err != nil {
		return models.Registration{}, errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal registration model", err)
	}

	_, err = c.ConnPool.Exec(context.Background(),
		sqlInsert(registryTableName, contentCol),
		dataBytes,
	)
	if err != nil {
		return models.Registration{}, pgClient.WrapDBError("failed to insert row to registry table", err)
	}

	return r, nil
}

// Registrations retrieves all the registry information from database
func (c *Client) Registrations() ([]models.Registration, errors.EdgeX) {
	rows, err := c.ConnPool.Query(context.Background(), sqlQueryContent(registryTableName))
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query rows from registry table", err)
	}

	registrations, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Registration, error) {
		var r models.Registration
		scanErr := row.Scan(&r)
		return r, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to Registration model", err)
	}

	return registrations, nil
}

// RegistrationByServiceId queries the registry by service id from database
func (c *Client) RegistrationByServiceId(serviceId string) (models.Registration, errors.EdgeX) {
	return queryRegistryByServiceId(c.ConnPool, serviceId)
}

// UpdateRegistration updates the registry information by service id from database
func (c *Client) UpdateRegistration(r models.Registration) errors.EdgeX {
	ctx := context.Background()

	oldRegistry, edgexErr := queryRegistryByServiceId(c.ConnPool, r.ServiceId)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	if r.Created == 0 {
		r.Created = oldRegistry.Created
	}
	r.Modified = time.Now().UTC().UnixMilli()

	dataBytes, err := json.Marshal(r)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal registration model", err)
	}

	queryObj := map[string]any{serviceIdField: r.ServiceId}
	_, err = c.ConnPool.Exec(ctx,
		sqlUpdateColsByJSONCondCol(registryTableName, contentCol),
		dataBytes,
		queryObj,
	)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to update row by service id '%s' from registry table", r.ServiceId), err)
	}

	return nil
}

// DeleteRegistrationByServiceId deletes the registry by service id from database
func (c *Client) DeleteRegistrationByServiceId(serviceId string) errors.EdgeX {
	ctx := context.Background()

	exists, edgexErr := checkRegistrationExists(c.ConnPool, ctx, serviceId)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("service id '%s' not found from registry table", serviceId), nil)
	}

	queryObj := map[string]any{serviceIdField: serviceId}

	_, err := c.ConnPool.Exec(context.Background(), sqlDeleteByJSONField(registryTableName), queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete row with service id '%s' from registry table", serviceId), err)
	}

	return nil
}

// checkRegistrationExists checks if registration exists by service id
func checkRegistrationExists(connPool *pgxpool.Pool, ctx context.Context, serviceId string) (bool, errors.EdgeX) {
	var exists bool

	queryObj := map[string]any{serviceIdField: serviceId}
	err := connPool.QueryRow(ctx, sqlCheckExistsByJSONField(registryTableName), queryObj).Scan(&exists)
	if err != nil {
		return exists, pgClient.WrapDBError(fmt.Sprintf("failed to query row by service id '%s' from registry table", serviceId), err)
	}

	return exists, nil
}

// queryRegistryByServiceId queries the registry by service id
func queryRegistryByServiceId(connPool *pgxpool.Pool, serviceId string) (models.Registration, errors.EdgeX) {
	var registry models.Registration
	sqlStmt := sqlQueryContentByJSONField(registryTableName)
	queryObj := map[string]any{serviceIdField: serviceId}

	row := connPool.QueryRow(context.Background(), sqlStmt, queryObj)

	err := row.Scan(&registry)
	if err != nil {
		if goErrors.Is(err, pgx.ErrNoRows) {
			return models.Registration{}, pgClient.WrapDBError(fmt.Sprintf("service id '%s' not found from registry table", serviceId), err)
		}
		return models.Registration{}, pgClient.WrapDBError(fmt.Sprintf("failed to query row by serviceId '%s' from registry table", serviceId), err)
	}
	return registry, nil
}
