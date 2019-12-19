/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2019 Intel Corporation
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
 * @author: Tingyu Zeng, Dell / Alain Pulluelo, ForgeRock AS
 *******************************************************************************/

package secretstoreclient

// InitRequest contains a Vault init request regarding the Shamir Secret Sharing (SSS) parameters
type InitRequest struct {
	SecretShares    int `json:"secret_shares"`
	SecretThreshold int `json:"secret_threshold"`
}

// InitResponse contains a Vault init response
type InitResponse struct {
	Keys       []string `json:"keys"`
	KeysBase64 []string `json:"keys_base64"`
	RootToken  string   `json:"root_token"`
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

// UpdateACLPolicyRequest contains a ACL policy create/update request
type UpdateACLPolicyRequest struct {
	Policy string `json:"policy"`
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

// TokenMetadata has introspection data about a token
type TokenMetadata struct {
	Accessor   string   `json:"accessor"`
	ExpireTime string   `json:"expire_time"`
	Path       string   `json:"path"`
	Policies   []string `json:"policies"`
}

// LookupAccessorRequest is used by accessor lookup API
type LookupAccessorRequest struct {
	Accessor string `json:"accessor"`
}

// TokenLookupResponse is the response to the token lookup API
type TokenLookupResponse struct {
	Data TokenMetadata
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

// ListSecretEnginesResponse is the response to GET /v1/sys/mounts
type ListSecretEnginesResponse struct {
	Data map[string]struct {
		Type string `json:"type"`
	} `json:"data"`
}

// EnableSecretsEngineRequest is the POST request to /v1/sys/mounts
type EnableSecretsEngineRequest struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Options     struct {
		Version string `json:"version"`
	} `json:"options"`
}
