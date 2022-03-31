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

/*

High level token maintenance flow is as follows:

1. Recover a root token from key shares
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
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/tokencreatable"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/v2/secrets"
)

type TokenMaintenance struct {
	logging      logger.LoggingClient
	secretClient secrets.SecretStoreClient
}

// NewTokenMaintenance creates a new TokenProvider
func NewTokenMaintenance(logging logger.LoggingClient, secretClient secrets.SecretStoreClient) *TokenMaintenance {
	return &TokenMaintenance{
		logging:      logging,
		secretClient: secretClient,
	}
}

// CreateTokenIssuingToken creates an admin token that
// allows the holder to create per-service tokens an policies.
// Requires a root token, returns a function that,
// if called, with revoke the token
func (tm *TokenMaintenance) CreateTokenIssuingToken(rootToken string) (map[string]interface{},
	tokencreatable.RevokeFunc, error) {

	err := tm.secretClient.InstallPolicy(rootToken, TokenCreatorPolicyName, TokenCreatorPolicy)
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
	createTokenResponse, err := tm.secretClient.CreateToken(rootToken, createTokenParameters)
	if err != nil {
		tm.logging.Errorf("failed creation of token-issuing-token: %s", err.Error())
		return nil, nil, err
	}

	newToken := createTokenResponse["auth"].(map[string]interface{})["client_token"].(string)
	revokeFunc := func() {
		tm.logging.Info("revoking token-issuing-token")
		if err2 := tm.secretClient.RevokeToken(newToken); err2 != nil {
			tm.logging.Warnf("failed revocation of token-issuing-token: %s", err2.Error())
		}
	}
	return createTokenResponse, revokeFunc, nil
}

// RevokeNonRootTokens revokes non-root tokens that may have been
// issued in previous EdgeX runs.  Should be called with a high-privileged token.
func (tm *TokenMaintenance) RevokeNonRootTokens(privilegedToken string) error {
	// First enumerate all accessors
	allAccessors, err := tm.secretClient.ListTokenAccessors(privilegedToken)
	if err != nil {
		return err // secret client already logged failure
	}

	selfMetadata, err := tm.secretClient.LookupToken(privilegedToken)
	if err != nil {
		return err // secret client already logged failure
	}
	selfAccessor := selfMetadata.Accessor

	// Lookup each accessor and figure out which ones are root tokens
	// add the non-root tokens to the list and also don't add ourself
	accessorsToRevoke := make([]string, 0)
	for _, accessor := range allAccessors {
		if accessor == selfAccessor {
			continue // don't revoke ourselves
		}
		tokenMetadata, err := tm.secretClient.LookupTokenAccessor(privilegedToken, accessor)
		if err != nil {
			return err // secret client already logged failure
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
		err = tm.secretClient.RevokeTokenAccessor(privilegedToken, accessor)
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
	allAccessors, err := tm.secretClient.ListTokenAccessors(privilegedToken)
	if err != nil {
		return err // secret client already logged failure
	}

	selfMetadata, err := tm.secretClient.LookupToken(privilegedToken)
	if err != nil {
		return err // secret client already logged failure
	}
	selfAccessor := selfMetadata.Accessor

	// Iterate and revoke any root tokens found that aren't ourselves
	for _, accessor := range allAccessors {
		if accessor == selfAccessor {
			continue // don't revoke ourselves
		}
		tokenMetadata, err := tm.secretClient.LookupTokenAccessor(privilegedToken, accessor)
		if err != nil {
			return err // secret client already logged failure
		}
		// Search attached policies: revoke root tokens
		for _, policy := range tokenMetadata.Policies {
			if policy == "root" {
				err = tm.secretClient.RevokeTokenAccessor(privilegedToken, accessor)
				if err != nil {
					return err // secret client already logged failure
				}
			}
		}
	}
	return nil
}
