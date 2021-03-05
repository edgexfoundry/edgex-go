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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/helper"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func TestNewCommand(t *testing.T) {
	// Arrange
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}

	tests := []struct {
		name        string
		cmdArgs     []string
		expectedErr bool
	}{
		{"Good:setupRegistryACL cmd empty option", []string{}, false},
		{"Bad:setupRegistryACL invalid option", []string{"--invalid=xxx"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := NewCommand(ctx, wg, lc, config, tt.cmdArgs)
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, command)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	type prepareTestFunc func(aclOkResponse bool, configAccessOkResponse bool, t *testing.T) (*config.ConfigurationStruct,
		*httptest.Server)

	// Arrange
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}

	tests := []struct {
		name                   string
		adminDir               string
		prepare                prepareTestFunc
		aclOkResponse          bool
		configAccessOkResponse bool
		expectedErr            bool
	}{
		{"Good:setupRegistryACL with ok response from server", "test1", prepareTestRegistryServer, true, true, false},
		{"Bad:setupRegistryACL with bootstrap ACL API failed response from server", "test2",
			prepareTestRegistryServer, false, false, true},
		{"Bad:setupRegistryACL with non-existing server", "test3",
			func(_ bool, _ bool, _ *testing.T) (*config.ConfigurationStruct, *httptest.Server) {
				return &config.ConfigurationStruct{
					StageGate: config.StageGateInfo{
						Registry: config.RegistryInfo{
							Host: "non-existing",
							Port: 10001,
							ACL:  config.ACLInfo{Protocol: "http"},
						},
					}}, httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			}, false, false, true},
		{"Bad:setupRegistryACL with empty api protocol", "test4",
			func(_ bool, _ bool, _ *testing.T) (*config.ConfigurationStruct, *httptest.Server) {
				return &config.ConfigurationStruct{
					StageGate: config.StageGateInfo{
						Registry: config.RegistryInfo{
							Host: "localhost",
							Port: 10001,
							ACL:  config.ACLInfo{Protocol: ""},
						},
					}}, httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			}, false, false, true},
		{"Bad:setupRegistryACL with timed out on waiting for secret token file", "",
			prepareTestRegistryServer, true, false, true},
		{"Bad:setupRegistryACL with config access API failed response from server", "test5",
			prepareTestRegistryServer, true, false, true},
	}

	for _, tt := range tests {
		test := tt // capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// prepare test
			conf, testServer := test.prepare(test.aclOkResponse, test.configAccessOkResponse, t)
			defer testServer.Close()
			// setup token related configs
			conf.StageGate.Registry.ACL.SecretsAdminTokenPath = filepath.Join(test.adminDir, "secret_token.json")
			conf.StageGate.Registry.ACL.BootstrapTokenPath = filepath.Join(test.adminDir, "bootstrap_token.json")
			conf.StageGate.Registry.ACL.SentinelFilePath = filepath.Join(test.adminDir, "sentinel_test_file")

			setupRegistryACL, err := NewCommand(ctx, wg, lc, conf, []string{})
			require.NoError(t, err)
			require.NotNil(t, setupRegistryACL)
			require.Equal(t, "setupRegistryACL", setupRegistryACL.GetCommandName())

			// create test secret token file
			if test.adminDir != "" {
				err = helper.CreateDirectoryIfNotExists(test.adminDir)
				require.NoError(t, err)
				err = ioutil.WriteFile(conf.StageGate.Registry.ACL.SecretsAdminTokenPath,
					[]byte(secretstoreTokenJsonStub), 0600)
				require.NoError(t, err)
			}

			// to speed up the test timeout
			localcmd := setupRegistryACL.(*cmd)
			localcmd.retryTimeout = 3 * time.Second
			statusCode, err := setupRegistryACL.Execute()
			defer func() {
				if test.adminDir == "" {
					// empty test dir case don't have the directory to clean up
					_ = os.Remove(conf.StageGate.Registry.ACL.BootstrapTokenPath)
				} else {
					_ = os.RemoveAll(test.adminDir)
				}
			}()

			if test.expectedErr {
				require.Error(t, err)
				require.Equal(t, interfaces.StatusCodeExitWithError, statusCode)
			} else {
				require.NoError(t, err)
				require.Equal(t, interfaces.StatusCodeExitNormal, statusCode)
				require.FileExists(t, conf.StageGate.Registry.ACL.BootstrapTokenPath)
				require.FileExists(t, conf.StageGate.Registry.ACL.SecretsAdminTokenPath)
				require.FileExists(t, conf.StageGate.Registry.ACL.SentinelFilePath)
			}
		})
	}
}

func prepareTestRegistryServer(aclOkResponse bool, configAccessOkResponse bool, t *testing.T) (*config.ConfigurationStruct,
	*httptest.Server) {
	registryTestConf := &config.ConfigurationStruct{}

	respCnt := 0
	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case consulGetLeaderAPI:
			require.Equal(t, http.MethodGet, r.Method)
			respCnt++
			w.WriteHeader(http.StatusOK)
			var err error
			if respCnt == 2 {
				_, err = w.Write([]byte("127.0.0.1:12345"))
			} else {
				_, err = w.Write([]byte(""))
			}
			require.NoError(t, err)
		case consulACLBootstrapAPI:
			require.Equal(t, http.MethodPut, r.Method)
			if aclOkResponse {
				w.WriteHeader(http.StatusOK)
				jsonResponse := map[string]interface{}{
					"AccessorID":  "bad060a9-0e2b-47ba-98d5-9d622e2322b5",
					"SecretID":    "7240fdd9-1665-419b-a8c5-5691ca03af7c",
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
			if configAccessOkResponse {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusForbidden)
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

	return registryTestConf, testSrv
}
