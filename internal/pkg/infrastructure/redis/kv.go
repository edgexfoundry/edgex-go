//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper/constants"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/gomodule/redigo/redis"
	"github.com/spf13/cast"
)

const KVCollection = "kp|kv"

// replaceKeyDelimiterForDB replace the key delimiter from slash(for EdgeX Keeper) to colon(for Redis)
func replaceKeyDelimiterForDB(wholeKey string) string {
	return strings.ReplaceAll(wholeKey, constants.KeyDelimiter, DBKeySeparator)
}

// replaceKeyDelimiterForKeeper replace the key delimiter from colon(for Redis) to slash(for EdgeX Keeper)
func replaceKeyDelimiterForKeeper(wholeKey string) string {
	return strings.ReplaceAll(wholeKey, DBKeySeparator, constants.KeyDelimiter)
}

// keeperKeys returns the value(s) stored in the specified key or keys with the same prefix
func keeperKeys(conn redis.Conn, key string, keyOnly bool, isRaw bool) (configs []models.KVResponse, edgeXerr errors.EdgeX) {
	configs, edgeXerr = getObjectsByKeyPrefix(conn, CreateKey(KVCollection, key), keyOnly, isRaw)

	if edgeXerr != nil {
		return configs, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return configs, edgeXerr
}

// addKeeperKeys stores the value in the specified key
func addKeeperKeys(conn redis.Conn, kv models.KVS, isFlatten bool) (keysResp []models.KeyOnly, edgeXerr errors.EdgeX) {
	key := kv.Key
	storedKey := CreateKey(KVCollection, key)

	// if the key (ex. core-data/Writable) already exists and is a hash type with child key(s) exist (ex. core-data/Writable/LogLevel)
	// the updated value is ony allowed to be a map to update the child keys
	if exists, _ := objectIdExists(conn, storedKey); exists {
		if keyType, _ := getKeyType(conn, storedKey); keyType == Hash {
			if _, ok := kv.Value.(map[string]any); !ok {
				return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "update key failed since child key(s) already exist", nil)
			}
		}
	}

	_ = conn.Send(MULTI)

	sendAddUpperLevelKeyCmds(conn, key)

	if isFlatten {
		keysResp, edgeXerr = sendCreateKeysByDataTypeCmds(conn, storedKey, kv.Value)
		if edgeXerr != nil {
			return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("send create/update key %s command failed", key), edgeXerr)
		}
	} else {
		// if the value type is map, convert the map to string
		storedValue := kv.Value
		if valueMap, ok := kv.Value.(map[string]interface{}); ok {
			vJSONBytes, err := json.Marshal(valueMap)
			if err != nil {
				return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal KV for Redis persistence", err)
			}
			storedValue = string(vJSONBytes)
		}
		keysResp, edgeXerr = sendCreateKeysByDataTypeCmds(conn, storedKey, storedValue)
		if edgeXerr != nil {
			return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("send create/update key %s command failed", key), edgeXerr)
		}
	}

	_, err := conn.Do(EXEC)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("create/update key %s failed", key), err)
	}

	return keysResp, nil
}

// sendAddUpperLevelKeyCmds send redis commands to add the fields in each upper level Hashes
// if the key is a:b:c:d, the corresponding hash fields will be added on the keys a:b:c, a:b and a
func sendAddUpperLevelKeyCmds(conn redis.Conn, key string) {
	prevKey := key
	suffixKey := ""

	for prevKey != "" {
		// check if any Hash exists in the upper level of the key
		// ex. a:b:c is the upper level key of a:b:c:uncompleted-key
		if idx := strings.LastIndex(prevKey, DBKeySeparator); idx != -1 {
			upperKey := prevKey[:idx]
			suffixKey = prevKey[idx+1:]
			_ = conn.Send(HSETNX, CreateKey(KVCollection, upperKey), suffixKey, CreateKey(KVCollection, prevKey))
			prevKey = upperKey
		} else {
			// create the Hash field on the root level
			_ = conn.Send(HSETNX, KVCollection, prevKey, CreateKey(KVCollection, prevKey))
			prevKey = ""
		}
	}
}

// sendCreateKeysByDataTypeCmds send redis commands to add the key in redis based on the value type
// if the value type is string, a key with string type will be created
// otherwise, if the value type is an object, a key with Hash type will be created along with fields corresponding to the object properties
func sendCreateKeysByDataTypeCmds(conn redis.Conn, key string, value interface{}) (keysResp []models.KeyOnly, edgeXerr errors.EdgeX) {
	switch v := value.(type) {
	case map[string]interface{}:
		for innerKey, element := range v {
			// if the element type is an empty map, do not add the inner key to the upper level Hash field
			if eleMap, ok := element.(map[string]interface{}); ok && len(eleMap) == 0 {
				continue
			}
			innerHashValue := CreateKey(key, innerKey)

			_ = conn.Send(HSET, key, innerKey, innerHashValue)

			// create the innerHashValue key at next level
			resp, sendErr := sendCreateKeysByDataTypeCmds(conn, innerHashValue, element)
			if sendErr != nil {
				return nil, sendErr
			}
			keysResp = append(keysResp, resp...)
		}
	case bool, int, int8, int16, int32, int64, float32, float64, string, []interface{}:
		var storedValueBytes []byte
		var err error

		if _, ok := value.([]interface{}); ok {
			// for key with array data type, covert the key to string with brackets and commas in bytes, ex. ["a","b"]
			storedValueBytes, err = json.Marshal(v)
			if err != nil {
				return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unable to encode key %s for Redis persistence", key), err)
			}
		} else {
			// for key with other data types, convert to string first and then byte array
			storedValueStr := cast.ToString(v)
			storedValueBytes = []byte(storedValueStr)
		}

		currentTimestamp := time.Now().UnixNano() / int64(time.Millisecond)
		// the StoredData struct will be saved to Redis with base64 encoded
		kv := models.StoredData{
			DBTimestamp: models.DBTimestamp{Created: currentTimestamp, Modified: currentTimestamp},
			Value:       storedValueBytes,
		}
		kvJSONBytes, err := json.Marshal(kv)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal KV for Redis persistence", err)
		}

		_ = conn.Send(SET, key, kvJSONBytes)

		// get the query key after KVCollection prefix (kp:)
		idx := strings.Index(key, DBKeySeparator)
		if idx == -1 {
			return keysResp, errors.NewCommonEdgeX(errors.KindDatabaseError, "retrieve updated key failed", nil)
		}
		queryKey := key[idx+1:]
		keysResp = []models.KeyOnly{models.KeyOnly(queryKey)}
	default:
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("unknown data type of key %s", key), nil)
	}
	return keysResp, nil
}

// deleteKeeperKeys delete the value in the specified key or keys with the same prefix
func deleteKeeperKeys(conn redis.Conn, key string, prefixMatch bool) (keys []models.KeyOnly, edgeXerr errors.EdgeX) {
	keys, edgeXerr = deleteByKeyPrefix(conn, CreateKey(KVCollection, key), prefixMatch)
	if edgeXerr != nil {
		return keys, edgeXerr
	}

	edgeXerr = deleteUpperLevelKeyFields(conn, key)
	if edgeXerr != nil {
		return keys, edgeXerr
	}
	return keys, edgeXerr
}

// deleteByKeyPrefix delete the specified key or keys with the same prefix
func deleteByKeyPrefix(conn redis.Conn, key string, prefixMatch bool) ([]models.KeyOnly, errors.EdgeX) {
	var keyResp []models.KeyOnly

	// check if the query key exists
	exists, err := objectIdExists(conn, key)
	if err != nil {
		return keyResp, errors.NewCommonEdgeXWrapper(err)
	}

	// key not exists in Redis, returns not found error
	if !exists {
		return keyResp, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("query key %s not exists", key), err)
	}

	// key exists in Redis, call deleteByKey to delete the key(s)
	return deleteByKey(conn, key, prefixMatch)
}

// deleteByKey is a recursive function that delete the specified key
// if the key is a Hash, it will traverse all the keys stored in the Hash fields and invoke the recursive function until the key is a String
func deleteByKey(conn redis.Conn, key string, prefixMatch bool) ([]models.KeyOnly, errors.EdgeX) {
	var keyResp []models.KeyOnly

	keyType, edgexErr := getKeyType(conn, key)
	if edgexErr != nil {
		return keyResp, edgexErr
	}

	switch keyType {
	case String:
		var resp models.KeyOnly
		// get the query key after KVCollection prefix (kp:)
		idx := strings.Index(key, DBKeySeparator)
		if idx == -1 {
			return keyResp, errors.NewCommonEdgeX(errors.KindDatabaseError, "retrieve query key failed", nil)
		}
		queryKey := key[idx+1:]

		_, err := conn.Do(DEL, key)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("key %s deletion failed", key), err)
		}
		resp = models.KeyOnly(queryKey)

		return []models.KeyOnly{resp}, nil
	case Hash:
		if !prefixMatch {
			return nil, errors.NewCommonEdgeX(errors.KindStatusConflict, fmt.Sprintf("keys having the same prefix %s exist and cannot be deleted", key), nil)
		}
		keyMap, mapKeyErr := getMapByKey(conn, key)
		if mapKeyErr != nil {
			return keyResp, mapKeyErr
		}
		for field, v := range keyMap {
			resp, deleteErr := deleteByKey(conn, v, prefixMatch)
			if deleteErr != nil {
				return keyResp, deleteErr
			}
			_, err := conn.Do(HDEL, key, field)
			if err != nil {
				return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("hash field %s in key %s deletion failed", field, key), err)
			}
			keyResp = append(keyResp, resp...)
		}
	}
	return keyResp, nil
}

// deleteUpperLevelKeyFields deletes the fields in the upper level Hashes
// if the key is a:b:c:d, the corresponding hash fields will be deleted on the keys a:b:c
func deleteUpperLevelKeyFields(conn redis.Conn, key string) errors.EdgeX {
	prevKey := key
	suffixKey := ""

	for prevKey != "" {
		// check if any Hash exists in the upper level of the key
		// ex. a:b:c is the upper level key of a:b:c:d
		if idx := strings.LastIndex(prevKey, DBKeySeparator); idx != -1 {
			upperKey := prevKey[:idx]
			suffixKey = prevKey[idx+1:]
			upperKeyPath := CreateKey(KVCollection, upperKey)
			// delete the upper level hash field with value is the delete key path
			// ex. field d is deleted in hash a:b:c
			_, err := conn.Do(HDEL, upperKeyPath, suffixKey)
			if err != nil {
				return errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("delete hash field %s of key %s failed", suffixKey, upperKey), err)
			}
			// check if the upper level hash(a:b:c) has other fields
			// if other field exists, the delete action is complete
			// if not, delete the upper level hash(a:b:c) and go to the next upper level hash(a:b)
			length, err := redis.Int(conn.Do(HLEN, upperKeyPath))
			if err != nil {
				return errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("check the length of hash key %s failed", upperKey), err)
			}
			if length != 0 {
				break
			}
			prevKey = upperKey
		} else {
			// delete the Hash field on the root level
			_, err := conn.Do(HDEL, KVCollection, prevKey)
			if err != nil {
				return errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("delete hash field %s of key %s failed", prevKey, KVCollection), err)
			}
			prevKey = ""
		}
	}
	return nil
}
