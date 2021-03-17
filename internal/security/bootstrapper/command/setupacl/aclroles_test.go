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

func TestCreateRole(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lc := logger.MockLogger{}
	testSecretStoreToken := "test-secretstore-token"
	testSinglePolicy := []Policy{
		{
			ID:   "test-ID",
			Name: "test-name",
		},
	}
	testMultiplePolicies := []Policy{
		{
			ID:   "test-ID1",
			Name: "test-name1",
		},
		{
			ID:   "test-ID2",
			Name: "test-name2",
		},
	}

	testRoleWithNilPolicy := NewRegistryRole("testRoleSingle", ClientType, nil, true)
	testRoleWithEmptyPolicy := NewRegistryRole("testRoleSingle", ClientType, []Policy{}, true)
	testRoleWithSinglePolicy := NewRegistryRole("testRoleSingle", ClientType, testSinglePolicy, true)
	testRoleWithMultiplePolicies := NewRegistryRole("testRoleMultiple", ClientType, testMultiplePolicies, true)
	testEmptyRoleName := NewRegistryRole("", ManagementType, testSinglePolicy, true)

	tests := []struct {
		name                string
		secretstoreToken    string
		registryRole        RegistryRole
		creatRoleOkResponse bool
		expectedErr         bool
	}{
		{"Good:create role with single policy ok", testSecretStoreToken, testRoleWithSinglePolicy, true, false},
		{"Good:create role with multiple policies ok", testSecretStoreToken, testRoleWithMultiplePolicies, true, false},
		{"Good:create role with empty policy ok", testSecretStoreToken, testRoleWithEmptyPolicy, true, false},
		{"Good:create role with nil policy ok", testSecretStoreToken, testRoleWithNilPolicy, true, false},
		{"Bad:create role bad response", testSecretStoreToken, testRoleWithSinglePolicy, false, true},
		{"Bad:empty secretstore token", "", testRoleWithMultiplePolicies, false, true},
		{"Bad:empty role name", testSecretStoreToken, testEmptyRoleName, false, true},
	}

	for _, tt := range tests {
		test := tt // capture as local copy
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// prepare test
			responseOpts := serverOptions{
				createRoleOk: test.creatRoleOkResponse,
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

			err = setupRegistryACL.createRole(test.secretstoreToken, test.registryRole)

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
