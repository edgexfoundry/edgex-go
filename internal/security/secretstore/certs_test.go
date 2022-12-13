/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Inc.
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
 * @author: Tingyu Zeng, Dell
 * @author: Daniel Harms, Dell
 *******************************************************************************/

package secretstore

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

func TestRetrieve(t *testing.T) {
	certPath := "testCertPath"
	expected := "token"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data": {"cert": "test-certificate", "key": "test-private-key"}}`))
		require.NoError(t, err)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, fmt.Sprintf("/%s", certPath), r.URL.EscapedPath())
		actual := r.Header.Get(VaultToken)
		assert.Equal(t, expected, actual)
	}))
	defer ts.Close()

	parsed, err := url.Parse(ts.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(parsed.Port())
	require.NoError(t, err)

	configuration := &config.ConfigurationStruct{}
	configuration.SecretStore = config.SecretStoreInfo{
		Host:     parsed.Hostname(),
		Port:     port,
		Protocol: "https",
	}

	mockLogger := logger.MockLogger{}
	cs := NewCerts(
		pkg.NewRequester(mockLogger).Insecure(),
		certPath,
		"token",
		configuration.SecretStore.GetBaseURL(),
		mockLogger)
	cp, err := cs.retrieve()
	require.NoError(t, err)

	if cp.Cert != "test-certificate" || cp.Key != "test-private-key" {
		t.Errorf("failed to parse certificate key pair")
	}
}
