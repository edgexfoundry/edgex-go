//
// Copyright (c) 2020-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package adduser

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelpBadArg tests unknown arg handler
func TestAddUserBadArg(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	badArgTestcases := [][]string{
		{},                              // missing token type
		{"-badarg"},                     // invalid arg
		{"--user"},                      // missing username
		{"--user", "foo", "--tokenTTL"}, // missing tokenTTL
		{"--user", "foo", "--jwtTTL"},   // missing jwtTTL
	}

	for _, args := range badArgTestcases {
		// Act
		command, err := NewCommand(lc, config, args)

		// Assert
		assert.Error(t, err, "Args: %v", args)
		assert.Nil(t, command)
	}
}

func addUserWithArgs(t *testing.T, args []string) {
	// Arrange
	lc := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/v1/sys/policies/acl/edgex-user-someuser":
			switch r.Method {
			case "PUT":
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("Unexpected call to %s %s", r.Method, r.URL.EscapedPath())
			}
		case "/v1/identity/entity/name/someuser":
			switch r.Method {
			case "POST":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"id":"someguid"}}`))
			default:
				t.Fatalf("Unexpected call to %s %s", r.Method, r.URL.EscapedPath())
			}
		case "/v1/sys/auth":
			switch r.Method {
			case "GET":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"userpass/":{"accessor":"someid"}}}`))
			default:
				t.Fatalf("Unexpected call to %s %s", r.Method, r.URL.EscapedPath())
			}
		case "/v1/auth/userpass/users/someuser":
			switch r.Method {
			case "POST":
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("Unexpected call to %s %s", r.Method, r.URL.EscapedPath())
			}
		case "/v1/identity/entity-alias":
			switch r.Method {
			case "POST":
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("Unexpected call to %s %s", r.Method, r.URL.EscapedPath())
			}
		case "/v1/identity/oidc/role/someuser":
			switch r.Method {
			case "POST":
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("Unexpected call to %s %s", r.Method, r.URL.EscapedPath())
			}
		default:
			t.Fatalf("Unexpected %s call to URL %s", r.Method, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	config := &config.ConfigurationStruct{}
	config.SecretStore.Host = tsURL.Hostname()
	p, _ := strconv.ParseInt(tsURL.Port(), 10, 32)
	config.SecretStore.Port = int(p)
	config.SecretStore.Protocol = "https"
	config.SecretStore.Type = "vault"
	config.SecretStore.TokenFolderPath = "testdata/"
	config.SecretStore.TokenFile = "token.json"

	// Act
	command, err := NewCommand(lc, config, args)
	require.NoError(t, err)

	code, err := command.Execute()

	// Assert
	require.NoError(t, err)
	require.Equal(t, interfaces.StatusCodeExitNormal, code)
}

// TestDelUser tests functionality of adduser command using JWT option
func TestAddUser(t *testing.T) {
	addUserWithArgs(t, []string{
		"--user", "someuser",
	})
}
