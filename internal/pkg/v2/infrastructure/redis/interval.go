//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/common"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/gomodule/redigo/redis"
)

const (
	IntervalCollection     = "ss|iv"
	IntervalCollectionName = IntervalCollection + DBKeySeparator + v2.Name
)

// intervalStoredKey return the interval's stored key which combines the collection name and object id
func intervalStoredKey(id string) string {
	return CreateKey(IntervalCollection, id)
}

// addInterval adds a new interval into DB
func addInterval(conn redis.Conn, interval models.Interval) (models.Interval, errors.EdgeX) {
	exists, edgeXerr := objectIdExists(conn, intervalStoredKey(interval.Id))
	if edgeXerr != nil {
		return interval, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return interval, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("interval id %s already exists", interval.Id), edgeXerr)
	}

	exists, edgeXerr = objectNameExists(conn, IntervalCollectionName, interval.Name)
	if edgeXerr != nil {
		return interval, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return interval, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("interval name %s already exists", interval.Name), edgeXerr)
	}

	ts := common.MakeTimestamp()
	if interval.Created == 0 {
		interval.Created = ts
	}
	interval.Modified = ts

	dsJSONBytes, err := json.Marshal(interval)
	if err != nil {
		return interval, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal interval for Redis persistence", err)
	}

	storedKey := intervalStoredKey(interval.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(SET, storedKey, dsJSONBytes)
	_ = conn.Send(ZADD, IntervalCollection, 0, storedKey)
	_ = conn.Send(HSET, IntervalCollectionName, interval.Name, storedKey)
	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "interval creation failed", err)
	}

	return interval, edgeXerr
}
