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
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
)

func TestConfigureSecureMessageBus(t *testing.T) {
	secureMessageBus := config.SecureMessageBusInfo{
		KuiperConfigPath:      "./testdata/edgex.yaml",
		KuiperConnectionsPath: "./testdata/connection.yaml",
	}
	expectedMqttConfigContent := `
application_conf:
  port: 5571
  protocol: tcp
  server: localhost
  topic: application
default:
  optional:
    Username: testUser
    Password: testPassword
  port: 1883
  protocol: tcp
  server: localhost
  connectionSelector: edgex.mqttMsgBus
  topic: rules-events
  type: mqtt
mqtt_conf:
  optional:
    ClientId: client1
  port: 1883
  protocol: tcp
  server: localhost
  topic: events
  type: mqtt
`
	expectedRedisConfigContent := `
application_conf:
  port: 5571
  protocol: tcp
  server: localhost
  topic: application
default:
  optional:
    Username: testUser
    Password: testPassword
  port: 6379
  protocol: redis
  server: localhost
  connectionSelector: edgex.redisMsgBus
  topic: rules-events
  type: redis
mqtt_conf:
  optional:
    ClientId: client1
  port: 1883
  protocol: tcp
  server: localhost
  topic: events
  type: mqtt
`

	expectedMqttConnectionsContent := `
edgex:
  mqttMsgBus: #connection key
    protocol: tcp
    server: localhost
    port: 1883
    type: mqtt
    optional:
      Username: testUser
      Password: testPassword
`
	expectedRedisConnectionsContent := `
edgex:
  redisMsgBus: #connection key
    protocol: redis
    server: localhost
    port: 6379
    type: redis
    optional:
      Username: testUser
      Password: testPassword
`
	creds := UserPasswordPair{
		User:     "testUser",
		Password: "testPassword",
	}

	tests := []struct {
		Name                 string
		Type                 string
		ConnectionFileExists bool
		Credentials          UserPasswordPair
		ExpectedConfig       *string
		ExpectedConnnection  *string
		ExpectError          bool
	}{
		{"valid redis - both files", redisSecureMessageBusType, true, creds, &expectedRedisConfigContent, &expectedRedisConnectionsContent, false},
		{"valid redis - no connection file", redisSecureMessageBusType, false, creds, &expectedRedisConfigContent, nil, false},
		{"valid mqtt - both files", mqttSecureMessageBusType, false, creds, &expectedMqttConfigContent, &expectedMqttConnectionsContent, false},
		{"valid mqtt - no connection file", mqttSecureMessageBusType, false, creds, &expectedMqttConfigContent, nil, false},
		{"valid blank", blankSecureMessageBusType, false, creds, nil, nil, false},
		{"valid none", noneSecureMessageBusType, false, creds, nil, nil, false},
		{"invalid type", "bogus", false, creds, nil, nil, true},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			_ = os.Remove(secureMessageBus.KuiperConfigPath)
			_ = os.Remove(secureMessageBus.KuiperConnectionsPath)

			defer func() {
				_ = os.Remove(secureMessageBus.KuiperConfigPath)
				_ = os.Remove(secureMessageBus.KuiperConnectionsPath)
			}()

			if test.ExpectedConfig != nil {
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

			if test.ExpectedConfig == nil {
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
			assert.Equal(t, *test.ExpectedConfig, string(contents))

			if test.ConnectionFileExists {
				// Connections file should have been written
				contents, err = os.ReadFile(secureMessageBus.KuiperConnectionsPath)
				require.NoError(t, err)
				assert.Equal(t, *test.ExpectedConnnection, string(contents))
			} else {
				// Connections file should not have been written
				_, err = os.Stat(secureMessageBus.KuiperConnectionsPath)
				require.True(t, os.IsNotExist(err))
			}
		})
	}
}
