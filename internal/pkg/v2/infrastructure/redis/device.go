//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/common"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/gomodule/redigo/redis"
)

const DeviceCollection = "v2:device"

// deviceStoredKey return the device's stored key which combines the collection name and object id
func deviceStoredKey(id string) string {
	return fmt.Sprintf("%s:%s", DeviceCollection, id)
}

// deviceNameExists whether the device exists by name
func deviceNameExists(conn redis.Conn, name string) (bool, errors.EdgeX) {
	exists, err := objectNameExists(conn, DeviceCollection+":name", name)
	if err != nil {
		return false, errors.NewCommonEdgeX(errors.KindDatabaseError, "device existence check by name failed", err)
	}
	return exists, nil
}

// deviceIdExists checks whether the device exists by id
func deviceIdExists(conn redis.Conn, id string) (bool, errors.EdgeX) {
	exists, err := objectIdExists(conn, deviceStoredKey(id))
	if err != nil {
		return false, errors.NewCommonEdgeX(errors.KindDatabaseError, "device existence check by id failed", err)
	}
	return exists, nil
}

// addDevice adds a new device into DB
func addDevice(conn redis.Conn, d model.Device) (model.Device, errors.EdgeX) {
	exists, edgeXerr := deviceIdExists(conn, d.Id)
	if edgeXerr != nil {
		return d, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return d, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device id %s already exists", d.Id), edgeXerr)
	}

	exists, edgeXerr = deviceNameExists(conn, d.Name)
	if edgeXerr != nil {
		return d, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return d, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device name %s already exists", d.Name), edgeXerr)
	}

	ts := common.MakeTimestamp()
	if d.Created == 0 {
		d.Created = ts
	}
	d.Modified = ts

	dsJSONBytes, err := json.Marshal(d)
	if err != nil {
		return d, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device for Redis persistence", err)
	}

	storedKey := deviceStoredKey(d.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(SET, storedKey, dsJSONBytes)
	_ = conn.Send(ZADD, DeviceCollection, 0, storedKey)
	_ = conn.Send(HSET, fmt.Sprintf("%s:name", DeviceCollection), d.Name, storedKey)
	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "device creation failed", err)
	}

	return d, edgeXerr
}
