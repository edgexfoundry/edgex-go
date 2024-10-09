//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
)

// ConfigurationStruct contains the configuration properties for the Support Cron Scheduler Service
type ConfigurationStruct struct {
	Writable   WritableInfo
	Database   bootstrapConfig.Database
	Registry   bootstrapConfig.RegistryInfo
	Service    bootstrapConfig.ServiceInfo
	Clients    bootstrapConfig.ClientsCollection
	MessageBus bootstrapConfig.MessageBusInfo
	Retention  RecordRetention
}

type WritableInfo struct {
	LogLevel        string
	InsecureSecrets bootstrapConfig.InsecureSecrets
	Telemetry       bootstrapConfig.TelemetryInfo
}

type RecordRetention struct {
	Enabled  bool
	Interval string
	MaxCap   uint32
	MinCap   uint32
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct which is
// then used to overwrite the service's existing configuration struct.
func (c *ConfigurationStruct) UpdateFromRaw(rawConfig any) bool {
	configuration, ok := rawConfig.(*ConfigurationStruct)
	if ok {
		*c = *configuration
	}
	return ok
}

// EmptyWritablePtr returns a pointer to a service-specific empty WritableInfo struct.  It is used by the bootstrap to
// provide the appropriate structure to registry.Client's WatchForChanges().
func (c *ConfigurationStruct) EmptyWritablePtr() any {
	return &WritableInfo{}
}

// GetWritablePtr returns pointer to the writable section
func (c *ConfigurationStruct) GetWritablePtr() any {
	return &c.Writable
}

// UpdateWritableFromRaw converts configuration received from the registry to a service-specific WritableInfo struct
// which is then used to overwrite the service's existing configuration's WritableInfo struct.
func (c *ConfigurationStruct) UpdateWritableFromRaw(rawWritable any) bool {
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
		Clients:    &c.Clients,
		Database:   &c.Database,
		MessageBus: &c.MessageBus,
		Registry:   &c.Registry,
		Service:    &c.Service,
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
func (c *ConfigurationStruct) GetDatabaseInfo() bootstrapConfig.Database {
	return c.Database
}

// GetInsecureSecrets returns the service's InsecureSecrets.
func (c *ConfigurationStruct) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return c.Writable.InsecureSecrets
}

// GetTelemetryInfo returns the service's Telemetry settings.
func (c *ConfigurationStruct) GetTelemetryInfo() *bootstrapConfig.TelemetryInfo {
	return &c.Writable.Telemetry
}
