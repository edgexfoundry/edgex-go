/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Corp.
 * Copyright (c) 2025 IOTech Ltd
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

package types

// InitResponse contains a Secret Store init response
type InitResponse struct {
	Keys          []string `json:"keys,omitempty"`
	KeysBase64    []string `json:"keys_base64,omitempty"`
	EncryptedKeys []string `json:"encrypted_keys,omitempty"`
	Nonces        []string `json:"nonces,omitempty"`
	RootToken     string   `json:"root_token,omitempty"`
}

// TokenMetadata has introspection data about a token and is the "data" sub-structure for token lookup,
// i.e. TokenLookupResponse, and token self-lookup
type TokenMetadata struct {
	Accessor   string   `json:"accessor"`
	ExpireTime string   `json:"expire_time"`
	Path       string   `json:"path"`
	Policies   []string `json:"policies"`
	Period     int      `json:"period"` // in seconds
	Renewable  bool     `json:"renewable"`
	Ttl        int      `json:"ttl"` // in seconds
}

// BootStrapACLTokenInfo is the key portion of the response metadata from consulACLBootstrapAPI
type BootStrapACLTokenInfo struct {
	SecretID string   `json:"SecretID"`
	Policies []Policy `json:"Policies"`
}

// Alias has introspection data about entity alias
type Alias struct {
	Name string `json:"name"`
}

// EntityMetadata has introspection data about entity
type EntityMetadata struct {
	Aliases  []Alias  `json:"aliases"`
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Policies []string `json:"policies"`
}
