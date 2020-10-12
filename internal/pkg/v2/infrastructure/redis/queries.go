//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/errors"

	"github.com/gomodule/redigo/redis"
)

func objectExistById(conn redis.Conn, id string) (bool, errors.EdgeX) {
	exists, err := redis.Bool(conn.Do(EXISTS, id))
	if err != nil {
		return false, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("query id %s from the database failed", id), err)
	}
	return exists, nil
}

func objectExistByHash(conn redis.Conn, hash string, field string) (bool, errors.EdgeX) {
	exists, err := redis.Bool(conn.Do(HEXISTS, hash, field))
	if err != nil {
		return false, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("query field %s from the database failed", field), err)
	}
	return exists, nil
}

func getObjectById(conn redis.Conn, id string, out interface{}) errors.EdgeX {
	obj, err := redis.Bytes(conn.Do(GET, id))
	if err == redis.ErrNil {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("object %T doesn't exist in the database", out), err)
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
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", field), err)
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
	ids, err := redis.Values(conn.Do(command, key, start, end))
	if err == redis.ErrNil {
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("objects under %s do not exist", key), err)
	} else if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "query object ids from database failed", err)
	}

	var result [][]byte
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
