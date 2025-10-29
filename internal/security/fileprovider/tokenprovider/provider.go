//
// Copyright (c) 2019-2023 Intel Corporation
// Copyright (c) 2025 IOTech Ltd
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
// SPDX-License-Identifier: Apache-2.0
//

package tokenprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/security/common"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore"
	secretstoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/go-mod-secrets/v4/secrets"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer"
)

// permissionable is the subset of the File API that allows setting file permissions
type permissionable interface {
	Chown(uid int, gid int) error
	Chmod(mode os.FileMode) error
}

// fileTokenProvider stores instance data
type fileTokenProvider struct {
	logger            logger.LoggingClient
	fileOpener        fileioperformer.FileIoPerformer
	tokenProvider     authtokenloader.AuthTokenLoader
	secretStoreClient secrets.SecretStoreClient
	secretStoreConfig secretstoreConfig.SecretStoreInfo
	tokenConfig       config.TokenFileProviderInfo
}

// NewTokenProvider creates a new TokenProvider
func NewTokenProvider(logger logger.LoggingClient,
	fileOpener fileioperformer.FileIoPerformer,
	tokenProvider authtokenloader.AuthTokenLoader,
	secretStoreClient secrets.SecretStoreClient) TokenProvider {
	return &fileTokenProvider{
		logger:            logger,
		fileOpener:        fileOpener,
		tokenProvider:     tokenProvider,
		secretStoreClient: secretStoreClient,
	}
}

// Set configuration
func (p *fileTokenProvider) SetConfiguration(secretStoreConfig secretstoreConfig.SecretStoreInfo, tokenConfig config.TokenFileProviderInfo) {
	p.secretStoreConfig = secretStoreConfig
	p.tokenConfig = tokenConfig
}

// Do whatever is needed
func (p *fileTokenProvider) Run() error {
	p.logger.Info("Generating Vault tokens")

	privilegedToken, err := p.tokenProvider.Load(p.tokenConfig.PrivilegedTokenPath)
	if err != nil {
		p.logger.Errorf("failed to read privileged access token: %s", err.Error())
		return err
	}

	tokenConfEnv, err := GetTokenConfigFromEnv()
	if err != nil {
		p.logger.Errorf("failed to get token config from environment variable %s with error: %s", addSecretstoreTokensEnvKey, err.Error())
		return err
	}

	var tokenConf TokenConfFile
	if err := LoadTokenConfig(p.fileOpener, p.tokenConfig.ConfigFile, &tokenConf); err != nil {
		p.logger.Errorf("failed to read token configuration file %s: %s", p.tokenConfig.ConfigFile, err.Error())
		return err
	}

	// merge the additional token configuration list from environment variable
	// note that the configuration file takes precedence, as the tokenConf will override
	// the tokenConfEnv with same duplicate keys
	// The tokenConfEnv only uses default settings.
	tokenConf = tokenConfEnv.mergeWith(tokenConf)

	credentialGenerator := secretstore.NewDefaultCredentialGenerator()

	userManager := common.NewUserManager(p.logger, p.secretStoreClient, p.tokenConfig.UserPassMountPoint, "edgex-identity",
		privilegedToken, p.tokenConfig.DefaultTokenTTL, p.tokenConfig.DefaultJWTAudience, p.tokenConfig.DefaultJWTTTL)

	for serviceName, serviceConfig := range tokenConf {
		p.logger.Infof("generating policy/token defaults for service %s", serviceName)

		servicePolicy := make(map[string]interface{})

		if serviceConfig.UseDefaults {
			p.logger.Infof("using policy/token defaults for service %s", serviceName)
			servicePolicy = common.MakeDefaultTokenPolicy(serviceName)
			defaultPolicyPaths := servicePolicy["path"].(map[string]interface{})
			for pathKey, policy := range defaultPolicyPaths {
				servicePolicy["path"].(map[string]interface{})[pathKey] = policy
			}
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

		// Generate a random password

		randomPassword, err := credentialGenerator.Generate(context.TODO())
		if err != nil {
			return err
		}

		// Create a user with the random password

		err = userManager.CreatePasswordUserWithPolicy(serviceName, randomPassword, "edgex-service-", servicePolicy)
		if err != nil {
			return err
		}

		// Immediately log in the user to get a vault token

		var createTokenResponse interface{}
		if createTokenResponse, err = p.secretStoreClient.InternalServiceLogin(privilegedToken, p.tokenConfig.UserPassMountPoint, serviceName, randomPassword); err != nil {
			return err
		}

		// Serialize the vault token to disk

		outputTokenDir := filepath.Join(p.tokenConfig.OutputDir, serviceName)
		outputTokenFilename := filepath.Join(outputTokenDir, p.tokenConfig.OutputFilename)
		if err := p.fileOpener.MkdirAll(outputTokenDir, os.FileMode(0700)); err != nil {
			p.logger.Errorf("failed to create base directory path(s) %s: %s", outputTokenDir, err.Error())
			return err
		}

		p.logger.Infof("creating token file %s", outputTokenFilename)
		writeCloser, err := p.fileOpener.OpenFileWriter(outputTokenFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600))
		if err != nil {
			p.logger.Errorf("failed open token file for writing %s: %s", outputTokenFilename, err.Error())
			return err
		}
		// writeCloser is writable file -- explicitly close() to ensure we catch errors writing to it

		permissionable, ok := writeCloser.(permissionable)
		if ok {
			if serviceConfig.FilePermissions != nil &&
				(serviceConfig.FilePermissions).ModeOctal != nil {
				// Use ParseUint because file modes are non-negative.
				// The bitSize of 32 correctly matches the underlying type of os.FileMode (uint32)
				mode, err := strconv.ParseUint(*(serviceConfig.FilePermissions).ModeOctal, 8, 32)
				if err != nil {
					_ = writeCloser.Close()
					p.logger.Errorf("invalid file mode %s: %s", *(serviceConfig.FilePermissions).ModeOctal, err.Error())
					return err
				}
				// The conversion from uint64 (returned by ParseUint) to os.FileMode (uint32) is now safe,
				// because ParseUint guarantees the value will fit within the specified bitSize (32)
				if err := permissionable.Chmod(os.FileMode(mode)); err != nil {
					_ = writeCloser.Close()
					p.logger.Errorf("failed to set file mode on %s: %s", outputTokenFilename, err.Error())
					return err
				}
			}
			if serviceConfig.FilePermissions != nil &&
				(serviceConfig.FilePermissions).Uid != nil &&
				(serviceConfig.FilePermissions).Gid != nil {
				err := permissionable.Chown(*(serviceConfig.FilePermissions).Uid, *(serviceConfig.FilePermissions).Gid)
				if err != nil {
					_ = writeCloser.Close()
					p.logger.Errorf("failed to set file user/group on %s: %s", outputTokenFilename, err.Error())
					return err
				}
			}
		}

		// Write resulting token
		if err := json.NewEncoder(writeCloser).Encode(createTokenResponse); err != nil {
			_ = writeCloser.Close()
			p.logger.Errorf("failed to write token file: %s", err.Error())
			return err
		}

		if err := writeCloser.Close(); err != nil {
			p.logger.Errorf("failed to close %s: %s", outputTokenFilename, err.Error())
			return err
		}
	}

	return nil
}

func (p *fileTokenProvider) RegenToken(entityId string) error {
	p.logger.Infof("Generating new OpenBao token associated with the entity id '%s'", entityId)

	privilegedToken, err := p.tokenProvider.Load(p.tokenConfig.PrivilegedTokenPath)
	if err != nil {
		p.logger.Errorf("failed to read privileged access token: %v", err)
		return err
	}

	tokenConfEnv, err := GetTokenConfigFromEnv()
	if err != nil {
		p.logger.Errorf("failed to get token config from environment variable %s with error: %v", addSecretstoreTokensEnvKey, err)
		return err
	}

	var tokenConf TokenConfFile
	if err := LoadTokenConfig(p.fileOpener, p.tokenConfig.ConfigFile, &tokenConf); err != nil {
		p.logger.Errorf("failed to read token configuration file %s: %v", p.tokenConfig.ConfigFile, err)
		return err
	}

	// merge the additional token configuration list from environment variable
	// note that the configuration file takes precedence, as the tokenConf will override
	// the tokenConfEnv with same duplicate keys
	// The tokenConfEnv only uses default settings.
	tokenConf = tokenConfEnv.mergeWith(tokenConf)

	var aliasName string
	var policies []string

	// retrieve the entity metadata that will be required while creating a token and associates with the entity
	getEntityResp, err := p.secretStoreClient.GetIdentityByEntityId(privilegedToken, entityId)
	if err != nil {
		p.logger.Errorf("failed to get entity by id '%s' from secretStoreClient: %v", entityId, err)
		return err
	}

	if len(getEntityResp.Aliases) == 0 {
		err = fmt.Errorf("alias name not defined in entity resp of entity id '%s'", entityId)
		p.logger.Error(err.Error())
		return err
	}
	aliasName = getEntityResp.Aliases[0].Name

	if len(getEntityResp.Policies) == 0 {
		policies = []string{"edgex-service-" + aliasName}
	} else {
		policies = getEntityResp.Policies
	}

	outputTokenDir := filepath.Join(p.tokenConfig.OutputDir, aliasName)
	outputTokenFilename := filepath.Join(outputTokenDir, p.tokenConfig.OutputFilename)

	p.logger.Infof("opening token file %s", outputTokenFilename)
	writeCloser, err := p.fileOpener.OpenFileWriter(outputTokenFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600))
	if err != nil {
		p.logger.Errorf("failed open token file for writing %s: %v", outputTokenFilename, err)
		return err
	}

	createTokenParameters := make(map[string]any)
	createTokenParameters["display_name"] = "token-" + aliasName
	createTokenParameters["role_name"] = aliasName
	createTokenParameters["entity_alias"] = aliasName
	createTokenParameters["no_parent"] = true
	createTokenParameters["policies"] = policies
	createTokenParameters["meta"] = map[string]string{"user": aliasName}
	createTokenParameters["ttl"] = p.tokenConfig.DefaultTokenTTL
	createTokenParameters["period"] = p.tokenConfig.DefaultTokenTTL
	createTokenParameters["renewable"] = true

	p.logger.Infof("creating token for %s ......", aliasName)

	p.logger.Debugf("invoking CreateTokenByRole method with parameters %v", createTokenParameters)

	createTokenResponse, err := p.secretStoreClient.CreateTokenByRole(privilegedToken, aliasName, createTokenParameters)
	if err != nil {
		p.logger.Errorf("failed creation of new token for '%s': %v", aliasName, err)
		return err
	}

	// Write resulting token
	if err := json.NewEncoder(writeCloser).Encode(createTokenResponse); err != nil {
		_ = writeCloser.Close()
		p.logger.Errorf("failed to write token file: %v", err)
		return err
	}

	if err := writeCloser.Close(); err != nil {
		p.logger.Errorf("failed to close %s: %v", outputTokenFilename, err)
		return err
	}

	return nil
}
