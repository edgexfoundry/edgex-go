/*******************************************************************************
 * Copyright 2019 Dell Inc.
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
	"net/url"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"
)

type ConfigurationStruct struct {
	Writable      WritableInfo
	Logging       bootstrapConfig.LoggingInfo
	KongURL       KongUrlInfo
	KongAuth      KongAuthInfo
	KongACL       KongAclInfo
	SecretStore   bootstrapConfig.SecretStoreInfo
	SecretService SecretServiceInfo
	Clients       map[string]bootstrapConfig.ClientInfo
}

type WritableInfo struct {
	LogLevel       string
	RequestTimeout int
}

type KongUrlInfo struct {
	Server             string
	AdminPort          int
	AdminPortSSL       int
	ApplicationPort    int
	ApplicationPortSSL int
}

func (k KongUrlInfo) GetProxyBaseURL() string {
	url := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%v", k.Server, k.AdminPort),
	}
	return url.String()
}

func (k KongUrlInfo) GetSecureURL() string {
	url := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%v", k.Server, k.ApplicationPortSSL),
	}
	return url.String()
}

type KongAuthInfo struct {
	Name       string
	TokenTTL   int
	Resource   string
	OutputPath string
}

type KongAclInfo struct {
	Name      string
	WhiteList string
}

type SecretServiceInfo struct {
	Server          string
	Port            int
	HealthcheckPath string
	CertPath        string
	TokenPath       string
	CACertPath      string
	SNIS            []string
}

func (s SecretServiceInfo) GetSecretSvcBaseURL() string {
	url := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%v", s.Server, s.Port),
	}
	return url.String()
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct which is
// then used to overwrite the service's existing configuration struct.
func (c *ConfigurationStruct) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ConfigurationStruct)
	if ok {
		// Check that information was successfully read from Registry
		if configuration.SecretService.Port == 0 {
			return false
		}
		*c = *configuration
	}
	return ok
}

// EmptyWritablePtr returns a pointer to a service-specific empty WritableInfo struct.  It is used by the bootstrap to
// provide the appropriate structure to registry.Client's WatchForChanges().
func (c *ConfigurationStruct) EmptyWritablePtr() interface{} {
	return &WritableInfo{}
}

// UpdateWritableFromRaw converts configuration received from the registry to a service-specific WritableInfo struct
// which is then used to overwrite the service's existing configuration's WritableInfo struct.
func (c *ConfigurationStruct) UpdateWritableFromRaw(rawWritable interface{}) bool {
	writable, ok := rawWritable.(*WritableInfo)
	if ok {
		c.Writable = *writable
	}
	return ok
}

// GetBootstrap returns the configuration elements required by the bootstrap.  Currently, a copy of the configuration
// data is returned.  This is intended to be temporary -- since ConfigurationStruct drives the configuration.toml's
// structure -- until we can make backwards-breaking configuration.toml changes (which would consolidate these fields
// into an interfaces.BootstrapConfiguration struct contained within ConfigurationStruct).
func (c *ConfigurationStruct) GetBootstrap() interfaces.BootstrapConfiguration {
	//To keep config file for proxy unchanged in Geneva we need to create a temporary SecretStore struct so that bootstrapHandler can use it to create a secret client
	//The config file may be changed in the future version and SecretStore can be used directly like other core services
	ss := bootstrapConfig.SecretStoreInfo{
		Host:                    c.SecretService.Server,
		Port:                    c.SecretService.Port,
		Path:                    c.SecretService.CertPath,
		Protocol:                "https",
		RootCaCertPath:          c.SecretService.CACertPath,
		ServerName:              c.SecretService.Server,
		Authentication:          vault.AuthenticationInfo{AuthType: "X-Vault-Token"},
		AdditionalRetryAttempts: 10,
		RetryWaitPeriod:         "5s",
		TokenFile:               c.SecretService.TokenPath,
	}

	// temporary until we can make backwards-breaking configuration.toml change
	return interfaces.BootstrapConfiguration{
		Clients:     c.Clients,
		Logging:     c.Logging,
		SecretStore: ss,
	}
}

// GetLogLevel returns the current ConfigurationStruct's log level.
func (c *ConfigurationStruct) GetLogLevel() string {
	return c.Writable.LogLevel
}

// GetRegistryInfo returns the RegistryInfo from the ConfigurationStruct.
func (c *ConfigurationStruct) GetRegistryInfo() bootstrapConfig.RegistryInfo {
	return bootstrapConfig.RegistryInfo{}
}

// GetDatabaseInfo returns a database information map.
func (c *ConfigurationStruct) GetDatabaseInfo() map[string]bootstrapConfig.Database {
	panic("GetDatabaseInfo() called unexpectedly.")
}
