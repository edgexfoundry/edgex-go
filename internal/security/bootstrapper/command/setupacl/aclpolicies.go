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
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/setupacl/share"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
)

const (
	// edgeXPolicyRules are rules for edgex services
	// in Phase 2, we will use the same policy for all EdgeX services
	// TODO: phase 3 will have more finer grained policies for each service
	edgeXPolicyRules = `
	# HCL definition of server agent policy for EdgeX
	agent "" {
		policy = "read"
	}
	agent_prefix "edgex" {
		policy = "write"
	}
	node "" {
  		policy = "read"
	}
	node_prefix "edgex" {
		policy = "write"
	}
	service "" {
		policy = "write"
	}
	service_prefix "" {
		policy = "write"
	}
	# allow key value store put
	# once the default_policy is switched to "deny",
	# this is needed if wants to allow updating Key/Value configuration
	key "" {
		policy = "write"
	}
	key_prefix "" {
		policy = "write"
	}
	`

	// edgeXServicePolicyName is the name of the agent policy for edgex
	edgeXServicePolicyName = "edgex-service-policy"

	consulCreatePolicyAPI     = "/v1/acl/policy"
	consulPolicyListAPI       = "/v1/acl/policies"
	consulReadPolicyByNameAPI = "/v1/acl/policy/name/%s"

	aclNotFoundMessage = "ACL not found"

	edgeXManagementPolicyRules = `
	# HCL definition of management policy for EdgeX
	agent "" {
		policy = "read"
	}
	agent_prefix "edgex" {
		policy = "write"
	}
	node "" {
  		policy = "read"
	}
	node_prefix "edgex" {
		policy = "write"
	}
	service "edgex-core-consul" {
		policy = "write"
	}
	service_prefix "" {
		policy = "write"
	}
	# allow key value store put
	# once the default_policy is switched to "deny",
	# this is needed if wants to allow updating Key/Value configuration
	key "" {
		policy = "write"
	}
	key_prefix "" {
		policy = "write"
	}
	`

	// edgeXServicePolicyName is the name of the management policy for edgex
	edgeXManagementPolicyName = "edgex-management-policy"
)

type PolicyListResponse []struct {
	Name string `json:"Name"`
}

// getOrCreateRegistryPolicy retrieves or creates a new policy
// it inserts a new policy if the policy name does not exist and returns a policy
// it returns the same policy if the policy name already exists
func (c *cmd) getOrCreateRegistryPolicy(tokenID, policyName, policyRules string) (*types.Policy, error) {
	// try to get the policy to see if it exists or not
	policy, err := c.getPolicyByName(tokenID, policyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy ID by name %s: %v", policyName, err)
	}

	if policy != nil {
		// policy exists, return this one
		return policy, nil
	}

	createPolicyURL, err := c.getRegistryApiUrl(consulCreatePolicyAPI)
	if err != nil {
		return nil, err
	}

	// payload struct for creating a new policy
	type CreatePolicy struct {
		Name        string `json:"Name"`
		Description string `json:"Description,omitempty"`
		Rules       string `json:"Rules,omitempty"`
	}

	createPolicy := &CreatePolicy{
		Name:        policyName,
		Description: "agent policy for EdgeX microservices",
		Rules:       policyRules,
	}

	jsonPayload, err := json.Marshal(createPolicy)
	c.loggingClient.Tracef("payload: %v", createPolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CreatePolicy JSON string payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, createPolicyURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to prepare create a new policy request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, tokenID)
	req.Header.Add(common.ContentType, common.ContentTypeJSON)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send create a new policy request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	createPolicyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read create a new policy response body: %w", err)
	}

	var created types.Policy

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(bytes.NewReader(createPolicyResp)).Decode(&created); err != nil {
			return nil, fmt.Errorf("failed to decode create a new policy response body: %v", err)
		}

		c.loggingClient.Infof("successfully created a new agent policy with name %s", policyName)

		return &created, nil
	default:
		return nil, fmt.Errorf("failed to create a new policy with name %s via URL [%s] and status code= %d: %s",
			policyName, consulCreatePolicyAPI, resp.StatusCode, string(createPolicyResp))
	}
}

// getPolicyByName gets policy by policy name, returns nil if not found
func (c *cmd) getPolicyByName(tokenID, policyName string) (*types.Policy, error) {
	policyExists, err := c.checkPolicyExists(tokenID, policyName)
	if err != nil {
		return nil, err
	}

	if !policyExists {
		return nil, nil
	}

	readPolicyByNameURL, err := c.getRegistryApiUrl(fmt.Sprintf(consulReadPolicyByNameAPI, policyName))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, readPolicyByNameURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare readPolicyByName request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, tokenID)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send readPolicyByName request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	readPolicyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read readPolicyByName response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var existing types.Policy
		if err := json.NewDecoder(bytes.NewReader(readPolicyResp)).Decode(&existing); err != nil {
			return nil, fmt.Errorf("failed to decode Policy json data: %v", err)
		}

		return &existing, nil
	case http.StatusForbidden:
		// when the policy cannot be found by the name, the body returns "ACL not found"
		// so we treat it as non-error case
		if strings.EqualFold(aclNotFoundMessage, string(readPolicyResp)) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read policy by name with error %s", string(readPolicyResp))
	default:
		return nil, fmt.Errorf("failed to read policy by name with status code= %d: %s",
			resp.StatusCode, string(readPolicyResp))
	}
}

func (c *cmd) checkPolicyExists(tokenID, policyName string) (bool, error) {
	policyListURL, err := c.getRegistryApiUrl(consulPolicyListAPI)
	if err != nil {
		return false, err
	}

	policyListReq, err := http.NewRequest(http.MethodGet, policyListURL, http.NoBody)
	if err != nil {
		return false, fmt.Errorf("Failed to prepare policyListReq request for http URL %s: %w", policyListURL, err)
	}

	policyListReq.Header.Add(share.ConsulTokenHeader, tokenID)
	policyListResp, err := c.client.Do(policyListReq)
	if err != nil {
		return false, fmt.Errorf("Failed to GET policy list request for http URL %s: %w", policyListURL, err)
	}
	defer policyListResp.Body.Close()

	var policyList PolicyListResponse

	err = json.NewDecoder(policyListResp.Body).Decode(&policyList)
	if err != nil {
		return false, fmt.Errorf("Failed to decode policy list reponse: %w", err)
	}

	switch policyListResp.StatusCode {
	case http.StatusOK:
		for _, policy := range policyList {
			// consul is case-sensitive
			if policy.Name == policyName {
				return true, nil
			}
		}
	default:
		return false, fmt.Errorf("Failed to get consul policy list from [%s] and status code= %d", consulPolicyListAPI,
			policyListResp.StatusCode)
	}
	return false, nil
}
