//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//
package fileprovider

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleJSON = `{
	"service-name": {
	  "edgex_use_defaults": true,
	  "custom_policy": {
		"path": {
		  "secret/non/standard/location/*": {
		    "capabilities": [ "list", "read" ]
		  }
		}
	  },
	  "custom_token_parameters": { "custom_option": "custom_vaule" }
	}
  }`

func TestLoadTokenConfig(t *testing.T) {
	stringReader := strings.NewReader(sampleJSON)
	mockFileIoPerformer := &mocks.FileIoPerformer{}
	mockFileIoPerformer.On("OpenFileReader", "dummy-file", os.O_RDONLY, os.FileMode(0400)).Return(stringReader, nil)

	var tokenConf TokenConfFile
	err := LoadTokenConfig(mockFileIoPerformer, "dummy-file", &tokenConf)
	assert.NoError(t, err)

	aService := tokenConf["service-name"]
	assert.NotNil(t, aService)
	assert.Equal(t, true, aService.UseDefaults)
	assert.Contains(t, aService.CustomPolicy, "path")
	var path = aService.CustomPolicy["path"].(map[string]interface{})
	assert.Contains(t, path, "secret/non/standard/location/*")
	// Don't need to go further down the type assertion rabbit hole to prove that this is working
	assert.Contains(t, aService.CustomTokenParameters, "custom_option")
}

func TestLoadTokenConfigError1(t *testing.T) {
	stringReader := strings.NewReader(sampleJSON)
	mockFileIoPerformer := &mocks.FileIoPerformer{}
	mockFileIoPerformer.On("OpenFileReader", "dummy-file", os.O_RDONLY, os.FileMode(0400)).Return(stringReader, errors.New("an error"))

	var tokenConf TokenConfFile
	err := LoadTokenConfig(mockFileIoPerformer, "dummy-file", &tokenConf)
	assert.Error(t, err)
}

func TestLoadTokenConfigError2(t *testing.T) {
	stringReader := strings.NewReader("in{valid")
	mockFileIoPerformer := &mocks.FileIoPerformer{}
	mockFileIoPerformer.On("OpenFileReader", "dummy-file", os.O_RDONLY, os.FileMode(0400)).Return(stringReader, nil)

	var tokenConf TokenConfFile
	err := LoadTokenConfig(mockFileIoPerformer, "dummy-file", &tokenConf)
	assert.Error(t, err)
}

func TestMergeWithNoDuplicates(t *testing.T) {
	map1 := TokenConfFile{
		"key1": ServiceKey{},
		"key2": ServiceKey{},
	}

	map2 := TokenConfFile{
		"key3": ServiceKey{},
		"key4": ServiceKey{},
		"key5": ServiceKey{},
	}

	// merge two maps with no duplicated keys from each other
	merged := map1.mergeWith(map2)

	assert.Equal(t, 5, len(merged), "expect 5 entries in merged map")

	// should get all keys in the merged map
	expectedKeys := []string{"key1", "key2", "key3", "key4", "key5"}
	for _, k := range expectedKeys {
		assert.True(t, merged.keyExists(k))
	}
}

func TestMergeWithEmptyKey(t *testing.T) {
	map1 := TokenConfFile{
		"key1": ServiceKey{},
		"key2": ServiceKey{},
	}
	// merge with an empty key case
	empty := TokenConfFile{}
	merged := map1.mergeWith(empty)
	assert.Equal(t, map1, merged)

	// empty key merges with non-empty map
	merged = empty.mergeWith(map1)
	assert.Equal(t, map1, merged)
}

func TestMergeWithDuplicateKeys(t *testing.T) {
	map1 := TokenConfFile{
		"key1": ServiceKey{},
		"key2": ServiceKey{},
	}

	map2 := TokenConfFile{
		"key3": ServiceKey{},
		"key4": ServiceKey{},
		"key5": ServiceKey{},
		"key7": ServiceKey{},
	}

	// some duplicate keys with other maps
	map3 := TokenConfFile{
		"key1": ServiceKey{UseDefaults: true},
		"key3": ServiceKey{UseDefaults: true},
		"key7": ServiceKey{UseDefaults: true},
	}

	// merge two maps with one duplicated key
	merged := map1.mergeWith(map3)

	assert.Equal(t, 4, len(merged), "expect 4 entries in merged map")
	expectedKeys := []string{"key1", "key2", "key3", "key7"}
	expectedServiceKeyValues := []ServiceKey{
		ServiceKey{UseDefaults: true},
		ServiceKey{UseDefaults: false},
		ServiceKey{UseDefaults: true},
		ServiceKey{UseDefaults: true}}
	for i, k := range expectedKeys {
		assert.True(t, merged.keyExists(k))
		assert.Equal(t, expectedServiceKeyValues[i].UseDefaults, merged[k].UseDefaults)
	}

	// use merged map and merge again with more than one duplicated keys
	merged = merged.mergeWith(map2)
	assert.Equal(t, 6, len(merged), "expect 6 entries in merged map")
	expectedKeys = []string{"key1", "key2", "key3", "key4", "key5", "key7"}
	expectedServiceKeyValues = []ServiceKey{
		ServiceKey{UseDefaults: true},
		ServiceKey{UseDefaults: false},
		ServiceKey{UseDefaults: false},
		ServiceKey{UseDefaults: false},
		ServiceKey{UseDefaults: false},
		ServiceKey{UseDefaults: false}}
	for i, k := range expectedKeys {
		assert.True(t, merged.keyExists(k))
		assert.Equal(t, expectedServiceKeyValues[i].UseDefaults, merged[k].UseDefaults)
	}
}

func TestGetTokenConfigFromEnv(t *testing.T) {
	oringEnv := os.Getenv(addSecretstoreTokensEnvKey)
	defer func() {
		os.Setenv(addSecretstoreTokensEnvKey, oringEnv)
	}()

	tests := []struct {
		name            string
		serviceNameList string
		expectedKeys    []string
		expectError     bool
	}{
		{
			name:            "Empty list",
			serviceNameList: "",
			expectedKeys:    []string{},
			expectError:     false,
		},
		{
			name:            "One service name",
			serviceNameList: "service1",
			expectedKeys:    []string{"service1"},
			expectError:     false,
		},
		{
			name:            "Two service names",
			serviceNameList: "service1, service2",
			expectedKeys:    []string{"service1", "service2"},
			expectError:     false,
		},
		{
			name:            "More than 2 service names",
			serviceNameList: "service1,service2, service3",
			expectedKeys:    []string{"service1", "service2", "service3"},
			expectError:     false,
		},
		{
			name:            "With some empty service names",
			serviceNameList: "service1, ,service3",
			expectedKeys:    []string{"service1", "service3"},
			expectError:     false,
		},
		{
			name:            "Malformed service name list",
			serviceNameList: "slash/,,backslash\\",
			expectedKeys:    []string{},
			expectError:     true,
		},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(test.name, func(t *testing.T) {
			os.Setenv(addSecretstoreTokensEnvKey, currentTest.serviceNameList)
			tokenFileMap, err := GetTokenConfigFromEnv()

			if currentTest.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, len(currentTest.expectedKeys), len(tokenFileMap))

			for _, expectedServiceName := range currentTest.expectedKeys {
				_, found := tokenFileMap[expectedServiceName]
				assert.Truef(t, found, "expected service name: %s", expectedServiceName)
			}
		})
	}
}
