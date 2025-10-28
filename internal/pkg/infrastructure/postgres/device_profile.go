//
// Copyright (C) 2024-2025 IOTech Ltd
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

// AddDeviceProfile adds a new device profile
func (c *Client) AddDeviceProfile(dp model.DeviceProfile) (model.DeviceProfile, errors.EdgeX) {
	ctx := context.Background()

	if len(dp.Id) == 0 {
		dp.Id = uuid.New().String()
	}

	// verify if device profile name is unique or not
	var exists bool
	exists, _ = deviceProfileNameExists(ctx, c.ConnPool, dp.Name)
	if exists {
		return model.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile name %s already exists", dp.Name), nil)
	}

	timestamp := pkgCommon.MakeTimestamp()
	dp.Created = timestamp
	dp.Modified = timestamp
	// Marshal the device profile to store it in the database
	deviceProfileJSONBytes, err := json.Marshal(dp)
	if err != nil {
		return model.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device profile for Postgres persistence", err)
	}

	_, err = c.ConnPool.Exec(ctx, sqlInsert(deviceProfileTableName, idCol, contentCol), dp.Id, deviceProfileJSONBytes)
	if err != nil {
		return model.DeviceProfile{}, pgClient.WrapDBError("failed to insert device profile", err)
	}

	return dp, nil
}

// UpdateDeviceProfile updates a new device profile
func (c *Client) UpdateDeviceProfile(dp model.DeviceProfile) errors.EdgeX {
	ctx := context.Background()

	// Check if the device profile exists
	exists, edgeXErr := deviceProfileNameExists(ctx, c.ConnPool, dp.Name)
	if edgeXErr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXErr)
	} else if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile '%s' does not exist", dp.Name), nil)
	}

	dp.Modified = pkgCommon.MakeTimestamp()

	// Marshal the device profile to store it in the database
	updatedDeviceProfileJSONBytes, err := json.Marshal(dp)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device profile for Postgres persistence", err)
	}

	queryObj := map[string]any{nameField: dp.Name}
	_, err = c.ConnPool.Exec(ctx, sqlUpdateColsByJSONCondCol(deviceProfileTableName, contentCol), updatedDeviceProfileJSONBytes, queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to update device profile by name '%s' from %s table", dp.Name, deviceProfileTableName), err)
	}

	return nil
}

// DeviceProfileById gets a device profile by id
func (c *Client) DeviceProfileById(id string) (model.DeviceProfile, errors.EdgeX) {
	ctx := context.Background()

	dp, err := queryOneDeviceProfile(ctx, c.ConnPool, sqlQueryContentById(deviceProfileTableName), id)
	if err != nil {
		if stdErrs.Is(err, pgx.ErrNoRows) {
			return dp, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no device profile with id '%s' found", id), err)
		}
		return dp, pgClient.WrapDBError("failed to scan row to device profile model", err)
	}
	return dp, nil
}

// DeviceProfileByName gets a device profile by name
func (c *Client) DeviceProfileByName(name string) (model.DeviceProfile, errors.EdgeX) {
	ctx := context.Background()

	queryObj := map[string]any{nameField: name}
	dp, err := queryOneDeviceProfile(ctx, c.ConnPool, sqlQueryContentByJSONField(deviceProfileTableName), queryObj)
	if err != nil {
		if stdErrs.Is(err, pgx.ErrNoRows) {
			return dp, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no device profile with name '%s' found", name), err)
		}
		return dp, pgClient.WrapDBError("failed to scan row to device profile model", err)
	}
	return dp, nil
}

// DeleteDeviceProfileById deletes a device profile by id
func (c *Client) DeleteDeviceProfileById(id string) errors.EdgeX {
	ctx := context.Background()

	_, err := c.ConnPool.Exec(ctx, sqlDeleteById(deviceProfileTableName), id)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete device profile by id %s", id), err)
	}
	return nil
}

// DeleteDeviceProfileByName deletes a device profile by name
func (c *Client) DeleteDeviceProfileByName(name string) errors.EdgeX {
	ctx := context.Background()

	queryObj := map[string]any{nameField: name}
	_, err := c.ConnPool.Exec(ctx, sqlDeleteByJSONField(deviceProfileTableName), queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete device profile by name %s", name), err)
	}
	return nil
}

// DeviceProfileNameExists checks the device profile exists by name
func (c *Client) DeviceProfileNameExists(name string) (bool, errors.EdgeX) {
	ctx := context.Background()
	return deviceProfileNameExists(ctx, c.ConnPool, name)
}

// AllDeviceProfiles query device profiles with offset, limit and labels
func (c *Client) AllDeviceProfiles(offset int, limit int, labels []string) (profiles []model.DeviceProfile, err errors.EdgeX) {
	ctx := context.Background()

	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	if len(labels) > 0 {
		c.loggingClient.Debugf("Querying device profiles by labels: %v", labels)
		queryObj := map[string]any{labelsField: labels}
		profiles, err = queryDeviceProfiles(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(deviceProfileTableName),
			pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
		if err != nil {
			return profiles, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all device profiles by labels", err)
		}
	} else {
		profiles, err = queryDeviceProfiles(ctx, c.ConnPool, sqlQueryContentWithPaginationAsNamedArgs(deviceProfileTableName),
			pgx.NamedArgs{offsetCondition: offset, limitCondition: validLimit})
		if err != nil {
			return profiles, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all device profiles", err)
		}
	}

	return profiles, nil
}

// DeviceProfilesByModel query device profiles with offset, limit and model
func (c *Client) DeviceProfilesByModel(offset int, limit int, model string) ([]model.DeviceProfile, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{modelField: model}
	return queryDeviceProfiles(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(deviceProfileTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
}

// DeviceProfilesByManufacturer query device profiles with offset, limit and manufacturer
func (c *Client) DeviceProfilesByManufacturer(offset int, limit int, manufacturer string) ([]model.DeviceProfile, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{manufacturerField: manufacturer}
	return queryDeviceProfiles(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(deviceProfileTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
}

// DeviceProfilesByManufacturerAndModel query device profiles with offset, limit, manufacturer and model
func (c *Client) DeviceProfilesByManufacturerAndModel(offset int, limit int, manufacturer string, model string) (profiles []model.DeviceProfile, err errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{modelField: model, manufacturerField: manufacturer}
	return queryDeviceProfiles(ctx, c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(deviceProfileTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
}

// DeviceProfileCountByLabels returns the total count of Device Profiles with labels specified.  If no label is specified, the total count of all device profiles will be returned.
func (c *Client) DeviceProfileCountByLabels(labels []string) (int64, errors.EdgeX) {
	ctx := context.Background()

	if len(labels) > 0 {
		queryObj := map[string]any{labelsField: labels}
		return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(deviceProfileTableName), queryObj)
	}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCount(deviceProfileTableName))
}

// DeviceProfileCountByManufacturer returns the count of Device Profiles associated with specified manufacturer
func (c *Client) DeviceProfileCountByManufacturer(manufacturer string) (int64, errors.EdgeX) {
	ctx := context.Background()
	queryObj := map[string]any{manufacturerField: manufacturer}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(deviceProfileTableName), queryObj)
}

// DeviceProfileCountByModel returns the count of Device Profiles associated with specified model
func (c *Client) DeviceProfileCountByModel(model string) (int64, errors.EdgeX) {
	ctx := context.Background()
	queryObj := map[string]any{modelField: model}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(deviceProfileTableName), queryObj)
}

// DeviceProfileCountByManufacturerAndModel returns the count of Device Profiles associated with specified manufacturer and model
func (c *Client) DeviceProfileCountByManufacturerAndModel(manufacturer, model string) (int64, errors.EdgeX) {
	ctx := context.Background()
	queryObj := map[string]any{manufacturerField: manufacturer, modelField: model}
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountByJSONField(deviceProfileTableName), queryObj)
}

// ResourceCount returns the total count of Resources
func (c *Client) InUseResourceCount() (int64, errors.EdgeX) {
	ctx := context.Background()
	return getTotalRowsCount(ctx, c.ConnPool, sqlQueryCountInUseResource())
}

func deviceProfileNameExists(ctx context.Context, connPool *pgxpool.Pool, name string) (bool, errors.EdgeX) {
	var exists bool
	queryObj := map[string]any{nameField: name}
	err := connPool.QueryRow(ctx, sqlCheckExistsByJSONField(deviceProfileTableName), queryObj).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError(fmt.Sprintf("failed to query device profile by name '%s' from %s table", name, deviceProfileTableName), err)
	}
	return exists, nil
}

func queryOneDeviceProfile(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (model.DeviceProfile, errors.EdgeX) {
	var dp model.DeviceProfile
	row := connPool.QueryRow(ctx, sql, args...)

	if err := row.Scan(&dp); err != nil {
		return dp, pgClient.WrapDBError("failed to query device profile", err)
	}
	return dp, nil
}

func queryDeviceProfiles(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]model.DeviceProfile, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query device profiles", err)
	}

	profiles, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.DeviceProfile, error) {
		var dp model.DeviceProfile
		scanErr := row.Scan(&dp)
		return dp, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to DeviceProfile model", err)
	}

	return profiles, nil
}
