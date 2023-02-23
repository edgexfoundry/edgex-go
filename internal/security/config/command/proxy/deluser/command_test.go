//
// Copyright (c) 2020-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package deluser

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
func TestDelUserBadArg(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	badArgTestcases := [][]string{
		{},          // missing token type
		{"-badarg"}, // invalid arg
		{"-user"},   // missing arg
	}

	for _, args := range badArgTestcases {
		// Act
		command, err := NewCommand(lc, config, args)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, command)
	}
}

func delUserWithArgs(t *testing.T, args []string) {
	// Arrange
	lc := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/v1/identity/entity/name/someuser":
			switch r.Method {
			case "GET":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"id":"someguid"}}`))
			case "DELETE":
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("Unexpected call to %s %s", r.Method, r.URL.EscapedPath())
			}
		case "/v1/auth/userpass/users/someuser":
			switch r.Method {
			case "DELETE":
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("Unexpected call to %s %s", r.Method, r.URL.EscapedPath())
			}
		default:
			t.Fatalf("Unexpected call to URL %s", r.URL.EscapedPath())
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
func TestDelUser(t *testing.T) {
	delUserWithArgs(t, []string{
		"--user", "someuser",
	})
}
