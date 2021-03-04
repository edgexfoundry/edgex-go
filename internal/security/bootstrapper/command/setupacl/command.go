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
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/helper"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
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
	emptyToken                 = ""
)

type cmd struct {
	loggingClient logger.LoggingClient
	client        internal.HttpCaller
	configuration *config.ConfigurationStruct

	// internal state
	retryTimeout time.Duration
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
// command setupRegistryACL sets up the ACL system of the registry, Consul in this case, preparing for generating
// Consul's agent tokens later on
func (c *cmd) Execute() (statusCode int, err error) {
	c.loggingClient.Infof("Security bootstrapper running %s", CommandName)

	// need to have a sentinel file to guard against the re-run of the command once we have successfully bootstrap ACL
	// if we already have a sentinelFile exists then skip this whole process since we already done this
	// process successfully before, otherwise Consul's ACL bootstrap will cause a panic
	sentinelFileAbsPath, err := filepath.Abs(c.configuration.StageGate.Registry.ACL.SentinelFilePath)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to get the absolute path of the sentinel file: %v", err)
	}

	if helper.CheckIfFileExists(sentinelFileAbsPath) {
		c.loggingClient.Info("Registry ACL had been setup successfully already, skip")
		return
	}

	if err := c.waitForNonEmptyConsulLeader(); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to wait for Consul leader: %v", err)
	}

	var bootstrapACLToken *BootStrapACLTokenInfo
	bootstrapACLToken, err = c.generateBootStrapACLToken()
	if err != nil {
		// although we have a leader, but it is a very very rare chance that we could hit an error on legacy mode
		// here we will sleep a bit of time and then retry once if there is error on Legacy ACL type of message
		// because Consul is still on its way to initialize the new ACL system internally
		// for the details of this issue, see related issue on Consul's Github website:
		// https://github.com/hashicorp/consul/issues/5218#issuecomment-457212336
		if !strings.Contains(err.Error(), consulLegacyACLModeError) {
			// other type of ACL bootstrapping error, cannot continue
			return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to bootstrap registry's ACL: %v", err)
		}

		c.loggingClient.Warnf("found Consul still in ACL legacy mode, will retry once again: %v", err)
		time.Sleep(5 * time.Second)
		bootstrapACLToken, err = c.generateBootStrapACLToken()
		if err != nil {
			return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to bootstrap registry's ACL: %v", err)
		}
	}

	c.loggingClient.Info("successfully bootstrap registry ACL")

	// Save the bootstrap token into the file so that it can be used later on
	if err := c.saveBootstrapACLToken(bootstrapACLToken); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to save registry's bootstrap ACL token: %v", err)
	}

	// retrieve the secretstore (Vault) token from the file produced by secretstore-setup
	secretstoreToken, err := c.getSecretStoreTokenFromFile()
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to retrieve secretstore token: %v", err)
	}

	c.loggingClient.Info("successfully get secretstore token and configuring the registry access for secretestore")

	// configure Consul access with both Secret Store token and consul's bootstrap acl token
	if err := c.configureConsulAccess(secretstoreToken, bootstrapACLToken.SecretID); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to configure Consul access: %v", err)
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
		c.loggingClient.Info("found Consul leader to bootstrap ACL")
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

	responseBody, err := ioutil.ReadAll(resp.Body)
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

func (c *cmd) getSecretStoreTokenFromFile() (string, error) {
	trimmedFilePath := strings.TrimSpace(c.configuration.StageGate.Registry.ACL.SecretsAdminTokenPath)
	if len(trimmedFilePath) == 0 {
		return emptyToken, errors.New("required StageGate_Registry_SecretsAdminTokenPath from configuration is empty")
	}

	tokenFileAbsPath, err := filepath.Abs(trimmedFilePath)
	if err != nil {
		return emptyToken, fmt.Errorf("failed to convert tokenFile to absolute path %s: %v", trimmedFilePath, err)
	}

	// since the secretstore token is created by another service, secretstore-setup,
	// so here we want to make sure we have the file
	if exists := helper.CheckIfFileExists(tokenFileAbsPath); !exists {
		return emptyToken, fmt.Errorf("secretstore token file %s not found", tokenFileAbsPath)
	}

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenLoader := authtokenloader.NewAuthTokenLoader(fileOpener)
	secretStoreToken, err := tokenLoader.Load(tokenFileAbsPath)
	if err != nil {
		return emptyToken, fmt.Errorf("tokenLoader failed to load secretstore token: %v", err)
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
	req.Header.Add(clients.ContentType, clients.ContentTypeJSON)
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
		body, err := ioutil.ReadAll(resp.Body)
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
