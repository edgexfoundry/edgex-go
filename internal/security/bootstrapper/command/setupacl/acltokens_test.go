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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
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
			responseOpts := serverOptions{
				enablePersistence:  test.enablePersist,
				consulCheckAgentOk: test.checkAgentOkResponse,
			}
			testSrv := newRegistryTestServer(responseOpts)
			conf := testSrv.getRegistryServerConf(t)
			defer testSrv.close()

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
				require.Equal(t, testSrv.serverOptions.enablePersistence, persistent)
			}
		})
	}
}

func TestCreateAgentToken(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}
	testBootstrapToken := types.BootStrapACLTokenInfo{
		SecretID: "test-bootstrap-token",
		Policies: []types.Policy{
			{
				ID:   "00000000-0000-0000-0000-000000000001",
				Name: "global-management",
			},
		},
	}

	tests := []struct {
		name                        string
		bootstrapToken              types.BootStrapACLTokenInfo
		listTokensOkResponse        bool
		listTokensRetriesOkResponse bool
		createTokenOkResponse       bool
		policyAlreadyExists         bool
		expectedErr                 bool
	}{
		{"Good:agent token ok response 1st time", testBootstrapToken, true, true, true, false, false},
		{"Good:agent token ok response 2nd time or later", testBootstrapToken, true, true, true, true, false},
		{"Bad:list tokens bad response", testBootstrapToken, false, false, true, false, true},
		{"Bad:create token bad response", testBootstrapToken, true, true, false, false, true},
		{"Bad:empty bootstrap token", types.BootStrapACLTokenInfo{}, false, false, false, false, true},
	}

	for _, tt := range tests {
		test := tt // capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// prepare test
			responseOpts := serverOptions{
				listTokensOk:        test.listTokensOkResponse,
				createTokenOk:       test.createTokenOkResponse,
				policyAlreadyExists: test.policyAlreadyExists,
				readPolicyByNameOk:  true,
				createNewPolicyOk:   true,
			}
			testSrv := newRegistryTestServer(responseOpts)
			conf := testSrv.getRegistryServerConf(t)
			defer testSrv.close()

			command, err := NewCommand(ctx, wg, lc, conf, []string{})
			require.NoError(t, err)
			require.NotNil(t, command)
			require.Equal(t, "setupRegistryACL", command.GetCommandName())
			setupRegistryACL := command.(*cmd)
			setupRegistryACL.retryTimeout = 3 * time.Second

			// first time we don't have the agent token yet
			agentToken1, err := setupRegistryACL.createAgentToken(test.bootstrapToken)

			if test.expectedErr {
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
	testBootstrapToken := types.BootStrapACLTokenInfo{
		SecretID: "test-bootstrap-token",
		Policies: []types.Policy{
			{
				ID:   "00000000-0000-0000-0000-000000000001",
				Name: "global-management",
			},
		},
	}
	testAgentToken := "test-agent-token"

	tests := []struct {
		name                    string
		bootstrapToken          types.BootStrapACLTokenInfo
		agentToken              string
		setAgentTokenOkResponse bool
		expectedErr             bool
	}{
		{"Good:set agent token ok response", testBootstrapToken, testAgentToken, true, false},
		{"Bad:set agent token bad response", testBootstrapToken, testAgentToken, false, true},
		{"Bad:empty bootstrap token", types.BootStrapACLTokenInfo{}, testAgentToken, false, true},
		{"Bad:empty agent token", testBootstrapToken, "", false, true},
	}

	for _, tt := range tests {
		test := tt // capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// prepare test
			responseOpts := serverOptions{
				setAgentTokenOk: test.setAgentTokenOkResponse,
			}
			testSrv := newRegistryTestServer(responseOpts)
			conf := testSrv.getRegistryServerConf(t)
			defer testSrv.close()

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
