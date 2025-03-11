/*******************************************************************************
 * Copyright (C) 2023 Intel Corporation
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

package interfaces

import (
	"net"
	"net/http"
	"time"
)

// SecretProvider defines the contract for secret provider implementations that
// allow secrets to be retrieved/stored from/to a services Secret Store and other secret related APIs.
// This interface is limited to the APIs that individual service code need.
type SecretProvider interface {
	// StoreSecret stores new secrets into the service's SecretStore at the specified secretName.
	StoreSecret(secretName string, secrets map[string]string) error

	// GetSecret retrieves secrets from the service's SecretStore at the specified secretName.
	GetSecret(secretName string, keys ...string) (map[string]string, error)

	// SecretsLastUpdated returns the last time secrets were updated
	SecretsLastUpdated() time.Time

	// ListSecretNames returns a list of secretNames for the current service from an insecure/secure secret store.
	ListSecretNames() ([]string, error)

	// HasSecret returns true if the service's SecretStore contains a secret at the specified secretName.
	HasSecret(secretName string) (bool, error)

	// RegisterSecretUpdatedCallback registers a callback for a secret. If you specify 'SecretNameWildcard'
	// as the secretName, then the callback will be called for any updated secret. Callbacks set for a specific
	// secretName are given a higher precedence over wildcard ones, and will be called instead of the wildcard one
	// if both are present.
	RegisterSecretUpdatedCallback(secretName string, callback func(secretName string)) error

	// DeregisterSecretUpdatedCallback removes a secret's registered callback secretName.
	DeregisterSecretUpdatedCallback(secretName string)
}

// SecretProviderExt defines the extended contract for secret provider implementations that
// provide additional APIs needed only from the bootstrap code.
type SecretProviderExt interface {
	SecretProvider

	// SecretsUpdated sets the secrets last updated time to current time.
	SecretsUpdated()

	// SecretUpdatedAtSecretName performs updates and callbacks for an updated secret or secretName.
	SecretUpdatedAtSecretName(secretName string)

	// GetMetricsToRegister returns all metric objects that needs to be registered.
	GetMetricsToRegister() map[string]interface{}

	// GetSelfJWT returns an encoded JWT for the current identity-based secret store token
	GetSelfJWT() (string, error)

	// IsJWTValid evaluates a given JWT and returns a true/false if the JWT is valid (i.e. belongs to us and current) or not
	IsJWTValid(jwt string) (bool, error)

	// HttpTransport returns the http.RoundTripper to be used by http-based clients
	HttpTransport() http.RoundTripper

	// SetHttpTransport sets the http.RoundTripper to be used by http-based clients
	SetHttpTransport(rt http.RoundTripper)

	// FallbackDialer returns the dialer to use to establish connections when there is no zero trust service found/authorized
	FallbackDialer() *net.Dialer

	// SetFallbackDialer sets the dialer to use to establish connections when there is no zero trust service found/authorized
	SetFallbackDialer(dialer *net.Dialer)

	// IsZeroTrustEnabled returns whether zero trust principles are enabled
	IsZeroTrustEnabled() bool

	// EnableZeroTrust marks the provider as being zero trust enabled
	EnableZeroTrust()
}
