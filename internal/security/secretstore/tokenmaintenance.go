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

/*

High level token maintenance flow is as follows:

1. Recover a root token from keyshares
2. Revoke all previous non-root tokens
3. Create new admin token from that root token
4. Create per-service tokens using admin token
5. Upload other bootstrapping access with root token (e.g. kong cert, db password)
6. Revoke all other root tokens (dangerous: breaks backward-compatibility)
   6a. Go back and retroactively remove root token from resp-init.json
7. Revoke self

End state is that only admin token and per-service tokens remain.
For Fuji, the root token in resp-init.json will also remain.
For Fuji.DOT/Geneva, all root tokens will be revoked.

*/

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type RevokeFunc func()

type TokenMaintenance struct {
	logging      logger.LoggingClient
	secretClient secretstoreclient.SecretStoreClient
}

// NewTokenMaintenance creates a new TokenProvider
func NewTokenMaintenance(logging logger.LoggingClient, secretClient secretstoreclient.SecretStoreClient) *TokenMaintenance {
	return &TokenMaintenance{
		logging:      logging,
		secretClient: secretClient,
	}
}

// CreateTokenIssuingToken creates an admin token that
// allows the holder to create per-service tokens an policies.
// Requires a root token, returns a function that,
// if called, with revoke the token
func (tm *TokenMaintenance) CreateTokenIssuingToken(rootToken string) (map[string]interface{}, RevokeFunc, error) {

	_, err := tm.secretClient.InstallPolicy(rootToken, TokenCreatorPolicyName, TokenCreatorPolicy)
	if err != nil {
		tm.logging.Error("failed installation of token-issuing-token policy")
		return nil, nil, err
	}

	createTokenParameters := make(map[string]interface{})
	createTokenParameters["display_name"] = TokenCreatorPolicyName
	createTokenParameters["no_parent"] = true
	createTokenParameters["period"] = "1h"
	createTokenParameters["policies"] = []string{TokenCreatorPolicyName}
	createTokenParameters["ttl"] = "1h"
	createTokenResponse := make(map[string]interface{})
	_, err = tm.secretClient.CreateToken(rootToken, createTokenParameters, &createTokenResponse)
	if err != nil {
		tm.logging.Error(fmt.Sprintf("failed creation of token-issuing-token: %s", err.Error()))
		return nil, nil, err
	}

	newToken := createTokenResponse["auth"].(map[string]interface{})["client_token"].(string)
	revokeFunc := func() {
		tm.logging.Info("revoking token-issuing-token")
		if _, err2 := tm.secretClient.RevokeSelf(newToken); err2 != nil {
			tm.logging.Warn("failed revokation of token-issuing-token: %s", err2.Error())
		}
	}
	return createTokenResponse, revokeFunc, nil
}

// RevokeNonRootTokens revokes non-root tokens that may have been
// issued in previous EdgeX runs.  Should be called with a high-privileged token.
func (tm *TokenMaintenance) RevokeNonRootTokens(privilegedToken string) error {
	// First enumerate all accessors
	allAccessors := make([]string, 0)
	_, err := tm.secretClient.ListAccessors(privilegedToken, &allAccessors)
	if err != nil {
		return err // secretclient already logged failure
	}

	var selfMetadata secretstoreclient.TokenMetadata
	_, err = tm.secretClient.LookupSelf(privilegedToken, &selfMetadata)
	if err != nil {
		return err // secretclient already logged failure
	}
	selfAccessor := selfMetadata.Accessor

	// Lookup each accessor and figure out which ones are root tokens
	// add the non-root tokens to the list and also don't add ourself
	accessorsToRevoke := make([]string, 0)
	for _, accessor := range allAccessors {
		if accessor == selfAccessor {
			continue // don't revoke ourselves
		}
		tokenMetadata := secretstoreclient.TokenMetadata{}
		_, err := tm.secretClient.LookupAccessor(privilegedToken, accessor, &tokenMetadata)
		if err != nil {
			return err // secretclient already logged failure
		}
		// Search attached policies: flag tokens with root policy attached
		var hasRootToken bool
		for _, policy := range tokenMetadata.Policies {
			if policy == "root" {
				hasRootToken = true
				break
			}
		}
		if !hasRootToken {
			accessorsToRevoke = append(accessorsToRevoke, accessor)
		}
	}

	var lastErr error

	// Revoke all the accessors in the above list
	for _, accessor := range accessorsToRevoke {
		// Revoke as many as we can despite errors
		_, err = tm.secretClient.RevokeAccessor(privilegedToken, accessor)
		if err != nil {
			lastErr = err
		}
	}

	return lastErr // return error if any revoke errored
}

// RevokeRootTokens revokes any root tokens found in the secret store.
// Should be called with a high-privileged token.
func (tm *TokenMaintenance) RevokeRootTokens(privilegedToken string) error {
	// First enumerate all accessors
	allAccessors := make([]string, 0)
	_, err := tm.secretClient.ListAccessors(privilegedToken, &allAccessors)
	if err != nil {
		return err // secretclient already logged failure
	}

	var selfMetadata secretstoreclient.TokenMetadata
	_, err = tm.secretClient.LookupSelf(privilegedToken, &selfMetadata)
	if err != nil {
		return err // secretclient already logged failure
	}
	selfAccessor := selfMetadata.Accessor

	// Iterate and revoke any root tokens found that aren't ourselves
	for _, accessor := range allAccessors {
		if accessor == selfAccessor {
			continue // don't revoke ourselves
		}
		tokenMetadata := secretstoreclient.TokenMetadata{}
		_, err := tm.secretClient.LookupAccessor(privilegedToken, accessor, &tokenMetadata)
		if err != nil {
			return err // secretclient already logged failure
		}
		// Search attached policies: revoke root tokens
		for _, policy := range tokenMetadata.Policies {
			if policy == "root" {
				_, err = tm.secretClient.RevokeAccessor(privilegedToken, accessor)
				if err != nil {
					return err // secretclient already logged failure
				}
			}
		}
	}
	return nil
}
