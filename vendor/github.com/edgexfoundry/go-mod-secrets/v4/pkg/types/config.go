/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2022 Intel Corp.
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

package types

import (
	"fmt"
	"net/url"
	"strings"
)

// SecretConfig contains configuration settings used to communicate with an HTTP based secret provider
type SecretConfig struct {
	Type string
	Host string
	Port int
	// BasePath is the base path to the secret's location in the secret store
	BasePath string
	// SecretsFile is path to optional JSON file containing secrets to seed into service's SecretStore
	SecretsFile    string
	Protocol       string
	Namespace      string
	RootCaCertPath string
	ServerName     string
	Authentication AuthenticationInfo
	// RuntimeTokenProvider could be optional if not using delayed start from a runtime token provider
	RuntimeTokenProvider RuntimeTokenProviderInfo
}

// BuildURL constructs a URL which can be used to identify a HTTP based secret provider
func (c SecretConfig) BuildURL(path string) (string, error) {
	return buildURL(c.Protocol, c.Host, path, c.Port)
}

// BuildSecretNameURL constructs a URL to the service's secret with in it's secret store
// secretName is the name of the secret in the service's secret store
func (c SecretConfig) BuildSecretNameURL(secretName string) (string, error) {
	return c.BuildURL(fmt.Sprintf("%s/%s", c.BasePath, secretName))
}

// BuildRequestURL constructs a request URL for send the a request to the secrets engine
func (c SecretConfig) BuildRequestURL(subPath string) (string, error) {
	return c.BuildURL(fmt.Sprintf("%s%s", c.BasePath, subPath))
}

// IsRuntimeProviderEnabled returns whether the token provider is using runtime token mechanism
func (c SecretConfig) IsRuntimeProviderEnabled() bool {
	return c.RuntimeTokenProvider.Enabled
}

// AuthenticationInfo contains authentication information to be used when communicating with an HTTP based provider
type AuthenticationInfo struct {
	AuthType  string
	AuthToken string
}

// RuntimeTokenProviderInfo contains the information about the server of a runtime secret token provider
type RuntimeTokenProviderInfo struct {
	Enabled        bool
	Protocol       string
	Host           string
	Port           int
	TrustDomain    string
	EndpointSocket string
	// comma-separated list of required secrets for the service
	// currently we have redis in a typical use case
	RequiredSecrets string
}

func (provider RuntimeTokenProviderInfo) BuildProviderURL(path string) (string, error) {
	return buildURL(provider.Protocol, provider.Host, path, provider.Port)
}

func buildURL(protocol, host, path string, portNum int) (string, error) {
	// Make sure there is not a trailing slash
	path = strings.TrimSuffix(path, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if len(protocol) == 0 {
		return "", fmt.Errorf("unable to build URL: Protocol not set. Please check configuration settings")
	}

	if len(host) == 0 {
		return "", fmt.Errorf("unable to build URL: Host not set. Please check configuration settings")
	}

	if portNum == 0 {
		return "", fmt.Errorf("unable to build URL: Port not set. Please check configuration settings")
	}

	builtUrl := fmt.Sprintf("%s://%s:%v%s", protocol, host, portNum, path)
	_, err := url.Parse(builtUrl)
	if err != nil {
		return "", fmt.Errorf(
			"URL '%s' built from settings is invalid: %s. Please check you configuration settings",
			builtUrl,
			err.Error())
	}

	return builtUrl, err
}
