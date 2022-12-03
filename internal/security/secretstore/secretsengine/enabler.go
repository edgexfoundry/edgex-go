/*******************************************************************************
* Copyright 2021-2023 Intel Corporation
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

package secretsengine

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v3/secrets"
)

const (
	// KVSecretsEngineMountPoint is the name of the mount point base for Vault's key-value secrets engine
	KVSecretsEngineMountPoint = "secret"
	// ConsulSecretEngineMountPoint is the name of the mount point base for Vault's Consul secrets engine
	ConsulSecretEngineMountPoint = "consul"

	// Vault's secrets engine type related constants
	KeyValue = "kv"
	Consul   = "consul"

	// kvVersion is the version of key-value secret storage used
	// currently we use version 1 from Vault
	kvVersion = "1"

	// defaultConsulTokenLeaseTtl is the default time-to-live value for consul token
	// currently we don't set any lease time-to-live limit for Consul tokens
	// this will be changed in future for phase 3 based on the ADR
	defaultConsulTokenLeaseTtl = "0"
)

// SecretsEngine is the metadata for secretstore secret engine enabler
type SecretsEngine struct {
	mountPoint string
	engineType string
}

// New creates an instance for SecretsEngine with mountPoint and engineType
func New(mountPoint string, engineType string) SecretsEngine {
	return SecretsEngine{mountPoint: mountPoint, engineType: engineType}
}

// Enable enables the specified secrets engine for the secretstore
// the rootToken is required and returns error if not provided or invalid token provided
// also returns error if unsupported / unknown secretsEngineType is used
func (eng SecretsEngine) Enable(rootToken *string,
	lc logger.LoggingClient,
	client secrets.SecretStoreClient) error {
	if rootToken == nil {
		return fmt.Errorf("rootToken is required")
	}

	// the data returned from GET of check installed secrets engine API of Vault is
	// the mountPoint with trailing slash(/), eg. "secret/" for kv's mountPoint "secret"
	checkMountPoint := eng.mountPoint + "/"
	installed, err := client.CheckSecretEngineInstalled(*rootToken, checkMountPoint, eng.engineType)
	if err != nil {
		return fmt.Errorf("failed call to check if %s secrets engine is installed: %s",
			eng.engineType, err.Error())
	}

	if !installed {
		lc.Infof("enabling %s secrets engine for the first time...", eng.engineType)
		switch eng.engineType {
		case KeyValue:
			// Enable KV storage version 1 at /v1/{eng.path} path (/v1 prefix supplied by Vault)
			if err := client.EnableKVSecretEngine(*rootToken, eng.mountPoint, kvVersion); err != nil {
				return fmt.Errorf("failed to enable KV version %s secrets engine: %w", kvVersion, err)
			}
			lc.Infof("KeyValue secrets engine with version %s enabled", kvVersion)
		case Consul:
			// Enable Consul secrets storage at /consul path
			if err := client.EnableConsulSecretEngine(*rootToken,
				eng.mountPoint, defaultConsulTokenLeaseTtl); err != nil {
				return fmt.Errorf("failed to enable Consul secrets engine: %w", err)
			}
			lc.Infof("Consul secrets engine with config default_ttl = %s enabled", defaultConsulTokenLeaseTtl)
		default:
			return fmt.Errorf("Unsupported secrets engine type: %s", eng.engineType)
		}
	} else {
		lc.Infof("%s secrets engine already enabled...", eng.engineType)
	}
	return nil
}
