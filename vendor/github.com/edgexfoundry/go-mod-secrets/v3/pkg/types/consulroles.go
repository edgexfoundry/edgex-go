/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2022 Intel Corp.
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

package types

import "strings"

type ConsulTokenType string

const (
	/*
	 * The following are available Consul token types that can be used for specifying in the role-based tokens
	 * created via /consul/creds secret engine Vault API.
	 * For the details, see reference https://www.vaultproject.io/api/secret/consul#create-update-role
	 */
	// ManagementType is the type of Consul role can be used to create tokens when role-based API /consul/creds is called
	// the management type of created tokens is automatically granted the built-in global management policy
	ManagementType ConsulTokenType = "management"
	// ClientType is the type of Consul role that can be used to create tokens when role-based API /consul/creds is called
	// the regular client type of created tokens is associated with custom policies
	ClientType ConsulTokenType = "client"
)

type ConsulRole struct {
	RoleName    string   `json:"name"`
	TokenType   string   `json:"token_type"`
	PolicyNames []string `json:"policies,omitempty"`
	Local       bool     `json:"local,omitempty"`
	TimeToLive  string   `json:"TTL,omitempty"`
}

type Policy struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

func NewConsulRole(name string, tokenType ConsulTokenType, policies []Policy, localUse bool) ConsulRole {
	// to conform to the payload of the Consul create role API,
	// we convert the slice of policies from type Policy to string and make it unique
	// as the policy name needs to be unique per API's requirement
	policyNames := make([]string, 0, len(policies))
	tempMap := make(map[string]bool)
	for _, policy := range policies {
		if _, exists := tempMap[policy.Name]; !exists {
			policyNames = append(policyNames, policy.Name)
		}
	}

	return ConsulRole{
		RoleName:    strings.TrimSpace(name),
		TokenType:   string(tokenType),
		PolicyNames: policyNames,
		Local:       localUse,
		// unlimited for now
		TimeToLive: "0s",
	}
}
