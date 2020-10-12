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

const DeviceServiceCollection = "v2:deviceService"

// deviceServiceStoredKey return the device service's stored key which combines the collection name and object id
func deviceServiceStoredKey(id string) string {
	return fmt.Sprintf("%s:%s", DeviceServiceCollection, id)
}

// addDeviceService adds a new device service into DB
func addDeviceService(conn redis.Conn, ds model.DeviceService) (addedDeviceService model.DeviceService, edgeXerr errors.EdgeX) {
	// retrieve Device Service by Id first to ensure there is no Id conflict; when Id exists, return duplicate error
	exists, edgeXerr := objectExistById(conn, deviceServiceStoredKey(ds.Id))
	if edgeXerr != nil {
		return addedDeviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return addedDeviceService, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device service id %s already exists", ds.Id), edgeXerr)
	}

	// verify if device service name is unique or not
	exists, edgeXerr = objectExistByHash(conn, fmt.Sprintf("%s:name", DeviceServiceCollection), ds.Name)
	if edgeXerr != nil {
		return addedDeviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return addedDeviceService, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device service name %s already exists", ds.Name), edgeXerr)
	}

	ts := common.MakeTimestamp()
	// For Redis DB, the PUT or PATCH operation will removes the old object and add the modified one,
	// so the Created is not zero value and we shouldn't set the timestamp again.
	if ds.Created == 0 {
		ds.Created = ts
	}
	ds.Modified = ts

	dsJSONBytes, err := json.Marshal(ds)
	if err != nil {
		return addedDeviceService, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device service for Redis persistence", err)
	}

	// redisKey represents the key stored in the redis, use the format of #{DeviceServiceCollection}:#{ds.Id}
	// as the redisKey to avoid data being accidentally deleted when other objects, e.g. device profiles, also
	// coincidentally have the same Id.
	redisKey := fmt.Sprintf("%s:%s", DeviceServiceCollection, ds.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(SET, redisKey, dsJSONBytes)
	_ = conn.Send(ZADD, DeviceServiceCollection, 0, redisKey)
	_ = conn.Send(HSET, fmt.Sprintf("%s:name", DeviceServiceCollection), ds.Name, redisKey)
	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "device service creation failed", err)
	}

	return ds, edgeXerr
}

// deviceServiceById query device service by id from DB
func deviceServiceById(conn redis.Conn, id string) (deviceService model.DeviceService, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectById(conn, deviceServiceStoredKey(id), &deviceService)
	if edgeXerr != nil {
		return deviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return
}

// deviceServiceByName query device service by name from DB
func deviceServiceByName(conn redis.Conn, name string) (deviceService model.DeviceService, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectByHash(conn, DeviceServiceCollection+":name", name, &deviceService)
	if edgeXerr != nil {
		return deviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return
}

func deleteDeviceService(conn redis.Conn, deviceService model.DeviceService) errors.EdgeX {
	storedKey := deviceServiceStoredKey(deviceService.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, DeviceServiceCollection, storedKey)
	_ = conn.Send(HDEL, DeviceServiceCollection+":name", deviceService.Name)
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "device service deletion failed", err)
	}
	return nil
}

// deleteDeviceServiceById deletes the device service by id
func deleteDeviceServiceById(conn redis.Conn, id string) errors.EdgeX {
	deviceService, err := deviceServiceById(conn, id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = deleteDeviceService(conn, deviceService)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}
