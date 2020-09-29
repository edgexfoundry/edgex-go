//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package adduser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelpBadArg tests unknown arg handler
func TestAddUserBadArg(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	badArgTestcases := [][]string{
		{},                          // missing token type
		{"-badarg"},                 // invalid arg
		{"--token-type", "invalid"}, // invalid token type
		{"--token-type", "jwt"},     // missing --user
		{"--token-type", "jwt", "--user", "someuser"},                           // missing --algorithm (jwt)
		{"--token-type", "jwt", "--user", "someuser", "--algorithm", "invalid"}, // invalid algorithm (jwt)
		{"--token-type", "jwt", "--user", "someuser", "--algorithm", "RS256"},   // missing public_key (jwt)
		{"--token-type", "oauth2"},                                              // missing --user
	}

	for _, args := range badArgTestcases {
		// Act
		command, err := NewCommand(lc, config, args)

		// Assert
		assert.Error(t, err, "Args: %v", args)
		assert.Nil(t, command)
	}
}

// TestAddUserJWT tests functionality of adduser command using JWT option
func TestAddUserJWT(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/consumers":
			require.Equal(t, "POST", r.Method)
			w.WriteHeader(http.StatusCreated)
		case "/consumers/someuser/acls":
			require.Equal(t, "POST", r.Method)
			w.WriteHeader(http.StatusCreated)
		case "/consumers/someuser/jwt":
			require.Equal(t, "POST", r.Method)
			w.WriteHeader(http.StatusCreated)
			jsonResponse := map[string]interface{}{
				"key": "bad060a9-0e2b-47ba-98d5-9d622e2322b5",
			}
			json.NewEncoder(w).Encode(jsonResponse)
		default:
			t.Fatal(fmt.Sprintf("Unexpected call to URL %s", r.URL.EscapedPath()))
		}
	}))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	config.KongURL.Server = tsURL.Hostname()
	config.KongURL.AdminPort, _ = strconv.Atoi(tsURL.Port())

	// Act
	command, err := NewCommand(lc, config, []string{
		"--token-type", "jwt",
		"--user", "someuser",
		"--algorithm", "RS256",
		"--public_key", "testdata/rsa.pub",
	})
	require.NoError(t, err)

	code, err := command.Execute()

	// Assert
	require.NoError(t, err)
	require.Equal(t, interfaces.StatusCodeExitNormal, code)
}

// TestAddUserOAuth2 tests functionality of adduser command using OAuth2 option
func TestAddUserOAuth2(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/consumers":
			require.Equal(t, "POST", r.Method)
			w.WriteHeader(http.StatusCreated)
		case "/consumers/someuser/acls":
			require.Equal(t, "POST", r.Method)
			w.WriteHeader(http.StatusCreated)
		case "/consumers/someuser/oauth2":
			require.Equal(t, "POST", r.Method)
			w.WriteHeader(http.StatusCreated)
			jsonResponse := map[string]interface{}{
				"key":           "bad060a9-0e2b-47ba-98d5-9d622e2322b5",
				"client_id":     "7240fdd9-1665-419b-a8c5-5691ca03af7c",
				"client_secret": "d3191db3-8468-4a3c-87fb-df4fccfee983",
			}
			json.NewEncoder(w).Encode(jsonResponse)
		default:
			t.Fatal(fmt.Sprintf("Unexpected call to URL %s", r.URL.EscapedPath()))
		}
	}))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	config.KongURL.Server = tsURL.Hostname()
	config.KongURL.AdminPort, _ = strconv.Atoi(tsURL.Port())

	// Act
	command, err := NewCommand(lc, config, []string{
		"--token-type", "oauth2",
		"--user", "someuser",
		"--redirect_uris", "https://placeholder",
	})
	require.NoError(t, err)

	code, err := command.Execute()

	// Assert
	require.NoError(t, err)
	require.Equal(t, interfaces.StatusCodeExitNormal, code)
}
