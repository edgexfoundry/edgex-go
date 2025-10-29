//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	stdErrs "errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"
	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// AddProvisionWatcher adds a new provision watcher
func (c *Client) AddProvisionWatcher(pw model.ProvisionWatcher) (model.ProvisionWatcher, errors.EdgeX) {
	ctx := context.Background()

	if len(pw.Id) == 0 {
		pw.Id = uuid.New().String()
	}

	exists, _ := provisionWatcherNameExists(ctx, c.ConnPool, pw.Name)
	if exists {
		return model.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("provision watcher name %s already exists", pw.Name), nil)
	}

	timestamp := pkgCommon.MakeTimestamp()
	pw.Created = timestamp
	pw.Modified = timestamp
	provisionWatcherJSONBytes, err := json.Marshal(pw)
	if err != nil {
		return model.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal provision watcher for Postgres persistence", err)
	}

	_, err = c.ConnPool.Exec(ctx, sqlInsert(provisionWatcherTableName, idCol, contentCol), pw.Id, provisionWatcherJSONBytes)
	if err != nil {
		return model.ProvisionWatcher{}, pgClient.WrapDBError("failed to insert provision watcher", err)
	}

	return pw, nil
}

// ProvisionWatcherById gets a provision watcher by id
func (c *Client) ProvisionWatcherById(id string) (model.ProvisionWatcher, errors.EdgeX) {
	ctx := context.Background()

	pw, err := queryOneProvisionWatcher(ctx, c.ConnPool, sqlQueryContentById(provisionWatcherTableName), id)
	if err != nil {
		if stdErrs.Is(err, pgx.ErrNoRows) {
			return pw, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no provision watcher with id '%s' found", id), err)
		}
		return pw, pgClient.WrapDBError("failed to scan row to provision watcher model", err)
	}
	return pw, nil
}

// ProvisionWatcherByName gets a provision watcher by name
func (c *Client) ProvisionWatcherByName(name string) (model.ProvisionWatcher, errors.EdgeX) {
	ctx := context.Background()

	queryObj := map[string]any{nameField: name}
	pw, err := queryOneProvisionWatcher(ctx, c.ConnPool, sqlQueryContentByJSONField(provisionWatcherTableName), queryObj)
	if err != nil {
		if stdErrs.Is(err, pgx.ErrNoRows) {
			return pw, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no provision watcher with name '%s' found", name), err)
		}
		return pw, pgClient.WrapDBError("failed to scan row to provision watcher model", err)
	}
	return pw, nil
}

// ProvisionWatchersByServiceName query provision watchers by offset, limit and service name
func (c *Client) ProvisionWatchersByServiceName(offset int, limit int, name string) ([]model.ProvisionWatcher, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{serviceNameField: name}
	return queryProvisionWatchers(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(provisionWatcherTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
}

// ProvisionWatchersByProfileName query provision watchers by offset, limit and profile name
func (c *Client) ProvisionWatchersByProfileName(offset int, limit int, name string) ([]model.ProvisionWatcher, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{"DiscoveredDevice": map[string]any{profileNameField: name}}
	return queryProvisionWatchers(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(provisionWatcherTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
}

// AllProvisionWatchers query provision watchers with offset, limit and labels
func (c *Client) AllProvisionWatchers(offset int, limit int, labels []string) (pws []model.ProvisionWatcher, err errors.EdgeX) {
	ctx := context.Background()

	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	if len(labels) > 0 {
		c.loggingClient.Debugf("Querying provision watchers by labels: %v", labels)
		queryObj := map[string]any{labelsField: labels}
		pws, err = queryProvisionWatchers(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(provisionWatcherTableName),
			pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
		if err != nil {
			return pws, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all provision watchers by labels", err)
		}
	} else {
		pws, err = queryProvisionWatchers(ctx, c.ConnPool, sqlQueryContentWithPagination(provisionWatcherTableName), offset, validLimit)
		if err != nil {
			return pws, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all provision watchers", err)
		}
	}

	return pws, nil
}

// DeleteProvisionWatcherByName deletes a provision watcher by name
func (c *Client) DeleteProvisionWatcherByName(name string) errors.EdgeX {
	ctx := context.Background()

	queryObj := map[string]any{nameField: name}
	_, err := c.ConnPool.Exec(ctx, sqlDeleteByJSONField(provisionWatcherTableName), queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete provision watcher by name %s", name), err)
	}
	return nil
}

// UpdateProvisionWatcher updates a provision watcher
func (c *Client) UpdateProvisionWatcher(pw model.ProvisionWatcher) errors.EdgeX {
	ctx := context.Background()

	exists, edgeXErr := provisionWatcherNameExists(ctx, c.ConnPool, pw.Name)
	if edgeXErr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXErr)
	} else if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("provision watcher '%s' does not exist", pw.Name), nil)
	}

	pw.Modified = pkgCommon.MakeTimestamp()

	updatedProvisionWatcherJSONBytes, err := json.Marshal(pw)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal provision watcher for Postgres persistence", err)
	}

	queryObj := map[string]any{nameField: pw.Name}
	_, err = c.ConnPool.Exec(ctx, sqlUpdateColsByJSONCondCol(provisionWatcherTableName, contentCol), updatedProvisionWatcherJSONBytes, queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to update provision watcher by name '%s' from %s table", pw.Name, provisionWatcherTableName), err)
	}

	return nil
}

// ProvisionWatcherCountByLabels returns the total count of Provision Watchers with labels specified.  If no label is specified, the total count of all provision watchers will be returned.
func (c *Client) ProvisionWatcherCountByLabels(labels []string) (int64, errors.EdgeX) {
	ctx := context.Background()

	if len(labels) > 0 {
		queryObj := map[string]any{labelsField: labels}
		return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(provisionWatcherTableName), queryObj)
	}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCount(provisionWatcherTableName))
}

// ProvisionWatcherCountByServiceName returns the count of Provision Watcher associated with specified service
func (c *Client) ProvisionWatcherCountByServiceName(name string) (int64, errors.EdgeX) {
	ctx := context.Background()
	queryObj := map[string]any{serviceNameField: name}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(provisionWatcherTableName), queryObj)
}

// ProvisionWatcherCountByProfileName returns the count of Provision Watcher associated with specified profile
func (c *Client) ProvisionWatcherCountByProfileName(name string) (int64, errors.EdgeX) {
	ctx := context.Background()
	queryObj := map[string]any{"DiscoveredDevice": map[string]any{profileNameField: name}}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(provisionWatcherTableName), queryObj)
}

func provisionWatcherNameExists(ctx context.Context, connPool *pgxpool.Pool, name string) (bool, errors.EdgeX) {
	var exists bool
	queryObj := map[string]any{nameField: name}
	err := connPool.QueryRow(ctx, sqlCheckExistsByJSONField(provisionWatcherTableName), queryObj).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError(fmt.Sprintf("failed to query provision watcher by name '%s' from %s table", name, provisionWatcherTableName), err)
	}
	return exists, nil
}

func queryOneProvisionWatcher(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (model.ProvisionWatcher, errors.EdgeX) {
	var pw model.ProvisionWatcher
	row := connPool.QueryRow(ctx, sql, args...)

	if err := row.Scan(&pw); err != nil {
		return pw, pgClient.WrapDBError("failed to query provision watcher", err)
	}
	return pw, nil
}

func queryProvisionWatchers(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]model.ProvisionWatcher, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query provision watchers", err)
	}

	pws, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.ProvisionWatcher, error) {
		var pw model.ProvisionWatcher
		scanErr := row.Scan(&pw)
		return pw, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to ProvisionWatcher model", err)
	}

	return pws, nil
}
