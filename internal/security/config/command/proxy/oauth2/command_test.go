//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package oauth2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOauth2BadArguments tests command line errors
func TestOauth2BadArguments(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}

	// Setup test cases
	badArgTestcases := [][]string{
		{},                        // missing all required arguments
		{"-badarg"},               // invalid arg
		{"--client_id", "someid"}, // missing --client_secret & --admin_api_jwt
		{"--client_id", "someid", "--client_secret", "somesecret"}, // missing --admin_api_jwt
	}

	for _, args := range badArgTestcases {
		// Act
		command, err := NewCommand(lc, config, args)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, command)
	}
}

// TestOauth2Generate tests succesful mock call to Kong
func TestOauth2Generate(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	args := []string{
		"--client_id", "myid",
		"--client_secret", "mysecret",
		"--jwt", "randomJWT",
	}
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "POST" {
			t.Errorf("expected GET request, got %s instead", r.Method)
		}

		if !strings.HasSuffix(r.URL.EscapedPath(), "/oauth2/token") {
			t.Errorf("expected request to /%s, got %s instead", ".../oauth2/token", r.URL.EscapedPath())
		}

		jsonResponse := map[string]interface{}{
			"access_token": "sometoken",
		}
		_ = json.NewEncoder(w).Encode(jsonResponse)
	}))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	config.KongURL.Server = tsURL.Hostname()
	config.KongURL.ApplicationPortSSL, _ = strconv.Atoi(tsURL.Port())

	// Act
	command, err := NewCommand(lc, config, args)
	require.NoError(t, err)
	code, err := command.Execute()

	// Assert
	require.NoError(t, err)
	require.Equal(t, interfaces.StatusCodeExitNormal, code)
}
