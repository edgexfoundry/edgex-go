//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/gomodule/redigo/redis"
)

const (
	IntervalCollection     = "ss|iv"
	IntervalCollectionName = IntervalCollection + DBKeySeparator + common.Name
)

// intervalStoredKey return the interval's stored key which combines the collection name and object id
func intervalStoredKey(id string) string {
	return CreateKey(IntervalCollection, id)
}

// sendAddIntervalCmd sends redis command for adding interval
func sendAddIntervalCmd(conn redis.Conn, storedKey string, interval models.Interval) errors.EdgeX {
	m, err := json.Marshal(interval)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal interval for Redis persistence", err)
	}
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, IntervalCollection, interval.Modified, storedKey)
	_ = conn.Send(HSET, IntervalCollectionName, interval.Name, storedKey)
	return nil
}

// addInterval adds a new interval into DB
func addInterval(conn redis.Conn, interval models.Interval) (models.Interval, errors.EdgeX) {
	exists, edgeXerr := objectIdExists(conn, intervalStoredKey(interval.Id))
	if edgeXerr != nil {
		return interval, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return interval, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("interval id %s already exists", interval.Id), edgeXerr)
	}

	exists, edgeXerr = intervalNameExists(conn, interval.Name)
	if edgeXerr != nil {
		return interval, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return interval, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("interval name %s already exists", interval.Name), edgeXerr)
	}

	ts := pkgCommon.MakeTimestamp()
	if interval.Created == 0 {
		interval.Created = ts
	}
	interval.Modified = ts

	storedKey := intervalStoredKey(interval.Id)
	_ = conn.Send(MULTI)
	edgeXerr = sendAddIntervalCmd(conn, storedKey, interval)
	if edgeXerr != nil {
		return interval, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "interval creation failed", err)
	}

	return interval, edgeXerr
}

// intervalByName query interval by name from DB
func intervalByName(conn redis.Conn, name string) (interval models.Interval, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectByHash(conn, IntervalCollectionName, name, &interval)
	if edgeXerr != nil {
		return interval, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to query interval by name %s", name), edgeXerr)
	}
	return
}

// intervalById query interval by id from DB
func intervalById(conn redis.Conn, id string) (interval models.Interval, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectById(conn, intervalStoredKey(id), &interval)
	if edgeXerr != nil {
		return interval, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to query interval by id %s", id), edgeXerr)
	}
	return
}

// allIntervals queries intervals by offset and limit
func allIntervals(conn redis.Conn, offset, limit int) (intervals []models.Interval, edgeXerr errors.EdgeX) {
	objects, edgeXerr := getObjectsByRevRange(conn, IntervalCollection, offset, limit)
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

// sendDeleteIntervalCmd sends redis command for deleting interval
func sendDeleteIntervalCmd(conn redis.Conn, storedKey string, interval models.Interval) {
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, IntervalCollection, storedKey)
	_ = conn.Send(HDEL, IntervalCollectionName, interval.Name)
}

// deleteIntervalByName deletes the interval by name
func deleteIntervalByName(conn redis.Conn, name string) errors.EdgeX {
	interval, edgeXerr := intervalByName(conn, name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	actions, edgeXerr := intervalActionsByIntervalName(conn, 0, 1, name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	if len(actions) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to delete the interval when associated intervalAction exists", nil)
	}
	storedKey := intervalStoredKey(interval.Id)
	_ = conn.Send(MULTI)
	sendDeleteIntervalCmd(conn, storedKey, interval)
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "interval deletion failed", err)
	}
	return nil
}

// updateInterval updates a interval
func updateInterval(conn redis.Conn, interval models.Interval) errors.EdgeX {
	oldInterval, edgeXerr := intervalByName(conn, interval.Name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	actions, edgeXerr := intervalActionsByIntervalName(conn, 0, 1, interval.Name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	if len(actions) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to patch the interval when associated intervalAction exists", nil)
	}

	interval.Modified = pkgCommon.MakeTimestamp()
	storedKey := intervalStoredKey(interval.Id)
	_ = conn.Send(MULTI)
	sendDeleteIntervalCmd(conn, storedKey, oldInterval)
	edgeXerr = sendAddIntervalCmd(conn, storedKey, interval)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "interval update failed", err)
	}

	return nil
}

// intervalNameExists whether the interval exists by name
func intervalNameExists(conn redis.Conn, name string) (bool, errors.EdgeX) {
	exists, err := objectNameExists(conn, IntervalCollectionName, name)
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	return exists, nil
}
