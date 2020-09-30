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

// addDeviceProfile adds a device profile to DB
func addDeviceProfile(conn redis.Conn, dp model.DeviceProfile) (addedDeviceProfile model.DeviceProfile, edgeXerr errors.EdgeX) {
	exists, err := redis.Bool(conn.Do("HEXISTS", DeviceProfileCollection+":name", dp.Name))
	if err != nil {
		return addedDeviceProfile, errors.NewCommonEdgeX(errors.KindDatabaseError, "device profile existence check failed", err)
	} else if exists {
		return addedDeviceProfile, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device profile %s already existence", dp.Name), err)
	}

	if dp.Created == 0 {
		dp.Created = common.MakeTimestamp()
	}

	m, err := json.Marshal(dp)
	if err != nil {
		return addedDeviceProfile, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal device profile for Redis persistence", err)
	}

	storeKey := fmt.Sprintf("%s:%s", DeviceProfileCollection, dp.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(SET, storeKey, m)
	_ = conn.Send(ZADD, DeviceProfileCollection, 0, storeKey)
	_ = conn.Send(HSET, DeviceProfileCollection+":name", dp.Name, storeKey)
	_ = conn.Send(SADD, DeviceProfileCollection+":manufacturer:"+dp.Manufacturer, storeKey)
	_ = conn.Send(SADD, DeviceProfileCollection+":model:"+dp.Model, storeKey)

	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "device profile creation failed", err)
	}

	return dp, edgeXerr
}
