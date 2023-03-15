/*******************************************************************************
 * Copyright 2022-2023 Intel Corporation
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
	fileProviderConfig "github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
)

// SpiffeInfo contains information for Spiffe configuration
type SpiffeInfo struct {
	// Path to workload API Unix domain socket (go_spiffe interprets env.SPIFFE_ENDPOINTSOCKET automatically)
	EndpointSocket string
	// trusted domain name
	TrustDomain string
}

type ConfigurationStruct struct {
	Writable    WritableInfo
	MessageBus  bootstrapConfig.MessageBusInfo
	Clients     map[string]bootstrapConfig.ClientInfo
	Database    bootstrapConfig.Database
	Registry    bootstrapConfig.RegistryInfo
	Service     bootstrapConfig.ServiceInfo
	TokenConfig fileProviderConfig.TokenFileProviderInfo
	Spiffe      SpiffeInfo
}

type WritableInfo struct {
	PersistData     bool
	LogLevel        string
	InsecureSecrets bootstrapConfig.InsecureSecrets
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct which is
// then used to overwrite the service's existing configuration struct.
func (c *ConfigurationStruct) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ConfigurationStruct)
	if ok {
		*c = *configuration
	}
	return ok
}

// EmptyWritablePtr returns a pointer to a service-specific empty WritableInfo struct.  It is used by the bootstrap to
// provide the appropriate structure to registry.Client's WatchForChanges().
func (c *ConfigurationStruct) EmptyWritablePtr() interface{} {
	return &WritableInfo{}
}

// GetWritablePtr returns pointer to the writable section
func (c *ConfigurationStruct) GetWritablePtr() any {
	return &c.Writable
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
// data is returned.  This is intended to be temporary -- since ConfigurationStruct drives the configuration.yaml's
// structure -- until we can make backwards-breaking configuration.yaml changes (which would consolidate these fields
// into an bootstrapConfig.BootstrapConfiguration struct contained within ConfigurationStruct).
func (c *ConfigurationStruct) GetBootstrap() bootstrapConfig.BootstrapConfiguration {
	// temporary until we can make backwards-breaking configuration.yaml change
	return bootstrapConfig.BootstrapConfiguration{
		Clients:  c.Clients,
		Service:  c.Service,
		Registry: c.Registry,
	}
}

// GetLogLevel returns the current ConfigurationStruct's log level.
func (c *ConfigurationStruct) GetLogLevel() string {
	return c.Writable.LogLevel
}

// GetRegistryInfo returns the RegistryInfo from the ConfigurationStruct.
func (c *ConfigurationStruct) GetRegistryInfo() bootstrapConfig.RegistryInfo {
	return c.Registry
}

// GetDatabaseInfo returns a database information.
func (c *ConfigurationStruct) GetDatabaseInfo() bootstrapConfig.Database {
	return c.Database
}

// GetInsecureSecrets returns the service's InsecureSecrets.
func (c *ConfigurationStruct) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return c.Writable.InsecureSecrets
}

// GetTelemetryInfo returns the service's Telemetry settings of which this service doesn't have. I.e. service has no metrics
func (c *ConfigurationStruct) GetTelemetryInfo() *bootstrapConfig.TelemetryInfo {
	return nil
}
