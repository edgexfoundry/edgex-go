/*******************************************************************************
 * Copyright 2021 Intel Corp.
 * Copyright 2025 IOTech Ltd
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

package openbao

import (
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/types"
)

const (
	KeyValue                   = "kv"
	UsernamePasswordAuthMethod = "userpass"
)

// InitRequest contains a secret store init request regarding the Shamir Secret Sharing (SSS) parameters
type InitRequest struct {
	SecretShares    int `json:"secret_shares"`
	SecretThreshold int `json:"secret_threshold"`
}

// UpdateACLPolicyRequest contains a ACL policy create/update request
type UpdateACLPolicyRequest struct {
	Policy string `json:"policy"`
}

// RootTokenControlResponse is the response to /v1/sys/generate-root/attempt
type RootTokenControlResponse struct {
	Complete bool   `json:"complete"`
	Nonce    string `json:"nonce"`
	Otp      string `json:"otp"`
}

// RootTokenRetrievalRequest is the request to /v1/sys/generate-root/update
type RootTokenRetrievalRequest struct {
	Key   string `json:"key"`
	Nonce string `json:"nonce"`
}

// RootTokenRetrievalResponse is the response to /v1/sys/generate-root/update
type RootTokenRetrievalResponse struct {
	Complete     bool   `json:"complete"`
	EncodedToken string `json:"encoded_token"`
}

type TokenLookupResponse struct {
	Data types.TokenMetadata
}

// ListTokenAccessorsResponse is the response to the list accessors API
type ListTokenAccessorsResponse struct {
	Data struct {
		Keys []string `json:"keys"`
	} `json:"data"`
}

// RevokeTokenAccessorRequest is the input to the revoke token by accessor API
type RevokeTokenAccessorRequest struct {
	Accessor string `json:"accessor"`
}

// LookupAccessorRequest is used by accessor lookup API
type LookupAccessorRequest struct {
	Accessor string `json:"accessor"`
}

// ListSecretEnginesResponse is the response to GET /v1/sys/mounts (and /v1/sys/auth)
type ListSecretEnginesResponse struct {
	Data map[string]struct {
		Type string `json:"type"`
	} `json:"data"`
}

// UnsealRequest contains a secret store unseal request
type UnsealRequest struct {
	Key   string `json:"key"`
	Reset bool   `json:"reset"`
}

// UnsealResponse contains a secret store unseal response
type UnsealResponse struct {
	Sealed   bool `json:"sealed"`
	T        int  `json:"t"`
	N        int  `json:"n"`
	Progress int  `json:"progress"`
}

type SecretsEngineOptions struct {
	Version string `json:"version"`
}

// SecretsEngineConfig is config for /v1/sys/mounts
type SecretsEngineConfig struct {
	DefaultLeaseTTLDuration string `json:"default_lease_ttl"`
}

// EnableSecretsEngineRequest is the POST request to /v1/sys/mounts
type EnableSecretsEngineRequest struct {
	Type        string                `json:"type"`
	Description string                `json:"description"`
	Options     *SecretsEngineOptions `json:"options,omitempty"`
	Config      *SecretsEngineConfig  `json:"config,omitempty"`
}

// CreateUpdateEntityRequest enables or updates a secret store Identity
type CreateUpdateEntityRequest struct {
	Metadata map[string]string `json:"metadata"`
	Policies []string          `json:"policies"`
}

// JsonID
type JsonID struct {
	ID string `json:"id"`
}

// CreateUpdateEntityResponse is the response to CreateUpdateEntityRequest
type CreateUpdateEntityResponse struct {
	Data JsonID `json:"data"`
}

// ReadEntityByNameResponse is the response to get entity by name
type ReadEntityByNameResponse struct {
	Data JsonID `json:"data"`
}

// ReadEntityByIdResponse is the response to get entity by id
type ReadEntityByIdResponse struct {
	Data types.EntityMetadata
}

// EnableAuthMethodRequest enables a secret store Identity authentication method
type EnableAuthMethodRequest struct {
	Type string `json:"type"`
}

// Accessor
type Accessor struct {
	Accessor string `json:"accessor"`
}

// ListAuthMethodsResponse is used to look up the accessor ID of an auth method
type ListAuthMethodsResponse struct {
	Data map[string]Accessor `json:"data"`
}

// CreateOrUpdateUserRequest is used to create a secret store login
type CreateOrUpdateUserRequest struct {
	Password      string   `json:"password"`
	TokenPeriod   string   `json:"token_period"`
	TokenPolicies []string `json:"token_policies"`
}

// CreateOrUpdateUserResponse is the response to get entity by name
type CreateOrUpdateUserResponse struct {
	Data JsonID `json:"data"`
}

// CreateEntityAliasRequest is used to bind an authenticator to an identity
type CreateEntityAliasRequest struct {
	// Name is the username in the authenticator
	Name string `json:"name"`
	// CanonicalID is the entity ID
	CanonicalID string `json:"canonical_id"`
	// MountAccessor is the id if the auth engine to use
	MountAccessor string `json:"mount_accessor"`
}

// UserPassLoginRequest is used to to log in an identity with the userpass auth engine
type UserPassLoginRequest struct {
	Password string `json:"password"`
}

// ListNamedKeysResponse is the response to LIST /v1/identity/oidc/key
type ListNamedKeysResponse struct {
	Data struct {
		Keys []string `json:"keys"`
	} `json:"data"`
}

// CreateNamedKeyRequest is the request to POST /v1/identity/oidc/key/:name:
type CreateNamedKeyRequest struct {
	AllowedClientIDs []string `json:"allowed_client_ids"`
	Algorithm        string   `json:"algorithm"`
}

// CreateOrUpdateIdentityRoleRequest is the request to POST /v1/identity/oidc/role/:name
type CreateOrUpdateIdentityRoleRequest struct {
	ClientID string  `json:"client_id,omitempty"`
	Key      string  `json:"key"`
	Template *string `json:"template,omitempty"`
	TokenTTL string  `json:"ttl"`
}
