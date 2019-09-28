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

package secretstore

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"net/url"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
)

type ConfigurationStruct struct {
	Writable      WritableInfo
	Clients       map[string]config.ClientInfo
	Logging       config.LoggingInfo
	Registry      config.RegistryInfo
	Service       config.ServiceInfo
	SecretService SecretServiceInfo
}

type WritableInfo struct {
	LogLevel string
	Title    string
}

type SecretServiceInfo struct {
	Scheme               string
	Server               string
	Port                 int
	CertPath             string
	CaFilePath           string
	CertFilePath         string
	KeyFilePath          string
	TokenFolderPath      string
	TokenFile            string
	VaultSecretShares    int
	VaultSecretThreshold int
}

func (s SecretServiceInfo) GetSecretSvcBaseURL() string {
	url := &url.URL{
		Scheme: s.Scheme,
		Host:   fmt.Sprintf("%s:%v", s.Server, s.Port),
		Path:   "/",
	}
	return url.String()
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct which is
// then used to overwrite the service's existing configuration struct.
func (c *ConfigurationStruct) UpdateFromRaw(rawConfig interface{}) bool {
	if configuration, ok := rawConfig.(*ConfigurationStruct); ok {
		c = configuration
		return true
	}
	return false
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
	// temporary until we can make backwards-breaking configuration.toml change
	return interfaces.BootstrapConfiguration{
		Clients:  c.Clients,
		Service:  c.Service,
		Registry: c.Registry,
		Logging:  c.Logging,
	}
}

// GetLogLevel returns the current ConfigurationStruct's log level.
func (c *ConfigurationStruct) GetLogLevel() string {
	return c.Writable.LogLevel
}

// SetLogLevel updates the log level in the ConfigurationStruct.
func (c *ConfigurationStruct) SetRegistryInfo(registryInfo config.RegistryInfo) {
	c.Registry = registryInfo
}
