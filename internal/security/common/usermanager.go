//
// Copyright (c) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package common

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v3/secrets"
)

type UserManager struct {
	logger             logger.LoggingClient
	secretStoreClient  secrets.SecretStoreClient
	userPassMountPoint string // userPassMountPoint is the name of the userpass mount point, almost always "userpass"
	jwtKeyName         string // jwtKeyName is the key identifier of the JWT signing key (e.g. edgex-identity)
	privilegedToken    string // privilegedToken is a Vault token that has permissions to do a lot of stuff like below
	tokenTTL           string // This is the TTL of the Vault token (which is renewable)
	jwtAudience        string // Value of "aud" claim in JWT's (passed in client_id field for creating JWT identity roles)
	jwtTTL             string // JWT's created using the Vault token have an independent validity period
}

func NewUserManager(
	logger logger.LoggingClient,
	secretStoreClient secrets.SecretStoreClient,
	userPassMountPoint string,
	jwtKeyName string,
	privilegedToken string,
	tokenTTL string,
	jwtAudience string,
	jwtTTL string,
) *UserManager {
	return &UserManager{
		logger,
		secretStoreClient,
		userPassMountPoint,
		jwtKeyName,
		privilegedToken,
		tokenTTL,
		jwtAudience,
		jwtTTL,
	}
}

// CreatePasswordUserWithPolicy creates a vault identity with an attached policy
// using userpass authentication engine.
// username should be the name of the user or service to be created
// password should be a random password to be assigned
// policyPrefix is prefixed to username and should be "edgex-user-" or policyPrefix
// policy is a map that will be serialized as a policy attached to the identity
func (m *UserManager) CreatePasswordUserWithPolicy(username string, password string, policyPrefix string, policy map[string]interface{}) error {

	m.logger.Infof("creating policy, identity, and userpass binding for %s", username)

	// Derive the name of the attached policy and create it

	policyName := policyPrefix + username

	policyBytes, err := json.Marshal(policy)
	if err != nil {
		m.logger.Errorf("failed encode policy for %s: %s", username, err.Error())
		return err
	}

	if err := m.secretStoreClient.InstallPolicy(m.privilegedToken, policyName, string(policyBytes)); err != nil {
		m.logger.Errorf("failed to install policy %s: %s", policyName, err.Error())
		return err
	}

	// Create or update underlying vault identity
	identityMetadata := map[string]string{
		// we will also put a name claim in any generated JWT's
		"name": username,
	}
	identityPolicies := []string{policyName}
	identityId, err := m.secretStoreClient.CreateOrUpdateIdentity(m.privilegedToken, username, identityMetadata, identityPolicies)
	if err != nil {
		return err
	}
	if identityId == "" {
		// Updating an entity doesn't return its ID (grr!), in that case, need to look it up
		identityId, err = m.secretStoreClient.LookupIdentity(m.privilegedToken, username)
		if err != nil {
			return err
		}
	}

	// When logging in "default" policy is added automatically to the token and identity_policy is inherited.
	err = m.secretStoreClient.CreateOrUpdateUser(m.privilegedToken, m.userPassMountPoint, username, password, m.tokenTTL, []string{})
	if err != nil {
		return err
	}

	authHandle, err := m.secretStoreClient.LookupAuthHandle(m.privilegedToken, m.userPassMountPoint)
	if err != nil {
		return err
	}

	err = m.secretStoreClient.BindUserToIdentity(m.privilegedToken, identityId, authHandle, username)
	if err != nil {
		return err
	}

	// See https://developer.hashicorm.com/vault/docs/secrets/identity/identity-token#token-contents-and-templates
	// for including custom claims.  We will include a claim identifying the calling EdgeX service
	// We will use the OIDC standard "name" claim.
	customClaims := fmt.Sprintf(`{"name": "%s"}`, username)
	err = m.secretStoreClient.CreateOrUpdateIdentityRole(m.privilegedToken, username, m.jwtKeyName, customClaims, m.jwtAudience, m.jwtTTL)
	if err != nil {
		return err
	}

	return nil
}

func (m *UserManager) DeletePasswordUser(username string) error {

	identityId, err := m.secretStoreClient.LookupIdentity(m.privilegedToken, username)
	if err != nil {
		return err
	}
	if identityId != "" {
		err = m.secretStoreClient.DeleteIdentity(m.privilegedToken, username)
		if err != nil {
			return err
		}
	}

	err = m.secretStoreClient.DeleteUser(m.privilegedToken, m.userPassMountPoint, username)
	if err != nil {
		return err
	}

	return nil
}
