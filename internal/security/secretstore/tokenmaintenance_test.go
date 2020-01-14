//
// Copyright (c) 2019 Intel Corporation
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
// SPDX-License-Identifier: Apache-2.0'
//

package secretstore

import (
	"net/http"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	. "github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTokenIssuingToken(t *testing.T) {
	// Arrange
	logging := logger.MockLogger{}
	secretClient := &MockSecretStoreClient{}
	tm := NewTokenMaintenance(logging, secretClient)

	secretClient.On("InstallPolicy", "root-token",
		TokenCreatorPolicyName, TokenCreatorPolicy).
		Return(http.StatusNoContent, nil)
	secretClient.On("CreateToken", "root-token", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			output := make(map[string]interface{})
			output["auth"] = make(map[string]interface{})
			output["auth"].(map[string]interface{})["client_token"] = "priv-token"
			*(args.Get(2)).(*map[string]interface{}) = output
		}).
		Return(http.StatusOK, nil)
	secretClient.On("RevokeSelf", "priv-token").
		Return(http.StatusNoContent, nil)

	// Act
	token, revoke, err := tm.CreateTokenIssuingToken("root-token")
	revoke()

	// Assert
	assert.Equal(t, "priv-token", token["auth"].(map[string]interface{})["client_token"])
	assert.Nil(t, err)
	secretClient.AssertExpectations(t)
}

func TestRevokeNonRootTokens(t *testing.T) {
	// Arrange
	logging := logger.MockLogger{}
	secretClient := &MockSecretStoreClient{}
	tm := NewTokenMaintenance(logging, secretClient)

	secretClient.On("ListAccessors", "priv-token", mock.Anything).
		Run(func(args mock.Arguments) {
			*(args.Get(1)).(*[]string) = []string{
				"rootaccessor",
				"nonrootaccessor",
				"priv-token-accessor",
			}
		}).
		Return(http.StatusOK, nil)
	secretClient.On("LookupSelf", "priv-token", mock.Anything).
		Run(func(args mock.Arguments) {
			*(args.Get(1)).(*secretstoreclient.TokenMetadata) = secretstoreclient.TokenMetadata{
				Accessor: "priv-token-accessor",
			}
		}).
		Return(http.StatusOK, nil)
	secretClient.On("LookupAccessor", "priv-token", "rootaccessor", mock.Anything).
		Run(func(args mock.Arguments) {
			*(args.Get(2)).(*secretstoreclient.TokenMetadata) = secretstoreclient.TokenMetadata{
				Accessor: "rootaccessor",
				Policies: []string{"root"},
			}
		}).
		Return(http.StatusOK, nil)
	secretClient.On("LookupAccessor", "priv-token", "nonrootaccessor", mock.Anything).
		Run(func(args mock.Arguments) {
			*(args.Get(2)).(*secretstoreclient.TokenMetadata) = secretstoreclient.TokenMetadata{
				Accessor: "nonrootaccessor",
			}
		}).
		Return(http.StatusOK, nil)
	secretClient.On("RevokeAccessor", "priv-token", "nonrootaccessor").
		Return(http.StatusNoContent, nil)

	// Act
	err := tm.RevokeNonRootTokens("priv-token")

	// Assert
	assert.Nil(t, err)
	secretClient.AssertExpectations(t)
}

func TestRevokeRootTokens(t *testing.T) {
	// Arrange
	logging := logger.MockLogger{}
	secretClient := &MockSecretStoreClient{}
	tm := NewTokenMaintenance(logging, secretClient)

	secretClient.On("ListAccessors", "priv-token", mock.Anything).
		Run(func(args mock.Arguments) {
			*(args.Get(1)).(*[]string) = []string{
				"rootaccessor",
				"nonrootaccessor",
				"priv-token-accessor",
			}
		}).
		Return(http.StatusOK, nil)
	secretClient.On("LookupSelf", "priv-token", mock.Anything).
		Run(func(args mock.Arguments) {
			*(args.Get(1)).(*secretstoreclient.TokenMetadata) = secretstoreclient.TokenMetadata{
				Accessor: "priv-token-accessor",
				Policies: []string{"root"}, // Make privileged token a root token for this test
			}
		}).
		Return(http.StatusOK, nil)
	secretClient.On("LookupAccessor", "priv-token", "rootaccessor", mock.Anything).
		Run(func(args mock.Arguments) {
			*(args.Get(2)).(*secretstoreclient.TokenMetadata) = secretstoreclient.TokenMetadata{
				Accessor: "rootaccessor",
				Policies: []string{"root"},
			}
		}).
		Return(http.StatusOK, nil)
	secretClient.On("LookupAccessor", "priv-token", "nonrootaccessor", mock.Anything).
		Run(func(args mock.Arguments) {
			*(args.Get(2)).(*secretstoreclient.TokenMetadata) = secretstoreclient.TokenMetadata{
				Accessor: "nonrootaccessor",
			}
		}).
		Return(http.StatusOK, nil)
	secretClient.On("RevokeAccessor", "priv-token", "rootaccessor").
		Return(http.StatusNoContent, nil)

	// Act
	err := tm.RevokeRootTokens("priv-token")

	// Assert
	assert.Nil(t, err)
	secretClient.AssertExpectations(t)
}
