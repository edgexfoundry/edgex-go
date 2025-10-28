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

// AddDeviceService adds a new device service
func (c *Client) AddDeviceService(ds model.DeviceService) (model.DeviceService, errors.EdgeX) {
	ctx := context.Background()

	if len(ds.Id) == 0 {
		ds.Id = uuid.New().String()
	}

	// verify if device service name is unique or not
	var exists bool
	exists, _ = deviceServiceNameExists(ctx, c.ConnPool, ds.Name)
	if exists {
		return model.DeviceService{}, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device service name %s already exists", ds.Name), nil)
	}

	timestamp := pkgCommon.MakeTimestamp()
	ds.Created = timestamp
	ds.Modified = timestamp
	// Marshal the device service to store it in the database
	deviceServiceJSONBytes, err := json.Marshal(ds)
	if err != nil {
		return model.DeviceService{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device service for Postgres persistence", err)
	}

	_, err = c.ConnPool.Exec(ctx, sqlInsert(deviceServiceTableName, idCol, contentCol), ds.Id, deviceServiceJSONBytes)
	if err != nil {
		return model.DeviceService{}, pgClient.WrapDBError("failed to insert device service", err)
	}

	return ds, nil
}

// DeviceServiceById gets a device service by id
func (c *Client) DeviceServiceById(id string) (deviceService model.DeviceService, edgeXerr errors.EdgeX) {
	ctx := context.Background()

	deviceService, err := queryOneDeviceService(ctx, c.ConnPool, sqlQueryContentById(deviceServiceTableName), id)
	if err != nil {
		if stdErrs.Is(err, pgx.ErrNoRows) {
			return deviceService, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no device service with id '%s' found", id), err)
		}
		return deviceService, pgClient.WrapDBError("failed to scan row to device service model", err)
	}
	return deviceService, nil
}

// DeviceServiceByName gets a device service by name
func (c *Client) DeviceServiceByName(name string) (deviceService model.DeviceService, edgeXerr errors.EdgeX) {
	ctx := context.Background()

	queryObj := map[string]any{nameField: name}
	deviceService, err := queryOneDeviceService(ctx, c.ConnPool, sqlQueryContentByJSONField(deviceServiceTableName), queryObj)
	if err != nil {
		if stdErrs.Is(err, pgx.ErrNoRows) {
			return deviceService, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no device service with name '%s' found", name), err)
		}
		return deviceService, pgClient.WrapDBError("failed to scan row to device service model", err)
	}
	return deviceService, nil
}

// DeleteDeviceServiceById deletes a device service by id
func (c *Client) DeleteDeviceServiceById(id string) errors.EdgeX {
	ctx := context.Background()

	_, err := c.ConnPool.Exec(ctx, sqlDeleteById(deviceServiceTableName), id)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete device service by id %s", id), err)
	}
	return nil
}

// DeleteDeviceServiceByName deletes a device service by name
func (c *Client) DeleteDeviceServiceByName(name string) errors.EdgeX {
	ctx := context.Background()

	queryObj := map[string]any{nameField: name}
	_, err := c.ConnPool.Exec(ctx, sqlDeleteByJSONField(deviceServiceTableName), queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete device service by name %s", name), err)
	}
	return nil
}

// DeviceServiceNameExists checks the device service exists by name
func (c *Client) DeviceServiceNameExists(name string) (bool, errors.EdgeX) {
	ctx := context.Background()
	return deviceServiceNameExists(ctx, c.ConnPool, name)
}

// AllDeviceServices returns multiple device services per query criteria, including
// offset: the number of items to skip before starting to collect the result set
// limit: The numbers of items to return
// labels: allows for querying a given object by associated user-defined labels
func (c *Client) AllDeviceServices(offset int, limit int, labels []string) (deviceServices []model.DeviceService, err errors.EdgeX) {
	ctx := context.Background()

	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	if len(labels) > 0 {
		c.loggingClient.Debugf("Querying device services by labels: %v", labels)
		queryObj := map[string]any{labelsField: labels}
		deviceServices, err = queryDeviceServices(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(deviceServiceTableName),
			pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
		if err != nil {
			return deviceServices, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all device services by labels", err)
		}
	} else {
		deviceServices, err = queryDeviceServices(ctx, c.ConnPool, sqlQueryContentWithPagination(deviceServiceTableName), offset, validLimit)
		if err != nil {
			return deviceServices, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all device services", err)
		}
	}

	return deviceServices, nil
}

// UpdateDeviceService updates a device service
func (c *Client) UpdateDeviceService(ds model.DeviceService) errors.EdgeX {
	ctx := context.Background()

	// Check if the device service exists
	exists, edgeXErr := deviceServiceNameExists(ctx, c.ConnPool, ds.Name)
	if edgeXErr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXErr)
	} else if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device service '%s' does not exist", ds.Name), nil)
	}

	ds.Modified = pkgCommon.MakeTimestamp()

	// Marshal the device service to store it in the database
	updatedDeviceServiceJSONBytes, err := json.Marshal(ds)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device service for Postgres persistence", err)
	}

	queryObj := map[string]any{nameField: ds.Name}
	_, err = c.ConnPool.Exec(ctx, sqlUpdateColsByJSONCondCol(deviceServiceTableName, contentCol), updatedDeviceServiceJSONBytes, queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to update device service by name '%s' from %s table", ds.Name, deviceServiceTableName), err)
	}

	return nil
}

// DeviceServiceCountByLabels returns the total count of Device Services with labels specified.  If no label is specified, the total count of all device services will be returned.
func (c *Client) DeviceServiceCountByLabels(labels []string) (int64, errors.EdgeX) {
	ctx := context.Background()

	if len(labels) > 0 {
		queryObj := map[string]any{labelsField: labels}
		return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(deviceServiceTableName), queryObj)
	}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCount(deviceServiceTableName))
}

func deviceServiceNameExists(ctx context.Context, connPool *pgxpool.Pool, name string) (bool, errors.EdgeX) {
	var exists bool
	queryObj := map[string]any{nameField: name}
	err := connPool.QueryRow(ctx, sqlCheckExistsByJSONField(deviceServiceTableName), queryObj).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError(fmt.Sprintf("failed to query device service by name '%s' from %s table", name, deviceServiceTableName), err)
	}
	return exists, nil
}

func queryOneDeviceService(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (model.DeviceService, errors.EdgeX) {
	var ds model.DeviceService
	row := connPool.QueryRow(ctx, sql, args...)

	if err := row.Scan(&ds); err != nil {
		return ds, pgClient.WrapDBError("failed to query device service", err)
	}
	return ds, nil
}

func queryDeviceServices(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]model.DeviceService, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query device services", err)
	}

	deviceServices, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.DeviceService, error) {
		var ds model.DeviceService
		scanErr := row.Scan(&ds)
		return ds, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to DeviceService model", err)
	}

	return deviceServices, nil
}
