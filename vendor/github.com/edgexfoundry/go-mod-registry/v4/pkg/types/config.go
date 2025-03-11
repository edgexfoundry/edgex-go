//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package types

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
)

// Config defines the information need to connect to the registry service and optionally register the service
// for discovery and health checks
type Config struct {
	// The Protocol that should be used to connect to the registry service. HTTP is used if not set.
	Protocol string
	// Host is the hostname or IP address of the registry service
	Host string
	// Port is the HTTP port of the registry service
	Port int
	// Type is the implementation type of the registry service, i.e. keeper
	Type string
	// ServiceKey is the key identifying the service for Registration and building the services base configuration path.
	ServiceKey string
	// ServiceHost is the hostname or IP address of the current running service using this module. May be left empty if not using registration
	ServiceHost string
	// ServicePort is the HTTP port of the current running service using this module. May be left unset if not using registration
	ServicePort int
	// The ServiceProtocol that should be used to call the current running service using this module. May be left empty if not using registration
	ServiceProtocol string
	// Health check callback route for the current running service using this module. May be left empty if not using registration
	CheckRoute string
	// Health check callback interval. May be left empty if not using registration
	CheckInterval string
	// AuthInjector is an interface to obtain a JWT and secure transport for remote service calls
	AuthInjector interfaces.AuthenticationInjector
	// EnableNameFieldEscape indicates whether enables NameFieldEscape in this service
	// The name field escape could allow the system to use special or Chinese characters in the different name fields, including device, profile, and so on.  If the EnableNameFieldEscape is false, some special characters might cause system error.
	// TODO: remove in EdgeX 4.0
	EnableNameFieldEscape bool
}

//
// A few helper functions for building URLs.
//

func (config Config) GetRegistryUrl() string {
	return fmt.Sprintf("%s://%s:%v", config.GetRegistryProtocol(), config.Host, config.Port)
}

func (config Config) GetHealthCheckUrl() string {
	return config.GetExpandedRoute(config.CheckRoute)
}

func (config Config) GetExpandedRoute(route string) string {
	return fmt.Sprintf("%s://%s:%v%s", config.GetServiceProtocol(), config.ServiceHost, config.ServicePort, route)
}

func (config Config) GetRegistryProtocol() string {
	if config.Protocol == "" {
		return "http"
	}

	return config.Protocol
}

func (config Config) GetServiceProtocol() string {
	if config.ServiceProtocol == "" {
		return "http"
	}

	return config.ServiceProtocol
}
