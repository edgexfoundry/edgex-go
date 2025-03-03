//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path"
	"time"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cast"
)

// KeeperKeys returns the values stored for the specified key or with the same key prefix
func (c *Client) KeeperKeys(key string, keyOnly bool, isRaw bool) ([]models.KVResponse, errors.EdgeX) {
	var result []models.KVResponse
	var sqlStatement string

	if keyOnly {
		sqlStatement = sqlQueryFieldsByColAndLikePat(configTableName, []string{keyCol}, keyCol)
	} else {
		sqlStatement = sqlQueryFieldsByColAndLikePat(configTableName, []string{keyCol, valueCol, createdCol, modifiedCol}, keyCol)
	}

	// Query the exact match key and all child level keys
	// e.g., key='edgex/v4/core-data' || key='edgex/v4/core-data/%'
	sqlStatement += fmt.Sprintf(" OR %s = $2", keyCol)
	rows, err := c.ConnPool.Query(context.Background(), sqlStatement, key+"/%", key)
	if err != nil {
		return nil, pgClient.WrapDBError(fmt.Sprintf("failed to query rows by key '%s'", key), err)
	}

	var kvKey string
	if keyOnly {
		_, err = pgx.ForEachRow(rows, []any{&kvKey}, func() error {
			keyOnlyModel := models.KeyOnly(kvKey)
			result = append(result, &keyOnlyModel)
			return nil
		})
	} else {
		var kvVal string
		var created, modified time.Time
		_, err = pgx.ForEachRow(rows, []any{&kvKey, &kvVal, &created, &modified}, func() error {
			var keyValue any
			if isRaw {
				decodeValue, decErr := base64.StdEncoding.DecodeString(kvVal)
				if decErr != nil {
					return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("decode the value of key %s failed", kvKey), err)
				}
				keyValue = string(decodeValue)
			} else {
				keyValue = kvVal
			}
			kvStore := models.KVS{
				Key: kvKey,
				StoredData: models.StoredData{
					DBTimestamp: models.DBTimestamp{Created: created.UnixMilli(), Modified: modified.UnixMilli()},
					Value:       keyValue,
				},
			}
			result = append(result, &kvStore)
			return nil
		})
	}
	if err != nil {
		return nil, pgClient.WrapDBError("failed to scan row to models.KVResponse", err)
	}

	if rows.CommandTag().RowsAffected() == 0 {
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no key starting with '%s' found", key), nil)
	}

	return result, nil
}

// AddKeeperKeys inserts or updates the key-value pair(s) based on the passed models.KVS
// if isFlatten is enabled, multiple key-value pair(s) will be updated based on the Value from models.KVS
func (c *Client) AddKeeperKeys(kv models.KVS, isFlatten bool) ([]models.KeyOnly, errors.EdgeX) {
	var keyReps []models.KeyOnly

	if isFlatten {
		// process the value map and convert the fields and store to multiple key-value pairs
		txErr := pgx.BeginFunc(context.Background(), c.ConnPool, func(tx pgx.Tx) error {
			var err error
			keyReps, err = updateMultiKVSInTx(tx, kv.Key, kv.Value)
			return err
		})
		if txErr != nil {
			return nil, errors.NewCommonEdgeXWrapper(txErr)
		}
	} else {
		// store the value in a single key
		err := updateKVS(c.ConnPool, kv.Key, kv.Value)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}

		keyReps = []models.KeyOnly{models.KeyOnly(kv.Key)}
	}
	return keyReps, nil
}

// DeleteKeeperKeys deletes one key or multiple keys(with isRecurse enabled)
func (c *Client) DeleteKeeperKeys(key string, isRecurse bool) ([]models.KeyOnly, errors.EdgeX) {
	var exists bool
	var resp []models.KeyOnly
	var childKeyCount uint32
	ctx := context.Background()
	queryPattern := key + "/%"

	// check if the exact same key exists
	err := c.ConnPool.QueryRow(
		context.Background(),
		sqlCheckExistsByCol(configTableName, keyCol),
		key,
	).Scan(&exists)
	if err != nil {
		return nil, pgClient.WrapDBError(fmt.Sprintf("failed to query row by key '%s'", key), err)
	}

	// check if the key(s) start with the keyPrefix exist and get the count of the result
	err = c.ConnPool.QueryRow(
		context.Background(),
		sqlQueryCountByColAndLikePat(configTableName, keyCol),
		queryPattern,
	).Scan(&childKeyCount)
	if err != nil {
		return nil, pgClient.WrapDBError(fmt.Sprintf("failed to query row by key starts with '%s'", key), err)
	}

	if exists {
		// delete the exact same key
		_, err = c.ConnPool.Exec(ctx, sqlDeleteByColumns(configTableName, keyCol), key)
		if err != nil {
			return nil, pgClient.WrapDBError(fmt.Sprintf("failed to query row by key '%s'", key), err)
		}
		resp = []models.KeyOnly{models.KeyOnly(key)}
	}

	if childKeyCount == 0 {
		if !exists {
			// key is not found and no other keys starts with key exist
			return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("key '%s' not found", key), nil)
		}
	} else {
		if isRecurse {
			// also delete the keys starts with the same key (e.g., edgex/v3/core-data/Writable, edgex/v3/core-data/Database all starts with edgex/v3/core-data)
			sqlStatement := sqlDeleteByColAndLikePat(configTableName, keyCol, keyCol)
			rows, err := c.ConnPool.Query(ctx, sqlStatement, queryPattern)
			if err != nil {
				return nil, pgClient.WrapDBError(fmt.Sprintf("failed to delete row by key starts with '%s'", key), err)
			}

			var returnedKey string
			_, err = pgx.ForEachRow(rows, []any{&returnedKey}, func() error {
				resp = append(resp, models.KeyOnly(returnedKey))
				return nil
			})
			if err != nil {
				return nil, pgClient.WrapDBError("failed to scan returned keys to models.KeyOnly", err)
			}
		} else {
			// if isRecurse is not enabled, don't delete the keys starts with the same key
			return nil, errors.NewCommonEdgeX(errors.KindStatusConflict, fmt.Sprintf("keys having the same prefix %s exist and cannot be deleted", key), nil)
		}
	}

	return resp, nil
}

// updateKVS insert or update a single key-value pair with value is simply a string or a map
func updateKVS(connPool *pgxpool.Pool, key string, value any) errors.EdgeX {
	ctx := context.Background()
	var storedValueBytes []byte

	switch v := value.(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string:
		storedValueStr := cast.ToString(v)
		storedValueBytes = []byte(storedValueStr)
	default:
		encBytes, err := json.Marshal(v)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to marshal stored value %v with key '%s'", value, key), err)
		}
		storedValueBytes = encBytes
	}

	// encode the value to a base64 string
	storedValue := base64.StdEncoding.EncodeToString(storedValueBytes)

	var exists bool
	err := connPool.QueryRow(ctx, sqlCheckExistsByCol(configTableName, keyCol), key).Scan(&exists)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to query value by key '%s'", key), err)
	}

	if exists {
		// update the key
		_, err = connPool.Exec(ctx, sqlUpdateColsByCondCol(configTableName, keyCol, valueCol, modifiedCol),
			storedValue,
			time.Now().UTC(),
			key,
		)
		if err != nil {
			return pgClient.WrapDBError(fmt.Sprintf("failed to modified value by key '%s'", key), err)
		}
	} else {
		// insert the key
		_, err = connPool.Exec(ctx, sqlInsert(configTableName, keyCol, valueCol),
			key,
			storedValue,
		)
		if err != nil {
			return pgClient.WrapDBError(fmt.Sprintf("failed to insert value by key '%s'", key), err)
		}
	}
	return nil
}

// updateMultiKVSInTx insert or update the key-value pairs in a map within a transaction
func updateMultiKVSInTx(tx pgx.Tx, currentKey string, value any) ([]models.KeyOnly, errors.EdgeX) {
	ctx := context.Background()
	var keyReps []models.KeyOnly

	switch v := value.(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, []any:
		var exists bool
		var sqlStatement string

		err := tx.QueryRow(
			context.Background(),
			sqlCheckExistsByCol(configTableName, keyCol),
			currentKey,
		).Scan(&exists)
		if err != nil {
			return nil, pgClient.WrapDBError(fmt.Sprintf("failed to query rows by key '%s'", currentKey), err)
		}

		storedValueStr := cast.ToString(v)
		encStr := base64.StdEncoding.EncodeToString([]byte(storedValueStr))
		if exists {
			sqlStatement = sqlUpdateColsByCondCol(configTableName, keyCol, valueCol, modifiedCol)
			_, err = tx.Exec(ctx, sqlStatement, encStr, time.Now().UTC(), currentKey)
			if err != nil {
				return nil, pgClient.WrapDBError(fmt.Sprintf("failed to update row by key '%s'", currentKey), err)
			}
		} else {
			sqlStatement = sqlInsert(configTableName, keyCol, valueCol)
			_, err = tx.Exec(ctx, sqlStatement, currentKey, encStr)
			if err != nil {
				return nil, pgClient.WrapDBError(fmt.Sprintf("failed to insert row by key '%s'", currentKey), err)
			}
		}
		keyReps = append(keyReps, models.KeyOnly(currentKey))
	case map[string]any:
		for innerKey, element := range v {
			// if the element type is an empty map, do not add the inner key to the upper level Hash field
			if eleMap, ok := element.(map[string]any); ok && len(eleMap) == 0 {
				continue
			}

			resp, err := updateMultiKVSInTx(tx, path.Join(currentKey, innerKey), element)
			if err != nil {
				return nil, errors.NewCommonEdgeXWrapper(err)
			}
			keyReps = append(keyReps, resp...)
		}
	}
	return keyReps, nil
}
