/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2019 Intel Corporation.
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

package config

import (
	secretstoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
)

type ConfigurationStruct struct {
	LogLevel          string
	SecretStore       secretstoreConfig.SecretStoreInfo
	TokenFileProvider TokenFileProviderInfo
}

type TokenFileProviderInfo struct {
	// Path to Vault authorization token to be used by the service
	PrivilegedTokenPath string
	// Configuration file used to control token creation (default: res/token-bootstrapConfig.json)
	ConfigFile string
	// Base directory for token file output
	OutputDir string
	// File name for token file (default: secrets-token.json)
	OutputFilename string
	// Default duration of issued tokens
	DefaultTokenTTL string
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct
// Not needed for this service, so return false
func (c *ConfigurationStruct) UpdateFromRaw(rawConfig interface{}) bool {
	return false
}

// EmptyWritablePtr returns a pointer to a service-specific empty WritableInfo struct.
// Not needed for this service, so return nil
func (c *ConfigurationStruct) EmptyWritablePtr() interface{} {
	return nil
}

// UpdateWritableFromRaw converts configuration received from the registry to a service-specific WritableInfo struct
// Not needed for this service, so return false
func (c *ConfigurationStruct) UpdateWritableFromRaw(rawWritable interface{}) bool {
	return false
}

// GetBootstrap returns the configuration elements required by the bootstrap.
// Not needed for this service, so return empty struct
func (c *ConfigurationStruct) GetBootstrap() bootstrapConfig.BootstrapConfiguration {
	return bootstrapConfig.BootstrapConfiguration{}
}

// GetLogLevel returns the current ConfigurationStruct's log level.
func (c *ConfigurationStruct) GetLogLevel() string {
	return c.LogLevel
}

// GetRegistryInfo returns the RegistryInfo from the ConfigurationStruct.
// Not needed for this service, so return empty struct
func (c *ConfigurationStruct) GetRegistryInfo() bootstrapConfig.RegistryInfo {
	return bootstrapConfig.RegistryInfo{}
}

// GetDatabaseInfo returns a database information map.
// Not needed for this service, so return nil
func (c *ConfigurationStruct) GetDatabaseInfo() map[string]bootstrapConfig.Database {
	return nil
}

// GetInsecureSecrets returns the service's InsecureSecrets which this service doesn't support
func (c *ConfigurationStruct) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return nil
}
