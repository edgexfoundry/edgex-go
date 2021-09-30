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
	"regexp"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/setupacl/share"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

// AgentTokenType is the type of token to be set on the Consul agent
type AgentTokenType string

const (
	/*
	 * The following are available agent token types that can be used for setting the token to Consul's agent
	 * For the details for each type, see reference https://www.consul.io/commands/acl/set-agent-token#token-types
	 */
	// DefaultType is agent token type "default" to be set
	DefaultType AgentTokenType = "default"
	// AgentType is agent token type "agent" to be set
	AgentType AgentTokenType = "agent"
	// MasterType is agent token type "master" to be set
	MasterType AgentTokenType = "master"
	// ReplicationType is agent token type "replication" to be set
	ReplicationType AgentTokenType = "replication"

	// consul API related:
	consulCheckAgentAPI    = "/v1/agent/self"
	consulSetAgentTokenAPI = "/v1/agent/token/%s" // nolint:gosec
	consulListTokensAPI    = "/v1/acl/tokens"     // nolint:gosec
	consulCreateTokenAPI   = "/v1/acl/token"      // nolint:gosec
	// RUD: Read Update Delete
	consulTokenRUDAPI = "/v1/acl/token/%s" //nolint:gosec
)

// CreateRegistryToken is the structure to create a new registry token
type CreateRegistryToken struct {
	Description string   `json:"Description"`
	Policies    []Policy `json:"Policies"`
	Local       bool     `json:"Local"`
	TTL         *string  `json:"ExpirationTTL,omitempty"`
}

// NewCreateRegistryToken instantiates a new CreateRegistryToken with a given inputs
func NewCreateRegistryToken(description string, policies []Policy, local bool, timeToLive *string) CreateRegistryToken {
	return CreateRegistryToken{
		Description: description,
		Policies:    policies,
		Local:       local,
		TTL:         timeToLive,
	}
}

// isACLTokenPersistent checks Consul agent's configuration property for EnablePersistence of ACLTokens
// it returns true if the token persistence is enabled; false otherwise
// once ACL rules are enforced, this call requires at least agent read permission and hence we use
// the bootstrap ACL token as that's the only token available before creating an agent token
// this determines whether we need to re-set the agent token every time Consul agent is restarted
func (c *cmd) isACLTokenPersistent(bootstrapACLToken string) (bool, error) {
	if len(bootstrapACLToken) == 0 {
		return false, errors.New("bootstrap ACL token is required for checking agent properties")
	}

	checkAgentURL, err := c.getRegistryApiUrl(consulCheckAgentAPI)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodGet, checkAgentURL, http.NoBody)
	if err != nil {
		return false, fmt.Errorf("Failed to prepare checkAgent request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, bootstrapACLToken)
	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("Failed to send checkAgent request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	checkAgentResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("Failed to read checkAgent response body: %w", err)
	}

	type AgentProperties struct {
		AgentDebugConfig struct {
			AgentACLTokens struct {
				EnablePersistence bool `json:"EnablePersistence"`
			} `json:"ACLTokens"`
		} `json:"DebugConfig"`
	}

	var agentProp AgentProperties

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(bytes.NewReader(checkAgentResp)).Decode(&agentProp); err != nil {
			return false, fmt.Errorf("failed to decode AgentProperties json data: %v", err)
		}

		c.loggingClient.Debugf("got AgentProperties: %v", agentProp)

		return agentProp.AgentDebugConfig.AgentACLTokens.EnablePersistence, nil
	default:
		return false, fmt.Errorf("failed to check agent ACL token persistence property with status code= %d: %s",
			resp.StatusCode, string(checkAgentResp))
	}
}

// createAgentToken uses bootstrapped ACL token to create an agent token
// this call requires ACL read/write permission and hence we use the bootstrap ACL token
// it checks whether there is an agent token already existing and re-uses it if so
// otherwise creates a new agent token
func (c *cmd) createAgentToken(bootstrapACLToken BootStrapACLTokenInfo) (string, error) {
	if len(bootstrapACLToken.SecretID) == 0 {
		return share.EmptyToken, errors.New("bootstrap ACL token is required for creating agent token")
	}

	// list tokens and search for the "edgex-core-consul" agent token if any
	// Note: the internal Consul ACL system may not be ready yet, hence we need to retry it as to
	// error on ACL in legacy mode until timed out otherwise
	timeoutInSec := int(c.retryTimeout.Seconds())
	timer := startup.NewTimer(timeoutInSec, 1)
	var agentToken string
	var err error
	for timer.HasNotElapsed() {
		agentToken, err = c.getEdgeXAgentToken(bootstrapACLToken)

		if err != nil && !strings.Contains(err.Error(), consulLegacyACLModeError) {
			// other type of request error, cannot continue
			return share.EmptyToken, fmt.Errorf("Failed to retrieve EdgeX agent token: %v", err)
		} else if err != nil {
			c.loggingClient.Warnf("found Consul still in ACL legacy mode, will retry once again: %v", err)
			timer.SleepForInterval()
			continue
		}

		// once reach here, Consul ACL system is ready
		c.loggingClient.Info("internal Consul ACL system is ready")
		break
	}

	// retries reached timeout, aborting
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to retrieve EdgeX agent token: %v", err)
	}

	if len(agentToken) == 0 {
		// need to create a new agent token as there is no matched one found
		agentToken, err = c.insertNewAgentToken(bootstrapACLToken)
		if err != nil {
			return share.EmptyToken, fmt.Errorf("Failed to insert a new EdgeX agent token: %v", err)
		}
	}
	return agentToken, nil
}

// getEdgeXAgentToken lists tokens and find the matched agent token by the expected key pattern
// it returns the first matched agent token if many tokens actually are matched
// it returns empty token if no matching found
func (c *cmd) getEdgeXAgentToken(bootstrapACLToken BootStrapACLTokenInfo) (string, error) {
	listTokensURL, err := c.getRegistryApiUrl(consulListTokensAPI)
	if err != nil {
		return share.EmptyToken, err
	}

	req, err := http.NewRequest(http.MethodGet, listTokensURL, http.NoBody)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to prepare getEdgeXAgentToken request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, bootstrapACLToken.SecretID)
	resp, err := c.client.Do(req)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to send getEdgeXAgentToken request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	listTokensResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to read getEdgeXAgentToken response body: %w", err)
	}

	type ListTokensInfo []struct {
		AccessorID  string `json:"AccessorID"`
		Description string `json:"Description"`
	}

	var listTokens ListTokensInfo

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(bytes.NewReader(listTokensResp)).Decode(&listTokens); err != nil {
			return share.EmptyToken, fmt.Errorf("failed to decode ListTokensInfo json data: %v", err)
		}

		edgexAgentToken := share.EmptyToken // initial value
		// use Description to match the substring to search for EdgeX's agent token
		// we cannot use policy to search yet as the agent token is created with the global policy
		// matching anything contains pattern "edgex[alphanumeric_with_space_or_dash] agent token" with case insensitive
		pattern := regexp.MustCompile(`(?i)edgex([0-9a-z\- ]+)agent token`)
		for _, token := range listTokens {
			if pattern.MatchString(token.Description) {
				tokenID, err := c.readTokenIDBy(bootstrapACLToken, strings.TrimSpace(token.AccessorID))
				if err != nil {
					return share.EmptyToken, err
				}
				edgexAgentToken = tokenID
				break
			}
		}

		return edgexAgentToken, nil
	default:
		return share.EmptyToken, fmt.Errorf("failed to list tokens with status code= %d: %s", resp.StatusCode,
			string(listTokensResp))
	}
}

// readTokenID reads the tokenID (i.e. SecretID) by given accessorID
// it returns the token's SecretID if found, otherwise empty string
func (c *cmd) readTokenIDBy(bootstrapACLToken BootStrapACLTokenInfo, accessorID string) (string, error) {
	if len(accessorID) == 0 {
		return share.EmptyToken, errors.New("accessorID is required and cannot be empty")
	}

	readTokenURL, err := c.getRegistryApiUrl(fmt.Sprintf(consulTokenRUDAPI, accessorID))
	if err != nil {
		return share.EmptyToken, err
	}

	req, err := http.NewRequest(http.MethodGet, readTokenURL, http.NoBody)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to prepare readTokenID request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, bootstrapACLToken.SecretID)
	resp, err := c.client.Do(req)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to send readTokenID request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	readTokenResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to read readTokenID response body: %w", err)
	}

	type TokenReadInfo struct {
		ID string `json:"SecretID"`
	}

	var tokenRead TokenReadInfo

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(bytes.NewReader(readTokenResp)).Decode(&tokenRead); err != nil {
			return share.EmptyToken, fmt.Errorf("failed to decode TokenReadInfo json data: %v", err)
		}

		return tokenRead.ID, nil
	default:
		return share.EmptyToken, fmt.Errorf("failed to read token ID with status code= %d: %s", resp.StatusCode,
			string(readTokenResp))
	}
}

// insertNewAgentToken creates a new Consul token
// it returns the token's ID and error if any error occurs
func (c *cmd) insertNewAgentToken(bootstrapACLToken BootStrapACLTokenInfo) (string, error) {
	// get a policy for this agent token to associate with
	edgexAgentPolicy, err := c.getOrCreateRegistryPolicy(bootstrapACLToken.SecretID,
		"edgex-agent-policy",
		edgeXPolicyRules)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("failed to create edgex agent policy: %v", err)
	}

	unlimitedDuration := "0s"
	createToken := NewCreateRegistryToken("edgex-core-consul agent token",
		[]Policy{
			*edgexAgentPolicy,
		}, true, &unlimitedDuration)
	newTokenInfo, err := c.createNewToken(bootstrapACLToken.SecretID, createToken)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("failed to insert new edgex agent token: %v", err)
	}

	var parsedTokenResponse map[string]interface{}
	if err := json.NewDecoder(strings.NewReader(newTokenInfo)).Decode(&parsedTokenResponse); err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to decode create token info: %v", err)
	}

	c.loggingClient.Info("successfully created a new agent token")

	return fmt.Sprintf("%s", parsedTokenResponse["SecretID"]), nil
}

// setAgentToken sets the ACL token currently in use by the agent
func (c *cmd) setAgentToken(bootstrapACLToken BootStrapACLTokenInfo, agentTokenID string,
	tokenType AgentTokenType) error {
	if len(bootstrapACLToken.SecretID) == 0 {
		return errors.New("bootstrap ACL token is required for setting agent token")
	}

	if len(agentTokenID) == 0 {
		return errors.New("agent token ID is required for setting agent token")
	}

	setAgentTokenURL, err := c.getRegistryApiUrl(fmt.Sprintf(consulSetAgentTokenAPI, tokenType))
	if err != nil {
		return err
	}

	type SetAgentToken struct {
		Token string `json:"Token"`
	}

	setToken := &SetAgentToken{
		Token: agentTokenID,
	}
	jsonPayload, err := json.Marshal(setToken)
	if err != nil {
		return fmt.Errorf("Failed to marshal SetAgentToken JSON string payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, setAgentTokenURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("Failed to prepare SetAgentToken request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, bootstrapACLToken.SecretID)
	req.Header.Add(common.ContentType, common.ContentTypeJSON)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send SetAgentToken request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		// no response body returned in this case
		c.loggingClient.Infof("agent token is set with type [%s]", tokenType)
	default:
		setAgentTokenResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Failed to read SetAgentToken response body with status code=%d: %w", resp.StatusCode, err)
		}

		return fmt.Errorf("failed to set agent token with type %s via URL [%s] and status code= %d: %s",
			tokenType, setAgentTokenURL, resp.StatusCode, string(setAgentTokenResp))
	}

	return nil
}

// createNewToken creates a new token based on the provided inputs
// it returns the whole json string containing the token and thus can be written to the file later
func (c *cmd) createNewToken(bootstrapACLTokenID string, createToken CreateRegistryToken) (string, error) {
	if len(bootstrapACLTokenID) == 0 {
		return share.EmptyToken, fmt.Errorf("bootstrap token ID cannot be empty")
	}

	createTokenURL, err := c.getRegistryApiUrl(consulCreateTokenAPI)
	if err != nil {
		return share.EmptyToken, err
	}

	jsonPayload, err := json.Marshal(&createToken)
	c.loggingClient.Tracef("payload: %v", createToken)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to marshal CreatRegistryToken JSON string payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, createTokenURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to prepare creat a new token request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, bootstrapACLTokenID)
	req.Header.Add(common.ContentType, common.ContentTypeJSON)
	resp, err := c.client.Do(req)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to send create a new token request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	createTokenResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("Failed to read create a new token response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		c.loggingClient.Info("successfully created a new registry token")
		return string(createTokenResp), nil
	default:
		return share.EmptyToken, fmt.Errorf("failed to create a new token via URL [%s] and status code= %d: %s",
			consulCreateTokenAPI, resp.StatusCode, string(createTokenResp))
	}
}
