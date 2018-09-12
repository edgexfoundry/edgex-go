/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
	"github.com/edgexfoundry/edgex-go/internal"
)

// Cannot use type string inside structs to be parsed into map[string]interface{} correctly
// For now using const literals for values
const (
	Consul = "consul"
)

// ServiceInfo contains configuration settings necessary for the basic operation of any EdgeX service.
type ServiceInfo struct {
	// BootTimeout indicates, in milliseconds, how long the service will retry connecting to upstream dependencies
	// before giving up. Default is 30,000.
	BootTimeout int
	// Health check interval
	CheckInterval string
	// Indicates the interval in milliseconds at which service clients should check for any configuration updates
	ClientMonitor int
	// Host is the hostname or IP address of the service.
	Host string
	// Port is the HTTP port of the service.
	Port int
	// The protocol that should be used to call this service
	Protocol string
	// StartupMsg specifies a string to log once service
	// initialization and startup is completed.
	StartupMsg string
	// ReadMaxLimit specifies the maximum size list supported
	// in response to REST calls to other services.
	ReadMaxLimit int
	// Timeout specifies a timeout (in milliseconds) for
	// processing REST calls from other services.
	Timeout int
}

// HealthCheck is a URL specifying a healthcheck REST endpoint used by the Registry to determine if the
// service is available.
func (s ServiceInfo) HealthCheck() string {
	hc := fmt.Sprintf("%s://%s:%v%s", s.Protocol, s.Host, s.Port, internal.ApiPingRoute)
	return hc
}

// Url provides a way to obtain the full url of the host service for use in initialization or, in some cases,
// responses to a caller.
func (s ServiceInfo) Url() string {
	url := fmt.Sprintf("%s://%s:%v", s.Protocol, s.Host, s.Port)
	return url
}

// RegistryInfo defines the type and location (via host/port) of the desired service registry (e.g. Consul, Eureka)
type RegistryInfo struct {
	Host string
	Port int
	Type string
}

// LoggingInfo provides basic parameters related to where logs should be written.
type LoggingInfo struct {
	EnableRemote bool
	File         string
	RemoteURL    string
}

// DatabaseInfo defines the parameters necessary for connecting to the desired persistence layer.
type DatabaseInfo struct {
	Type     string
	Timeout  int
	Host     string
	Port     int
	Username string
	Password string
	Name     string
}

// clientInfo provides the host and port of another service in the eco-system.
type clientInfo struct {
	// Host is the hostname or IP address of a service.
	Host string
	// Port defines the port on which to access a given service
	Port int
	// Protocol indicates the protocol to use when accessing a given service
	Protocol string
}

func (c clientInfo) Url() string {
	url := fmt.Sprintf("%s://%s:%v", c.Protocol, c.Host, c.Port)
	return url
}

// BaseConfig is a struct which provides the basic types that all services will need
// for operation. Extend/wrap this type according to the needs of your service.
type BaseConfig struct {
	// Clients is a map of services used by a DS.
	Clients map[string]clientInfo
	// Logging contains logging-specific configuration settings.
	Logging LoggingInfo
	// Registry contains registry-specific settings.
	Registry RegistryInfo
	// Service contains settings specific to the service being run.
	Service ServiceInfo
}
