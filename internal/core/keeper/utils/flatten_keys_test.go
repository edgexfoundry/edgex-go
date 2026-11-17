//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenKeys(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       any
		expected    []string
		expectError bool
	}{
		{
			"valid scalar value creates the requested key only",
			"edgex/v4/core-data/Writable/LogLevel", "DEBUG",
			[]string{"edgex/v4/core-data/Writable/LogLevel"}, false,
		},
		{
			"valid nested object creates one key per leaf",
			"edgex/v4/core-data/Writable",
			map[string]any{"MaxEventSize": 25000, "LogLevel": "DEBUG"},
			[]string{"edgex/v4/core-data/Writable/MaxEventSize", "edgex/v4/core-data/Writable/LogLevel"},
			false,
		},
		{
			"valid deep nesting",
			"edgex/v4/test",
			map[string]any{"a": map[string]any{"b": map[string]any{"c": true}}},
			[]string{"edgex/v4/test/a/b/c"},
			false,
		},
		{
			"empty inner object creates no keys",
			"edgex/v4/test", map[string]any{"empty": map[string]any{}},
			[]string(nil), false,
		},
		{
			"field name with colon is rejected",
			"edgex/v4/test", map[string]any{"Log:Level": "DEBUG"},
			nil, true,
		},
		{
			"field name with MQTT wildcard is rejected",
			"edgex/v4/test", map[string]any{"foo#": "DEBUG"},
			nil, true,
		},
		{
			"field name with whitespace is rejected",
			"edgex/v4/test", map[string]any{"Log Level": "DEBUG"},
			nil, true,
		},
		{
			"empty field name is rejected",
			"edgex/v4/test", map[string]any{"": "DEBUG"},
			nil, true,
		},
		{
			"field name with invalid character in nested level is rejected",
			"edgex/v4/test", map[string]any{"ok": map[string]any{"bad+name": "DEBUG"}},
			nil, true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			keys, err := FlattenKeys(testCase.key, testCase.value)
			if testCase.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			sort.Strings(keys)
			sort.Strings(testCase.expected)
			if len(testCase.expected) == 0 && len(keys) == 0 {
				assert.Len(t, keys, 0)
				return
			}
			assert.Equal(t, testCase.expected, keys)
		})
	}
}
