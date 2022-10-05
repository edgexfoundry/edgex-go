/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright 2022 IOTech Ltd.
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
 *******************************************************************************/

package config

import (
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
)

// ConfigurationStruct contains the configuration properties for the core-command service.
type ConfigurationStruct struct {
	//TODO: Remove in EdgeX 3.0 - Is needed now for backward compatability in 2.0
	RequireMessageBus bool
	Writable          WritableInfo
	Clients           map[string]bootstrapConfig.ClientInfo
	Databases         map[string]bootstrapConfig.Database
	Registry          bootstrapConfig.RegistryInfo
	Service           bootstrapConfig.ServiceInfo
	MessageQueue      MessageQueue
	SecretStore       bootstrapConfig.SecretStoreInfo
}

// WritableInfo contains configuration properties that can be updated and applied without restarting the service.
type WritableInfo struct {
	LogLevel        string
	InsecureSecrets bootstrapConfig.InsecureSecrets
	Telemetry       bootstrapConfig.TelemetryInfo
}

type MessageQueue struct {
	Internal bootstrapConfig.MessageBusInfo
	External bootstrapConfig.ExternalMQTTInfo
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct which is
// then used to overwrite the service's existing configuration struct.
func (c *ConfigurationStruct) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ConfigurationStruct)
	if ok {
		// Check that information was successfully read from Registry
		if configuration.Service.Port == 0 {
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
// into an bootstrapConfig.BootstrapConfiguration struct contained within ConfigurationStruct).
func (c *ConfigurationStruct) GetBootstrap() bootstrapConfig.BootstrapConfiguration {
	return bootstrapConfig.BootstrapConfiguration{
		Clients:      c.Clients,
		Service:      c.Service,
		Registry:     c.Registry,
		SecretStore:  c.SecretStore,
		MessageQueue: c.MessageQueue.Internal,
		ExternalMQTT: c.MessageQueue.External,
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

// GetDatabaseInfo returns a database information map.
func (c *ConfigurationStruct) GetDatabaseInfo() map[string]bootstrapConfig.Database {
	return c.Databases
}

// GetInsecureSecrets returns the service's InsecureSecrets.
func (c *ConfigurationStruct) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return c.Writable.InsecureSecrets
}

// GetTelemetryInfo returns the service's Telemetry settings.
func (c *ConfigurationStruct) GetTelemetryInfo() *bootstrapConfig.TelemetryInfo {
	return &c.Writable.Telemetry
}
