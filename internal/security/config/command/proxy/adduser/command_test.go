//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package adduser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

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

	// Create a mock logger and grab the default configuration struct
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}

	// Test variables
	tokenType := "jwt"
	jwt := "s0meJWT"
	user := "someuser"
	algorithm := "RS256"
	publicKey := "testdata/rsa.pub"

	// Create a mock server for handling the command requests
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check to make sure the JWT authorization header exists
		assert.NotNil(t, r.Header.Values(internal.AuthHeaderTitle))
		require.Equal(t, internal.BearerLabel+jwt, r.Header.Values(internal.AuthHeaderTitle)[0])

		switch r.URL.EscapedPath() {

		// Testing --> add a consumer
		case "/admin/consumers":
			require.Equal(t, "POST", r.Method)
			w.WriteHeader(http.StatusCreated)

		// Testing --> enable ACL plugin for specific consumer
		case "/admin/consumers/someuser/acls":
			require.Equal(t, "POST", r.Method)
			w.WriteHeader(http.StatusCreated)

		// Testing --> enable JWT plugin for specific consumer
		case "/admin/consumers/someuser/jwt":
			require.Equal(t, "POST", r.Method)
			w.WriteHeader(http.StatusCreated)
			jsonResponse := map[string]interface{}{
				"key": "bad060a9-0e2b-47ba-98d5-9d622e2322b5",
			}
			err := json.NewEncoder(w).Encode(jsonResponse)
			require.NoError(t, err)

		// Testing --> fail if we don't recognize the URL in the request
		default:
			t.Fatalf("Unexpected call to URL %s", r.URL.EscapedPath())
		}
	}))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	config.KongURL.Server = tsURL.Hostname()
	config.KongURL.ApplicationPortSSL, _ = strconv.Atoi(tsURL.Port())

	// Setup command "addUser w/JWT"
	command, err := NewCommand(lc, config, []string{
		"--token-type", tokenType,
		"--user", user,
		"--algorithm", algorithm,
		"--public_key", publicKey,
		"--jwt", jwt,
	})

	require.NoError(t, err)

	// Execute command "addUser w/JWT"
	code, err := command.Execute()

	// Assert execution return
	require.NoError(t, err)
	require.Equal(t, interfaces.StatusCodeExitNormal, code)
}

// #3564: Deprecate for Ireland; commenting in case user community needs back in Jakarta stabilization release
// TestAddUserOAuth2 tests functionality of adduser command using OAuth2 option
// func TestAddUserOAuth2(t *testing.T) {

// 	// Create a mock logger and grab the default configuration struct
// 	lc := logger.MockLogger{}
// 	config := &config.ConfigurationStruct{}

// 	// Test variables
// 	tokenType := "oauth2"
// 	jwt := "s0meJWT"
// 	user := "someuser"
// 	redirectUris := "https://placeholder"

// 	// (placeholder)
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		// Check to make sure the JWT authorization header exists
// 		assert.NotNil(t, r.Header.Values(internal.AuthHeaderTitle))
// 		require.Equal(t, internal.BearerLabel+jwt, r.Header.Values(internal.AuthHeaderTitle)[0])

// 		switch r.URL.EscapedPath() {

// 		// Testing --> add a consumer
// 		case "/admin/consumers":
// 			require.Equal(t, "POST", r.Method)
// 			w.WriteHeader(http.StatusCreated)

// 		// Testing --> enable ACL plugin for specific consumer
// 		case "/admin/consumers/someuser/acls":
// 			require.Equal(t, "POST", r.Method)
// 			w.WriteHeader(http.StatusCreated)

// 		// Testing --> enable JWT plugin for specific consumer
// 		case "/admin/consumers/someuser/oauth2":
// 			require.Equal(t, "POST", r.Method)
// 			w.WriteHeader(http.StatusCreated)
// 			jsonResponse := map[string]interface{}{
// 				"key":           "bad060a9-0e2b-47ba-98d5-9d622e2322b5",
// 				"client_id":     "7240fdd9-1665-419b-a8c5-5691ca03af7c",
// 				"client_secret": "d3191db3-8468-4a3c-87fb-df4fccfee983",
// 			}
// 			json.NewEncoder(w).Encode(jsonResponse)

// 		// Testing --> fail if we don't recognize the URL in the request
// 		default:
// 			t.Fatal(fmt.Sprintf("Unexpected call to URL %s", r.URL.EscapedPath()))
// 		}
// 	}))
// 	defer ts.Close()
// 	tsURL, err := url.Parse(ts.URL)
// 	require.NoError(t, err)

// 	config.KongURL.Server = tsURL.Hostname()
// 	config.KongURL.ApplicationPort, _ = strconv.Atoi(tsURL.Port())

// 	// Setup command "addUser w/OAuth2"
// 	command, err := NewCommand(lc, config, []string{
// 		"--token-type", tokenType,
// 		"--user", user,
// 		"--redirect_uris", redirectUris,
// 		"--jwt", jwt,
// 	})

// 	require.NoError(t, err)

// 	// Execute command "addUser w/JWT"
// 	code, err := command.Execute()

// 	// Assert execution return
// 	require.NoError(t, err)
// 	require.Equal(t, interfaces.StatusCodeExitNormal, code)
// }
