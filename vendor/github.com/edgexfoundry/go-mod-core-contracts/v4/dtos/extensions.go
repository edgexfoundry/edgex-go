//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// mergeExtensions merges the extensions map into the already-marshaled data bytes at top level.
// The extensions keys will overwrite existing keys if there is a conflict.
func mergeExtensions(data []byte, extensions map[string]any, unmarshalFn func([]byte, any) error,
	marshalFn func(any) ([]byte, error)) ([]byte, error) {
	var m map[string]any
	if err := unmarshalFn(data, &m); err != nil {
		return nil, err
	}
	for k, v := range extensions {
		m[k] = v
	}
	return marshalFn(m)
}

// jsonUnmarshalUseNumber unmarshals JSON with UseNumber enabled to preserve numeric precision.
func jsonUnmarshalUseNumber(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(v); err != nil {
		return err
	}
	// Ensure no trailing non-whitespace content remains
	if err := dec.Decode(new(struct{})); err != io.EOF {
		if err == nil {
			return fmt.Errorf("invalid JSON: extra data after first value")
		}
		return err
	}
	return nil
}

// convertJSONNumbers recursively converts json.Number values in a map/slice to float64,
// matching the behavior of standard json.Unmarshal when decoding numbers into any-typed targets.
// This is called after origin is extracted (which uses json.Number.Int64() directly to preserve
// int64 precision), so the remaining numeric values in the map can safely use float64.
func convertJSONNumbers(v any) any {
	switch val := v.(type) {
	case json.Number:
		if f, err := val.Float64(); err == nil {
			return f
		}
		return val.String()
	case map[string]any:
		for k, v := range val {
			val[k] = convertJSONNumbers(v)
		}
		return val
	case []any:
		for i, v := range val {
			val[i] = convertJSONNumbers(v)
		}
		return val
	default:
		return v
	}
}

// normalizeMap recursively converts map[any]any (produced by CBOR decoding into any)
// to map[string]any for consistency with JSON-decoded maps. No-op for JSON-decoded data.
func normalizeMap(v any) any {
	switch val := v.(type) {
	case map[any]any:
		result := make(map[string]any, len(val))
		for k, v := range val {
			ks, ok := k.(string)
			if !ok {
				ks = fmt.Sprint(k)
			}
			result[ks] = normalizeMap(v)
		}
		return result
	case map[string]any:
		for k, v := range val {
			val[k] = normalizeMap(v)
		}
		return val
	case []any:
		for i, v := range val {
			val[i] = normalizeMap(v)
		}
		return val
	default:
		return v
	}
}

// popKey removes the key from the map and returns the value.
func popKey(m map[string]any, key string) any {
	v := m[key]
	delete(m, key)
	return v
}

// popStringValueFromKey removes the key from the map and returns its value as a string.
// Returns an error if the key is present but not a string type.
func popStringValueFromKey(m map[string]any, key string) (string, error) {
	v := popKey(m, key)
	switch val := v.(type) {
	case string:
		return val, nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("field %q must be a string, got %T", key, v)
	}
}
