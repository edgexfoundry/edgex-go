/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright 2023 Intel Corporation
 * Copyright 2021-2025 IOTech Ltd.
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
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/v4/secrets"
)

const (
	DefaultHttpProtocol = "http"
)

const (
	ServiceTypeApp    = "app-service"
	ServiceTypeDevice = "device-service"
	ServiceTypeOther  = "other"
	//TOOD: add security-service to use in place of useSecretProvider
)

const (
	CommonConfigDone = "IsCommonConfigReady"
)

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
	// MaxRequestSize defines the maximum size of http request body in kilobytes
	MaxRequestSize int64
	// RequestTimeout specifies a timeout (in milliseconds) for
	// processing REST request calls from other services.
	RequestTimeout string
	// EnableNameFieldEscape indicates whether enables NameFieldEscape in this service
	// The name field escape could allow the system to use special or Chinese characters in the different name fields, including device, profile, and so on.  If the EnableNameFieldEscape is false, some special characters might cause system error.
	// TODO: remove in EdgeX 4.0
	EnableNameFieldEscape bool
	// CORSConfiguration defines the cross-origin resource sharing related settings
	CORSConfiguration CORSConfigurationInfo
	// SecurityOptions is a key/value map, used for configuring hosted services. Currently used for zero trust but
	// could be for other options additional security related configuration
	SecurityOptions map[string]string
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

// RegistryInfo defines the type and location (via host/port) of the desired service registry (e.g. keeper, Eureka)
type RegistryInfo struct {
	// Host is the host name where the Registry client is running
	Host string
	// Port is the port number that the Registry client is listening
	Port int
	// Type is the type of Registry client to use, i.e. 'keeper'
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
	// UseMessageBus indicates weather to use Messaging version of client
	UseMessageBus bool
	// SecurityOptions is a key/value map, used for configuring clients. Currently used for zero trust but
	// could be for other options additional security related configuration
	SecurityOptions map[string]string
}

func (c ClientInfo) Url() string {
	url := fmt.Sprintf("%s://%s:%v", c.Protocol, c.Host, c.Port)
	return url
}

func NewSecretStoreSetupClientInfo() *ClientsCollection {
	secretStoreStepClient := ClientsCollection{
		common.SecuritySecretStoreSetupServiceKey: &ClientInfo{
			Host:     "localhost",
			Port:     59843,
			Protocol: "http",
		}}
	return &secretStoreStepClient
}

// SecretStoreInfo encapsulates configuration properties used to create a SecretClient.
type SecretStoreInfo struct {
	Type           string
	Host           string
	Port           int
	StoreName      string
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

	// RuntimeTokenProvider is optional if not using delayed start from spiffe-token provider
	RuntimeTokenProvider types.RuntimeTokenProviderInfo
}

func NewSecretStoreInfo(serviceKey string) SecretStoreInfo {
	return SecretStoreInfo{
		Type:                    secrets.DefaultSecretStore,
		Protocol:                "http",
		Host:                    "localhost",
		Port:                    8200,
		StoreName:               serviceKey,
		TokenFile:               fmt.Sprintf("/tmp/edgex/secrets/%s/secrets-token.json", serviceKey),
		DisableScrubSecretsFile: false,
		Namespace:               "",
		RootCaCertPath:          "",
		ServerName:              "",
		SecretsFile:             "",
		Authentication: types.AuthenticationInfo{
			AuthType:  "X-Vault-Token",
			AuthToken: "",
		},
		RuntimeTokenProvider: types.RuntimeTokenProviderInfo{
			Enabled:         false,
			Protocol:        "https",
			Host:            "localhost",
			Port:            59841,
			TrustDomain:     "edgexfoundry.org",
			EndpointSocket:  "/tmp/edgex/secrets/spiffe/public/api.sock",
			RequiredSecrets: "redisdb",
		},
	}
}

type Database struct {
	Type    string
	Timeout string
	Host    string
	Port    int
	Name    string
}

// Credentials encapsulates username-password attributes.
type Credentials struct {
	Username string
	Password string
}

// CertKeyPair encapsulates public certificate/private key pair for an SSL certificate
type CertKeyPair struct {
	Cert string
	Key  string
}

// InsecureSecrets is used to hold the secrets stored in the configuration
type InsecureSecrets map[string]InsecureSecretsInfo

// InsecureSecretsInfo encapsulates info used to retrieve insecure secrets
type InsecureSecretsInfo struct {
	SecretName string
	SecretData map[string]string
}

// ClientsCollection is a collection of Client information for communicating to dependent clients.
type ClientsCollection map[string]*ClientInfo

// BootstrapConfiguration defines the configuration elements required by the bootstrap.
type BootstrapConfiguration struct {
	Clients      *ClientsCollection
	Service      *ServiceInfo
	Config       *ConfigProviderInfo
	Registry     *RegistryInfo
	MessageBus   *MessageBusInfo
	Database     *Database
	ExternalMQTT *ExternalMQTTInfo
}

// MessageBusInfo provides parameters related to connecting to the EdgeX MessageBus
type MessageBusInfo struct {
	// Disabled indicates if the use of the EdgeX MessageBus is disabled.
	Disabled bool
	// Indicates the message bus implementation to use, i.e. mqtt, nats...
	Type string
	// Protocol indicates the protocol to use when accessing the message bus.
	Protocol string
	// Host is the hostname or IP address of the broker, if applicable.
	Host string
	// Port defines the port on which to access the message bus.
	Port int
	// AuthMode specifies the type of secure connection to the message bus which are 'none', 'usernamepassword'
	// 'clientcert' or 'cacert'. MQTT and NATS support all options.
	AuthMode string
	// SecretName is the name of the secret in the SecretStore that contains the Auth Credentials. The credential are
	// dynamically loaded using this name and store the Option property below where the implementation expected to
	// find them.
	SecretName string
	// BaseTopicPrefix is the base topic prefix that all topics start with.
	// If not set the DefaultBaseTopic constant is used.
	BaseTopicPrefix string
	// Provides additional configuration properties which do not fit within the existing field.
	// Typically, the key is the name of the configuration property and the value is a string representation of the
	// desired value for the configuration property.
	Optional map[string]string
}

func (m MessageBusInfo) GetBaseTopicPrefix() string {
	if len(m.BaseTopicPrefix) == 0 {
		return common.DefaultBaseTopic
	}

	return m.BaseTopicPrefix
}

type ExternalMQTTInfo struct {
	// Url contains the fully qualified URL to connect to the MQTT broker
	Url string
	// SubscribeTopics is a comma separated list of topics in which to subscribe
	SubscribeTopics string
	// PublishTopic is the topic to publish pipeline output (if any)
	PublishTopic string
	// Topics allows ExternalMQTTInfo to be more flexible with respect to topics.
	// TODO: move PublishTopic and SubscribeTopics to Topics in EdgeX 3.0
	Topics map[string]string
	// ClientId to connect to the broker with.
	ClientId string
	// ConnectTimeout is a time duration indicating how long to wait timing out on the broker connection
	ConnectTimeout string
	// AutoReconnect indicated whether to retry connection if disconnected
	AutoReconnect bool
	// KeepAlive is seconds between client ping when no active data flowing to avoid client being disconnected
	KeepAlive int64
	// QoS for MQTT Connection
	QoS byte
	// Retain setting for MQTT Connection
	Retain bool
	// SkipCertVerify indicates if the certificate verification should be skipped
	SkipCertVerify bool
	// SecretName is the name of the secret in secret provider to retrieve your secrets
	SecretName string
	// AuthMode indicates what to use when connecting to the broker. Options are "none", "cacert" , "usernamepassword", "clientcert".
	// If a CA Cert exists in the SecretPath then it will be used for all modes except "none".
	AuthMode string
	// RetryDuration indicates how long (in seconds) to wait timing out on the MQTT client creation
	RetryDuration int
	// RetryInterval indicates the time (in seconds) that will be waited between attempts to create MQTT client
	RetryInterval int
	// Enabled determines whether the service needs to connect to the external MQTT broker
	Enabled bool
}

// URL constructs a URL from the protocol, host and port and returns that as a string.
func (p MessageBusInfo) URL() string {
	return fmt.Sprintf("%s://%s:%v", p.Protocol, p.Host, p.Port)
}

// TelemetryInfo contains the configuration for a service's metrics collection
type TelemetryInfo struct {
	// Interval is the time duration in which to collect and report the service's metrics
	Interval string
	// Metrics is the list of service's metrics that can be collected. Each of the service's metrics must be in the list
	// and set to true if enable or false if disabled.
	Metrics map[string]bool
	// Tags is a list of service level tags that are attached to every metric reported for the service
	// Example: Gateway = "Gateway123"
	Tags map[string]string
}

// GetEnabledMetricName returns the matching configured Metric name and if it is enabled.
func (t *TelemetryInfo) GetEnabledMetricName(metricName string) (string, bool) {
	for configMetricName, enabled := range t.Metrics {
		// Match on config metric name as prefix of passed in metric name (service's metric item name)
		// This allows for a class of Metrics to be enabled with one configured metric name.
		// App SDK uses this for PipelineMetrics by appending the pipeline ID to the name
		// of the metric(s) it is collecting for multiple function pipelines.
		if !strings.HasPrefix(metricName, configMetricName) {
			continue
		}

		return configMetricName, enabled
	}

	// Service's metric name did not match any config Metric name.
	return "", false
}
