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
	"github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/gomodule/redigo/redis"
)

const DeviceServiceCollection = "v2:deviceService"

// deviceServiceStoredKey return the device service's stored key which combines the collection name and object id
func deviceServiceStoredKey(id string) string {
	return fmt.Sprintf("%s:%s", DeviceServiceCollection, id)
}

// deviceServiceNameExist whether the device service exists by name
func deviceServiceNameExist(conn redis.Conn, name string) (bool, errors.EdgeX) {
	exists, err := objectNameExists(conn, fmt.Sprintf("%s:%s", DeviceServiceCollection, v2.Name), name)
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	return exists, nil
}

// addDeviceService adds a new device service into DB
func addDeviceService(conn redis.Conn, ds models.DeviceService) (addedDeviceService models.DeviceService, edgeXerr errors.EdgeX) {
	// retrieve Device Service by Id first to ensure there is no Id conflict; when Id exists, return duplicate error
	exists, edgeXerr := objectIdExists(conn, deviceServiceStoredKey(ds.Id))
	if edgeXerr != nil {
		return addedDeviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return addedDeviceService, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device service id %s already exists", ds.Id), edgeXerr)
	}

	// verify if device service name is unique or not
	exists, edgeXerr = objectNameExists(conn, fmt.Sprintf("%s:%s", DeviceServiceCollection, v2.Name), ds.Name)
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
	// query API will sort the result based on Modified, so even newly created device service shall specify Modified as Created
	ds.Modified = ds.Created

	dsJSONBytes, err := json.Marshal(ds)
	if err != nil {
		return addedDeviceService, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device service for Redis persistence", err)
	}

	// redisKey represents the key stored in the redis, use the format of #{DeviceServiceCollection}:#{ds.Id}
	// as the redisKey to avoid data being accidentally deleted when other objects, e.g. device profiles, also
	// coincidentally have the same Id.
	redisKey := deviceServiceStoredKey(ds.Id)
	_ = conn.Send(MULTI)
	// Set the redisKey to associate with object byte array for later retrieval
	_ = conn.Send(SET, redisKey, dsJSONBytes)
	// Store the redisKey into a Sorted Set with Modified as the score for order
	_ = conn.Send(ZADD, DeviceServiceCollection, ds.Modified, redisKey)
	// Store the ds.Name into a Hash for later Name existence check
	_ = conn.Send(HSET, fmt.Sprintf("%s:%s", DeviceServiceCollection, v2.Name), ds.Name, redisKey)
	for _, label := range ds.Labels { // Store the redisKey into Sorted Set of labels with Modified as the score for order
		_ = conn.Send(ZADD, fmt.Sprintf("%s:%s:%s", DeviceServiceCollection, v2.Label, label), ds.Modified, redisKey)
	}
	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "device service creation failed", err)
	}

	return ds, edgeXerr
}

// deviceServiceById query device service by id from DB
func deviceServiceById(conn redis.Conn, id string) (deviceService models.DeviceService, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectById(conn, deviceServiceStoredKey(id), &deviceService)
	if edgeXerr != nil {
		return deviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return
}

// deviceServiceByName query device service by name from DB
func deviceServiceByName(conn redis.Conn, name string) (deviceService models.DeviceService, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectByHash(conn, fmt.Sprintf("%s:%s", DeviceServiceCollection, v2.Name), name, &deviceService)
	if edgeXerr != nil {
		return deviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return
}

func deleteDeviceService(conn redis.Conn, deviceService models.DeviceService) errors.EdgeX {
	storedKey := deviceServiceStoredKey(deviceService.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, DeviceServiceCollection, storedKey)
	_ = conn.Send(HDEL, fmt.Sprintf("%s:%s", DeviceServiceCollection, v2.Name), deviceService.Name)
	for _, label := range deviceService.Labels {
		_ = conn.Send(ZREM, fmt.Sprintf("%s:label:%s", DeviceServiceCollection, label), storedKey)
	}

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

// deleteDeviceServiceByName deletes the device service by name
func deleteDeviceServiceByName(conn redis.Conn, name string) errors.EdgeX {
	deviceService, err := deviceServiceByName(conn, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = deleteDeviceService(conn, deviceService)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// deviceServicesByLabels query multiple device services from DB per labels
func deviceServicesByLabels(conn redis.Conn, offset int, limit int, labels []string) (deviceServices []models.DeviceService, edgeXerr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, err := getObjectsByLabelsAndSomeRange(conn, ZREVRANGE, DeviceServiceCollection, labels, offset, end)
	if err != nil {
		return deviceServices, errors.NewCommonEdgeXWrapper(err)
	}

	deviceServices = make([]models.DeviceService, len(objects))
	for i, in := range objects {
		s := models.DeviceService{}
		err := json.Unmarshal(in, &s)
		if err != nil {
			return []models.DeviceService{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "device service format parsing failed from the database", err)
		}
		deviceServices[i] = s
	}
	return deviceServices, nil
}
