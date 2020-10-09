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

// addDeviceService adds a new device service into DB
func addDeviceService(conn redis.Conn, ds model.DeviceService) (addedDeviceService model.DeviceService, edgeXerr errors.EdgeX) {
	// retrieve Device Service by Id first to ensure there is no Id conflict; when Id exists, return duplicate error
	exists, err := redis.Bool(conn.Do(EXISTS, fmt.Sprintf("%s:%s", DeviceServiceCollection, ds.Id)))
	if err != nil {
		return addedDeviceService, errors.NewCommonEdgeX(errors.KindDatabaseError, "device service Id existence check failed", err)
	} else if exists {
		return addedDeviceService, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device service id %s already exists", ds.Id), err)
	}

	// verify if device service name is unique or not
	exists, err = redis.Bool(conn.Do(HEXISTS, fmt.Sprintf("%s:name", DeviceServiceCollection), ds.Name))
	if err != nil {
		return addedDeviceService, errors.NewCommonEdgeX(errors.KindDatabaseError, "device service name existence check failed", err)
	} else if exists {
		return addedDeviceService, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device service name %s already exists", ds.Name), err)
	}

	if ds.Created == 0 {
		ds.Created = common.MakeTimestamp()
	}

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

// deviceServiceByName query device service by name from DB
func deviceServiceByName(conn redis.Conn, name string) (deviceService model.DeviceService, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectByHash(conn, DeviceServiceCollection+":name", name, &deviceService)
	if edgeXerr != nil {
		return deviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return
}
