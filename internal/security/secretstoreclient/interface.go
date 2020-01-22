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

package secretstoreclient

// SecretStoreClient is interface to Vault
type SecretStoreClient interface {
	HealthCheck() (statusCode int, err error)
	Init(secretThreshold int, secretShares int, initResponse *InitResponse) (statusCode int, err error)
	Unseal(initResponse *InitResponse) (statusCode int, err error)
	InstallPolicy(token string,
		policyName string, policyDocument string) (statusCode int, err error)
	CreateToken(token string,
		parameters map[string]interface{}, response interface{}) (statusCode int, err error)
	ListAccessors(token string, accessors *[]string) (statusCode int, err error)
	RevokeAccessor(token string, accessor string) (statusCode int, err error)
	LookupAccessor(token string, accessor string, tokenMetadata *TokenMetadata) (statusCode int, err error)
	LookupSelf(token string, tokenMetadata *TokenMetadata) (statusCode int, err error)
	RevokeSelf(token string) (statusCode int, err error)
	RegenRootToken(initResponse *InitResponse, rootToken *string) (err error)
	CheckSecretEngineInstalled(token string, mountPoint string, engine string) (isInstalled bool, err error)
	EnableKVSecretEngine(token string, mountPoint string, kvVersion string) (statusCode int, err error)
}
