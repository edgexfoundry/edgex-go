/*******************************************************************************
 * Copyright 2021 Intel Corporation
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

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
)

type ConfigurationStruct struct {
	LogLevel       string
	RequestTimeout int
	KongURL        KongUrlInfo
	KongAuth       KongAuthInfo
	KongACL        KongAclInfo
	SecretStore    SecretStoreInfo
	Clients        map[string]bootstrapConfig.ClientInfo
}

type KongUrlInfo struct {
	Server             string
	AdminPort          int
	AdminPortSSL       int
	ApplicationPort    int
	ApplicationPortSSL int
}

func (k KongUrlInfo) GetProxyBaseURL() string {
	baseUrl := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%v", k.Server, k.AdminPort),
	}
	return baseUrl.String()
}

func (k KongUrlInfo) GetSecureURL() string {
	secureUrl := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%v", k.Server, k.ApplicationPortSSL),
	}
	return secureUrl.String()
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

type SecretStoreInfo struct {
	Type            string
	Protocol        string
	Host            string
	Port            int
	HealthCheckPath string
	CertPath        string
	TokenPath       string
	CACertPath      string
	SNIS            []string
}

// GetBaseURL builds and returns the base URL for the SecretStore service
func (s SecretStoreInfo) GetBaseURL() string {
	return fmt.Sprintf("%s://%s:%d", s.Protocol, s.Host, s.Port)
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct
// Not needed for this service, so just return false
func (c *ConfigurationStruct) UpdateFromRaw(_ interface{}) bool {
	return false
}

// EmptyWritablePtr returns a pointer to a service-specific empty WritableInfo struct.
// Not needed for this service, so return nil
func (c *ConfigurationStruct) EmptyWritablePtr() interface{} {
	return nil
}

// UpdateWritableFromRaw converts configuration received from the registry to a service-specific WritableInfo struct
// Not needed for this service, so just return false
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
// Not needed for this service, so return nil
func (c *ConfigurationStruct) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return nil
}
