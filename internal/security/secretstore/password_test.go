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
 *
 *******************************************************************************/

package secretstore

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateWithDefaults(t *testing.T) {
	rootToken := "s.Ga5jyNq6kNfRMVQk2LY1j9iu" // nolint:gosec
	mockLogger := logger.MockLogger{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gk := NewPasswordGenerator(mockLogger, "", []string{})
	cr := NewCred(&http.Client{}, rootToken, gk, "", logger.MockLogger{})

	p1, err := cr.GeneratePassword(ctx)
	require.NoError(t, err, "failed to create credential")
	p2, err := cr.GeneratePassword(ctx)
	require.NoError(t, err, "failed to create credential")
	assert.NotEqual(t, p1, p2, "each call to GeneratePassword should return a new password")
}

func TestRetrieveCred(t *testing.T) {
	credPath := "testCredPath"
	expected := "token"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data": {"username": "test-user", "password": "test-password"}}`))
		require.NoError(t, err)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, fmt.Sprintf("/%s", credPath), r.URL.EscapedPath())
		actual := r.Header.Get(VaultToken)
		assert.Equal(t, expected, actual)
	}))
	defer ts.Close()

	parsed, err := url.Parse(ts.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(parsed.Port())
	require.NoError(t, err)

	configuration := &config.ConfigurationStruct{
		SecretStore: config.SecretStoreInfo{
			Host:     parsed.Hostname(),
			Port:     port,
			Protocol: "https",
		},
	}

	mockLogger := logger.MockLogger{}
	cr := NewCred(
		pkg.NewRequester(mockLogger).Insecure(),
		"token",
		NewPasswordGenerator(mockLogger, "", []string{}),
		configuration.SecretStore.GetBaseURL(),
		mockLogger)
	pair, err := cr.retrieve(credPath)
	require.NoError(t, err)

	if pair.User != "test-user" || pair.Password != "test-password" {
		t.Errorf("failed to parse credential pair")
	}
}
