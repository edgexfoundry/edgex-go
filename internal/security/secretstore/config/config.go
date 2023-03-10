/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright (C) 2023 Intel Corporation
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
	"fmt"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
)

type ConfigurationStruct struct {
	LogLevel         string
	SecretStore      SecretStoreInfo
	Databases        map[string]Database
	SecureMessageBus SecureMessageBusInfo
}

type Database struct {
	Username string
	Service  string
}

type SecureMessageBusInfo struct {
	Type                  string
	KuiperConfigPath      string
	KuiperConnectionsPath string
	Services              map[string]ServiceInfo
}

type ServiceInfo struct {
	Service string
}

type SecretStoreInfo struct {
	Type                        string
	Protocol                    string
	Host                        string
	Port                        int
	ServerName                  string
	CertPath                    string
	CaFilePath                  string
	CertFilePath                string
	KeyFilePath                 string
	TokenFolderPath             string
	TokenFile                   string
	VaultSecretShares           int
	VaultSecretThreshold        int
	TokenProvider               string
	TokenProviderArgs           []string
	TokenProviderType           string
	TokenProviderAdminTokenPath string
	PasswordProvider            string
	PasswordProviderArgs        []string
	RevokeRootTokens            bool
	ConsulSecretsAdminTokenPath string
}

// GetBaseURL builds and returns the base URL for the SecretStore service
func (s SecretStoreInfo) GetBaseURL() string {
	return fmt.Sprintf("%s://%s:%d/", s.Protocol, s.Host, s.Port)
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct
// Not needed for this service, so always return false
func (c *ConfigurationStruct) UpdateFromRaw(_ interface{}) bool {
	return false
}

// EmptyWritablePtr returns a pointer to a service-specific empty WritableInfo struct.
// Not needed for this service, so return nil
func (c *ConfigurationStruct) EmptyWritablePtr() interface{} {
	return nil
}

// GetWritablePtr returns pointer to the writable section
// Not needed for this service, so return nil
func (c *ConfigurationStruct) GetWritablePtr() any {
	return nil
}

// UpdateWritableFromRaw converts configuration received from the registry to a service-specific WritableInfo struct
// Not needed for this service, so always return false
func (c *ConfigurationStruct) UpdateWritableFromRaw(_ interface{}) bool {
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

// GetInsecureSecrets returns the service's InsecureSecrets
// Not used for this service, so return nil
func (c *ConfigurationStruct) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return nil
}

// GetTelemetryInfo returns the service's Telemetry settings of which this service doesn't have. I.e. service has no metrics
func (c *ConfigurationStruct) GetTelemetryInfo() *bootstrapConfig.TelemetryInfo {
	return nil
}
