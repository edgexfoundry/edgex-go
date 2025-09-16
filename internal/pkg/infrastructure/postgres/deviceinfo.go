//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
	dbModels "github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/models"
	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/postgres/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// AllDeviceInfos query all deviceInfos
func (c *Client) AllDeviceInfos(offset int, limit int) ([]dbModels.DeviceInfo, errors.EdgeX) {
	sqlStmt := sqlQueryAllWithPaginationAsNamedArgs(deviceInfoTableName)
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	rows, err := c.ConnPool.Query(context.Background(), sqlStmt, pgx.NamedArgs{offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query deviceInfos", err)
	}
	deviceInfos, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (dbModels.DeviceInfo, error) {
		d, scanErr := pgx.RowToStructByNameLax[dbModels.DeviceInfo](row)
		if err != nil {
			return dbModels.DeviceInfo{}, err
		}
		return d, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to DeviceInfo model", err)
	}
	return deviceInfos, nil
}

func (c *Client) deviceInfoIdByEvent(e model.Event) (int, errors.EdgeX) {
	deviceInfo := dbModels.DeviceInfo{
		DeviceName:  e.DeviceName,
		ProfileName: e.ProfileName,
		SourceName:  e.SourceName,
		Tags:        e.Tags,
	}
	return c.deviceInfoIdFromCache(deviceInfo)
}

func (c *Client) deviceInfoIdByReading(r models.Reading) (int, errors.EdgeX) {
	deviceInfo := dbModels.DeviceInfo{
		DeviceName:   r.DeviceName,
		ProfileName:  r.ProfileName,
		ResourceName: r.ResourceName,
		ValueType:    r.ValueType,
		Units:        r.Units,
		Tags:         r.Tags,
	}
	if r.MediaType != nil {
		deviceInfo.MediaType = *r.MediaType
	}
	return c.deviceInfoIdFromCache(deviceInfo)
}

func (c *Client) deviceInfoIdFromCache(deviceInfo dbModels.DeviceInfo) (int, errors.EdgeX) {
	deviceInfoCache := container.DeviceInfoCacheFrom(c.dic.Get)
	var exists bool
	var err error
	// get deviceInfo id from the cache
	id, exists := deviceInfoCache.GetDeviceInfoId(deviceInfo)
	if !exists {
		// get deviceInfo id from the DB
		id, err = c.deviceInfoId(deviceInfo)
		if err != nil && errors.Kind(err) == errors.KindEntityDoesNotExist {
			// add new deviceInfo if not exist in the DB
			id, err = c.addDeviceInfo(deviceInfo)
			if err != nil {
				return 0, errors.NewCommonEdgeXWrapper(err)
			}
			deviceInfo.Id = id

		} else if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
		deviceInfo.Id = id
		// keep in the cache
		deviceInfoCache.Add(deviceInfo)
	}
	return id, nil
}

func (c *Client) deviceInfoId(deviceInfo dbModels.DeviceInfo) (int, errors.EdgeX) {
	var id int
	tagsBytes, err := json.Marshal(deviceInfo.Tags)
	if err != nil {
		return 0, errors.NewCommonEdgeX(errors.KindServerError, "unable to JSON marshal deviceInfo tags", err)
	}
	row := c.ConnPool.QueryRow(
		context.Background(),
		sqlQueryFieldsByCol(
			deviceInfoTableName,
			[]string{idCol},
			deviceNameCol, profileNameCol, sourceNameCol, tagsCol,
			resourceNameCol, valueTypeCol, unitsCol, mediaTypeCol),
		deviceInfo.DeviceName, deviceInfo.ProfileName, deviceInfo.SourceName, tagsBytes,
		deviceInfo.ResourceName, deviceInfo.ValueType, deviceInfo.Units, deviceInfo.MediaType,
	)

	if err := row.Scan(&id); err != nil {
		return 0, pgClient.WrapDBError("failed to query deviceInfo", err)
	}
	return id, nil
}

// addDeviceInfo adds a new deviceInfo
func (c *Client) addDeviceInfo(deviceInfo dbModels.DeviceInfo) (int, errors.EdgeX) {
	tagsBytes, err := json.Marshal(deviceInfo.Tags)
	if err != nil {
		return 0, errors.NewCommonEdgeX(errors.KindServerError, "unable to JSON marshal deviceInfo tags", err)
	}

	var id int
	err = c.ConnPool.QueryRow(context.Background(),
		sqlInsert(
			deviceInfoTableName, deviceNameCol, profileNameCol, sourceNameCol, tagsCol,
			resourceNameCol, valueTypeCol, unitsCol, mediaTypeCol,
		)+" RETURNING id",
		deviceInfo.DeviceName, deviceInfo.ProfileName, deviceInfo.SourceName, tagsBytes,
		deviceInfo.ResourceName, deviceInfo.ValueType, deviceInfo.Units, deviceInfo.MediaType,
	).Scan(&id)
	if err != nil {
		return 0, pgClient.WrapDBError("failed to insert deviceInfo", err)
	}
	return id, nil
}

// deviceInfosByConds query deviceInfos by specified conditions
func (c *Client) deviceInfosByConds(cols []string, values pgx.NamedArgs) ([]dbModels.DeviceInfo, errors.EdgeX) {
	sqlStmt := sqlQueryAllWithNamedArgConds(deviceInfoTableName, cols...)
	rows, err := c.ConnPool.Query(context.Background(), sqlStmt, values)
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query deviceInfos", err)
	}
	deviceInfos, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (dbModels.DeviceInfo, error) {
		d, scanErr := pgx.RowToStructByNameLax[dbModels.DeviceInfo](row)
		if err != nil {
			return dbModels.DeviceInfo{}, err
		}
		return d, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to DeviceInfo model", err)
	}
	return deviceInfos, nil
}

// deleteDeviceInfoById deletes a deviceInfo by id
// only calling this func when corresponding events/readings have been removed, otherwise may
// potentially break referential integrity between events/readings/device_info records
func deleteDeviceInfoById(ctx context.Context, tx pgx.Tx, id int) errors.EdgeX {
	_, err := tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = @%s", deviceInfoTableName, idCol, idCol), pgx.NamedArgs{idCol: id})
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete deviceInfo by id %d", id), err)
	}
	return nil
}

// updateDeviceInfosDeletableByDeviceName updates deviceInfos by deviceName as deletable
func (c *Client) updateDeviceInfosDeletableByDeviceName(deviceName string) errors.EdgeX {
	_, err := c.ConnPool.Exec(context.Background(), fmt.Sprintf("UPDATE %s SET %s = true WHERE %s=$1", deviceInfoTableName, markDeletedCol, deviceNameCol), deviceName)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("update %s to deletable by devicename '%s'", deviceInfoTableName, deviceName), err)
	}
	return nil
}
