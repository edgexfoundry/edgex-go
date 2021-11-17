/*******************************************************************************
 * Copyright 2021 Intel Corp.
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

package vault

import (
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
)

const (
	KeyValue = "kv"
	Consul   = "consul"
)

// InitRequest contains a Vault init request regarding the Shamir Secret Sharing (SSS) parameters
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

// ListSecretEnginesResponse is the response to GET /v1/sys/mounts
type ListSecretEnginesResponse struct {
	Data map[string]struct {
		Type string `json:"type"`
	} `json:"data"`
}

// UnsealRequest contains a Vault unseal request
type UnsealRequest struct {
	Key   string `json:"key"`
	Reset bool   `json:"reset"`
}

// UnsealResponse contains a Vault unseal response
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
