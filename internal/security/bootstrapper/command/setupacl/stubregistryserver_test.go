/*******************************************************************************
 * Copyright 2023 Intel Corporation
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
)

type registryTestServer struct {
	serverOptions serverOptions
	server        *httptest.Server
}

type serverOptions struct {
	aclBootstrapOkResponse  bool
	configAccessOkResponse  bool
	enablePersistence       bool
	consulCheckAgentOk      bool
	listTokensOk            bool
	listTokensWithRetriesOk bool
	createTokenOk           bool
	readTokenOk             bool
	setAgentTokenOk         bool
	readPolicyByNameOk      bool
	policyAlreadyExists     bool
	createNewPolicyOk       bool
	createRoleOk            bool
}

func newRegistryTestServer(respOpts serverOptions) *registryTestServer {
	return &registryTestServer{
		serverOptions: respOpts,
	}
}

func (registry *registryTestServer) close() {
	if registry.server != nil {
		registry.server.Close()
	}
}

func (registry *registryTestServer) getRegistryServerConf(t *testing.T) *config.ConfigurationStruct {
	registryTestConf := &config.ConfigurationStruct{}
	testAgentTokenAccessorID := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	testEdgeXPolicyID := "eeeeeeeee-eeee-eeee-eeee-eeeeeeeee"
	tokens := []map[string]interface{}{
		{
			"AccessorID":  "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy",
			"Description": "some other type of agent token",
			"SecretID":    "000000000000000000000000000",
			"Policies": []map[string]interface{}{
				{
					"ID":   "0000",
					"Name": "p1",
				},
				{
					"ID":   "1111",
					"Name": "p2",
				},
			},
		},
		{
			"AccessorID":  "00000000-0000-0000-0000-000000000002",
			"Description": "Anonymous Token",
			"SecretID":    "11111111111111111111111111",
			"Policies": []map[string]interface{}{
				{
					"ID":   "0000",
					"Name": "p1",
				},
				{
					"ID":   "1111",
					"Name": "p2",
				},
			},
		},
		{
			"AccessorID":  "mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm",
			"Description": "Bootstrap Token (Global Management)",
			"SecretID":    "2222222222222222222222222",
			"Policies": []map[string]interface{}{
				{
					"ID":   "0000",
					"Name": "p1",
				},
				{
					"ID":   "1111",
					"Name": "p2",
				},
			},
		},
	}
	respCnt := 0
	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathBase := path.Base(r.URL.Path)
		switch r.URL.EscapedPath() {
		case consulGetLeaderAPI:
			require.Equal(t, http.MethodGet, r.Method)
			respCnt++
			w.WriteHeader(http.StatusOK)
			var err error
			if respCnt >= 2 {
				_, err = w.Write([]byte("127.0.0.1:12345"))
			} else {
				_, err = w.Write([]byte(""))
			}
			require.NoError(t, err)
		case consulACLBootstrapAPI:
			require.Equal(t, http.MethodPut, r.Method)
			if registry.serverOptions.aclBootstrapOkResponse {
				w.WriteHeader(http.StatusOK)
				jsonResponse := map[string]interface{}{
					"AccessorID":  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"SecretID":    "22222222-bbbb-3333-cccc-444444444444",
					"Description": "Bootstrap Token (Global Management)",
					"Policies": []map[string]interface{}{
						{
							"ID":   "00000000-0000-0000-0000-000000000001",
							"Name": "global-management",
						},
					},
					"Local":      false,
					"CreateTime": "2021-03-01T10:34:20.843397-07:00",
				}
				err := json.NewEncoder(w).Encode(jsonResponse)
				require.NoError(t, err)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("The ACL system is currently in legacy mode."))
			}
		case consulConfigAccessVaultAPI:
			require.Equal(t, http.MethodPost, r.Method)
			if registry.serverOptions.configAccessOkResponse {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
		case fmt.Sprintf("/v1/consul/roles/%s", pathBase):
			require.Equal(t, http.MethodPost, r.Method)
			if registry.serverOptions.createRoleOk {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
		case consulCheckAgentAPI:
			require.Equal(t, http.MethodGet, r.Method)
			if registry.serverOptions.consulCheckAgentOk {
				w.WriteHeader(http.StatusOK)
				jsonResponse := map[string]interface{}{
					"DebugConfig": map[string]interface{}{
						"ACLDatacenter":    "dc1",
						"ACLDefaultPolicy": "allow",
						"ACLDisabledTTL":   "2m0s",
						"ACLTokens": map[string]interface{}{
							"EnablePersistence": registry.serverOptions.enablePersistence,
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
			if registry.serverOptions.listTokensOk {
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(tokens)
				require.NoError(t, err)
			} else if registry.serverOptions.listTokensWithRetriesOk {
				if respCnt >= 2 {
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(tokens)
					require.NoError(t, err)
				} else {
					respCnt++
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(consulLegacyACLModeError))
				}
			} else if !registry.serverOptions.listTokensWithRetriesOk {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(consulLegacyACLModeError))
			} else {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("permission denied"))
			}
		case consulCreateTokenAPI:
			require.Equal(t, http.MethodPut, r.Method)
			if registry.serverOptions.createTokenOk {
				w.WriteHeader(http.StatusOK)
				jsonResponse := map[string]interface{}{
					"AccessorID":  testAgentTokenAccessorID,
					"Description": "edgex-core-consul agent token",
					"SecretID":    "12121212121212121212121",
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
			if r.Method == http.MethodGet && registry.serverOptions.readTokenOk {
				w.WriteHeader(http.StatusOK)
				jsonResponse := map[string]interface{}{
					"AccessorID":  testAgentTokenAccessorID,
					"Description": "edgex-core-consul agent token",
					"SecretID":    "888888888888888888888888888888888888",
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
				t.Fatalf("Unexpected method %s to URL %s", r.Method, r.URL.EscapedPath())
			}
		case fmt.Sprintf(consulSetAgentTokenAPI, AgentType):
			require.Equal(t, http.MethodPut, r.Method)
			if registry.serverOptions.setAgentTokenOk {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("agent token set successfully"))
			} else {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("permission denied"))
			}
		case fmt.Sprintf(consulReadPolicyByNameAPI, ""):
			require.Equal(t, http.MethodGet, r.Method)
			// the policy name is empty
			w.WriteHeader(http.StatusBadRequest)
		case fmt.Sprintf(consulReadPolicyByNameAPI, pathBase):
			require.Equal(t, http.MethodGet, r.Method)
			if registry.serverOptions.readPolicyByNameOk && registry.serverOptions.policyAlreadyExists {
				w.WriteHeader(http.StatusOK)
				jsonResponse := map[string]interface{}{
					"ID":          testEdgeXPolicyID,
					"Name":        pathBase,
					"Description": "test edgex policy",
					"Rules":       edgeXPolicyRules,
				}

				err := json.NewEncoder(w).Encode(jsonResponse)
				require.NoError(t, err)
			} else if registry.serverOptions.readPolicyByNameOk {
				// no existing policy
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(aclNotFoundMessage))
			} else {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("permission denied"))
			}
		case consulCreatePolicyAPI:
			require.Equal(t, http.MethodPut, r.Method)
			if registry.serverOptions.createNewPolicyOk {
				w.WriteHeader(http.StatusOK)
				reqBody, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				var policyMap map[string]interface{}
				err = json.Unmarshal(reqBody, &policyMap)
				require.NoError(t, err)
				jsonResponse := map[string]interface{}{
					"ID":          testEdgeXPolicyID,
					"Name":        policyMap["Name"],
					"Description": "test edgex policy",
					"Rules":       policyMap["Rules"],
				}

				err = json.NewEncoder(w).Encode(jsonResponse)
				require.NoError(t, err)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Invalid Policy: A Policy with Name " + edgeXServicePolicyName + " already exists"))
			}
		case consulPolicyListAPI:
			require.Equal(t, http.MethodGet, r.Method)
			w.WriteHeader(http.StatusOK)
			jsonResponse := []map[string]interface{}{
				{
					"Name": "global-management",
				},
				{
					"Name": "node-read",
				},
				{
					"Name": "test-policy-name",
				},
			}
			err := json.NewEncoder(w).Encode(jsonResponse)
			require.NoError(t, err)
		default:
			t.Fatalf("Unexpected call to URL %s", r.URL.EscapedPath())
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
	os.Setenv("SECRETSTORE_PROTOCOL", tsURL.Scheme)
	os.Setenv("SECRETSTORE_HOST", tsURL.Hostname())
	os.Setenv("SECRETSTORE_PORT", tsURL.Port())

	registry.server = testSrv
	return registryTestConf

}
