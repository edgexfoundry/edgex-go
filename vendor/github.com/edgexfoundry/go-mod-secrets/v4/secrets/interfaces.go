/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Corp.
 * Copyright 2024-2025 IOTech Ltd
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
 *******************************************************************************/

package secrets

import (
	"context"

	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/types"
)

// SecretClient provides a contract for storing and retrieving secrets from a secret store provider.
type SecretClient interface {
	// GetSecret retrieves secret from a secret store.
	// secretName specifies the type or location of the secret to retrieve. If specified it is appended
	// to the base path from the SecretConfig
	// keys specifies the secret data to retrieve. If no keys are provided then all the keys associated with the
	// specified path will be returned.
	GetSecret(secretName string, keys ...string) (map[string]string, error)

	// StoreSecret stores the secret to a secret store.
	// it sets the values requested at provided keys
	// secretName specifies the type or location of the secret to store.
	// data map specifies the "key": "value" pairs of secret data to store
	StoreSecret(secretName string, data map[string]string) error

	// SetAuthToken sets the internal Auth Token with the new value specified.
	SetAuthToken(ctx context.Context, token string) error

	// GetSecretNames retrieves the secret names currently in service's secret store.
	GetSecretNames() ([]string, error)

	// GetSelfJWT returns an encoded JWT for the current identity-based secret store token
	GetSelfJWT(serviceKey string) (string, error)

	// IsJWTValid evaluates a given JWT and returns a true/false if the JWT is valid (i.e. belongs to us and current) or not
	IsJWTValid(jwt string) (bool, error)
}

// SecretStoreClient provides a contract for managing a Secret Store from a secret store provider.
type SecretStoreClient interface {
	HealthCheck() (int, error)
	Init(secretThreshold int, secretShares int) (types.InitResponse, error)
	Unseal(keysBase64 []string) error
	InstallPolicy(token string, policyName string, policyDocument string) error
	CheckSecretEngineInstalled(token string, mountPoint string, engine string) (bool, error)
	EnableKVSecretEngine(token string, mountPoint string, kvVersion string) error
	RegenRootToken(keys []string) (string, error)
	CreateToken(token string, parameters map[string]any) (map[string]any, error)
	CreateTokenByRole(token string, role string, parameters map[string]any) (map[string]any, error)
	ListTokenAccessors(token string) ([]string, error)
	RevokeTokenAccessor(token string, accessor string) error
	LookupTokenAccessor(token string, accessor string) (types.TokenMetadata, error)
	LookupToken(token string) (types.TokenMetadata, error)
	RevokeToken(token string) error
	CreateOrUpdateTokenRole(token string, roleName string, parameters map[string]any) error
	CreateOrUpdateIdentity(token string, name string, metadata map[string]string, policies []string) (string, error)
	DeleteIdentity(token string, name string) error
	LookupIdentity(token string, name string) (string, error)
	GetIdentityByEntityId(token string, entityId string) (types.EntityMetadata, error)
	CheckAuthMethodEnabled(token string, mountPoint string, authType string) (bool, error)
	EnablePasswordAuth(token string, mountPoint string) error
	LookupAuthHandle(token string, mountPoint string) (string, error)
	CreateOrUpdateUser(token string, mountPoint string, username string, password string, tokenTTL string, tokenPolicies []string) error
	DeleteUser(token string, mountPoint string, username string) error
	BindUserToIdentity(token string, identityId string, authHandle string, username string) error
	InternalServiceLogin(token string, authEngine string, username string, password string) (map[string]any, error)
	CheckIdentityKeyExists(token string, keyName string) (bool, error)
	CreateNamedIdentityKey(token string, keyName string, algorithm string) error
	CreateOrUpdateIdentityRole(token string, roleName string, keyName string, template string, audience string, jwtTTL string) error
}
