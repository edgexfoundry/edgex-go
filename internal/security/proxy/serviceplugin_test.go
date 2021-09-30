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

package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/require"
)

const (
	testServiceName = "testService"
)

func TestAddConsulHeaderTo(t *testing.T) {
	mockLogger := logger.MockLogger{}
	kongService := &models.KongService{
		Name: testServiceName,
	}
	// setup access token file
	// nolint:gosec
	tokenData := `{
		"SecretID":"test-access-token",
		"Policies": [
			{
				"ID": "test-policy-id",
				"Name":"test-policy-name"
			}
		]
	}
	`

	tests := []struct {
		name                         string
		getPluginIDOkResponse        bool
		emptyPluginIDResponse        bool
		deletePluginOkResponse       bool
		accessTokenFileCreate        bool
		createHeaderPluginOkResponse bool
		expectedErr                  bool
	}{
		{"Good:add consul header with ok empty pluginID response", true, true, false, true, true, false},
		{"Good:add consul header with ok some pluginID and delete response", true, false, true, true, true, false},
		{"Bad:add consul header with ok some pluginID but bad delete response", true, false, false, false, false, true},
		{"Bad:add consul header with missing access token file", true, false, true, false, false, true},
		{"Bad:add consul header with bad get pluginID response", false, true, false, true, false, true},
		{"Bad:add consul header with bad create pluginID response", true, false, true, true, false, true},
		{"Bad:add consul header with bad create pluginID response- empty pluginID", true, true, true, true, false, true},
	}

	for _, tt := range tests {
		test := tt // capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// prepare test
			if test.accessTokenFileCreate {
				// use test name as the config name so that it is unique
				err := os.WriteFile(test.name, []byte(tokenData), 0600)
				require.NoError(t, err)
			}

			testSrvOptions := serverOptions{
				emptyServicePlugin:                  test.emptyPluginIDResponse,
				getServicePluginIDResponseOk:        test.getPluginIDOkResponse,
				deleteServicePluginResponseOk:       test.deletePluginOkResponse,
				createServicePluginHeaderResponseOk: test.createHeaderPluginOkResponse,
			}
			testSrv := newKongTestServer(testSrvOptions)
			defer testSrv.close()

			testConfig := testSrv.getStubKongServerConfig(t)
			testConfig.AccessTokenFile = test.name
			service := NewService(NewRequestor(true, 10, "", mockLogger), mockLogger, testConfig)
			err := service.addConsulTokenHeaderTo(kongService)

			defer func() {
				_ = os.Remove(test.name)
			}()

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type kongTestServer struct {
	serverOptions serverOptions
	server        *httptest.Server
}

type serverOptions struct {
	emptyServicePlugin                  bool
	getServicePluginIDResponseOk        bool
	deleteServicePluginResponseOk       bool
	createServicePluginHeaderResponseOk bool
}

func newKongTestServer(srvOptions serverOptions) kongTestServer {
	return kongTestServer{
		serverOptions: srvOptions,
	}
}

func (s kongTestServer) close() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s kongTestServer) getStubKongServerConfig(t *testing.T) *config.ConfigurationStruct {
	kongTestConfig := &config.ConfigurationStruct{}
	testServicePluginID := "test-plugin-id"
	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case path.Join("/admin", ServicesPath, testServiceName, PluginsPath):
			switch r.Method {
			case http.MethodGet:
				if s.serverOptions.getServicePluginIDResponseOk && s.serverOptions.emptyServicePlugin {
					w.WriteHeader(http.StatusOK)
					jsonResponse := map[string]interface{}{
						"data": []map[string]interface{}{},
					}
					err := json.NewEncoder(w).Encode(jsonResponse)
					require.NoError(t, err)
				} else if s.serverOptions.getServicePluginIDResponseOk {
					w.WriteHeader(http.StatusOK)
					jsonResponse := map[string]interface{}{
						"data": []map[string]interface{}{
							{
								"id":   testServicePluginID,
								"name": requestTransformerPlugin,
								"service": map[string]interface{}{
									"id": "test-service-id",
								},
								"config": map[string]interface{}{
									"add": map[string]interface{}{
										"querystring": []interface{}{},
										"headers": []interface{}{
											"X-Consul-Token:test-access-token",
										},
										"body": []interface{}{},
									},
								},
							},
						},
					}
					err := json.NewEncoder(w).Encode(jsonResponse)
					require.NoError(t, err)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("failed to get data for service " + testServiceName))
				}
			case http.MethodPost:
				if s.serverOptions.createServicePluginHeaderResponseOk {
					w.WriteHeader(http.StatusCreated)
				} else {
					w.WriteHeader(http.StatusConflict)
					_, _ = w.Write([]byte(`{name": "unique constraint violation}`))
				}
			default:
				t.Errorf("Unsupported method %s to URL %s", r.Method, r.URL.EscapedPath())
			}
		case path.Join("/admin", ServicesPath, testServiceName, PluginsPath, testServicePluginID):
			require.Equal(t, http.MethodDelete, r.Method)
			if s.serverOptions.deleteServicePluginResponseOk {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("could not find plugin " + testServicePluginID))
			}
		default:
			t.Errorf("Unexpected call to URL %s", r.URL.EscapedPath())
		}
	}))
	tsURL, err := url.Parse(testSrv.URL)
	require.NoError(t, err)
	portNum, _ := strconv.Atoi(tsURL.Port())
	kongTestConfig.KongURL.Server = tsURL.Hostname()
	kongTestConfig.KongURL.ApplicationPort = portNum
	s.server = testSrv

	return kongTestConfig
}
