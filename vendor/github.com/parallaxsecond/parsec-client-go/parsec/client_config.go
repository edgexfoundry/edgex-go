// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0
package parsec

import (
	"github.com/parallaxsecond/parsec-client-go/interface/auth"
	"github.com/parallaxsecond/parsec-client-go/interface/connection"
)

// ClientConfig holds a configuration for the basic client to be passed to InitClient
// ClientConfig's methods use the Builder pattern to build configurations, e.g:
// config := NewClientConfig().DirectAuthConfigData("myapp").Connection(myConn)
type ClientConfig struct {
	authenticatorData map[auth.AuthenticationType]interface{}
	connection        connection.Connection
	defaultProvider   *ProviderID
	authenticator     Authenticator
}

// NewClientConfig ceates a ClientConfig with defaults
func NewClientConfig() *ClientConfig {
	config := ClientConfig{
		authenticatorData: make(map[auth.AuthenticationType]interface{}),
	}
	return &config
}

// DirectAuthConfigData creates a new ClientConfig with the appName parameter set for Direct Authentication
func DirectAuthConfigData(appName string) *ClientConfig {
	config := NewClientConfig()
	config.authenticatorData[auth.AuthDirect] = appName
	return config
}

// DirectAuthConfigData sets the appName parameter to use when using Direct Authentication
func (config *ClientConfig) DirectAuthConfigData(appName string) *ClientConfig {
	config.authenticatorData[auth.AuthDirect] = appName
	return config
}

// Connection sets the conn.Connection object to use when connecting to the parsec service.
// This is primarily used for testing purposes, to allow for mocking of the parsec service.
func (config *ClientConfig) Connection(conn connection.Connection) *ClientConfig {
	config.connection = conn
	return config
}

// Provider set the provider to use.  If this is set the basic client won't attempt to auto select
// a provider, even if this one is not supported by the parsec service.
func (config *ClientConfig) Provider(provider ProviderID) *ClientConfig {
	config.defaultProvider = &provider
	return config
}

// Authenticator sets the authenticator to use.  If this is set, the basic client won't attempt to
// auto select an authenticator even if this one is not supported by the parsec service
func (config *ClientConfig) Authenticator(authenticator Authenticator) *ClientConfig {
	config.authenticator = authenticator
	return config
}
