//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"path"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper/constants"
)

// FlattenKeys enumerates the child keys that AddKeeperKeys derives from a
// flatten=true request and validates each of them. It mirrors the traversal of
// the database clients (updateMultiKVSInTx for Postgres and its Redis
// counterpart), which builds the child keys by joining the requested key with
// the field names of the JSON object without ever applying ValidateKeys to
// them. Without this check a flatten request can create keys containing
// characters that the direct (non-flatten) write path rejects, e.g. a field
// name with ':', '+', '#' or whitespace. Such keys cannot be read, updated or
// deleted through the API again (every request with an invalid key is
// rejected), and a poisoned key inside a watched configuration subtree breaks
// the configuration decoding of the services watching it.
func FlattenKeys(key string, value any) ([]string, errors.EdgeX) {
	var keys []string
	var offending string
	if err := flattenKeys(key, value, &keys, &offending); err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("flatten value contains an invalid key field name '%s' for key '%s': the key only allows characters between a to z, A to Z, 0 to 9, or one of -_ ~;=./%%", offending, key), nil)
	}
	if len(keys) == 0 {
		// the value is an empty JSON object, which creates no keys
		return keys, nil
	}
	return keys, nil
}

func flattenKeys(currentKey string, value any, keys *[]string, offending *string) error {
	switch v := value.(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, []any:
		if !constants.KeyAllowedCharsRegex.MatchString(currentKey) {
			*offending = currentKey
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid key", nil)
		}
		*keys = append(*keys, currentKey)
	case map[string]any:
		for innerKey, element := range v {
			if innerKey == "" {
				*offending = currentKey
				return errors.NewCommonEdgeX(errors.KindContractInvalid, "empty field name", nil)
			}
			// skip empty inner objects, mirroring the database traversal which
			// does not create keys for them
			if eleMap, ok := element.(map[string]any); ok && len(eleMap) == 0 {
				continue
			}
			if err := flattenKeys(path.Join(currentKey, innerKey), element, keys, offending); err != nil {
				return err
			}
		}
	default:
		*offending = currentKey
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unsupported value type", nil)
	}
	return nil
}
