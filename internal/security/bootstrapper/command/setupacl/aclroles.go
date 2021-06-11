/*******************************************************************************
 * Copyright 2021 Intel Corporation
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
 *******************************************************************************/

package setupacl

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

// RegistryTokenType is the type of registry tokens that will be created when the role is using to call token creds API
type RegistryTokenType string

const (
	/*
	 * The following are available registry token types that can be used for specifying in the role-based tokens
	 * created via /consul/creds secret engine Vault API.
	 * For the details, see reference https://www.vaultproject.io/api/secret/consul#create-update-role
	 */
	// ManagementType is the type of registry role can be used to create tokens when role-based API /consul/creds is called
	// the management type of created tokens is automatically granted the built-in global management policy
	ManagementType RegistryTokenType = "management"
	// ClientType is the type of registry role that can be used to create tokens when role-based API /consul/creds is called
	// the regular client type of created tokens is associated with custom policies
	ClientType RegistryTokenType = "client"

	createConsulRoleVaultAPI = "/v1/consul/roles/%s"
)

// RegistryRole is the meta definition for creating registry's role
type RegistryRole struct {
	RoleName    string   `json:"name"`
	TokenType   string   `json:"token_type"`
	PolicyNames []string `json:"policies,omitempty"`
	Local       bool     `json:"local,omitempty"`
	TimeToLive  string   `json:"TTL,omitempty"`
}

// NewRegistryRole instantiates a new RegistryRole with the given inputs
func NewRegistryRole(name string, tokenType RegistryTokenType, policies []Policy, localUse bool) RegistryRole {
	// to conform to the payload of the registry create role API,
	// we convert the slice of policies from type Policy to string and make it unique
	// as the policy name needs to be unique per API's requirement
	policyNames := make([]string, 0, len(policies))
	tempMap := make(map[string]bool)
	for _, policy := range policies {
		if _, exists := tempMap[policy.Name]; !exists {
			policyNames = append(policyNames, policy.Name)
		}
	}

	return RegistryRole{
		RoleName:    strings.TrimSpace(name),
		TokenType:   string(tokenType),
		PolicyNames: policyNames,
		Local:       localUse,
		// unlimited for now
		TimeToLive: "0s",
	}
}

// createRole creates a secret store role that can be used to generate registry tokens
// and part of elements for the role ties up with the registry policies in which it dictates
// the permission of accesses to the registry kv store or agent etc.
func (c *cmd) createRole(secretStoreToken string, registryRole RegistryRole) error {
	if len(secretStoreToken) == 0 {
		return errors.New("required secret store token is empty")
	}

	if len(registryRole.RoleName) == 0 {
		return errors.New("required role name cannot be empty")
	}

	createRoleURL := fmt.Sprintf("%s://%s:%d%s", c.configuration.SecretStore.Protocol,
		c.configuration.SecretStore.Host, c.configuration.SecretStore.Port,
		fmt.Sprintf(createConsulRoleVaultAPI, registryRole.RoleName))
	_, err := url.Parse(createRoleURL)
	if err != nil {
		return fmt.Errorf("failed to parse create role URL: %v", err)
	}

	c.loggingClient.Debugf("createRoleURL: %s", createRoleURL)

	jsonPayload, err := json.Marshal(&registryRole)

	if err != nil {
		return fmt.Errorf("failed to marshal JSON string payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, createRoleURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to prepare POST request for http URL: %w", err)
	}

	req.Header.Add("X-Vault-Token", secretStoreToken)
	req.Header.Add(common.ContentType, common.ContentTypeJSON)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusNoContent:
		// no response body returned in this case
		c.loggingClient.Infof("successfully created a role [%s] for secretstore", registryRole.RoleName)
		return nil
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.loggingClient.Errorf("cannot read resp.Body: %v", err)
		}
		return fmt.Errorf("failed to create a role %s for secretstore via URL [%s] and status code= %d: %s",
			registryRole.RoleName, createRoleURL, resp.StatusCode, string(body))
	}
}
