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
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/setupacl/share"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/helper"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

const (
	// the command name for setting up registry's ACL
	CommandName string = "setupRegistryACL"

	consulGetLeaderAPI         = "/v1/status/leader"
	consulACLBootstrapAPI      = "/v1/acl/bootstrap"
	consulConfigAccessVaultAPI = "/v1/consul/config/access"
	consulLegacyACLModeError   = "The ACL system is currently in legacy mode"
	defaultRetryTimeout        = 30 * time.Second
	emptyLeader                = `""`

	// environment variable contains a comma separated list of registry role names to be added
	addRegistryRolesEnvKey = "ADD_REGISTRY_ACL_ROLES"
)

type cmd struct {
	loggingClient logger.LoggingClient
	client        internal.HttpCaller
	configuration *config.ConfigurationStruct

	// internal state
	retryTimeout           time.Duration
	bootstrapACLTokenCache *BootStrapACLTokenInfo
	secretstoreTokenCache  string
}

// NewCommand creates a new cmd and parses through options if any
func NewCommand(
	_ context.Context,
	_ *sync.WaitGroup,
	lc logger.LoggingClient,
	conf *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {
	cmd := cmd{
		loggingClient: lc,
		client:        pkg.NewRequester(lc).Insecure(),
		configuration: conf,
		retryTimeout:  defaultRetryTimeout,
	}
	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "confdir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}

	return &cmd, nil
}

// Execute implements Command and runs this command
// command setupRegistryACL sets up the ACL system of the registry, Consul in this case, bootstrap ACL system,
// configure Consul access for the secret store, create agent token, and set up the agent token to agent
func (c *cmd) Execute() (statusCode int, err error) {
	c.loggingClient.Infof("Security bootstrapper running %s", CommandName)

	// need to have a sentinel file to guard against the re-run of the command once we have successfully bootstrap ACL
	// if we already have a sentinelFile exists then skip this whole process since we already done this
	// process successfully before, otherwise Consul's ACL bootstrap will cause a panic
	sentinelFileAbsPath, err := filepath.Abs(c.configuration.StageGate.Registry.ACL.SentinelFilePath)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to get the absolute path of the sentinel file: %v", err)
	}

	// a non-empty leader is a prerequisite for any agent related API operations
	if err := c.waitForNonEmptyConsulLeader(); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to wait for Consul leader: %v", err)
	}

	if helper.CheckIfFileExists(sentinelFileAbsPath) {
		// run through any needed to be re-set up on every restart of this call
		if err := c.reSetup(); err != nil {
			return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to re-setup registry ACL: %v", err)
		}

		c.loggingClient.Info("setupRegistryACL successfully done")

		return
	}

	bootstrapACLToken, err := c.createBootstrapACLToken()
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to create bootstrap ACL token: %v", err)
	}

	// retrieve the secretstore (Vault) token from the file produced by secretstore-setup
	secretstoreToken, err := c.retrieveSecretStoreTokenFromFile()
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to retrieve secretstore token: %v", err)
	}

	// configure Consul access with both Secret Store token and consul's bootstrap acl token
	if err := c.configureConsulAccess(secretstoreToken, bootstrapACLToken.SecretID); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to configure Consul access: %v", err)
	}

	c.loggingClient.Info("successfully get secretstore token and configuring the registry access for secretestore")

	if err := c.createEdgeXACLTokenRoles(bootstrapACLToken.SecretID, secretstoreToken); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to createEdgeXACLTokenRoles: %v", err)
	}

	// set up agent token to agent for the first time
	if err := c.setupAgentToken(bootstrapACLToken); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to set up agent token: %v", err)
	}

	if err := c.saveACLTokens(bootstrapACLToken); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to save ACL tokens: %v", err)
	}

	// write a sentinel file to indicate Consul ACL bootstrap is done so that we don't bootstrap ACL again,
	// this is to avoid re-bootstrapping error and that error can cause the snap crash if restart this process
	if err := c.writeSentinelFile(); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to write sentinel file: %v", err)
	}

	c.loggingClient.Info("setupRegistryACL successfully done")

	return
}

// GetCommandName returns the name of this command
func (c *cmd) GetCommandName() string {
	return CommandName
}

// reSetup calls when anything is running 2nd time or later, in order to re-set up the registry ACL
func (c *cmd) reSetup() error {
	// although we may have done setup ACL successfully already previous times,
	// if token persistence is not enabled, we have to re-set agent token every time agent is restarted,
	// i.e., when this subcommand is called every time, regardless whether it is first time or not
	if err := c.setupAgentToken(nil); err != nil {
		return fmt.Errorf("on 2nd time or later, failed to re-set up agent token: %v", err)
	}

	// set up roles for both static and dynamic again in case there're changes
	if err := c.reSetupEdgeXACLTokenRoles(); err != nil {
		return fmt.Errorf("on 2nd time or later, failed to re-set up roles: %v", err)
	}

	return nil
}

func (c *cmd) reSetupEdgeXACLTokenRoles() error {
	bootstrapACLToken, err := c.reconstructBootstrapACLToken()
	if err != nil {
		return fmt.Errorf("failed to reconstruct bootstrap ACL token: %v", err)
	} else if len(bootstrapACLToken.SecretID) == 0 {
		return errors.New("bootstrapACLToken.SecretID is empty")
	}

	secretstoreToken, err := c.getSecretStoreToken()
	if err != nil {
		return fmt.Errorf("failed to retrieve secretstore token: %v", err)
	}

	if err := c.createEdgeXACLTokenRoles(bootstrapACLToken.SecretID, secretstoreToken); err != nil {
		return fmt.Errorf("failed to create EdgeX roles: %v", err)
	}

	return nil
}

func (c *cmd) createBootstrapACLToken() (*BootStrapACLTokenInfo, error) {
	bootstrapACLToken, err := c.generateBootStrapACLToken()
	if err != nil {
		// although we have a leader, but it is a very very rare chance that we could hit an error on legacy mode
		// here we will sleep a bit of time and then retry once if there is error on Legacy ACL type of message
		// because Consul is still on its way to initialize the new ACL system internally
		// for the details of this issue, see related issue on Consul's Github website:
		// https://github.com/hashicorp/consul/issues/5218#issuecomment-457212336
		if !strings.Contains(err.Error(), consulLegacyACLModeError) {
			// other type of ACL bootstrapping error, cannot continue
			return nil, fmt.Errorf("failed to bootstrap registry's ACL: %v", err)
		}

		c.loggingClient.Warnf("found Consul still in ACL legacy mode, will retry once again: %v", err)
		time.Sleep(5 * time.Second)
		bootstrapACLToken, err = c.generateBootStrapACLToken()
		if err != nil {
			return nil, fmt.Errorf("failed to bootstrap registry's ACL: %v", err)
		}
	}

	c.loggingClient.Info("successfully bootstrap registry ACL")

	return bootstrapACLToken, nil
}

func (c *cmd) saveACLTokens(bootstrapACLToken *BootStrapACLTokenInfo) error {
	// Save the bootstrap ACL token into json file so that it can be used later on
	if err := c.saveBootstrapACLToken(bootstrapACLToken); err != nil {
		return fmt.Errorf("failed to save registry's bootstrap ACL token: %v", err)
	}

	return nil
}

// createEdgeXACLTokenRoles creates secret store roles that can be used for genearting registry tokens
// via Consul secret engine API /consul/creds/[role_name] later on for all EdgeX microservices
func (c *cmd) createEdgeXACLTokenRoles(bootstrapACLTokenID, secretstoreToken string) error {
	edgexServicePolicy, err := c.getOrCreateRegistryPolicy(bootstrapACLTokenID, edgeXServicePolicyName, edgeXPolicyRules)
	if err != nil {
		return fmt.Errorf("failed to create edgex service policy: %v", err)
	}

	roleNames, err := c.getUniqueRoleNames()
	if err != nil {
		return fmt.Errorf("failed to get unique role names: %v", err)
	}

	// create registry roles for EdgeX
	for roleName := range roleNames {
		// create roles based on the service keys as the role names
		// in phase 2, we are using the same policy rule for all services
		edgexACLTokenRole := NewRegistryRole(roleName, ClientType, []Policy{
			*edgexServicePolicy,
			// localUse set to false as some EdgeX services may be running in a different node
		}, false)

		// fail all if any one of the role creation failed
		if err := c.createRole(secretstoreToken, edgexACLTokenRole); err != nil {
			return fmt.Errorf("failed to create edgex role: %v", err)
		}
	}

	return nil
}

func (c *cmd) getUniqueRoleNames() (map[string]struct{}, error) {
	roleNamesFromConfig := c.configuration.StageGate.Registry.ACL.GetACLRoleNames()
	if len(roleNamesFromConfig) == 0 {
		return nil, errors.New("no ACL role names defined in configuration")
	}

	// there may be some role names to be added, and we want to add only unique role names
	// use empty struct{} as value so that there is no really memory allocated for values
	rolesFromConfig := make(map[string]struct{})
	for _, roleName := range roleNamesFromConfig {
		rolesFromConfig[roleName] = struct{}{}
	}

	rolesFromEnv, err := getUniqueRolesFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get unique roles from env: %v", err)
	}

	// merge roles from config and from env
	uniqueRoleNames := c.mergeUniqueRoles(rolesFromConfig, rolesFromEnv)

	c.loggingClient.Infof("successfully got unique role names, total size=%d", len(uniqueRoleNames))

	return uniqueRoleNames, nil
}

func (c *cmd) mergeUniqueRoles(configRoles, envRoles map[string]struct{}) map[string]struct{} {
	if len(envRoles) == 0 {
		return configRoles
	}

	uniqueRoleNames := make(map[string]struct{})
	for key, val := range configRoles {
		uniqueRoleNames[key] = val
	}

	c.loggingClient.Infof("Adding role names from environment variable %s", addRegistryRolesEnvKey)
	for roleName, val := range envRoles {
		if _, exists := uniqueRoleNames[roleName]; exists {
			c.loggingClient.Warnf("the service key %s from env already exists in config roles, skip", roleName)
			continue
		}
		uniqueRoleNames[roleName] = val
	}

	return uniqueRoleNames
}

func getUniqueRolesFromEnv() (map[string]struct{}, error) {
	uniqueRolesEnv := make(map[string]struct{})
	// read a list of service keys as role names from environment variable if any
	addRoleList := os.Getenv(addRegistryRolesEnvKey)
	if len(strings.TrimSpace(addRoleList)) == 0 {
		return uniqueRolesEnv, nil
	}

	// the list of service keys is comma-separated
	serviceKeyList := strings.Split(addRoleList, ",")

	// regex for valid service key as role name
	// the service key eventually becomes part of the URL to Vault's create/read registry role APIs call
	// according to the specs, the registry role name can only contain the followings:
	// alphanumeric characters, dashes -, and underscores _
	// also role name must be unique
	// we also limit the the length of the name up to 512 characters
	roleNameValidateRegx := regexp.MustCompile(`^[\w\-\_]{1,512}$`)

	// do name validation before add it to the final list
	for _, serviceKey := range serviceKeyList {
		roleName := strings.ToLower(strings.TrimSpace(serviceKey))
		if len(roleName) == 0 {
			// skipping the empty cases, ie. treating it as no role
			continue
		}

		if !roleNameValidateRegx.MatchString(roleName) {
			return nil, fmt.Errorf("invalid service key as registry role name %s from env %s", roleName, addRegistryRolesEnvKey)
		}

		uniqueRolesEnv[roleName] = struct{}{}
	}

	return uniqueRolesEnv, nil
}

// setupAgentToken is to set up the agent token using the inputToken to the running agent if haven't set up yet
// if the inputToken is nil then it will try to reconstruct from the saved file
func (c *cmd) setupAgentToken(inputToken *BootStrapACLTokenInfo) error {
	var err error
	setupAlreadyPrevious := false
	bootstrapACLToken := inputToken
	if inputToken == nil {
		// this may be the case that re-run with a different configuration like token persistence is changed
		// reconstruct the bootstrapACLToken from the file
		bootstrapACLToken, err = c.reconstructBootstrapACLToken()
		if err != nil {
			return fmt.Errorf("failed to reconstruct bootstrap ACL token: %v", err)
		}
		setupAlreadyPrevious = true
	}

	persistent, err := c.isACLTokenPersistent(bootstrapACLToken.SecretID)
	if err != nil {
		return fmt.Errorf("failed to check the agent token persistence: %v", err)
	}

	// if property token persistence is not enabled, we have to re-set agent token every time agent is restarted
	// i.e., when this subcommand is called, regardless whether it is first time or not
	// furthermore, we also need to set agent token if it is the first time to set up the registry ACL
	if !persistent || !setupAlreadyPrevious {
		agentToken, err := c.createAgentToken(*bootstrapACLToken)
		if err != nil {
			return fmt.Errorf("setupAgentToken failed: %v", err)
		}
		err = c.setAgentToken(*bootstrapACLToken, agentToken, AgentType)
		if err != nil {
			return fmt.Errorf("failed to set agent token: %v", err)
		}

		c.loggingClient.Info("successfully set up agent token into agent")
	} else {
		// we had already done all necessary setups
		c.loggingClient.Info("setupAgentToken had been done before and agent token is persistent, skip")
	}

	return nil
}

// reconstructBootstrapACLToken reads bootstrap ACL token from the saved file and reconstruct it into BootStrapACLTokenInfo
func (c *cmd) reconstructBootstrapACLToken() (*BootStrapACLTokenInfo, error) {
	if c.bootstrapACLTokenCache != nil {
		// re-use the cached one
		return c.bootstrapACLTokenCache, nil
	}

	bootstrapTokenFilePath := strings.TrimSpace(c.configuration.StageGate.Registry.ACL.BootstrapTokenPath)
	if len(bootstrapTokenFilePath) == 0 {
		return nil, errors.New("required StageGate_Registry_ACL_BootstrapTokenPath from configuration is empty")
	}

	tokenFileAbsPath, err := filepath.Abs(bootstrapTokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tokenFile to absolute path %s: %v", bootstrapTokenFilePath, err)
	}

	// make sure we have the file
	if exists := helper.CheckIfFileExists(tokenFileAbsPath); !exists {
		return nil, fmt.Errorf("registry bootstrap ACL token file %s not found", tokenFileAbsPath)
	}

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenReader, err := fileOpener.OpenFileReader(tokenFileAbsPath, os.O_RDONLY, 0400)
	if err != nil {
		return nil, fmt.Errorf("failed to open file reader: %v", err)
	}

	var bootstrapACLToken BootStrapACLTokenInfo
	if err := json.NewDecoder(tokenReader).Decode(&bootstrapACLToken); err != nil {
		return nil, fmt.Errorf("failed to parse token data into BootStrapACLTokenInfo: %v", err)
	}

	// cache it for later use
	c.bootstrapACLTokenCache = &bootstrapACLToken

	c.loggingClient.Infof("successfully reconstructed bootstrap ACL token from %s", bootstrapTokenFilePath)

	return &bootstrapACLToken, nil
}

func (c *cmd) getRegistryApiUrl(api string) (string, error) {
	apiURL := fmt.Sprintf("%s://%s:%d%s", c.configuration.StageGate.Registry.ACL.Protocol,
		c.configuration.StageGate.Registry.Host, c.configuration.StageGate.Registry.Port, api)
	_, err := url.Parse(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse API URL: %v", err)
	}
	return apiURL, nil
}

// waitForNonEmptyConsulLeader is a special waitFor function on waiting for "non-empty" leader being available
// the ordinary http waitFor won't work as the returned http status code from API call is 200 even when Consul's leader
// is an empty string ("") but we need an non-empty leader; so 200 doesn't mean we have a leader
func (c *cmd) waitForNonEmptyConsulLeader() error {
	timeoutInSec := int(c.retryTimeout.Seconds())
	timer := startup.NewTimer(timeoutInSec, 1)
	for timer.HasNotElapsed() {
		if err := c.getNonEmptyConsulLeader(); err != nil {
			c.loggingClient.Warnf("error from getting Consul leader API call, will retry it again: %v", err)
			timer.SleepForInterval()
			continue
		}
		c.loggingClient.Info("found Consul leader to set up ACL")
		return nil
	}

	return errors.New("timed out to get non-empty Consul leader")
}

// getNonEmptyConsulLeader makes http request call to get the registry Consul leader
// the response of getting leader call could be an empty leader (represented by "")
// even if the http status code is 200 when Consul is just booting up and
// it will take a bit of time to elect the raft leader
func (c *cmd) getNonEmptyConsulLeader() error {
	getLeaderURL, err := c.getRegistryApiUrl(consulGetLeaderAPI)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, getLeaderURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("Failed to prepare request for http URL: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response body to get leader: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		trimmedResp := strings.TrimSpace(string(responseBody))
		// Consul's raft leader election process could take a bit of time
		// before responds back with a non-empty leader
		if len(trimmedResp) == 0 || emptyLeader == trimmedResp {
			return errors.New("no leader yet")
		}
		// now we have a cluster raft leader
		c.loggingClient.Infof("leader [%s] is elected", trimmedResp)
		return nil

	// almost unlikely for this case unless URL is incorrect
	default:
		return fmt.Errorf("get Consul leader request failed with status code= %d: %s", resp.StatusCode, string(responseBody))
	}
}

func (c *cmd) getSecretStoreToken() (string, error) {
	if len(c.secretstoreTokenCache) > 0 {
		return c.secretstoreTokenCache, nil
	}

	if secretStoreToken, err := c.retrieveSecretStoreTokenFromFile(); err == nil {
		c.secretstoreTokenCache = secretStoreToken
		return secretStoreToken, nil
	} else {
		return share.EmptyToken, err
	}
}

func (c *cmd) retrieveSecretStoreTokenFromFile() (string, error) {
	trimmedFilePath := strings.TrimSpace(c.configuration.StageGate.Registry.ACL.SecretsAdminTokenPath)
	if len(trimmedFilePath) == 0 {
		return share.EmptyToken, errors.New("required StageGate_Registry_ACL_SecretsAdminTokenPath from configuration is empty")
	}

	tokenFileAbsPath, err := filepath.Abs(trimmedFilePath)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("failed to convert tokenFile to absolute path %s: %v", trimmedFilePath, err)
	}

	// since the secretstore token is created by another service, secretstore-setup,
	// so here we want to make sure we have the file
	if exists := helper.CheckIfFileExists(tokenFileAbsPath); !exists {
		return share.EmptyToken, fmt.Errorf("secretstore token file %s not found", tokenFileAbsPath)
	}

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenLoader := authtokenloader.NewAuthTokenLoader(fileOpener)
	secretStoreToken, err := tokenLoader.Load(tokenFileAbsPath)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("tokenLoader failed to load secretstore token: %v", err)
	}

	c.loggingClient.Infof("successfully retrieved secretstore management token from %s", trimmedFilePath)

	return secretStoreToken, nil
}

// configureConsulAccess is to enable the Consul config access to the SecretStore via consul/config/access API
// see the reference: https://www.vaultproject.io/api-docs/secret/consul#configure-access
func (c *cmd) configureConsulAccess(secretStoreToken string, bootstrapACLToken string) error {
	configAccessURL := fmt.Sprintf("%s://%s:%d%s", c.configuration.SecretStore.Protocol,
		c.configuration.SecretStore.Host, c.configuration.SecretStore.Port, consulConfigAccessVaultAPI)
	_, err := url.Parse(configAccessURL)
	if err != nil {
		return fmt.Errorf("failed to parse config Access URL: %v", err)
	}

	c.loggingClient.Debugf("configAccessURL: %s", configAccessURL)

	type ConfigAccess struct {
		RegistryAddress   string `json:"address"`
		BootstrapACLToken string `json:"token"`
	}

	payload := &ConfigAccess{
		RegistryAddress:   fmt.Sprintf("%s:%d", c.configuration.StageGate.Registry.Host, c.configuration.StageGate.Registry.Port),
		BootstrapACLToken: bootstrapACLToken,
	}

	jsonPayload, err := json.Marshal(payload)
	c.loggingClient.Tracef("payload: %v", payload)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON string payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, configAccessURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("Failed to prepare POST request for http URL: %w", err)
	}

	req.Header.Add("X-Vault-Token", secretStoreToken)
	req.Header.Add(common.ContentType, common.ContentTypeJSON)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusNoContent:
		// no response body returned in this case
		c.loggingClient.Info("successfully configure Consul access for secretstore")
		return nil
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.loggingClient.Errorf("cannot read resp.Body: %v", err)
		}
		return fmt.Errorf("failed to configure Consul access for secretstore via URL [%s] and status code= %d: %s",
			configAccessURL, resp.StatusCode, string(body))
	}
}

func (c *cmd) writeSentinelFile() error {
	absPath, err := filepath.Abs(c.configuration.StageGate.Registry.ACL.SentinelFilePath)
	if err != nil {
		return fmt.Errorf("failed to get the absolute path of the sentinel file: %v", err)
	}

	dirToWrite := filepath.Dir(absPath)
	filePerformer := fileioperformer.NewDefaultFileIoPerformer()
	if err := filePerformer.MkdirAll(dirToWrite, 0700); err != nil {
		return fmt.Errorf("failed to create sentinel base dir: %s", err.Error())
	}

	writer, err := filePerformer.OpenFileWriter(absPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file writer %s: %s", absPath, err.Error())
	}
	defer func() { _ = writer.Close() }()

	if _, err := writer.Write([]byte("done")); err != nil {
		return fmt.Errorf("failed to write out to sentinel file %s: %v", absPath, err)
	}

	return nil
}
