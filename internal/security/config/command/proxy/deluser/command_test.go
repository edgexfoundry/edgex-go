//
// Copyright (c) 2020 Intel Corporation
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
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

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
	config := &config.ConfigurationStruct{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/admin/consumers/someuser":
			require.Equal(t, "DELETE", r.Method)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("Unexpected call to URL %s", r.URL.EscapedPath())
		}
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

// TestDelUser tests functionality of adduser command using JWT option
func TestDelUser(t *testing.T) {
	delUserWithArgs(t, []string{
		"--user", "someuser",
		"--jwt", "someRandomJWT",
	})
}
