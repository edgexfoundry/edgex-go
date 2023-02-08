/*******************************************************************************
 * Copyright 2021-2023 Intel Corporation
 * Copyright 2019 Dell Inc.
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
 * @author: Daniel Harms, Dell
 * @author: Tingyu Zeng, Dell
 *******************************************************************************/

package secretstore

const (
	// ServiceNameValidationRegx is regex string for valid service name as key
	// the service name eventually becomes part of the URL to Vault's API call
	// Based upon the RFC 3986: https://tools.ietf.org/html/rfc3986#page-12,
	// the following characters are reserved characters for URI's and thus NOT allowed:
	// gen-delims  = ":" / "/" / "?" / "#" / "[" / "]" / "@"
	// sub-delims  = "!" / "$" / "&" / "'" / "(" / ")" / "*" / "+" / "," / ";" / "="
	// backslash (\) also is not allowed due to being as a delimiter for URI directory in Windows
	// percent symbol (%) also is not allowed due to being used to encode the reserved characters in URI
	// and the regular alphanumeric characters like A to Z, a to z, 0 to 9, underscore (_),
	// -, ~, ^,  {, }, |, <, >, . are allowed.
	// and the the length of the name is upto 512 characters
	ServiceNameValidationRegx = `^[\w. \~\^\-\|\<\>\{\}]{1,512}$`

	VaultToken             = "X-Vault-Token" // nolint:gosec
	TokenCreatorPolicyName = "privileged-token-creator"

	// This is an admin token policy that allow for creation of
	// per-service tokens and policies
	// nolint:gosec
	TokenCreatorPolicy = `
path "identity/entity/name" {
  capabilities = ["list"]
}

path "identity/entity/name/*" {
  capabilities = ["create", "update", "read"]
}

path "identity/entity-alias" {
  capabilities = ["create", "update"]
}

path "identity/oidc/role" {
  capabilities = ["list"]
}

path "identity/oidc/role/*" {
  capabilities = ["create", "update"]
}
  
path "auth/userpass/users/*" {
	capabilities = ["create", "update"]
  }
  
path "sys/auth" {
  capabilities = ["read"]
}
  
path "sys/policies/acl/edgex-service-*" {
  capabilities = ["create", "read", "update", "delete" ]
}

path "sys/policies/acl" {
  capabilities = ["list"]
}
`

	// UPAuthMountPoint is where the username/password auth engine is mounted
	UPAuthMountPoint = "userpass"
	// UserPassAuthEngine is the auth engine name
	UserPassAuthEngine = "userpass"
)
