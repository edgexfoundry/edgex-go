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

package configuration

import (
	"os"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"

	"github.com/stretchr/testify/assert"
)

const (
	envValue = "consul://localhost:8500"

	expectedTypeValue = "consul"
	expectedHostValue = "localhost"
	expectedPortValue = 8500

	defaultHostValue = "defaultHost"
	defaultPortValue = 987654321
	defaultTypeValue = "defaultType"

	defaultStartupDuration = 30
	defaultStartupInterval = 1
)

func initializeTest(t *testing.T) (config.RegistryInfo, config.StartupInfo) {
	os.Clearenv()
	regInfo := config.RegistryInfo{
		Host: defaultHostValue,
		Port: defaultPortValue,
		Type: defaultTypeValue,
	}
	startupInfo := config.StartupInfo{
		Duration: defaultStartupDuration,
		Interval: defaultStartupInterval,
	}

	return regInfo, startupInfo
}

func TestEnvVariableUpdatesRegistryInfo(t *testing.T) {
	registryInfo, startupInfo := initializeTest(t)

	if err := os.Setenv(envKeyUrl, envValue); err != nil {
		t.Fail()
	}
	registryInfo, startupInfo = OverrideFromEnvironment(registryInfo, startupInfo)

	assert.Equal(t, registryInfo.Host, expectedHostValue)
	assert.Equal(t, registryInfo.Port, expectedPortValue)
	assert.Equal(t, registryInfo.Type, expectedTypeValue)
}

func TestNoEnvVariableDoesNotUpdateRegistryInfo(t *testing.T) {
	registryInfo, startupInfo := initializeTest(t)

	registryInfo, startupInfo = OverrideFromEnvironment(registryInfo, startupInfo)

	assert.Equal(t, registryInfo.Host, defaultHostValue)
	assert.Equal(t, registryInfo.Port, defaultPortValue)
	assert.Equal(t, registryInfo.Type, defaultTypeValue)
}
