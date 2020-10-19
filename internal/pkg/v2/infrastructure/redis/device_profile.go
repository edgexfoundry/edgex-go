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

const DeviceProfileCollection = "v2:deviceProfile"

// deviceProfileStoredKey return the device profile's stored key which combines the collection name and object id
func deviceProfileStoredKey(id string) string {
	return fmt.Sprintf("%s:%s", DeviceProfileCollection, id)
}

// deviceProfileNameExists whether the device profile exists by name
func deviceProfileNameExists(conn redis.Conn, name string) (bool, errors.EdgeX) {
	exists, err := objectNameExists(conn, DeviceProfileCollection+":name", name)
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	return exists, nil
}

// deviceProfileIdExists checks whether the device profile exists by id
func deviceProfileIdExists(conn redis.Conn, id string) (bool, errors.EdgeX) {
	exists, err := objectIdExists(conn, deviceProfileStoredKey(id))
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	return exists, nil
}

// addDeviceProfile adds a device profile to DB
func addDeviceProfile(conn redis.Conn, dp models.DeviceProfile) (addedDeviceProfile models.DeviceProfile, edgeXerr errors.EdgeX) {
	// query device profile name and id to avoid the conflict
	exists, edgeXerr := deviceProfileIdExists(conn, dp.Id)
	if edgeXerr != nil {
		return addedDeviceProfile, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return addedDeviceProfile, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile id %s exists", dp.Id), edgeXerr)
	}

	exists, edgeXerr = deviceProfileNameExists(conn, dp.Name)
	if edgeXerr != nil {
		return addedDeviceProfile, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return addedDeviceProfile, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile name %s exists", dp.Name), edgeXerr)
	}

	ts := common.MakeTimestamp()
	// For Redis DB, the PUT or PATCH operation will removes the old object and add the modified one,
	// so the Created is not zero value and we shouldn't set the timestamp again.
	if dp.Created == 0 {
		dp.Created = ts
	}
	dp.Modified = ts

	m, err := json.Marshal(dp)
	if err != nil {
		return addedDeviceProfile, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device profile for Redis persistence", err)
	}

	storedKey := deviceProfileStoredKey(dp.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, DeviceProfileCollection, 0, storedKey)
	_ = conn.Send(HSET, fmt.Sprintf("%s:%s", DeviceProfileCollection, v2.Name), dp.Name, storedKey)
	_ = conn.Send(SADD, fmt.Sprintf("%s:%s:%s", DeviceProfileCollection, v2.Manufacturer, dp.Manufacturer), storedKey)
	_ = conn.Send(SADD, fmt.Sprintf("%s:%s:%s", DeviceProfileCollection, v2.Model, dp.Model), storedKey)
	for _, label := range dp.Labels {
		_ = conn.Send(ZADD, fmt.Sprintf("%s:%s:%s", DeviceProfileCollection, v2.Label, label), dp.Modified, storedKey)
	}

	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "device profile creation failed", err)
	}

	return dp, edgeXerr
}

// deviceProfileById query device profile by id from DB
func deviceProfileById(conn redis.Conn, id string) (deviceProfile models.DeviceProfile, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectById(conn, deviceProfileStoredKey(id), &deviceProfile)
	if edgeXerr != nil {
		return deviceProfile, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return
}

// deviceProfileByName query device profile by name from DB
func deviceProfileByName(conn redis.Conn, name string) (deviceProfile models.DeviceProfile, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectByHash(conn, fmt.Sprintf("%s:%s", DeviceProfileCollection, v2.Name), name, &deviceProfile)
	if edgeXerr != nil {
		return deviceProfile, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return
}

func deleteDeviceProfile(conn redis.Conn, dp models.DeviceProfile) errors.EdgeX {
	storedKey := deviceProfileStoredKey(dp.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, DeviceProfileCollection, storedKey)
	_ = conn.Send(HDEL, fmt.Sprintf("%s:%s", DeviceProfileCollection, v2.Name), dp.Name)
	_ = conn.Send(SREM, fmt.Sprintf("%s:%s:%s", DeviceProfileCollection, v2.Manufacturer, dp.Manufacturer), storedKey)
	_ = conn.Send(SREM, fmt.Sprintf("%s:%s:%s", DeviceProfileCollection, v2.Model, dp.Model), storedKey)
	for _, label := range dp.Labels {
		_ = conn.Send(ZREM, fmt.Sprintf("%s:%s:%s", DeviceProfileCollection, v2.Label, label), storedKey)
	}

	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "device profile deletion failed", err)
	}
	return nil
}

// updateDeviceProfile updates a device profile to DB
func updateDeviceProfile(conn redis.Conn, dp models.DeviceProfile) (edgeXerr errors.EdgeX) {
	var oldDeviceProfile models.DeviceProfile
	oldDeviceProfile, edgeXerr = deviceProfileById(conn, dp.Id)
	if edgeXerr == nil {
		if dp.Name != oldDeviceProfile.Name {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device profile name '%s' not match the exsting '%s' ", dp.Name, oldDeviceProfile.Name), nil)
		}
	} else {
		oldDeviceProfile, edgeXerr = deviceProfileByName(conn, dp.Name)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	}

	edgeXerr = deleteDeviceProfile(conn, oldDeviceProfile)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	// Add new one
	dp.Id = oldDeviceProfile.Id
	dp.Created = oldDeviceProfile.Created
	dp.Modified = common.MakeTimestamp()
	_, edgeXerr = addDeviceProfile(conn, dp)
	if edgeXerr != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "device profile updating failed", edgeXerr)
	}

	return edgeXerr
}

// deleteDeviceProfileById deletes the device profile by id
func deleteDeviceProfileById(conn redis.Conn, id string) errors.EdgeX {
	deviceProfile, err := deviceProfileById(conn, id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = deleteDeviceProfile(conn, deviceProfile)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// deleteDeviceProfileByName deletes the device profile by name
func deleteDeviceProfileByName(conn redis.Conn, name string) errors.EdgeX {
	deviceProfile, err := deviceProfileByName(conn, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = deleteDeviceProfile(conn, deviceProfile)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// deviceProfilesByLabels query device profile with offset and limit
func deviceProfilesByLabels(conn redis.Conn, offset int, limit int, labels []string) (deviceProfiles []models.DeviceProfile, edgeXerr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, edgeXerr := getObjectsByLabelsAndSomeRange(conn, ZREVRANGE, DeviceProfileCollection, labels, offset, end)
	if edgeXerr != nil {
		return deviceProfiles, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	deviceProfiles = make([]models.DeviceProfile, len(objects))
	for i, in := range objects {
		dp := models.DeviceProfile{}
		err := json.Unmarshal(in, &dp)
		if err != nil {
			return deviceProfiles, errors.NewCommonEdgeX(errors.KindContractInvalid, "device profile parsing failed", err)
		}
		deviceProfiles[i] = dp
	}
	return deviceProfiles, nil
}
