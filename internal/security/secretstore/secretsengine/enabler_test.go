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

package secretsengine

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v2/secrets/mocks"
)

func TestNewSecretsEngine(t *testing.T) {
	tests := []struct {
		name       string
		mountPath  string
		engineType string
	}{
		{"New kv type of secrets engine", "kv-1-test/", KeyValue},
		{"New consul type of secrets engine", "consul-test/", Consul},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := New(tt.mountPath, tt.engineType)
			require.Equal(t, tt.mountPath, instance.mountPoint)
			require.Equal(t, tt.engineType, instance.engineType)
		})
	}
}

func TestEnableSecretsEngine(t *testing.T) {
	lc := logger.MockLogger{}
	testToken := "fake-token"
	unsupportedEngTypeErr := errors.New("Unsupported secrets engine type")

	tests := []struct {
		name             string
		rootToken        *string
		mountPoint       string
		engineType       string
		kvInstalled      bool
		consulInstalled  bool
		clientCallFailed bool
		expectError      bool
	}{
		{"Ok:Enable kv secrets engine not installed yet with client call ok", &testToken, "kv-1-test",
			KeyValue, false, false, false, false},
		{"Ok:Enable consul secrets engine not installed yet with client call ok", &testToken, "consul-test",
			Consul, false, false, false, false},
		{"Ok:Enable kv secrets engine already installed with client call ok (1)", &testToken, "kv-1-test",
			KeyValue, true, false, false, false},
		{"Ok:Enable consul secrets engine already installed with client call ok (1)", &testToken, "consul-test",
			Consul, false, true, false, false},
		{"Ok:Enable kv secrets engine already installed with client call ok (2)", &testToken, "kv-1-test",
			KeyValue, true, true, false, false},
		{"Ok:Enable consul secrets engine already installed with client call ok (2)", &testToken, "consul-test",
			Consul, true, true, false, false},
		{"Bad:Enable kv secrets engine not installed yet but client call failed", &testToken, "kv-1-test",
			KeyValue, false, false, true, true},
		{"Bad:Enable consul secrets engine not installed yet but client call failed", &testToken, "consul-test",
			Consul, false, false, true, true},
		{"Bad:Enable kv secrets engine already installed but client call failed (1)", &testToken, "kv-1-test",
			KeyValue, true, false, true, true},
		{"Bad:Enable consul secrets engine already installed but client call failed (1)", &testToken, "consul-test",
			Consul, false, true, true, true},
		{"Bad:Enable kv secrets engine already installed but client call failed (2)", &testToken, "kv-1-test",
			KeyValue, true, true, true, true},
		{"Bad:Enable consul secrets engine already installed but client call failed (2)", &testToken, "consul-test",
			Consul, true, true, true, true},
		{"Bad:Enable kv secrets engine with nil token", nil, "kv-1-test",
			KeyValue, false, true, false, true},
		{"Bad:Enable consul secrets engine with nil token", nil, "consul-test",
			Consul, true, false, false, true},
		{"Bad:Unsupported secrets engine type", &testToken, "whatever",
			"unsupported", false, false, false, true},
	}

	for _, test := range tests {
		// this local copy is to ensure test is thread-safe as we are running in parallel
		localTest := test
		t.Run(localTest.name, func(t *testing.T) {
			// run all tests in parallel
			t.Parallel()

			var chkErr error
			var enableClientErr error

			// to simplify testing, assume both errors when client calls failed
			if localTest.clientCallFailed {
				chkErr = errors.New("CheckSecretEngineInstalled called failed")
				enableClientErr = errors.New("EnableKVSecretEngine called failed")
			}

			mockClient := &mocks.SecretStoreClient{}
			mockClient.On("CheckSecretEngineInstalled", mock.Anything, mock.Anything, KeyValue).
				Return(localTest.kvInstalled, chkErr)
			mockClient.On("CheckSecretEngineInstalled", mock.Anything, mock.Anything, Consul).
				Return(localTest.consulInstalled, chkErr)
			mockClient.On("CheckSecretEngineInstalled", mock.Anything, mock.Anything, mock.Anything).
				Return(false, chkErr)
			mockClient.On("EnableKVSecretEngine", mock.Anything, localTest.mountPoint, kvVersion).
				Return(enableClientErr)
			mockClient.On("EnableKVSecretEngine", mock.Anything, mock.Anything, mock.Anything).
				Return(unsupportedEngTypeErr)
			mockClient.On("EnableConsulSecretEngine", mock.Anything, localTest.mountPoint, defaultConsulTokenLeaseTtl).
				Return(enableClientErr)
			mockClient.On("EnableConsulSecretEngine", mock.Anything, mock.Anything, mock.Anything).
				Return(unsupportedEngTypeErr)

			err := New(localTest.mountPoint, localTest.engineType).
				Enable(localTest.rootToken, lc, mockClient)

			if localTest.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
