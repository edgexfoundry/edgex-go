/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

const (
	envValueString = "envValue"
	envValueInt    = 123456789

	defaultHostValue = "defaultHost"
	defaultPortValue = 987654321
	defaultTypeValue = "defaultType"
)

func initializeTest(t *testing.T) RegistryInfo {
	os.Clearenv()
	return RegistryInfo{
		Host: defaultHostValue,
		Port: defaultPortValue,
		Type: defaultTypeValue,
	}
}

func TestEnvVariableUpdatesRegistryInfo(t *testing.T) {
	var tests = []struct {
		name              string
		envKey            string
		envValue          string
		expectedHostValue string
		expectedPortValue int
		expectedTypeValue string
	}{
		{"host", envKeyHost, envValueString, envValueString, defaultPortValue, defaultTypeValue},
		{"port", envKeyPort, strconv.Itoa(envValueInt), defaultHostValue, envValueInt, defaultTypeValue},
		{"type", envKeyType, envValueString, defaultHostValue, defaultPortValue, envValueString},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registryInfo := initializeTest(t)
			if err := os.Setenv(test.envKey, test.envValue); err != nil {
				t.Fail()
			}

			registryInfo = OverrideFromEnvironment(registryInfo)

			assert.Equal(t, registryInfo.Host, test.expectedHostValue)
			assert.Equal(t, registryInfo.Port, test.expectedPortValue)
			assert.Equal(t, registryInfo.Type, test.expectedTypeValue)
		})
	}
}

func TestNoEnvVariableDoesNotUpdateRegistryInfo(t *testing.T) {
	registryInfo := initializeTest(t)

	registryInfo = OverrideFromEnvironment(registryInfo)

	assert.Equal(t, registryInfo.Host, defaultHostValue)
	assert.Equal(t, registryInfo.Port, defaultPortValue)
	assert.Equal(t, registryInfo.Type, defaultTypeValue)
}
