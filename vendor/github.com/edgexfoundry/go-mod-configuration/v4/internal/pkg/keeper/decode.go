//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"errors"
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/mitchellh/mapstructure"
)

// decode converts the key-value pairs from core keeper to the target configuration data type
func decode(prefix string, pairs []models.KVS, configTarget interface{}) error {
	// check if the prefix ends with the '/' char
	if !strings.HasSuffix(prefix, KeyDelimiter) {
		prefix += KeyDelimiter
	}

	raw := make(map[string]interface{})
	for _, p := range pairs {
		// Trim the prefix off our key first
		key := strings.TrimPrefix(p.Key, prefix)

		// Determine what map we're writing the value to. We split by '/'
		// to determine any sub-maps that need to be created.
		m := raw
		children := strings.Split(key, KeyDelimiter)
		if len(children) > 0 {
			key = children[len(children)-1]
			children = children[:len(children)-1]
			for _, child := range children {
				if m[child] == nil {
					m[child] = make(map[string]interface{})
				}

				subm, ok := m[child].(map[string]interface{})
				if !ok {
					return fmt.Errorf("child is both a data item and dir: %s", child)
				}

				m = subm
			}
		}
		value := p.Value
		switch value.(type) {
		case bool:
			m[key] = value
		case int:
			m[key] = value
		case int8:
			m[key] = value
		case int16:
			m[key] = value
		case int32:
			m[key] = value
		case int64:
			m[key] = value
		case float32:
			m[key] = value
		case float64:
			m[key] = value
		case string:
			m[key] = value
		default:
			return errors.New("unknown data type of the stored value")
		}
	}

	// Now decode into it
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:         nil,
		WeaklyTypedInput: true,
		Result:           configTarget,
	})
	if err != nil {
		return fmt.Errorf("json decoding failed, err: %v", err)
	}
	if err := decoder.Decode(raw); err != nil {
		return fmt.Errorf("json decoding failed, err: %v", err)
	}

	return nil
}
