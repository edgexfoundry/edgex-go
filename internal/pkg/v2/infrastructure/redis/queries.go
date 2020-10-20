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

	"github.com/gomodule/redigo/redis"
)

func getObjectById(conn redis.Conn, id string, out interface{}) errors.EdgeX {
	obj, err := redis.Bytes(conn.Do(GET, id))
	if err == redis.ErrNil {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("fail to query object %T, because id: %s doesn't exist in the database", out, id), err)
	} else if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("query object %T by id from the database failed", out), err)
	}

	err = json.Unmarshal(obj, out)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("object %T format parsing failed from the database", out), err)
	}

	return nil
}

// getObjectByHash retrieves the id with associated field from the hash stored and then retrieves the object by id
func getObjectByHash(conn redis.Conn, hash string, field string, out interface{}) errors.EdgeX {
	id, err := redis.String(conn.Do(HGET, hash, field))
	if err == redis.ErrNil {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("fail to query object %T, because %s: %s doesn't exist in the database", out, field, hash), err)
	} else if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("query %s from the database failed", field), err)
	}

	return getObjectById(conn, id, out)
}

// getObjectsByRange retrieves the entries for keys enumerated in a sorted set.
// The entries are retrieved in the sorted set order.
func getObjectsByRange(conn redis.Conn, key string, start, end int) ([][]byte, errors.EdgeX) {
	return getObjectsBySomeRange(conn, ZRANGE, key, start, end)
}

// getObjectsByRevRange retrieves the entries for keys enumerated in a sorted set.
// The entries are retrieved in the reverse sorted set order.
func getObjectsByRevRange(conn redis.Conn, key string, start int, end int) ([][]byte, errors.EdgeX) {
	return getObjectsBySomeRange(conn, ZREVRANGE, key, start, end)
}

// getObjectsBySomeRange retrieves the entries for keys enumerated in a sorted set using the specified Redis range
// command (i.e. RANGE, REVRANGE). The entries are retrieved in the order specified by the supplied Redis command.
func getObjectsBySomeRange(conn redis.Conn, command string, key string, start int, end int) ([][]byte, errors.EdgeX) {
	count, err := redis.Int(conn.Do(ZCOUNT, key, InfiniteMin, InfiniteMax))
	if count == 0 { // return nil slice when there is no records in the DB
		return nil, nil
	} else if count > 0 && start > count { // return RangeNotSatisfiable error when start is out of range
		return nil, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, fmt.Sprintf("query objects bounds out of range. length:%v", count), nil)
	}
	ids, err := redis.Values(conn.Do(command, key, start, end))
	if err == redis.ErrNil {
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("objects under %s do not exist", key), err)
	} else if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "query object ids from database failed", err)
	}

	return getObjectsByIds(conn, ids)
}

// getObjectsByLabelsAndSomeRange retrieves the entries for keys enumerated in a sorted set using the specified Redis range
// command (i.e. RANGE, REVRANGE). The entries are retrieved in the order specified by the supplied Redis command.
func getObjectsByLabelsAndSomeRange(conn redis.Conn, command string, key string, labels []string, start int, end int) ([][]byte, errors.EdgeX) {
	if labels == nil || len(labels) == 0 { //if no labels specified, simply return getObjectsBySomeRange
		return getObjectsBySomeRange(conn, command, key, start, end)
	}

	idsSlice := make([][]string, len(labels))
	for i, label := range labels { //iterate each labels to retrieve Ids associated with labels
		idsWithLabel, err := redis.Strings(conn.Do(command, fmt.Sprintf("%s:label:%s", key, label), 0, -1))
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("query object ids by label %s from database failed", label), err)
		}
		idSlice := make([]string, len(idsWithLabel))
		for i, v := range idsWithLabel {
			idSlice[i] = fmt.Sprint(v)
		}
		idsSlice[i] = idSlice
	}

	//find common Ids among two-dimension Ids slice associated with labels
	commonIds := common.FindCommonStrings(idsSlice...)
	if start > len(commonIds) {
		return nil, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, fmt.Sprintf("query objects bounds out of range. length:%v", len(commonIds)), nil)
	}
	if end > len(commonIds) {
		commonIds = commonIds[start:]
	} else { // as end index in golang re-slice is exclusive, increment the end index to ensure the end could be inclusive
		commonIds = commonIds[start : end+1]
	}

	return getObjectsByIds(conn, common.ConvertStringsToInterfaces(commonIds))
}

// getObjectsByIds retrieves the entries with Ids
func getObjectsByIds(conn redis.Conn, ids []interface{}) ([][]byte, errors.EdgeX) {
	var result [][]byte
	var err error
	if len(ids) > 0 {
		result, err = redis.ByteSlices(conn.Do(MGET, ids...))
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "query objects from database failed", err)
		}
	}

	var objects [][]byte
	for _, obj := range result {
		if obj != nil {
			objects = append(objects, obj)
		}
	}

	return objects, nil
}

// objectNameExists checks whether the object name exists or not in the specified hashKey
func objectNameExists(conn redis.Conn, hashKey string, name string) (bool, errors.EdgeX) {
	exists, err := redis.Bool(conn.Do(HEXISTS, hashKey, name))
	if err != nil {
		return false, errors.NewCommonEdgeX(errors.KindDatabaseError, "object name existence check failed", err)
	}
	return exists, nil
}

// objectIdExists checks whether the object id exists or not
func objectIdExists(conn redis.Conn, id string) (bool, errors.EdgeX) {
	exists, err := redis.Bool(conn.Do(EXISTS, id))
	if err != nil {
		return false, errors.NewCommonEdgeX(errors.KindDatabaseError, "object Id existence check failed", err)
	}
	return exists, nil
}
