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

type prepareTestFunc func(aclOkResponse bool, configAccessOkResponse bool, t *testing.T) (*config.ConfigurationStruct,
	*httptest.Server)

func TestExecute(t *testing.T) {
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

func TestMultipleExecuteCalls(t *testing.T) {
	// this test is to simulate the restarting of consul agent multiple times
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}

	expectedBootstrapTokenID := "22222222-bbbb-3333-cccc-444444444444"

	tests := []struct {
		name       string
		adminDir   string
		prepare    prepareTestFunc
		numOfTimes int
	}{
		{"Good:setupRegistryACL with calling Execute() 2 times", "test1", prepareTestRegistryServer, 2},
		{"Good:setupRegistryACL with calling Execute() 3 times", "test2", prepareTestRegistryServer, 3},
	}

	for _, tt := range tests {
		test := tt // capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// prepare test
			conf, testServer := test.prepare(true, true, t)
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
			localcmd.retryTimeout = 2 * time.Second
			statusCode, err := setupRegistryACL.Execute()

			defer func() {
				if test.adminDir == "" {
					// empty test dir case don't have the directory to clean up
					_ = os.Remove(conf.StageGate.Registry.ACL.BootstrapTokenPath)
				} else {
					_ = os.RemoveAll(test.adminDir)
				}
			}()

			require.NoError(t, err)
			require.Equal(t, interfaces.StatusCodeExitNormal, statusCode)
			require.FileExists(t, conf.StageGate.Registry.ACL.BootstrapTokenPath)
			require.FileExists(t, conf.StageGate.Registry.ACL.SecretsAdminTokenPath)
			require.FileExists(t, conf.StageGate.Registry.ACL.SentinelFilePath)

			for i := 1; i < test.numOfTimes; i++ {
				statusCode, err = setupRegistryACL.Execute()

				require.NoError(t, err)
				require.Equal(t, interfaces.StatusCodeExitNormal, statusCode)
				require.FileExists(t, conf.StageGate.Registry.ACL.BootstrapTokenPath)
				require.FileExists(t, conf.StageGate.Registry.ACL.SecretsAdminTokenPath)
				require.FileExists(t, conf.StageGate.Registry.ACL.SentinelFilePath)

				bootstrappedToken, err := localcmd.reconstructBootstrapACLToken()
				require.NoError(t, err)
				require.Equal(t, expectedBootstrapTokenID, bootstrappedToken.SecretID)
			}
		})
	}
}

func prepareTestRegistryServer(aclOkResponse bool, configAccessOkResponse bool, t *testing.T) (*config.ConfigurationStruct,
	*httptest.Server) {
	registryTestConf := &config.ConfigurationStruct{}

	testAgentTokenAccessorID := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	respCnt := 0
	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			if aclOkResponse {
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
			if configAccessOkResponse {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
		case consulCheckAgentAPI:
			require.Equal(t, http.MethodGet, r.Method)
			w.WriteHeader(http.StatusOK)
			jsonResponse := map[string]interface{}{
				"DebugConfig": map[string]interface{}{
					"ACLDatacenter":    "dc1",
					"ACLDefaultPolicy": "allow",
					"ACLDisabledTTL":   "2m0s",
					"ACLTokens": map[string]interface{}{
						"EnablePersistence": true,
					},
					"ACLsEnabled": true,
				},
			}
			err := json.NewEncoder(w).Encode(jsonResponse)
			require.NoError(t, err)
		case consulListTokensAPI:
			require.Equal(t, http.MethodGet, r.Method)
			w.WriteHeader(http.StatusOK)
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
			err := json.NewEncoder(w).Encode(tokens)
			require.NoError(t, err)
		case consulCreateTokenAPI:
			require.Equal(t, http.MethodPut, r.Method)
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
		case fmt.Sprintf(consulTokenRUDAPI, testAgentTokenAccessorID):
			if r.Method == http.MethodGet {
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
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				t.Fatal(fmt.Sprintf("Unexpected method %s to URL %s", r.Method, r.URL.EscapedPath()))
			}
		case fmt.Sprintf(consulSetAgentTokenAPI, AgentType):
			require.Equal(t, http.MethodPut, r.Method)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("agent token set successfully"))
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
