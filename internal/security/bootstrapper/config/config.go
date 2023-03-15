/*******************************************************************************
 * Copyright 2023 Intel Corporation
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
	"strings"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
)

type ConfigurationStruct struct {
	LogLevel  string
	StageGate StageGateInfo
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct which is
// then used to overwrite the service's existing configuration struct.
func (c *ConfigurationStruct) UpdateFromRaw(rawConfig interface{}) bool {
	return false
}

// EmptyWritablePtr returns a pointer to a service-specific empty WritableInfo struct.  It is used by the bootstrap to
// provide the appropriate structure to registry.Client's WatchForChanges().
func (c *ConfigurationStruct) EmptyWritablePtr() interface{} {
	return nil
}

// GetWritablePtr returns pointer to the writable section
// Not needed for this service, so return nil
func (c *ConfigurationStruct) GetWritablePtr() any {
	return nil
}

// UpdateWritableFromRaw converts configuration received from the registry to a service-specific WritableInfo struct
// which is then used to overwrite the service's existing configuration's WritableInfo struct.
func (c *ConfigurationStruct) UpdateWritableFromRaw(rawWritable interface{}) bool {
	return false
}

// GetBootstrap returns the configuration elements required by the bootstrap.  Currently, a copy of the configuration
// data is returned.  This is intended to be temporary -- since ConfigurationStruct drives the configuration.yaml's
// structure -- until we can make backwards-breaking configuration.yaml changes (which would consolidate these fields
// into an bootstrapConfig.BootstrapConfiguration struct contained within ConfigurationStruct).
func (c *ConfigurationStruct) GetBootstrap() bootstrapConfig.BootstrapConfiguration {
	// temporary until we can make backwards-breaking configuration.yaml change
	return bootstrapConfig.BootstrapConfiguration{}
}

// GetLogLevel returns the current ConfigurationStruct's log level.
func (c *ConfigurationStruct) GetLogLevel() string {
	return c.LogLevel
}

// GetRegistryInfo returns the RegistryInfo from the ConfigurationStruct.
func (c *ConfigurationStruct) GetRegistryInfo() bootstrapConfig.RegistryInfo {
	return bootstrapConfig.RegistryInfo{}
}

// GetDatabaseInfo returns a database information map.
func (c *ConfigurationStruct) GetDatabaseInfo() map[string]bootstrapConfig.Database {
	return nil
}

// GetInsecureSecrets returns the service's InsecureSecrets.
func (c *ConfigurationStruct) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return nil
}

// GetTelemetryInfo returns the service's Telemetry settings of which this service doesn't have. I.e. service has no metrics
func (c *ConfigurationStruct) GetTelemetryInfo() *bootstrapConfig.TelemetryInfo {
	return nil
}

// GetRoleNames gets the slice of the keys (i.e. the service keys) from map Roles as ACL role names
func (acl ACLInfo) GetACLRoleNames() []string {
	roleNames := make([]string, 0, len(acl.Roles))
	for serviceKey := range acl.Roles {
		// always converts to lower cases by design
		roleNames = append(roleNames, strings.ToLower(serviceKey))
	}
	return roleNames
}
