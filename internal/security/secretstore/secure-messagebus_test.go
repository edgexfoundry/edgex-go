/*******************************************************************************
 * Copyright 2021 Intel Corporation
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
 *
 *******************************************************************************/

package secretstore

import (
	"os"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
)

func TestConfigureSecureMessageBus(t *testing.T) {
	secureMessageBus := config.SecureMessageBusInfo{
		KuiperConfigPath:      "./testdata/edgex.yaml",
		KuiperConnectionsPath: "./testdata/connection.yaml",
	}

	validExpected := UserPasswordPair{
		User:     "testUser",
		Password: "testPassword",
	}

	tests := []struct {
		Name                 string
		Type                 string
		ConnectionFileExists bool
		Credentials          UserPasswordPair
		Expected             *UserPasswordPair
		ExpectError          bool
	}{
		{"valid redis - both files", redisSecureMessageBusType, true, validExpected, &validExpected, false},
		{"valid redis - no connection file", redisSecureMessageBusType, false, validExpected, &validExpected, false},
		{"valid blank", blankSecureMessageBusType, false, validExpected, nil, false},
		{"valid none", noneSecureMessageBusType, false, validExpected, nil, false},
		{"invalid type", "bogus", false, validExpected, nil, true},
		{"invalid mqtt", mqttSecureMessageBusType, false, validExpected, nil, true},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			_ = os.Remove(secureMessageBus.KuiperConfigPath)
			_ = os.Remove(secureMessageBus.KuiperConnectionsPath)

			defer func() {
				_ = os.Remove(secureMessageBus.KuiperConfigPath)
				_ = os.Remove(secureMessageBus.KuiperConnectionsPath)
			}()

			if test.Expected != nil {
				_, err := os.Create(secureMessageBus.KuiperConfigPath)
				require.NoError(t, err)

				if test.ConnectionFileExists {
					_, err := os.Create(secureMessageBus.KuiperConnectionsPath)
					require.NoError(t, err)
				}
			}

			secureMessageBus.Type = test.Type
			err := ConfigureSecureMessageBus(secureMessageBus, test.Credentials, logger.NewMockClient())
			if test.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if test.Expected == nil {
				// Source Config file should not have been written
				_, err = os.Stat(secureMessageBus.KuiperConfigPath)
				require.True(t, os.IsNotExist(err))

				// Connections file should not have been written
				_, err = os.Stat(secureMessageBus.KuiperConnectionsPath)
				require.True(t, os.IsNotExist(err))

				return
			}

			// Source Config file should have been written
			contents, err := os.ReadFile(secureMessageBus.KuiperConfigPath)
			require.NoError(t, err)
			assert.True(t, strings.Contains(string(contents), test.Expected.User))
			assert.True(t, strings.Contains(string(contents), test.Expected.Password))

			if test.ConnectionFileExists {
				// Connections file should have been written
				contents, err = os.ReadFile(secureMessageBus.KuiperConnectionsPath)
				require.NoError(t, err)
				assert.True(t, strings.Contains(string(contents), test.Expected.User))
				assert.True(t, strings.Contains(string(contents), test.Expected.Password))
			} else {
				// Connections file should not have been written
				_, err = os.Stat(secureMessageBus.KuiperConnectionsPath)
				require.True(t, os.IsNotExist(err))
			}
		})
	}
}
