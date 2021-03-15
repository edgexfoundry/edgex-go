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

package setupacl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func TestIsACLTokenPersistent(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}
	testBootstrapToken := "test-bootstrap-token"

	tests := []struct {
		name                 string
		bootstrapToken       string
		enablePersist        bool
		checkAgentOkResponse bool
		expectedErr          bool
	}{
		{"Good:persist enabled ok response", testBootstrapToken, true, true, false},
		{"Good:persist disabled ok response", testBootstrapToken, false, true, false},
		{"Bad:persist check bad response", testBootstrapToken, false, false, true},
		{"Bad:empty bootstrap token", "", false, false, true},
	}

	for _, tt := range tests {
		test := tt // capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// prepare test
			responseOpts := responseOptions{
				enablePersistence:  test.enablePersist,
				consulCheckAgentOk: test.checkAgentOkResponse,
			}
			testSrv := newRegistryTestServer(responseOpts)
			conf := testSrv.getRegistryServerConf(t)

			command, err := NewCommand(ctx, wg, lc, conf, []string{})
			require.NoError(t, err)
			require.NotNil(t, command)
			require.Equal(t, "setupRegistryACL", command.GetCommandName())
			setupRegistryACL := command.(*cmd)
			setupRegistryACL.retryTimeout = 3 * time.Second

			persistent, err := setupRegistryACL.isACLTokenPersistent(test.bootstrapToken)

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testSrv.responseOptions.enablePersistence, persistent)
			}
		})
	}
}

func TestCreateAgentToken(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}
	testBootstrapToken := BootStrapACLTokenInfo{
		SecretID: "test-bootstrap-token",
		Policies: []Policy{
			{
				ID:   "00000000-0000-0000-0000-000000000001",
				Name: "global-management",
			},
		},
	}

	tests := []struct {
		name                        string
		bootstrapToken              BootStrapACLTokenInfo
		listTokensOkResponse        bool
		listTokensRetriesOkResponse bool
		createTokenOkResponse       bool
		readTokenOkResponse         bool
		expectedErr                 bool
	}{
		{"Good:agent token ok response", testBootstrapToken, true, true, true, true, false},
		{"Bad:list tokens bad response", testBootstrapToken, false, false, true, true, true},
		{"Bad:create token bad response", testBootstrapToken, true, true, false, true, true},
		{"Bad:read token bad response", testBootstrapToken, true, true, true, false, true},
		{"Bad:empty bootstrap token", BootStrapACLTokenInfo{}, false, false, false, true, true},
	}

	for _, tt := range tests {
		test := tt // capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// prepare test
			responseOpts := responseOptions{
				listTokensOk:  test.listTokensOkResponse,
				createTokenOk: test.createTokenOkResponse,
				readTokenOk:   test.readTokenOkResponse,
			}
			testSrv := newRegistryTestServer(responseOpts)
			conf := testSrv.getRegistryServerConf(t)

			command, err := NewCommand(ctx, wg, lc, conf, []string{})
			require.NoError(t, err)
			require.NotNil(t, command)
			require.Equal(t, "setupRegistryACL", command.GetCommandName())
			setupRegistryACL := command.(*cmd)
			setupRegistryACL.retryTimeout = 3 * time.Second

			// first time we don't have the agent token yet
			agentToken1, err := setupRegistryACL.createAgentToken(test.bootstrapToken)

			// readToken only being executed only on the second time, so the first time should not expected an error
			if test.expectedErr && test.readTokenOkResponse {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, agentToken1)
			}

			// re-run create to test the existing token route
			agentToken2, err := setupRegistryACL.createAgentToken(test.bootstrapToken)
			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, agentToken2)
				require.Equal(t, agentToken1, agentToken2)
			}
		})
	}
}

func TestSetAgentTokenToAgent(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}
	testBootstrapToken := BootStrapACLTokenInfo{
		SecretID: "test-bootstrap-token",
		Policies: []Policy{
			{
				ID:   "00000000-0000-0000-0000-000000000001",
				Name: "global-management",
			},
		},
	}
	testAgentToken := "test-agent-token"

	tests := []struct {
		name                    string
		bootstrapToken          BootStrapACLTokenInfo
		agentToken              string
		setAgentTokenOkResponse bool
		expectedErr             bool
	}{
		{"Good:set agent token ok response", testBootstrapToken, testAgentToken, true, false},
		{"Bad:set agent token bad response", testBootstrapToken, testAgentToken, false, true},
		{"Bad:empty bootstrap token", BootStrapACLTokenInfo{}, testAgentToken, false, true},
		{"Bad:empty agent token", testBootstrapToken, "", false, true},
	}

	for _, tt := range tests {
		test := tt // capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// prepare test
			responseOpts := responseOptions{
				setAgentTokenOk: test.setAgentTokenOkResponse,
			}
			testSrv := newRegistryTestServer(responseOpts)
			conf := testSrv.getRegistryServerConf(t)

			command, err := NewCommand(ctx, wg, lc, conf, []string{})
			require.NoError(t, err)
			require.NotNil(t, command)
			require.Equal(t, "setupRegistryACL", command.GetCommandName())
			setupRegistryACL := command.(*cmd)
			setupRegistryACL.retryTimeout = 3 * time.Second

			err = setupRegistryACL.setAgentToken(test.bootstrapToken, test.agentToken, AgentType)

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type registryTestServer struct {
	config          *config.ConfigurationStruct
	responseOptions responseOptions
}

type responseOptions struct {
	enablePersistence       bool
	consulCheckAgentOk      bool
	listTokensOk            bool
	listTokensWithRetriesOk bool
	createTokenOk           bool
	readTokenOk             bool
	setAgentTokenOk         bool
}

func newRegistryTestServer(respOpts responseOptions) *registryTestServer {
	return &registryTestServer{
		config:          &config.ConfigurationStruct{},
		responseOptions: respOpts,
	}
}

func (registry *registryTestServer) getRegistryServerConf(t *testing.T) *config.ConfigurationStruct {
	registryTestConf := registry.config
	testAgentTokenAccessorID := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	tokens := []map[string]interface{}{
		{
			"AccessorID":  "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy",
			"Description": "some other type of agent token",
		},
		{
			"AccessorID":  "00000000-0000-0000-0000-000000000002",
			"Description": "Anonymous Token",
		},
		{
			"AccessorID":  "mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm",
			"Description": "Bootstrap Token (Global Management)",
		},
	}
	respCnt := 0
	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case consulCheckAgentAPI:
			require.Equal(t, http.MethodGet, r.Method)
			if registry.responseOptions.consulCheckAgentOk {
				w.WriteHeader(http.StatusOK)
				jsonResponse := map[string]interface{}{
					"DebugConfig": map[string]interface{}{
						"ACLDatacenter":    "dc1",
						"ACLDefaultPolicy": "allow",
						"ACLDisabledTTL":   "2m0s",
						"ACLTokens": map[string]interface{}{
							"EnablePersistence": registry.responseOptions.enablePersistence,
						},
						"ACLsEnabled": true,
					},
				}
				err := json.NewEncoder(w).Encode(jsonResponse)
				require.NoError(t, err)
			} else {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("permission denied"))
			}
		case consulListTokensAPI:
			require.Equal(t, http.MethodGet, r.Method)
			if registry.responseOptions.listTokensOk {
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(tokens)
				require.NoError(t, err)
			} else if registry.responseOptions.listTokensWithRetriesOk {
				if respCnt >= 2 {
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(tokens)
					require.NoError(t, err)
				} else {
					respCnt++
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(consulLegacyACLModeError))
				}
			} else if !registry.responseOptions.listTokensWithRetriesOk {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(consulLegacyACLModeError))
			} else {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("permission denied"))
			}
		case consulCreateTokenAPI:
			require.Equal(t, http.MethodPut, r.Method)
			if registry.responseOptions.createTokenOk {
				w.WriteHeader(http.StatusOK)
				jsonResponse := map[string]interface{}{
					"AccessorID":  testAgentTokenAccessorID,
					"Description": "edgex-core-consul agent token",
					"Policies": []map[string]interface{}{
						{
							"ID":   "00000000-0000-0000-0000-000000000001",
							"Name": "global-management",
						},
						{
							"ID":   "rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr",
							"Name": "node-read policy",
						},
					},
					"Local":       true,
					"CreateTime":  "2021-03-10T12:25:06.123456-07:00",
					"Hash":        "UuiRkOQPRCvoRZHRtUxxbrmwZ5crYrOdZ0Z1FTFbTbA=",
					"CreateIndex": 59,
					"ModifyIndex": 59,
				}

				err := json.NewEncoder(w).Encode(jsonResponse)
				require.NoError(t, err)
				tokens = append(tokens, jsonResponse)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("cannot create token"))
			}
		case fmt.Sprintf(consulTokenRUDAPI, testAgentTokenAccessorID):
			if r.Method == http.MethodGet && registry.responseOptions.readTokenOk {
				w.WriteHeader(http.StatusOK)
				jsonResponse := map[string]interface{}{
					"AccessorID":  testAgentTokenAccessorID,
					"Description": "edgex-core-consul agent token",
					"Policies": []map[string]interface{}{
						{
							"ID":   "00000000-0000-0000-0000-000000000001",
							"Name": "global-management",
						},
						{
							"ID":   "rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr",
							"Name": "node-read policy",
						},
					},
					"Local":       false,
					"CreateTime":  "2021-03-10T12:25:06.123456-07:00",
					"Hash":        "UuiRkOQPRCvoRZHRtUxxbrmwZ5crYrOdZ0Z1FTFbTbA=",
					"CreateIndex": 59,
					"ModifyIndex": 59,
				}

				err := json.NewEncoder(w).Encode(jsonResponse)
				require.NoError(t, err)
			} else if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("permission denied"))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				t.Fatal(fmt.Sprintf("Unexpected method %s to URL %s", r.Method, r.URL.EscapedPath()))
			}
		case fmt.Sprintf(consulSetAgentTokenAPI, AgentType):
			require.Equal(t, http.MethodPut, r.Method)
			if registry.responseOptions.setAgentTokenOk {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("agent token set successfully"))
			} else {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("permission denied"))
			}
		default:
			t.Fatal(fmt.Sprintf("Unexpected call to URL %s", r.URL.EscapedPath()))
		}
	}))
	tsURL, err := url.Parse(testSrv.URL)
	require.NoError(t, err)
	portNum, _ := strconv.Atoi(tsURL.Port())
	registryTestConf.StageGate.Registry.ACL.Protocol = tsURL.Scheme
	registryTestConf.StageGate.Registry.Host = tsURL.Hostname()
	registryTestConf.StageGate.Registry.Port = portNum
	registryTestConf.StageGate.WaitFor.Timeout = "1m"
	registryTestConf.StageGate.WaitFor.RetryInterval = "1s"
	// for the sake of simplicity, we use the same test server as the secret store server
	registryTestConf.SecretStore.Protocol = tsURL.Scheme
	registryTestConf.SecretStore.Host = tsURL.Hostname()
	registryTestConf.SecretStore.Port = portNum
	return registryTestConf
}
