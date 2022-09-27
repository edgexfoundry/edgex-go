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
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/setupacl/share"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
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
	Description string         `json:"Description"`
	Policies    []types.Policy `json:"Policies"`
	Local       bool           `json:"Local"`
	TTL         *string        `json:"ExpirationTTL,omitempty"`
}

// ACLTokenInfo is the key portion of the response metadata from consulCreateTokenAPI
type ACLTokenInfo struct {
	SecretID    string         `json:"SecretID"`
	AccessorID  string         `json:"AccessorID"`
	Policies    []types.Policy `json:"Policies"`
	Description string         `json:"Description"`
}

// ManagementACLTokenInfo is the key portion of the response metadata from consulCreateTokenAPI for management acl token
type ManagementACLTokenInfo struct {
	SecretID string         `json:"SecretID"`
	Policies []types.Policy `json:"Policies"`
}

// NewCreateRegistryToken instantiates a new CreateRegistryToken with a given inputs
func NewCreateRegistryToken(description string, policies []types.Policy, local bool, timeToLive *string) CreateRegistryToken {
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
		return false, fmt.Errorf("failed to prepare checkAgent request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, bootstrapACLToken)
	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send checkAgent request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	checkAgentResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read checkAgent response body: %w", err)
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
func (c *cmd) createAgentToken(bootstrapACLToken types.BootStrapACLTokenInfo) (string, error) {
	if len(bootstrapACLToken.SecretID) == 0 {
		return share.EmptyToken, errors.New("bootstrap ACL token is required for creating agent token")
	}

	// list tokens and search for the "edgex-core-consul" agent token if any
	// Note: the internal Consul ACL system may not be ready yet, hence we need to retry it as to
	// error on ACL in legacy mode until timed out otherwise
	timeoutInSec := int(c.retryTimeout.Seconds())
	timer := startup.NewTimer(timeoutInSec, 1)
	var aclTokenInfo *ACLTokenInfo
	var err error
	for timer.HasNotElapsed() {
		pattern := regexp.MustCompile(`(?i)edgex([0-9a-z\- ]+)agent token`)
		aclTokenInfo, err = c.getEdgeXTokenByPattern(bootstrapACLToken, pattern)

		if err != nil && !strings.Contains(err.Error(), consulLegacyACLModeError) {
			// other type of request error, cannot continue
			return share.EmptyToken, fmt.Errorf("failed to retrieve EdgeX agent token: %v", err)
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
		return share.EmptyToken, fmt.Errorf("failed to retrieve EdgeX agent token: %v", err)
	}

	var agentToken string
	// when search by pattern not found it shall return empty token struct
	if reflect.DeepEqual(*aclTokenInfo, ACLTokenInfo{}) {
		// need to create a new agent token as there is no matched one found
		agentToken, err = c.insertNewAgentToken(bootstrapACLToken)
		if err != nil {
			return share.EmptyToken, fmt.Errorf("failed to insert a new EdgeX agent token: %v", err)
		}
	} else {
		agentToken = aclTokenInfo.SecretID
	}
	return agentToken, nil
}

// getEdgeXTokenByPattern lists tokens and find the matched ACL Token info that contains token by the expected key pattern
// it returns the first matched ACL Token info that contains token if many tokens actually are matched
// it returns empty ACLTokenInfo if no matching found
func (c *cmd) getEdgeXTokenByPattern(bootstrapACLToken types.BootStrapACLTokenInfo, pattern *regexp.Regexp) (*ACLTokenInfo, error) {
	listTokensURL, err := c.getRegistryApiUrl(consulListTokensAPI)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, listTokensURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare getEdgeXTokenByPattern request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, bootstrapACLToken.SecretID)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send getEdgeXTokenByPattern request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	listTokensResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read getEdgeXTokenByPattern response body: %w", err)
	}

	type ListTokensInfo []struct {
		ACLTokenInfo
	}

	var listTokens ListTokensInfo

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(bytes.NewReader(listTokensResp)).Decode(&listTokens); err != nil {
			return nil, fmt.Errorf("failed to decode ListTokensInfo json data: %v", err)
		}

		edgexTokenByPattern := ACLTokenInfo{} // initial value
		// use Description to match the substring to search for EdgeX's token
		// we cannot use policy to search yet as the token is created with the global policy
		// matching anything contains pattern "edgex[alphanumeric_with_space_or_dash] <example> token" with case insensitive
		// <example> can be agent or management
		for _, token := range listTokens {
			if pattern.MatchString(token.Description) {
				edgexTokenByPattern = token.ACLTokenInfo
				break
			}
		}

		return &edgexTokenByPattern, nil
	default:
		return nil, fmt.Errorf("failed to list tokens with status code= %d: %s", resp.StatusCode,
			string(listTokensResp))
	}
}

// insertNewAgentToken creates a new Consul token
// it returns the token's ID and error if any error occurs
func (c *cmd) insertNewAgentToken(bootstrapACLToken types.BootStrapACLTokenInfo) (string, error) {
	// get a policy for this agent token to associate with
	edgexAgentPolicy, err := c.getOrCreateRegistryPolicy(bootstrapACLToken.SecretID,
		"edgex-agent-policy",
		edgeXPolicyRules)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("failed to create edgex agent policy: %v", err)
	}

	unlimitedDuration := "0s"
	createToken := NewCreateRegistryToken("edgex-core-consul agent token",
		[]types.Policy{
			*edgexAgentPolicy,
		}, true, &unlimitedDuration)
	newTokenInfo, err := c.createNewToken(bootstrapACLToken.SecretID, createToken)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("failed to insert new edgex agent token: %v", err)
	}
	c.loggingClient.Info("successfully created a new agent token")

	return newTokenInfo.SecretID, nil
}

// setAgentToken sets the ACL token currently in use by the agent
func (c *cmd) setAgentToken(bootstrapACLToken types.BootStrapACLTokenInfo, agentTokenID string,
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
		return fmt.Errorf("failed to marshal SetAgentToken JSON string payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, setAgentTokenURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to prepare SetAgentToken request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, bootstrapACLToken.SecretID)
	req.Header.Add(common.ContentType, common.ContentTypeJSON)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SetAgentToken request for http URL: %w", err)
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
			return fmt.Errorf("failed to read SetAgentToken response body with status code=%d: %w", resp.StatusCode, err)
		}

		return fmt.Errorf("failed to set agent token with type %s via URL [%s] and status code= %d: %s",
			tokenType, setAgentTokenURL, resp.StatusCode, string(setAgentTokenResp))
	}

	return nil
}

// createNewToken creates a new token based on the provided inputs
// it returns ACLTokenInfo object and thus can be written to the file later
func (c *cmd) createNewToken(bootstrapACLTokenID string, createToken CreateRegistryToken) (*ACLTokenInfo, error) {
	if len(bootstrapACLTokenID) == 0 {
		return nil, fmt.Errorf("bootstrap token ID cannot be empty")
	}

	createTokenURL, err := c.getRegistryApiUrl(consulCreateTokenAPI)
	if err != nil {
		return nil, err
	}

	jsonPayload, err := json.Marshal(&createToken)
	c.loggingClient.Tracef("payload: %v", createToken)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CreatRegistryToken JSON string payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, createTokenURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to prepare creat a new token request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, bootstrapACLTokenID)
	req.Header.Add(common.ContentType, common.ContentTypeJSON)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send create a new token request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	createTokenResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read create a new token response body: %w", err)
	}

	var aclTokenInfo ACLTokenInfo
	switch resp.StatusCode {
	case http.StatusOK:
		err := json.Unmarshal(createTokenResp, &aclTokenInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal new token response body: %w", err)
		}
		c.loggingClient.Info("successfully created a new registry token")
		return &aclTokenInfo, nil
	default:
		return nil, fmt.Errorf("failed to create a new token via URL [%s] and status code= %d: %s",
			consulCreateTokenAPI, resp.StatusCode, string(createTokenResp))
	}
}

// insertNewManagementToken creates a new Consul token
// it returns the ACLTokenInfo and error if any error occurs
func (c *cmd) insertNewManagementToken(bootstrapACLToken types.BootStrapACLTokenInfo) (*ACLTokenInfo, error) {
	// get a policy for this consul token to associate with
	edgexManagementPolicy, err := c.getOrCreateRegistryPolicy(bootstrapACLToken.SecretID,
		edgeXManagementPolicyName,
		edgeXManagementPolicyRules)
	if err != nil {
		return nil, fmt.Errorf("failed to create edgex management policy: %v", err)
	}

	unlimitedDuration := "0s"
	createToken := NewCreateRegistryToken("edgex-core-consul management token",
		[]types.Policy{
			*edgexManagementPolicy,
		}, true, &unlimitedDuration)
	newTokenInfo, err := c.createNewToken(bootstrapACLToken.SecretID, createToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create new edgex management token: %v", err)
	}

	c.loggingClient.Info("successfully created a new management token")

	return newTokenInfo, nil
}

// createManagementToken uses bootstrapped ACL token to create an manatement token
// this call requires ACL read/write permission and hence we use the bootstrap ACL token
// it checks whether there is an management token already existing and re-uses it if so
// otherwise creates a new management token
func (c *cmd) createManagementToken(bootstrapACLToken types.BootStrapACLTokenInfo) (*ManagementACLTokenInfo, error) {
	if len(bootstrapACLToken.SecretID) == 0 {
		return nil, errors.New("bootstrap ACL token is required for creating management token")
	}

	// list tokens and search for the "edgex-core-consul" management token if any
	pattern := regexp.MustCompile(`(?i)edgex([0-9a-z\- ]+)management token`)
	aclTokenInfo, err := c.getEdgeXTokenByPattern(bootstrapACLToken, pattern)

	if err != nil {
		c.loggingClient.Errorf("failed to retrieve EdgeX management token from pattern %s, error %s", pattern, err.Error())
		return nil, err
	}

	if reflect.DeepEqual(*aclTokenInfo, ACLTokenInfo{}) {
		// need to create a new agent token as there is no matched one found
		aclTokenInfo, err = c.insertNewManagementToken(bootstrapACLToken)
		if err != nil {
			return nil, fmt.Errorf("failed to insert a new EdgeX management token: %v", err)
		}
	}

	managementACLTokenInfo := ManagementACLTokenInfo{
		SecretID: aclTokenInfo.SecretID,
		Policies: aclTokenInfo.Policies,
	}
	return &managementACLTokenInfo, nil
}

// saveManagementACLToken will save the token information to file
// parameters: tokenInfoToBeSaved for managementACLToken
func (c *cmd) saveManagementACLToken(tokenInfoToBeSaved *ManagementACLTokenInfo) error {
	// Write the token to the specified file
	tokenFileAbsPath, err := filepath.Abs(c.configuration.StageGate.Registry.ACL.ManagementTokenPath)
	if err != nil {
		return fmt.Errorf("failed to convert tokenFile to absolute path %s: %s",
			c.configuration.StageGate.Registry.ACL.ManagementTokenPath, err.Error())
	}

	// create the directory of tokenfile if not exists yet
	dirOfToken := filepath.Dir(tokenFileAbsPath)
	fileIoPerformer := fileioperformer.NewDefaultFileIoPerformer()
	if err := fileIoPerformer.MkdirAll(dirOfToken, 0700); err != nil {
		return fmt.Errorf("failed to create tokenpath base dir: %s", err.Error())
	}

	fileWriter, err := fileIoPerformer.OpenFileWriter(tokenFileAbsPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file writer %s: %s", tokenFileAbsPath, err.Error())
	}

	if err := json.NewEncoder(fileWriter).Encode(tokenInfoToBeSaved); err != nil {
		_ = fileWriter.Close()
		return fmt.Errorf("failed to write management token: %s", err.Error())
	}

	if err := fileWriter.Close(); err != nil {
		return fmt.Errorf("failed to close token file: %s", err.Error())
	}

	c.loggingClient.Infof("management token is written to %s", tokenFileAbsPath)

	return nil
}
