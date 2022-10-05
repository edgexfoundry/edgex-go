//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/gomodule/redigo/redis"
)

const (
	DeviceServiceCollection      = "md|ds"
	DeviceServiceCollectionName  = DeviceServiceCollection + DBKeySeparator + common.Name
	DeviceServiceCollectionLabel = DeviceServiceCollection + DBKeySeparator + common.Label
)

// deviceServiceStoredKey return the device service's stored key which combines the collection name and object id
func deviceServiceStoredKey(id string) string {
	return CreateKey(DeviceServiceCollection, id)
}

// deviceServiceNameExist whether the device service exists by name
func deviceServiceNameExist(conn redis.Conn, name string) (bool, errors.EdgeX) {
	exists, err := objectNameExists(conn, DeviceServiceCollectionName, name)
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	return exists, nil
}

// sendAddDeviceServiceCmd send redis command for adding device service
func sendAddDeviceServiceCmd(conn redis.Conn, storedKey string, ds models.DeviceService) errors.EdgeX {
	m, err := json.Marshal(ds)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device service for Redis persistence", err)
	}
	// Set the storedKey to associate with object byte array for later retrieval
	_ = conn.Send(SET, storedKey, m)
	// Store the storedKey into a Sorted Set with Modified as the score for order
	_ = conn.Send(ZADD, DeviceServiceCollection, ds.Modified, storedKey)
	// Store the ds.Name into a Hash for later Name existence check
	_ = conn.Send(HSET, DeviceServiceCollectionName, ds.Name, storedKey)
	for _, label := range ds.Labels { // Store the redisKey into Sorted Set of labels with Modified as the score for order
		_ = conn.Send(ZADD, CreateKey(DeviceServiceCollectionLabel, label), ds.Modified, storedKey)
	}
	return nil
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
	exists, edgeXerr = objectNameExists(conn, DeviceServiceCollectionName, ds.Name)
	if edgeXerr != nil {
		return addedDeviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return addedDeviceService, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device service name %s already exists", ds.Name), edgeXerr)
	}

	// For Redis DB, the PUT or PATCH operation will removes the old object and add the modified one,
	// so the Created is not zero value and we shouldn't set the timestamp again.
	if ds.Created == 0 {
		ds.Created = pkgCommon.MakeTimestamp()
	}
	// query API will sort the result based on Modified, so even newly created device service shall specify Modified as Created
	ds.Modified = ds.Created

	// storedKey represents the key stored in the redis, use the format of #{DeviceServiceCollection}:#{ds.Id}
	// as the redisKey to avoid data being accidentally deleted when other objects, e.g. device profiles, also
	// coincidentally have the same Id.
	storedKey := deviceServiceStoredKey(ds.Id)
	_ = conn.Send(MULTI)
	edgeXerr = sendAddDeviceServiceCmd(conn, storedKey, ds)
	if edgeXerr != nil {
		return ds, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "device service creation failed", err)
	}

	return ds, edgeXerr
}

// deviceServiceById query device service by id from DB
func deviceServiceById(conn redis.Conn, id string) (deviceService models.DeviceService, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectById(conn, deviceServiceStoredKey(id), &deviceService)
	if edgeXerr != nil {
		return deviceService, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to query device service by id %s", id), edgeXerr)
	}
	return
}

// deviceServiceByName query device service by name from DB
func deviceServiceByName(conn redis.Conn, name string) (deviceService models.DeviceService, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectByHash(conn, DeviceServiceCollectionName, name, &deviceService)
	if edgeXerr != nil {
		return deviceService, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to query device service by name %s", name), edgeXerr)
	}
	return
}

// sendDeleteDeviceServiceCmd send redis command for deleting device service
func sendDeleteDeviceServiceCmd(conn redis.Conn, storedKey string, ds models.DeviceService) {
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, DeviceServiceCollection, storedKey)
	_ = conn.Send(HDEL, DeviceServiceCollectionName, ds.Name)
	for _, label := range ds.Labels {
		_ = conn.Send(ZREM, CreateKey(DeviceServiceCollectionLabel, label), storedKey)
	}
}

func deleteDeviceService(conn redis.Conn, ds models.DeviceService) errors.EdgeX {
	storedKey := deviceServiceStoredKey(ds.Id)
	_ = conn.Send(MULTI)
	sendDeleteDeviceServiceCmd(conn, storedKey, ds)
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

	// Check the associated Device and ProvisionWatcher existence
	devices, err := devicesByServiceName(conn, 0, 1, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if len(devices) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to delete the device service when associated device exists", nil)
	}
	provisionWatchers, err := provisionWatchersByServiceName(conn, 0, 1, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if len(provisionWatchers) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to delete the device service when associated provisionWatcher exists", nil)
	}

	err = deleteDeviceService(conn, deviceService)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// deviceServicesByLabels query multiple device services from DB per labels
func deviceServicesByLabels(conn redis.Conn, offset int, limit int, labels []string) (deviceServices []models.DeviceService, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByLabelsAndSomeRange(conn, ZREVRANGE, DeviceServiceCollection, labels, offset, limit)
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

func updateDeviceService(conn redis.Conn, ds models.DeviceService) errors.EdgeX {
	oldDeviceService, edgeXerr := deviceServiceByName(conn, ds.Name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	ds.Modified = pkgCommon.MakeTimestamp()
	storedKey := deviceServiceStoredKey(ds.Id)
	_ = conn.Send(MULTI)
	sendDeleteDeviceServiceCmd(conn, storedKey, oldDeviceService)
	edgeXerr = sendAddDeviceServiceCmd(conn, storedKey, ds)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "device service update failed", err)
	}

	return nil
}
