//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//

package secretstore

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/v2/secrets/mocks"
)

func TestCreateTokenIssuingToken(t *testing.T) {
	// Arrange
	logging := logger.MockLogger{}
	secretClient := &mocks.SecretStoreClient{}
	tm := NewTokenMaintenance(logging, secretClient)

	expectedToken := "priv-token"
	secretClient.On("InstallPolicy", "root-token",
		TokenCreatorPolicyName, TokenCreatorPolicy).
		Return(nil)

	createTokenResult := make(map[string]interface{})
	createTokenResult["auth"] = make(map[string]interface{})
	createTokenResult["auth"].(map[string]interface{})["client_token"] = expectedToken
	secretClient.On("CreateToken", "root-token", mock.Anything).
		Return(createTokenResult, nil)

	secretClient.On("RevokeToken", "priv-token").
		Return(nil)

	// Act
	token, revoke, err := tm.CreateTokenIssuingToken("root-token")
	require.NoError(t, err)
	revoke()

	// Assert
	assert.Equal(t, expectedToken, token["auth"].(map[string]interface{})["client_token"])
	assert.Nil(t, err)
	secretClient.AssertExpectations(t)
}

func TestRevokeNonRootTokens(t *testing.T) {
	// Arrange
	logging := logger.MockLogger{}
	secretClient := &mocks.SecretStoreClient{}
	tm := NewTokenMaintenance(logging, secretClient)

	secretClient.On("ListTokenAccessors", "priv-token").
		Return([]string{"rootaccessor", "nonrootaccessor", "priv-token-accessor"}, nil)
	secretClient.On("LookupToken", "priv-token").
		Return(types.TokenMetadata{Accessor: "priv-token-accessor"}, nil)
	secretClient.On("LookupTokenAccessor", "priv-token", "rootaccessor").
		Return(types.TokenMetadata{Accessor: "rootaccessor", Policies: []string{"root"}}, nil)
	secretClient.On("LookupTokenAccessor", "priv-token", "nonrootaccessor").
		Return(types.TokenMetadata{Accessor: "nonrootaccessor"}, nil)
	secretClient.On("RevokeTokenAccessor", "priv-token", "nonrootaccessor").
		Return(nil)

	// Act
	err := tm.RevokeNonRootTokens("priv-token")

	// Assert
	assert.NoError(t, err)
	secretClient.AssertExpectations(t)
}

func TestRevokeRootTokens(t *testing.T) {
	// Arrange
	logging := logger.MockLogger{}
	secretClient := &mocks.SecretStoreClient{}
	tm := NewTokenMaintenance(logging, secretClient)

	secretClient.On("ListTokenAccessors", "priv-token").
		Return([]string{"rootaccessor", "nonrootaccessor", "priv-token-accessor"}, nil)
	// Make privileged token a root token for this test
	secretClient.On("LookupToken", "priv-token").
		Return(types.TokenMetadata{Accessor: "priv-token-accessor", Policies: []string{"root"}}, nil)
	secretClient.On("LookupTokenAccessor", "priv-token", "rootaccessor").
		Return(types.TokenMetadata{Accessor: "rootaccessor", Policies: []string{"root"}}, nil)
	secretClient.On("LookupTokenAccessor", "priv-token", "nonrootaccessor").
		Return(types.TokenMetadata{Accessor: "nonrootaccessor"}, nil)
	secretClient.On("RevokeTokenAccessor", "priv-token", "rootaccessor").
		Return(nil)

	// Act
	err := tm.RevokeRootTokens("priv-token")

	// Assert
	assert.NoError(t, err)
	secretClient.AssertExpectations(t)
}
