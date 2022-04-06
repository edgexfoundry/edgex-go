//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
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
func getObjectsByRevRange(conn redis.Conn, key string, offset int, limit int) ([][]byte, errors.EdgeX) {
	return getObjectsBySomeRange(conn, ZREVRANGE, key, offset, limit)
}

// getObjectsBySomeRange retrieves the entries for keys enumerated in a sorted set using the specified Redis range
// command (i.e. RANGE, REVRANGE). The entries are retrieved in the order specified by the supplied Redis command.
func getObjectsBySomeRange(conn redis.Conn, command string, key string, offset int, limit int) ([][]byte, errors.EdgeX) {
	if limit == 0 {
		return [][]byte{}, nil
	}
	start := offset
	end := start + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	count, _ := redis.Int(conn.Do(ZCOUNT, key, InfiniteMin, InfiniteMax))
	if start > count { // return RangeNotSatisfiable error when start is out of range
		return nil, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, fmt.Sprintf("query objects bounds out of range. length:%v", count), nil)
	}
	if count == 0 { // return nil slice when there is no records satisfied with the score range in the DB, so there is no need to query the DB further
		return nil, nil
	}
	ids, err := redis.Values(conn.Do(command, key, start, end))
	if err == redis.ErrNil {
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("objects under %s do not exist", key), err)
	} else if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "query object ids from database failed", err)
	}

	return getObjectsByIds(conn, ids)
}

// getObjectsByScoreRange query objects by specified key's score range, offset, and limit.  Note that the specified key must be a sorted set.
func getObjectsByScoreRange(conn redis.Conn, key string, start int, end int, offset int, limit int) (objects [][]byte, edgeXerr errors.EdgeX) {
	if limit == 0 {
		return
	}
	count, _ := redis.Int(conn.Do(ZCOUNT, key, start, end))
	if offset > count { // return RangeNotSatisfiable error when offset is out of range
		return nil, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, fmt.Sprintf("query objects bounds out of range. length:%v offset:%v", count, offset), nil)
	}
	if count == 0 { // return nil slice when there is no records satisfied with the score range in the DB, so there is no need to query the DB further
		return nil, nil
	}
	// Use following redis command to retrieve the id of objects satisfied with score range/offset/limit
	// ZREVRANGEBYSCORE key max min LIMIT offset count
	objIds, err := redis.Strings(conn.Do(ZREVRANGEBYSCORE, key, end, start, LIMIT, offset, limit))
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return getObjectsByIds(conn, pkgCommon.ConvertStringsToInterfaces(objIds))
}

// getObjectsByLabelsAndSomeRange retrieves the entries for keys enumerated in a sorted set using the specified Redis range
// command (i.e. RANGE, REVRANGE). The entries are retrieved in the order specified by the supplied Redis command.
func getObjectsByLabelsAndSomeRange(conn redis.Conn, command string, key string, labels []string, offset int, limit int) ([][]byte, errors.EdgeX) {
	if len(labels) == 0 { //if no labels specified, simply return getObjectsBySomeRange
		return getObjectsBySomeRange(conn, command, key, offset, limit)
	}

	if limit == 0 {
		return [][]byte{}, nil
	}
	idsSlice := make([][]string, len(labels))
	for i, label := range labels { //iterate each labels to retrieve Ids associated with labels
		idsWithLabel, err := redis.Strings(conn.Do(command, CreateKey(key, common.Label, label), 0, -1))
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("query object ids by label %s from database failed", label), err)
		}
		idsSlice[i] = idsWithLabel
	}
	start := offset
	end := start + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	//find common Ids among two-dimension Ids slice associated with labels
	commonIds := pkgCommon.FindCommonStrings(idsSlice...)
	if start > len(commonIds) {
		return nil, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, fmt.Sprintf("query objects bounds out of range. length:%v", len(commonIds)), nil)
	}
	if end >= len(commonIds) {
		commonIds = commonIds[start:]
	} else { // as end index in golang re-slice is exclusive, increment the end index to ensure the end could be inclusive
		commonIds = commonIds[start : end+1]
	}

	return getObjectsByIds(conn, pkgCommon.ConvertStringsToInterfaces(commonIds))
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

func getMemberCountByScoreRange(conn redis.Conn, key string, start int, end int) (uint32, errors.EdgeX) {
	count, err := redis.Int(conn.Do(ZCOUNT, key, start, end))
	if err != nil {
		return 0, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to get member count from %s between score range %v to %v", key, start, end), err)
	}

	return uint32(count), nil
}

// getMemberCountByLabels return the record count of key with labels specified.
func getMemberCountByLabels(conn redis.Conn, command string, key string, labels []string) (uint32, errors.EdgeX) {
	if len(labels) == 0 { //if no labels specified, simply return the count of record for the key
		return getMemberNumber(conn, ZCARD, key)
	}

	idsSlice := make([][]string, len(labels))
	for i, label := range labels { //iterate each labels to retrieve Ids associated with labels
		idsWithLabel, err := redis.Strings(conn.Do(command, CreateKey(key, common.Label, label), 0, -1))
		if err != nil {
			return 0, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("query object ids by label %s from database failed", label), err)
		}
		idsSlice[i] = idsWithLabel
	}
	//find common Ids among two-dimension Ids slice associated with labels
	commonIds := pkgCommon.FindCommonStrings(idsSlice...)
	return uint32(len(commonIds)), nil
}

func getMemberNumber(conn redis.Conn, command string, key string) (uint32, errors.EdgeX) {
	count, err := redis.Int(conn.Do(command, key))
	if err != nil {
		return 0, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to get member number with command %s from %s", command, key), err)
	}

	return uint32(count), nil
}

// unionObjectsByValues returns the keys of the set resulting from the union of all the given sets.
func unionObjectsByKeys(conn redis.Conn, offset int, limit int, redisKeys ...string) ([][]byte, errors.EdgeX) {
	return objectsByKeys(conn, ZUNIONSTORE, offset, limit, redisKeys...)
}

// intersectionObjectsByKeys returns the keys of the set resulting from the intersection of all the given sets.
func intersectionObjectsByKeys(conn redis.Conn, offset int, limit int, redisKeys ...string) ([][]byte, errors.EdgeX) {
	return objectsByKeys(conn, ZINTERSTORE, offset, limit, redisKeys...)
}

// objectsByKeys returns the keys of the set resulting from the all the given sets. The data set method could be ZINTERSTORE or ZUNIONSTORE
func objectsByKeys(conn redis.Conn, setMethod string, offset int, limit int, redisKeys ...string) ([][]byte, errors.EdgeX) {
	if limit == 0 {
		return [][]byte{}, nil
	}
	end := offset + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	args := redis.Args{}
	cacheSet := uuid.New().String()
	args = append(args, cacheSet)
	args = append(args, strconv.Itoa(len(redisKeys)))
	for _, key := range redisKeys {
		args = args.Add(key)
	}
	_, err := conn.Do(setMethod, args...)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to execute %s command with args %v", setMethod, args), err)
	}
	storeKeys, err := redis.Values(conn.Do(ZREVRANGE, cacheSet, 0, -1))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query storeKeys", err)
	}
	count := len(storeKeys)
	if offset > count { // return RangeNotSatisfiable error when offset is out of range
		return nil, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, fmt.Sprintf("query objects bounds out of range. length:%v", count), nil)
	} else if count == 0 || count == offset {
		return nil, nil
	} else if end >= count || end == -1 {
		storeKeys = storeKeys[offset:]
	} else { // as end index in golang re-slice is exclusive, increment the end index to ensure the end could be inclusive
		storeKeys = storeKeys[offset : end+1]
	}
	objects, err := redis.ByteSlices(conn.Do(MGET, storeKeys...))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "query objects from database failed", err)
	}

	// clean up unused cache set
	_, err = redis.Int(conn.Do(DEL, cacheSet))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "cache set deletion failed", err)
	}

	return objects, nil
}

// unionObjectsByKeysAndScoreRange returns objects resulting from the union of all the given sets with specified score range, offset, and limit
func unionObjectsByKeysAndScoreRange(conn redis.Conn, start, end, offset, limit int, redisKeys ...string) ([][]byte, uint32, errors.EdgeX) {
	return objectsByKeysAndScoreRange(conn, ZUNIONSTORE, start, end, offset, limit, redisKeys...)
}

// objectsByKeysAndScoreRange returns objects resulting from the set method of all the given sets with specified score range, offset, and limit.  The data set method could be either ZINTERSTORE or ZUNIONSTORE
func objectsByKeysAndScoreRange(conn redis.Conn, setMethod string, start, end, offset, limit int, redisKeys ...string) (objects [][]byte, totalCount uint32, edgeXerr errors.EdgeX) {
	// build up the redis command arguments
	args := redis.Args{}
	cacheSet := uuid.New().String()
	args = append(args, cacheSet)
	args = append(args, strconv.Itoa(len(redisKeys)))
	for _, key := range redisKeys {
		args = args.Add(key)
	}

	// create a temporary sorted set, cacheSet, resulting from the specified setMethod
	_, err := conn.Do(setMethod, args...)
	if err != nil {
		return nil, totalCount, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to execute %s command with args %v", setMethod, args), err)
	}

	// get the total count of the temporary sorted set
	if totalCount, edgeXerr = getMemberCountByScoreRange(conn, cacheSet, start, end); edgeXerr != nil {
		return nil, totalCount, edgeXerr
	}

	// get objects from the temporary sorted set
	if objects, edgeXerr = getObjectsByScoreRange(conn, cacheSet, start, end, offset, limit); edgeXerr != nil {
		return nil, totalCount, edgeXerr
	}

	// clean up unused temporary sorted set
	_, err = redis.Int(conn.Do(DEL, cacheSet))
	if err != nil {
		return nil, totalCount, errors.NewCommonEdgeX(errors.KindDatabaseError, "cache set deletion failed", err)
	}

	return objects, totalCount, nil
}

// idFromStoredKey extracts Id from the store key
func idFromStoredKey(storeKey string) string {
	substrings := strings.Split(storeKey, DBKeySeparator)
	return substrings[len(substrings)-1]
}
