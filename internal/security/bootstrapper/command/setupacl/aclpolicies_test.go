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
)

func TestGetOrCreatePolicy(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}
	testBootstrapToken := "test-bootstrap-token"
	testPolicyName := "test-policy-name"

	tests := []struct {
		name                    string
		bootstrapToken          string
		policyName              string
		getPolicyOkResponse     bool
		policyNameAlreadyExists bool
		createPolicyOkResponse  bool
		expectedErr             bool
	}{
		{"Good:create policy ok with non-existing name yet", testBootstrapToken, testPolicyName, true, false, true, false},
		{"Good:get or create policy ok with pre-existing name", testBootstrapToken, testPolicyName, true, true, true, false},
		{"Bad:get policy bad response", testBootstrapToken, testPolicyName, false, false, false, true},
		{"Bad:create policy bad response", testBootstrapToken, testPolicyName, true, false, false, true},
		{"Bad:empty bootstrap token", "", testPolicyName, false, false, false, true},
		{"Bad:empty policy name", testBootstrapToken, "", false, false, false, true},
	}

	for _, tt := range tests {
		test := tt // capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// prepare test
			responseOpts := serverOptions{
				readPolicyByNameOk:  test.getPolicyOkResponse,
				policyAlreadyExists: test.policyNameAlreadyExists,
				createNewPolicyOk:   test.createPolicyOkResponse,
			}
			testSrv := newRegistryTestServer(responseOpts)
			conf := testSrv.getRegistryServerConf(t)
			defer testSrv.close()

			command, err := NewCommand(ctx, wg, lc, conf, []string{})
			require.NoError(t, err)
			require.NotNil(t, command)
			require.Equal(t, "setupRegistryACL", command.GetCommandName())
			setupRegistryACL := command.(*cmd)
			setupRegistryACL.retryTimeout = 2 * time.Second

			policyActual, err := setupRegistryACL.getOrCreateRegistryPolicy(test.bootstrapToken, test.policyName, edgeXPolicyRules)

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, policyActual)
				require.Equal(t, test.policyName, policyActual.Name)
			}
		})
	}
}
