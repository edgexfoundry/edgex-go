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
	_ = conn.Send(ZADD, IntervalCollection, interval.Modified, storedKey)
	_ = conn.Send(HSET, IntervalCollectionName, interval.Name, storedKey)
	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "interval creation failed", err)
	}

	return interval, edgeXerr
}

// intervalByName query interval by name from DB
func intervalByName(conn redis.Conn, name string) (interval models.Interval, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectByHash(conn, IntervalCollectionName, name, &interval)
	if edgeXerr != nil {
		return interval, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return
}

// allIntervals queries intervals by offset and limit
func allIntervals(conn redis.Conn, offset, limit int) (intervals []models.Interval, edgeXerr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, edgeXerr := getObjectsByRevRange(conn, IntervalCollection, offset, end)
	if edgeXerr != nil {
		return intervals, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	intervals = make([]models.Interval, len(objects))
	for i, o := range objects {
		s := models.Interval{}
		err := json.Unmarshal(o, &s)
		if err != nil {
			return []models.Interval{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "interval format parsing failed from the database", err)
		}
		intervals[i] = s
	}
	return intervals, nil
}

// deleteIntervalByName deletes the interval by name
func deleteIntervalByName(conn redis.Conn, name string) errors.EdgeX {
	interval, edgeXerr := intervalByName(conn, name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	storedKey := intervalStoredKey(interval.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, IntervalCollection, storedKey)
	_ = conn.Send(HDEL, IntervalCollectionName, interval.Name)
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "interval deletion failed", err)
	}
	return nil
}
