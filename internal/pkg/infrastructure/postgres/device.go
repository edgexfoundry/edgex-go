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
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"
	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// AddDevice adds a new device
func (c *Client) AddDevice(d model.Device) (model.Device, errors.EdgeX) {
	ctx := context.Background()

	if len(d.Id) == 0 {
		d.Id = uuid.New().String()
	}

	exists, _ := deviceNameExists(ctx, c.ConnPool, d.Name)
	if exists {
		return model.Device{}, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device name %s already exists", d.Name), nil)
	}

	timestamp := pkgCommon.MakeTimestamp()
	d.Created = timestamp
	d.Modified = timestamp
	// Marshal the device to store it in the database
	deviceJSONBytes, err := json.Marshal(d)
	if err != nil {
		return model.Device{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device for Postgres persistence", err)
	}

	_, err = c.ConnPool.Exec(ctx, sqlInsert(deviceTableName, idCol, contentCol), d.Id, deviceJSONBytes)
	if err != nil {
		return model.Device{}, pgClient.WrapDBError("failed to insert device", err)
	}

	return d, nil
}

// DeleteDeviceById deletes a device by id
func (c *Client) DeleteDeviceById(id string) errors.EdgeX {
	ctx := context.Background()

	_, err := c.ConnPool.Exec(ctx, sqlDeleteById(deviceTableName), id)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete device by id %s", id), err)
	}
	return nil
}

// DeleteDeviceByName deletes a device by name
func (c *Client) DeleteDeviceByName(name string) errors.EdgeX {
	ctx := context.Background()

	queryObj := map[string]any{nameField: name}
	_, err := c.ConnPool.Exec(ctx, sqlDeleteByJSONField(deviceTableName), queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete device by name %s", name), err)
	}
	return nil
}

// DevicesByServiceName query devices by offset, limit and name
func (c *Client) DevicesByServiceName(offset int, limit int, name string) ([]model.Device, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{serviceNameField: name}
	return queryDevices(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPagination(deviceTableName), queryObj, offset, validLimit)
}

// DeviceIdExists checks the device existence by id
func (c *Client) DeviceIdExists(id string) (bool, errors.EdgeX) {
	ctx := context.Background()
	return deviceIdExists(ctx, c.ConnPool, id)
}

// DeviceNameExists checks the device existence by name
func (c *Client) DeviceNameExists(name string) (bool, errors.EdgeX) {
	ctx := context.Background()
	return deviceNameExists(ctx, c.ConnPool, name)
}

// DeviceById gets a device by id
func (c *Client) DeviceById(id string) (model.Device, errors.EdgeX) {
	ctx := context.Background()

	d, err := queryOneDevice(ctx, c.ConnPool, sqlQueryContentById(deviceTableName), id)
	if err != nil {
		if stdErrs.Is(err, pgx.ErrNoRows) {
			return d, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no device with id '%s' found", id), err)
		}
		return d, pgClient.WrapDBError("failed to scan row to device model", err)
	}
	return d, nil
}

// DeviceByName gets a device by name
func (c *Client) DeviceByName(name string) (model.Device, errors.EdgeX) {
	ctx := context.Background()

	queryObj := map[string]any{nameField: name}
	d, err := queryOneDevice(ctx, c.ConnPool, sqlQueryContentByJSONField(deviceTableName), queryObj)
	if err != nil {
		if stdErrs.Is(err, pgx.ErrNoRows) {
			return d, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no device with name '%s' found", name), err)
		}
		return d, pgClient.WrapDBError("failed to scan row to device model", err)
	}
	return d, nil
}

// AllDevices query the devices with offset, limit, and labels
func (c *Client) AllDevices(offset int, limit int, labels []string) (devices []model.Device, err errors.EdgeX) {
	ctx := context.Background()

	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	if len(labels) > 0 {
		c.loggingClient.Debugf("Querying devices by labels: %v", labels)
		queryObj := map[string]any{labelsField: labels}
		devices, err = queryDevices(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPagination(deviceTableName), queryObj, offset, validLimit)
		if err != nil {
			return devices, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all devices by labels", err)
		}
	} else {
		devices, err = queryDevices(ctx, c.ConnPool, sqlQueryContentWithPagination(deviceTableName), offset, validLimit)
		if err != nil {
			return devices, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all devices", err)
		}
	}

	return devices, nil
}

// DevicesByProfileName query devices by offset, limit and profile name
func (c *Client) DevicesByProfileName(offset int, limit int, profileName string) ([]model.Device, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{profileNameField: profileName}
	return queryDevices(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPagination(deviceTableName), queryObj, offset, validLimit)
}

// UpdateDevice updates a device
func (c *Client) UpdateDevice(d model.Device) errors.EdgeX {
	ctx := context.Background()

	// Check if the device exists
	exists, edgeXErr := deviceNameExists(ctx, c.ConnPool, d.Name)
	if edgeXErr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXErr)
	} else if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device '%s' does not exist", d.Name), nil)
	}

	d.Modified = pkgCommon.MakeTimestamp()

	// Marshal the device to store it in the database
	updatedDeviceJSONBytes, err := json.Marshal(d)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device for Postgres persistence", err)
	}

	queryObj := map[string]any{nameField: d.Name}
	_, err = c.ConnPool.Exec(ctx, sqlUpdateColsByJSONCondCol(deviceTableName, contentCol), updatedDeviceJSONBytes, queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to update device by name '%s' from %s table", d.Name, deviceTableName), err)
	}

	return nil
}

// DeviceCountByLabels returns the total count of Devices with labels specified.  If no label is specified, the total count of all devices will be returned.
func (c *Client) DeviceCountByLabels(labels []string) (uint32, errors.EdgeX) {
	ctx := context.Background()

	if len(labels) > 0 {
		queryObj := map[string]any{labelsField: labels}
		return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(deviceTableName), queryObj)
	}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCount(deviceTableName))
}

// DeviceCountByProfileName returns the count of Devices associated with specified profile
func (c *Client) DeviceCountByProfileName(profileName string) (uint32, errors.EdgeX) {
	ctx := context.Background()
	queryObj := map[string]any{profileNameField: profileName}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(deviceTableName), queryObj)
}

// DeviceCountByServiceName returns the count of Devices associated with specified service
func (c *Client) DeviceCountByServiceName(serviceName string) (uint32, errors.EdgeX) {
	ctx := context.Background()
	queryObj := map[string]any{serviceNameField: serviceName}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(deviceTableName), queryObj)
}

// Get device objects with matching parent and labels (one level of the tree).
func deviceTreeLevel(ctx context.Context, connPool *pgxpool.Pool, parent string, labels []string) ([]model.Device, errors.EdgeX) {
	queryObj := map[string]any{parentField: parent}
	if len(labels) != 0 {
		queryObj[labelsField] = labels
	}
	return queryDevices(ctx, connPool, sqlQueryContentByJSONField(deviceTableName), queryObj)
}

// Get the entire subtree starting with the given parent, descending at most the given number of levels.
func deviceSubTree(ctx context.Context, connPool *pgxpool.Pool, parent string, levels int, labels []string) ([]model.Device, errors.EdgeX) {
	var emptyList = []model.Device{}
	if levels <= 0 {
		return emptyList, nil
	}
	topLevelList, err := deviceTreeLevel(ctx, connPool, parent, labels)
	if err != nil {
		return emptyList, err
	}
	if levels == 1 {
		return topLevelList, nil
	}
	var subtreesAtThisLevel []model.Device
	for _, device := range topLevelList {
		if device.Name == device.Parent {
			message := "Device " + device.Name + " is its own parent, stopping tree query"
			return emptyList, errors.NewCommonEdgeX(errors.KindDatabaseError, message, nil)
		}
		subtree, err := deviceSubTree(ctx, connPool, device.Name, levels-1, labels)
		if err != nil {
			return emptyList, err
		}
		subtreesAtThisLevel = append(subtreesAtThisLevel, subtree...)
	}
	return append(topLevelList, subtreesAtThisLevel...), nil
}

// Get the full result-set since that's the only way to correctly get totalCount.
// Then return the subset of the result-set that corresponds to the requested offset and limit.
func (c *Client) DeviceTree(parent string, levels int, offset int, limit int, labels []string) (uint32, []model.Device, errors.EdgeX) {
	var maxLevels int
	var emptyList = []model.Device{}
	if levels <= 0 {
		maxLevels = math.MaxInt
	} else {
		maxLevels = levels
	}
	allDevices, err := deviceSubTree(context.Background(), c.ConnPool, parent, maxLevels, labels)
	if err != nil {
		return 0, emptyList, err
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= len(allDevices) {
		return uint32(len(allDevices)), emptyList, nil
	}
	numToReturn := len(allDevices) - offset
	if limit > 0 && limit < numToReturn {
		numToReturn = limit
	}
	return uint32(len(allDevices)), allDevices[offset : offset+numToReturn], nil
}

func deviceNameExists(ctx context.Context, connPool *pgxpool.Pool, name string) (bool, errors.EdgeX) {
	var exists bool
	queryObj := map[string]any{nameField: name}
	err := connPool.QueryRow(ctx, sqlCheckExistsByJSONField(deviceTableName), queryObj).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError(fmt.Sprintf("failed to query device by name '%s' from %s table", name, deviceTableName), err)
	}
	return exists, nil
}

func deviceIdExists(ctx context.Context, connPool *pgxpool.Pool, id string) (bool, errors.EdgeX) {
	var exists bool
	err := connPool.QueryRow(ctx, sqlCheckExistsByCol(deviceTableName, idCol), id).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError(fmt.Sprintf("failed to query device by id '%s' from %s table", id, deviceTableName), err)
	}
	return exists, nil
}

func queryOneDevice(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (model.Device, errors.EdgeX) {
	var d model.Device
	row := connPool.QueryRow(ctx, sql, args...)

	if err := row.Scan(&d); err != nil {
		return d, pgClient.WrapDBError("failed to query devicee", err)
	}
	return d, nil
}

func queryDevices(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]model.Device, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query devices", err)
	}

	devices, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Device, error) {
		var d model.Device
		scanErr := row.Scan(&d)
		return d, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to Device model", err)
	}

	return devices, nil
}
