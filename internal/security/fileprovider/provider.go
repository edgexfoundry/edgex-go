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

package fileprovider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/pkg/token/fileioperformer"
)

// permissionable is the subset of the File API that allows setting file permissions
type permissionable interface {
	Chown(uid int, gid int) error
	Chmod(mode os.FileMode) error
}

// fileTokenProvider stores instance data
type fileTokenProvider struct {
	logger        logger.LoggingClient
	fileOpener    fileioperformer.FileIoPerformer
	tokenProvider authtokenloader.AuthTokenLoader
	vaultClient   secretstoreclient.SecretStoreClient
	secretConfig  secretstoreclient.SecretServiceInfo
	tokenConfig   config.TokenFileProviderInfo
}

// NewTokenProvider creates a new TokenProvider
func NewTokenProvider(logger logger.LoggingClient,
	fileOpener fileioperformer.FileIoPerformer,
	tokenProvider authtokenloader.AuthTokenLoader,
	vaultClient secretstoreclient.SecretStoreClient) TokenProvider {
	return &fileTokenProvider{
		logger:        logger,
		fileOpener:    fileOpener,
		tokenProvider: tokenProvider,
		vaultClient:   vaultClient,
	}
}

// Set configuration
func (p *fileTokenProvider) SetConfiguration(secretConfig secretstoreclient.SecretServiceInfo, tokenConfig config.TokenFileProviderInfo) {
	p.secretConfig = secretConfig
	p.tokenConfig = tokenConfig
}

// Do whatever is needed
func (p *fileTokenProvider) Run() error {
	p.logger.Info("Generating Vault tokens")

	privilegedToken, err := p.tokenProvider.Load(p.tokenConfig.PrivilegedTokenPath)
	if err != nil {
		p.logger.Error(fmt.Sprintf("failed to read privileged access token: %s", err.Error()))
		return err
	}

	tokenConfEnv, err := GetTokenConfigFromEnv()
	if err != nil {
		p.logger.Error(fmt.Sprintf("failed to get token config from environment variable %s with error: %s", addSecretstoreTokensEnvKey, err.Error()))
		return err
	}

	var tokenConf TokenConfFile
	if err := LoadTokenConfig(p.fileOpener, p.tokenConfig.ConfigFile, &tokenConf); err != nil {
		p.logger.Error(fmt.Sprintf("failed to read token configuration file %s: %s", p.tokenConfig.ConfigFile, err.Error()))
		return err
	}

	// merge the additional token configuration list from environment variable
	// note that the configuration file takes precedence, as the tokenConf will override
	// the tokenConfEnv with same duplicate keys
	// The tokenConfEnv only uses default settings.
	tokenConf = tokenConfEnv.mergeWith(tokenConf)

	for serviceName, serviceConfig := range tokenConf {
		p.logger.Info(fmt.Sprintf("generating policy/token defaults for service %s", serviceName))

		servicePolicy := make(map[string]interface{})
		createTokenParameters := make(map[string]interface{})

		if serviceConfig.UseDefaults {
			p.logger.Info(fmt.Sprintf("using policy/token defaults for service %s", serviceName))
			servicePolicy = makeDefaultTokenPolicy(serviceName)
			createTokenParameters = makeDefaultTokenParameters(serviceName)
		}

		if serviceConfig.CustomPolicy != nil {
			customPolicy := serviceConfig.CustomPolicy
			if customPolicy["path"] != nil {
				customPaths := customPolicy["path"].(map[string]interface{})
				if servicePolicy["path"] == nil {
					servicePolicy["path"] = make(map[string]interface{})
				}
				for k, v := range customPaths {
					(servicePolicy["path"]).(map[string]interface{})[k] = v
				}
			}
		}

		if serviceConfig.CustomTokenParameters != nil {
			// Custom token parameters override the defaults
			createTokenParameters = mergeMaps(createTokenParameters, serviceConfig.CustomTokenParameters)
		}

		// Set a meta property that consuming serices can use to automatically scope secret queries
		createTokenParameters["meta"] = map[string]interface{}{
			"edgex-service-name": serviceName,
		}

		// Always create a policy with this name
		policyName := "edgex-service-" + serviceName

		policyBytes, err := json.Marshal(servicePolicy)
		if err != nil {
			p.logger.Error(fmt.Sprintf("failed encode service policy for %s: %s", serviceName, err.Error()))
			return err
		}

		if _, err := p.vaultClient.InstallPolicy(privilegedToken, policyName, string(policyBytes)); err != nil {
			p.logger.Error(fmt.Sprintf("failed to install policy %s: %s", policyName, err.Error()))
			return err
		}

		var createTokenResponse interface{}

		if _, err = p.vaultClient.CreateToken(privilegedToken, createTokenParameters, &createTokenResponse); err != nil {
			p.logger.Error(fmt.Sprintf("failed to create vault token for service %s: %s", serviceName, err.Error()))
			return err
		}

		outputTokenDir := filepath.Join(p.tokenConfig.OutputDir, serviceName)
		outputTokenFilename := filepath.Join(outputTokenDir, p.tokenConfig.OutputFilename)
		if err := p.fileOpener.MkdirAll(outputTokenDir, os.FileMode(0700)); err != nil {
			p.logger.Error(fmt.Sprintf("failed to create base directory path(s) %s: %s", outputTokenDir, err.Error()))
			return err
		}

		p.logger.Info(fmt.Sprintf("creating token file %s", outputTokenFilename))
		writeCloser, err := p.fileOpener.OpenFileWriter(outputTokenFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600))
		if err != nil {
			p.logger.Error(fmt.Sprintf("failed open token file for writing %s: %s", outputTokenFilename, err.Error()))
			return err
		}
		// writeCloser is writable file -- explicitly close() to ensure we catch errors writing to it

		permissionable, ok := writeCloser.(permissionable)
		if ok {
			if serviceConfig.FilePermissions != nil &&
				(serviceConfig.FilePermissions).ModeOctal != nil {
				mode, err := strconv.ParseInt(*(serviceConfig.FilePermissions).ModeOctal, 8, 32)
				if err != nil {
					_ = writeCloser.Close()
					p.logger.Error(fmt.Sprintf("invalid file mode %s: %s", *(serviceConfig.FilePermissions).ModeOctal, err.Error()))
					return err
				}
				if err := permissionable.Chmod(os.FileMode(mode)); err != nil {
					_ = writeCloser.Close()
					p.logger.Error(fmt.Sprintf("failed to set file mode on %s: %s", outputTokenFilename, err.Error()))
					return err
				}
			}
			if serviceConfig.FilePermissions != nil &&
				(serviceConfig.FilePermissions).Uid != nil &&
				(serviceConfig.FilePermissions).Gid != nil {
				err := permissionable.Chown(*(serviceConfig.FilePermissions).Uid, *(serviceConfig.FilePermissions).Gid)
				if err != nil {
					_ = writeCloser.Close()
					p.logger.Error(fmt.Sprintf("failed to set file user/group on %s: %s", outputTokenFilename, err.Error()))
					return err
				}
			}
		}

		encoder := json.NewEncoder(writeCloser)
		if encoder == nil {
			_ = writeCloser.Close()
			err = fmt.Errorf("unable to create JSON output encoder")
			return err
		}

		// Write resulting token
		if err := encoder.Encode(createTokenResponse); err != nil {
			_ = writeCloser.Close()
			p.logger.Error(fmt.Sprintf("failed to write token file: %s", err.Error()))
			return err
		}

		if err := writeCloser.Close(); err != nil {
			p.logger.Error(fmt.Sprintf("failed to close %s: %s", outputTokenFilename, err.Error()))
			return err
		}
	}

	return nil
}
