/*******************************************************************************
 * Copyright 2018 Redis Labs Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package redis

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/gomodule/redigo/redis"
)

func getObjectById(conn redis.Conn, id string, unmarshal unmarshalFunc, out interface{}) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	return unmarshal(object, out)
}

func getObjectByHash(conn redis.Conn, hash string, field string, unmarshal unmarshalFunc, out interface{}) error {
	id, err := redis.String(conn.Do("HGET", hash, field))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	object, err := redis.Bytes(conn.Do("GET", id))
	if err != nil {
		return err
	}

	return unmarshal(object, out)
}

func getObjectsByValue(conn redis.Conn, v string) (objects [][]byte, err error) {
	ids, err := redis.Values(conn.Do("SMEMBERS", v))
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	objects, err = redis.ByteSlices(conn.Do("MGET", ids...))
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func getObjectsByValues(conn redis.Conn, vals ...string) (objects [][]byte, err error) {
	args := redis.Args{}
	for _, v := range vals {
		args = args.Add(v)
	}
	ids, err := redis.Values(conn.Do("SINTER", args...))
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	objects, err = redis.ByteSlices(conn.Do("MGET", ids...))
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func getObjectsByRange(conn redis.Conn, key string, start, end int) (objects [][]byte, err error) {
	ids, err := redis.Values(conn.Do("ZRANGE", key, start, end))
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	if len(ids) > 0 {
		objects, err = redis.ByteSlices(conn.Do("MGET", ids...))
		if err != nil {
			return nil, err
		}
	}
	return objects, nil
}

// Return objects by a score from a zset
// if limit is 0, all are returned
// if end is negative, it is considered as positive infinity
func getObjectsByRangeFilter(conn redis.Conn, key string, filter string, start, end int) (objects [][]byte, err error) {
	ids, err := redis.Values(conn.Do("ZRANGE", key, start, end))
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
	fids := ids[:0]
	if len(ids) > 0 {
		for _, id := range ids {
			err := conn.Send("ZSCORE", filter, id)
			if err != nil {
				return nil, err
			}
		}
		scores, err := redis.Strings(conn.Do(""))
		if err != nil {
			return nil, err
		}

		for i, score := range scores {
			if score != "" {
				fids = append(fids, ids[i])
			}
		}

		objects, err = redis.ByteSlices(conn.Do("MGET", fids...))
		if err != nil {
			return nil, err
		}
	}
	return objects, nil
}

func getObjectsByScore(conn redis.Conn, key string, start, end int64, limit int) (objects [][]byte, err error) {
	args := []interface{}{key, start}
	if end < 0 {
		args = append(args, "+inf")
	} else {
		args = append(args, end)
	}
	if limit != 0 {
		args = append(args, "LIMIT")
		args = append(args, 0)
		args = append(args, limit)
	}
	ids, err := redis.Values(conn.Do("ZRANGEBYSCORE", args...))
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	if len(ids) > 0 {
		objects, err = redis.ByteSlices(conn.Do("MGET", ids...))
		if err != nil {
			return nil, err
		}
	}
	return objects, nil
}
