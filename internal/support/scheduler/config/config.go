/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
 *******************************************************************************/

package config

import (
	"fmt"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
)

// Configuration V2 for the Support Scheduler Service
type ConfigurationStruct struct {
	Writable        WritableInfo
	Clients         map[string]bootstrapConfig.ClientInfo
	Database        bootstrapConfig.Database
	Registry        bootstrapConfig.RegistryInfo
	Service         bootstrapConfig.ServiceInfo
	MessageBus      bootstrapConfig.MessageBusInfo
	Intervals       map[string]IntervalInfo
	IntervalActions map[string]IntervalActionInfo
	// ScheduleIntervalTime is a time(Millisecond) to create a ticker to delay the scheduler loop
	ScheduleIntervalTime int
}

type WritableInfo struct {
	LogLevel        string
	InsecureSecrets bootstrapConfig.InsecureSecrets
	Telemetry       bootstrapConfig.TelemetryInfo
}

type IntervalInfo struct {
	// Name of the schedule must be unique?
	Name string
	// Start time in ISO 8601 format YYYYMMDD'T'HHmmss
	Start string
	// End time in ISO 8601 format YYYYMMDD'T'HHmmss
	End string
	// Periodicity of the schedule
	Interval string
	// Cron style regular expression indicating how often the action under schedule should occur.
	// Use either runOnce, frequency or cron and not all.
	Cron string
	// Boolean indicating that this schedules runs one time - at the time indicated by the start
	RunOnce bool
}

type IntervalActionInfo struct {
	// Host is the hostname or IP address of a service.
	Host string
	// Port defines the port on which to access a given service
	Port int
	// Protocol indicates the protocol to use when accessing a given service
	Protocol string
	// Action name
	Name string
	// Action http method *const prob*
	Method string
	// Action target parameters
	Parameters string
	// Action target request body
	RequestBody string
	// Action target API path
	Path string
	// Associated Schedule for the Event
	Interval string
	// Content is the actual content to be sent as the body of the notification
	Content string
	// ContentType indicates the MIME type/Content-type of the notification's content.
	ContentType string
	// Administrative state (LOCKED/UNLOCKED)
	AdminState string
	// AuthMethod indicates how to authenticate the outbound URL -- "none" (default) or "jwt"
	AuthMethod string
}

const (
	AuthMethodNone = "NONE"
	AuthMethodJWT  = "JWT"
)

// URI constructs a URI from the protocol, host and port and returns that as a string.
func (e IntervalActionInfo) URL() string {
	return fmt.Sprintf("%s://%s:%v", e.Protocol, e.Host, e.Port)
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
		Clients:    c.Clients,
		Service:    c.Service,
		Registry:   c.Registry,
		MessageBus: c.MessageBus,
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
