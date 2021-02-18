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
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient/mocks"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func TestNewSecretsEngine(t *testing.T) {
	tests := []struct {
		name       string
		mountPath  string
		engineType string
	}{
		{"New kv type of secrets engine", "kv-1-test/", secretstoreclient.KeyValue},
		{"New consul type of secrets engine", "consul-test/", secretstoreclient.Consul},
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
			secretstoreclient.KeyValue, false, false, false, false},
		{"Ok:Enable consul secrets engine not installed yet with client call ok", &testToken, "consul-test",
			secretstoreclient.Consul, false, false, false, false},
		{"Ok:Enable kv secrets engine already installed with client call ok (1)", &testToken, "kv-1-test",
			secretstoreclient.KeyValue, true, false, false, false},
		{"Ok:Enable consul secrets engine already installed with client call ok (1)", &testToken, "consul-test",
			secretstoreclient.Consul, false, true, false, false},
		{"Ok:Enable kv secrets engine already installed with client call ok (2)", &testToken, "kv-1-test",
			secretstoreclient.KeyValue, true, true, false, false},
		{"Ok:Enable consul secrets engine already installed with client call ok (2)", &testToken, "consul-test",
			secretstoreclient.Consul, true, true, false, false},
		{"Bad:Enable kv secrets engine not installed yet but client call failed", &testToken, "kv-1-test",
			secretstoreclient.KeyValue, false, false, true, true},
		{"Bad:Enable consul secrets engine not installed yet but client call failed", &testToken, "consul-test",
			secretstoreclient.Consul, false, false, true, true},
		{"Bad:Enable kv secrets engine already installed but client call failed (1)", &testToken, "kv-1-test",
			secretstoreclient.KeyValue, true, false, true, true},
		{"Bad:Enable consul secrets engine already installed but client call failed (1)", &testToken, "consul-test",
			secretstoreclient.Consul, false, true, true, true},
		{"Bad:Enable kv secrets engine already installed but client call failed (2)", &testToken, "kv-1-test",
			secretstoreclient.KeyValue, true, true, true, true},
		{"Bad:Enable consul secrets engine already installed but client call failed (2)", &testToken, "consul-test",
			secretstoreclient.Consul, true, true, true, true},
		{"Bad:Enable kv secrets engine with nil token", nil, "kv-1-test",
			secretstoreclient.KeyValue, false, true, false, true},
		{"Bad:Enable consul secrets engine with nil token", nil, "consul-test",
			secretstoreclient.Consul, true, false, false, true},
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
			var enableStatusCode int
			var enableClientErr error

			// to simplify testing, assume both errors when client calls failed
			if localTest.clientCallFailed {
				chkErr = errors.New("CheckSecretEngineInstalled called failed")
				enableClientErr = errors.New("EnableKVSecretEngine called failed")
				enableStatusCode = http.StatusBadRequest
			} else {
				enableStatusCode = http.StatusNoContent
			}

			mockClient := &mocks.MockSecretStoreClient{}
			mockClient.On("CheckSecretEngineInstalled", mock.Anything, mock.Anything, secretstoreclient.KeyValue).
				Return(localTest.kvInstalled, chkErr)
			mockClient.On("CheckSecretEngineInstalled", mock.Anything, mock.Anything, secretstoreclient.Consul).
				Return(localTest.consulInstalled, chkErr)
			mockClient.On("CheckSecretEngineInstalled", mock.Anything, mock.Anything, mock.Anything).
				Return(false, chkErr)
			mockClient.On("EnableKVSecretEngine", mock.Anything, localTest.mountPoint, kvVersion).
				Return(enableStatusCode, enableClientErr)
			mockClient.On("EnableKVSecretEngine", mock.Anything, mock.Anything, mock.Anything).
				Return(enableStatusCode, unsupportedEngTypeErr)
			mockClient.On("EnableConsulSecretEngine", mock.Anything, localTest.mountPoint, defaultConsulTokenLeaseTtl).
				Return(enableStatusCode, enableClientErr)
			mockClient.On("EnableConsulSecretEngine", mock.Anything, mock.Anything, mock.Anything).
				Return(enableStatusCode, unsupportedEngTypeErr)

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
