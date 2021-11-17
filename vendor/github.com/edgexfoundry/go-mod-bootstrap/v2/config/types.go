/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright 2020 Intel Inc.
 * Copyright 2021 IOTech Ltd.
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

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
)

const DefaultHttpProtocol = "http"

// ServiceInfo contains configuration settings necessary for the basic operation of any EdgeX service.
type ServiceInfo struct {
	// HealthCheckInterval is the interval for Registry heal check callback
	HealthCheckInterval string
	// Host is the hostname or IP address of the service.
	Host string
	// Port is the HTTP port of the service.
	Port int
	// ServerBindAddr specifies an IP address or hostname
	// for ListenAndServe to bind to, such as 0.0.0.0
	ServerBindAddr string
	// StartupMsg specifies a string to log once service
	// initialization and startup is completed.
	StartupMsg string
	// MaxResultCount specifies the maximum size list supported
	// in response to REST calls to other services.
	MaxResultCount int
	// MaxRequestSize defines the maximum size of http request body in bytes
	MaxRequestSize int64
	// RequestTimeout specifies a timeout (in milliseconds) for
	// processing REST request calls from other services.
	RequestTimeout string
	// CORSConfiguration defines the cross-origin resource sharing related settings
	CORSConfiguration CORSConfigurationInfo
}

// HealthCheck is a URL specifying a health check REST endpoint used by the Registry to determine if the
// service is available.
func (s ServiceInfo) HealthCheck() string {
	hc := fmt.Sprintf("%s://%s:%v%s", "http", s.Host, s.Port, common.ApiPingRoute)
	return hc
}

// Url provides a way to obtain the full url of the host service for use in initialization or, in some cases,
// responses to a caller.
func (s ServiceInfo) Url() string {
	url := fmt.Sprintf("%s://%s:%v", DefaultHttpProtocol, s.Host, s.Port)
	return url
}

// CORSConfigurationInfo defines the cross-origin resource sharing related settings
type CORSConfigurationInfo struct {
	// EnableCORS indicates whether enables CORS in this service
	EnableCORS bool
	// CORSAllowCredentials defines the value of Access-Control-Allow-Credentials in the response header.
	// The Access-Control-Allow-Credentials response header tells browsers whether to expose the response
	// to the frontend JavaScript code when the request's credentials mode (Request.credentials) is included.
	CORSAllowCredentials bool
	// CORSAllowedOrigin defines the value of Access-Control-Allow-Origin in the response header.
	// The Access-Control-Allow-Origin response header indicates whether the response can be shared
	// with requesting code from the given origin.
	CORSAllowedOrigin string
	// CORSAllowedMethods defines the value of Access-Control-Allow-Methods in the response header.
	// The Access-Control-Allow-Methods response header specifies one or more methods allowed when
	// accessing a resource in response to a preflight request.
	CORSAllowedMethods string
	// CORSAllowedHeaders defines the value of Access-Control-Allow-Headers in the response header.
	// The Access-Control-Allow-Headers response header is used in response to a preflight request which
	// includes the Access-Control-Request-Headers to indicate which HTTP headers can be used during the actual request.
	CORSAllowedHeaders string
	// CORSExposeHeaders defines the value of Access-Control-Expose-Headers in the response header
	// The Access-Control-Expose-Headers response header allows a server to indicate which response headers
	// should be made available to scripts running in the browser, in response to a cross-origin request.
	CORSExposeHeaders string
	// CORSMaxAge defines the value of Access-Control-Max-Age in the response header.
	// The Access-Control-Max-Age response header indicates how long the results of a preflight request can be cached.
	CORSMaxAge int
}

// ConfigProviderInfo defines the type and location (via host/port) of the desired configuration provider (e.g. Consul, Eureka)
type ConfigProviderInfo struct {
	Host string
	Port int
	Type string
}

// RegistryInfo defines the type and location (via host/port) of the desired service registry (e.g. Consul, Eureka)
type RegistryInfo struct {
	// Host is the host name where the Registry client is running
	Host string
	// Port is the port number that the Registry client is listening
	Port int
	// Type is the type of Registry client to use, i.e. 'consul'
	Type string
}

// ClientInfo provides the host and port of another service in the eco-system.
type ClientInfo struct {
	// Host is the hostname or IP address of a service.
	Host string
	// Port defines the port on which to access a given service
	Port int
	// Protocol indicates the protocol to use when accessing a given service
	Protocol string
}

func (c ClientInfo) Url() string {
	url := fmt.Sprintf("%s://%s:%v", c.Protocol, c.Host, c.Port)
	return url
}

// SecretStoreInfo encapsulates configuration properties used to create a SecretClient.
type SecretStoreInfo struct {
	Type           string
	Host           string
	Port           int
	Path           string
	Protocol       string
	Namespace      string
	RootCaCertPath string
	ServerName     string
	Authentication types.AuthenticationInfo
	// TokenFile provides a location to a token file.
	TokenFile string
	// SecretsFile is optional Path to JSON file containing secrets to seed into service's SecretStore
	SecretsFile string
	// DisableScrubSecretsFile specifies to not scrub secrets file after importing. Service will fail start-up if
	// not disabled and file can not be written.
	DisableScrubSecretsFile bool
}

type Database struct {
	Type    string
	Timeout int
	Host    string
	Port    int
	Name    string
}

// Credentials encapsulates username-password attributes.
type Credentials struct {
	Username string
	Password string
}

//CertKeyPair encapsulates public certificate/private key pair for an SSL certificate
type CertKeyPair struct {
	Cert string
	Key  string
}

// InsecureSecrets is used to hold the secrets stored in the configuration
type InsecureSecrets map[string]InsecureSecretsInfo

// InsecureSecretsInfo encapsulates info used to retrieve insecure secrets
type InsecureSecretsInfo struct {
	Path    string
	Secrets map[string]string
}

// BootstrapConfiguration defines the configuration elements required by the bootstrap.
type BootstrapConfiguration struct {
	Clients     map[string]ClientInfo
	Service     ServiceInfo
	Config      ConfigProviderInfo
	Registry    RegistryInfo
	SecretStore SecretStoreInfo
}

// MessageBusInfo provides parameters related to connecting to a message bus as a publisher
type MessageBusInfo struct {
	// Indicates the message bus implementation to use, i.e. zero, mqtt, redisstreams...
	Type string
	// Protocol indicates the protocol to use when accessing the message bus.
	Protocol string
	// Host is the hostname or IP address of the broker, if applicable.
	Host string
	// Port defines the port on which to access the message bus.
	Port int
	// PublishTopicPrefix indicates the topic prefix the data is published to.
	PublishTopicPrefix string
	// SubscribeTopic indicates the topic in which to subscribe.
	SubscribeTopic string
	// AuthMode specifies the type of secure connection to the message bus which are 'none', 'usernamepassword'
	// 'clientcert' or 'cacert'. Not all option supported by each implementation.
	// ZMQ doesn't support any Authmode beyond 'none', RedisStreams only supports 'none' & 'usernamepassword'
	// while MQTT supports all options.
	AuthMode string
	// SecretName is the name of the secret in the SecretStore that contains the Auth Credentials. The credential are
	// dynamically loaded using this name and store the Option property below where the implementation expected to
	// find them.
	SecretName string
	// Provides additional configuration properties which do not fit within the existing field.
	// Typically the key is the name of the configuration property and the value is a string representation of the
	// desired value for the configuration property.
	Optional map[string]string
	// SubscribeEnabled indicates whether enable the subscription to the Message Queue
	SubscribeEnabled bool
}

// URL constructs a URL from the protocol, host and port and returns that as a string.
func (p MessageBusInfo) URL() string {
	return fmt.Sprintf("%s://%s:%v", p.Protocol, p.Host, p.Port)
}
