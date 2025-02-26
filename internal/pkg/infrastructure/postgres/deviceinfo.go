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

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/postgres/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

func (c *Client) deviceInfoIdByEvent(e model.Event) (int, errors.EdgeX) {
	deviceInfo := models.DeviceInfo{
		DeviceName:  e.DeviceName,
		ProfileName: e.ProfileName,
		SourceName:  e.SourceName,
		Tags:        e.Tags,
	}
	return c.deviceInfoIdFromCache(deviceInfo)
}

func (c *Client) deviceInfoIdByReading(r models.Reading) (int, errors.EdgeX) {
	deviceInfo := models.DeviceInfo{
		DeviceName:   r.DeviceName,
		ProfileName:  r.ProfileName,
		ResourceName: r.ResourceName,
		ValueType:    r.ValueType,
		Unit:         r.Units,
		Tags:         r.Tags,
	}
	if r.MediaType != nil {
		deviceInfo.MediaType = *r.MediaType
	}
	return c.deviceInfoIdFromCache(deviceInfo)
}

func (c *Client) deviceInfoIdFromCache(deviceInfo models.DeviceInfo) (int, errors.EdgeX) {
	var exists bool
	var err error
	// get deviceInfo id from the cache
	id, exists := c.deviceInfoIdCache.Get(deviceInfo)
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
		c.deviceInfoIdCache.Add(deviceInfo)
	}
	return id, nil
}

func (c *Client) deviceInfoId(deviceInfo models.DeviceInfo) (int, errors.EdgeX) {
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
		deviceInfo.ResourceName, deviceInfo.ValueType, deviceInfo.Unit, deviceInfo.MediaType,
	)

	if err := row.Scan(&id); err != nil {
		return 0, pgClient.WrapDBError("failed to query deviceInfo", err)
	}
	return id, nil
}

// addDeviceInfo adds a new deviceInfo
func (c *Client) addDeviceInfo(deviceInfo models.DeviceInfo) (int, errors.EdgeX) {
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
		deviceInfo.ResourceName, deviceInfo.ValueType, deviceInfo.Unit, deviceInfo.MediaType,
	).Scan(&id)
	if err != nil {
		return 0, pgClient.WrapDBError("failed to insert deviceInfo", err)
	}
	return id, nil
}

// RemoveDeviceInfosByDeviceName removes deviceInfos by the deviceName
func (c *Client) RemoveDeviceInfosByDeviceName(deviceName string) errors.EdgeX {
	ctx := context.Background()
	sqlStmt := sqlQueryAllWithConds(deviceInfoTableName, deviceNameCol)
	rows, err := c.ConnPool.Query(ctx, sqlStmt, deviceName)
	if err != nil {
		return pgClient.WrapDBError("failed to query deviceInfos", err)
	}
	deviceInfos, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.DeviceInfo, error) {
		var d models.DeviceInfo
		scanErr := row.Scan(&d)
		return d, scanErr
	})
	if err != nil {
		return pgClient.WrapDBError("failed to collect rows to DeviceInfo model", err)
	}

	for _, deviceInfo := range deviceInfos {
		err = c.deleteDeviceInfoById(deviceInfo.Id)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		c.deviceInfoIdCache.Remove(deviceInfo)
	}
	return nil
}

// deleteDeviceInfoById deletes a deviceInfo by id
func (c *Client) deleteDeviceInfoById(id int) errors.EdgeX {
	ctx := context.Background()

	_, err := c.ConnPool.Exec(ctx, sqlDeleteById(deviceInfoTableName), id)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete deviceInfo by id %d", id), err)
	}
	return nil
}
