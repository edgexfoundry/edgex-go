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

const DeviceProfileCollection = "v2:deviceProfile"

// deviceProfileStoredKey return the device profile's stored key which combines the collection name and object id
func deviceProfileStoredKey(id string) string {
	return fmt.Sprintf("%s:%s", DeviceProfileCollection, id)
}

// deviceProfileExistByName whether the device profile exists by name
func deviceProfileExistByName(conn redis.Conn, name string) (bool, errors.EdgeX) {
	exists, err := redis.Bool(conn.Do(HEXISTS, DeviceProfileCollection+":name", name))
	if err != nil {
		return false, errors.NewCommonEdgeX(errors.KindDatabaseError, "device profile existence check by name failed", err)
	} else if exists {
		return true, nil
	}
	return false, nil
}

// deviceProfileExistById checks whether the device profile exists by id
func deviceProfileExistById(conn redis.Conn, id string) (bool, errors.EdgeX) {
	exists, err := redis.Bool(conn.Do(EXISTS, deviceProfileStoredKey(id)))
	if err != nil {
		return false, errors.NewCommonEdgeX(errors.KindDatabaseError, "device profile existence check by id failed", err)
	} else if exists {
		return true, nil
	}
	return false, nil
}

// addDeviceProfile adds a device profile to DB
func addDeviceProfile(conn redis.Conn, dp model.DeviceProfile) (addedDeviceProfile model.DeviceProfile, edgeXerr errors.EdgeX) {
	// query device profile name and id to avoid the conflict
	exists, edgeXerr := deviceProfileExistById(conn, dp.Id)
	if edgeXerr != nil {
		return addedDeviceProfile, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return addedDeviceProfile, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile id %s exists", dp.Id), edgeXerr)
	}

	exists, edgeXerr = deviceProfileExistByName(conn, dp.Name)
	if edgeXerr != nil {
		return addedDeviceProfile, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return addedDeviceProfile, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile name %s exists", dp.Name), edgeXerr)
	}

	if dp.Created == 0 {
		dp.Created = common.MakeTimestamp()
	}

	m, err := json.Marshal(dp)
	if err != nil {
		return addedDeviceProfile, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device profile for Redis persistence", err)
	}

	storedKey := deviceProfileStoredKey(dp.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, DeviceProfileCollection, 0, storedKey)
	_ = conn.Send(HSET, DeviceProfileCollection+":name", dp.Name, storedKey)
	_ = conn.Send(SADD, DeviceProfileCollection+":manufacturer:"+dp.Manufacturer, storedKey)
	_ = conn.Send(SADD, DeviceProfileCollection+":model:"+dp.Model, storedKey)

	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "device profile creation failed", err)
	}

	return dp, edgeXerr
}
