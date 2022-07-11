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
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type prepareTestFunc func(serverOptions serverOptions, t *testing.T) (*config.ConfigurationStruct,
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
			func(_ serverOptions, _ *testing.T) (*config.ConfigurationStruct, *httptest.Server) {
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
			func(_ serverOptions, _ *testing.T) (*config.ConfigurationStruct, *httptest.Server) {
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
			testSrvOptions := serverOptions{
				aclBootstrapOkResponse:  test.aclOkResponse,
				configAccessOkResponse:  test.configAccessOkResponse,
				enablePersistence:       true,
				consulCheckAgentOk:      true,
				listTokensOk:            true,
				listTokensWithRetriesOk: true,
				createTokenOk:           true,
				readTokenOk:             true,
				setAgentTokenOk:         true,
				readPolicyByNameOk:      true,
				policyAlreadyExists:     true,
				createNewPolicyOk:       true,
				createRoleOk:            true,
			}
			conf, testServer := test.prepare(testSrvOptions, t)
			defer testServer.Close()
			// setup token related configs
			conf.StageGate.Registry.ACL.SecretsAdminTokenPath = filepath.Join(test.adminDir, "secret_token.json")
			conf.StageGate.Registry.ACL.BootstrapTokenPath = filepath.Join(test.adminDir, "bootstrap_token.json")
			conf.StageGate.Registry.ACL.SentinelFilePath = filepath.Join(test.adminDir, "sentinel_test_file")
			conf.StageGate.Registry.ACL.ManagementTokenPath = filepath.Join(test.adminDir, "mgmt_token.json")

			setupRegistryACL, err := NewCommand(ctx, wg, lc, conf, []string{})
			require.NoError(t, err)
			require.NotNil(t, setupRegistryACL)
			require.Equal(t, "setupRegistryACL", setupRegistryACL.GetCommandName())

			// create test secret token file
			if test.adminDir != "" {
				err = helper.CreateDirectoryIfNotExists(test.adminDir)
				require.NoError(t, err)
				err = os.WriteFile(conf.StageGate.Registry.ACL.SecretsAdminTokenPath,
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
					for roleName := range conf.StageGate.Registry.ACL.Roles {
						curDir, _ := os.Getwd()
						_ = os.Remove(filepath.Join(curDir, strings.ToLower(roleName)))
					}
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
			testSrvOptions := serverOptions{
				aclBootstrapOkResponse:  true,
				configAccessOkResponse:  true,
				enablePersistence:       true,
				consulCheckAgentOk:      true,
				listTokensOk:            true,
				listTokensWithRetriesOk: true,
				createTokenOk:           true,
				readTokenOk:             true,
				setAgentTokenOk:         true,
				readPolicyByNameOk:      true,
				policyAlreadyExists:     true,
				createNewPolicyOk:       true,
				createRoleOk:            true,
			}
			conf, testServer := test.prepare(testSrvOptions, t)
			defer testServer.Close()
			// setup token related configs
			conf.StageGate.Registry.ACL.SecretsAdminTokenPath = filepath.Join(test.adminDir, "secret_token.json")
			conf.StageGate.Registry.ACL.BootstrapTokenPath = filepath.Join(test.adminDir, "bootstrap_token.json")
			conf.StageGate.Registry.ACL.SentinelFilePath = filepath.Join(test.adminDir, "sentinel_test_file")
			conf.StageGate.Registry.ACL.ManagementTokenPath = filepath.Join(test.adminDir, "mgmt_token.json")

			setupRegistryACL, err := NewCommand(ctx, wg, lc, conf, []string{})
			require.NoError(t, err)
			require.NotNil(t, setupRegistryACL)
			require.Equal(t, "setupRegistryACL", setupRegistryACL.GetCommandName())

			// create test secret token file
			if test.adminDir != "" {
				err = helper.CreateDirectoryIfNotExists(test.adminDir)
				require.NoError(t, err)
				err = os.WriteFile(conf.StageGate.Registry.ACL.SecretsAdminTokenPath,
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
					for roleName := range conf.StageGate.Registry.ACL.Roles {
						curDir, _ := os.Getwd()
						_ = os.Remove(filepath.Join(curDir, strings.ToLower(roleName)))
					}
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

func TestGetUniqueRoleNames(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}

	testConfigOneRole := make(map[string]config.ACLRoleInfo)
	testConfigOneRole["testRole1"] = config.ACLRoleInfo{Description: "role1"}

	// random number of roles between 2 and 4
	numOfConfigRoles := rand.Intn(3)*3 + 2 // nolint:gosec
	testConfigMultipleRoles := make(map[string]config.ACLRoleInfo)
	for i := 0; i < numOfConfigRoles; i++ {
		roleName := "testRole" + strconv.Itoa(i+1)
		testConfigMultipleRoles[roleName] = config.ACLRoleInfo{Description: "role for " + roleName}
	}

	emptyAddRoleEnv := ""

	tests := []struct {
		name              string
		configRoles       map[string]config.ACLRoleInfo
		addRolesFromEnv   string
		expectedNumRoles  int
		spotTestRoleNames []string
		expectedError     bool
	}{
		{"Ok:getUniqueRoles with 1 config role only", testConfigOneRole, emptyAddRoleEnv, 1, []string{"testrole1"}, false},
		{"Ok:getUniqueRoles with multiple config roles", testConfigMultipleRoles, emptyAddRoleEnv, numOfConfigRoles,
			[]string{"testrole1", "testrole2"}, false},
		{"Ok:getUniqueRoles with 1 config role and 1 added role from env", testConfigOneRole, "envrole-1", 2,
			[]string{"testrole1", "envrole-1"}, false},
		{"Ok:getUniqueRoles with 1 config role and multiple added roles from env", testConfigOneRole,
			"envrole-1, envrole-2, envrole-3", 4, []string{"testrole1", "envrole-1", "envrole-3"}, false},
		{"Ok:getUniqueRoles with multiple config roles and 1 added role from env", testConfigMultipleRoles, "envrole-1",
			numOfConfigRoles + 1, []string{"testrole1", "testrole2", "envrole-1"}, false},
		{"Ok:getUniqueRoles with multiple config roles and multiple added roles from env", testConfigMultipleRoles,
			"envrole-1, envrole-2, envrole-3", numOfConfigRoles + 3, []string{"testrole1", "testrole2", "envrole-1", "envrole-2"}, false},
		{"Ok:getUniqueRoles with duplicate roles", testConfigMultipleRoles,
			"envrole-1, testrole2, envrole-1", numOfConfigRoles + 1, []string{"testrole1", "testrole2", "envrole-1"}, false},
		{"Ok:getUniqueRoles with empty role name", testConfigMultipleRoles,
			" , envrole-1,", numOfConfigRoles + 1, []string{"testrole1", "testrole2", "envrole-1"}, false},
		{"Bad:getUniqueRoles with invalid role name: space", testConfigMultipleRoles,
			"a role for,", numOfConfigRoles, []string{"testrole1", "testrole2"}, true},
		{"Bad:getUniqueRoles with invalid role name: special characters", testConfigMultipleRoles,
			"$Role , ~arole!,^@#%*&=+|/;:.", numOfConfigRoles, []string{"testrole1", "testrole2"}, true},
		{"Bad:getUniqueRoles with invalid role name: invalid []<>(){}", testConfigMultipleRoles,
			"[],< >,(), {}", numOfConfigRoles, []string{"testrole1", "testrole2"}, true},
		{"Bad:getUniqueRoles with empty config role", make(map[string]config.ACLRoleInfo), emptyAddRoleEnv, 0, nil, true},
		{"Bad:getUniqueRoles with nil config role", nil, emptyAddRoleEnv, 0, nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// prepare test
			err := os.Setenv(addRegistryRolesEnvKey, test.addRolesFromEnv)
			require.NoError(t, err)
			conf := &config.ConfigurationStruct{}
			conf.StageGate.Registry.ACL.Roles = test.configRoles

			setupRegistryACL, err := NewCommand(ctx, wg, lc, conf, []string{})
			require.NoError(t, err)
			require.NotNil(t, setupRegistryACL)
			require.Equal(t, "setupRegistryACL", setupRegistryACL.GetCommandName())

			localcmd := setupRegistryACL.(*cmd)
			actualRoleNames, err := localcmd.getUniqueRoleNames()

			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedNumRoles, len(actualRoleNames))
				for _, checkRoleName := range test.spotTestRoleNames {
					if _, exists := actualRoleNames[checkRoleName]; !exists {
						require.Fail(t, fmt.Sprintf("missing the expected role name: %s", checkRoleName))
					}
				}
			}
		})
	}
}

func prepareTestRegistryServer(testSrvOptions serverOptions, t *testing.T) (*config.ConfigurationStruct,
	*httptest.Server) {
	testSrv := newRegistryTestServer(testSrvOptions)
	conf := testSrv.getRegistryServerConf(t)
	testRoles := make(map[string]config.ACLRoleInfo)
	testRoles["Role1"] = config.ACLRoleInfo{
		Description: "test for role 1",
	}
	testRoles["Role2"] = config.ACLRoleInfo{
		Description: "test for role 2",
	}
	conf.StageGate.Registry.ACL.Roles = testRoles
	return conf, testSrv.server
}
