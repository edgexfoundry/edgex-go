/*******************************************************************************
 * Copyright 2020-2023 Intel Corporation
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

package secret

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	gometrics "github.com/rcrowley/go-metrics"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// InsecureProvider implements the SecretProvider interface for insecure secrets
type InsecureProvider struct {
	lc                        logger.LoggingClient
	configuration             interfaces.Configuration
	lastUpdated               time.Time
	registeredSecretCallbacks map[string]func(secretName string)
	securitySecretsRequested  gometrics.Counter
	securitySecretsStored     gometrics.Counter
	dic                       *di.Container
}

// NewInsecureProvider creates, initializes Provider for insecure secrets.
func NewInsecureProvider(config interfaces.Configuration, lc logger.LoggingClient, dic *di.Container) *InsecureProvider {
	return &InsecureProvider{
		configuration:             config,
		lc:                        lc,
		lastUpdated:               time.Now(),
		registeredSecretCallbacks: make(map[string]func(secretName string)),
		securitySecretsRequested:  gometrics.NewCounter(),
		securitySecretsStored:     gometrics.NewCounter(),
		dic:                       dic,
	}
}

// GetSecret retrieves secrets from a Insecure Secrets secret store.
// secretName specifies the type or location of the secrets to retrieve.
// keys specifies the secrets which to retrieve. If no keys are provided then all the keys associated with the
// specified secretName will be returned.
func (p *InsecureProvider) GetSecret(secretName string, keys ...string) (map[string]string, error) {
	p.securitySecretsRequested.Inc(1)

	results := make(map[string]string)
	secretNameExists := false
	var missingKeys []string

	insecureSecrets := p.configuration.GetInsecureSecrets()
	if insecureSecrets == nil {
		err := fmt.Errorf("InsecureSecrets missing from configuration")
		return nil, err
	}

	for _, insecureSecret := range insecureSecrets {
		if insecureSecret.SecretName == secretName {
			if len(keys) == 0 {
				// If no keys are provided then all the keys associated with the specified secretName will be returned
				for k, v := range insecureSecret.SecretData {
					results[k] = v
				}
				return results, nil
			}

			secretNameExists = true
			for _, key := range keys {
				value, keyExists := insecureSecret.SecretData[key]
				if !keyExists {
					missingKeys = append(missingKeys, key)
					continue
				}
				results[key] = value
			}
		}
	}

	if len(missingKeys) > 0 {
		err := fmt.Errorf("No value for the keys: [%s] exists", strings.Join(missingKeys, ","))
		return nil, err
	}

	if !secretNameExists {
		// if secretName is not in secret store
		err := fmt.Errorf("Error, secretName (%v) doesn't exist in secret store", secretName)
		return nil, err
	}

	return results, nil
}

// StoreSecret attempts to store the secrets in the ConfigurationProvider's InsecureSecrets. If no ConfigurationProvider
// is in use, it will return an error.
//
// Note: This does not call SecretUpdatedAtSecretName, SecretsUpdated, or increase the secrets stored metric because
// those will all occur once the ConfigurationProvider tells the service that the configuration has updated.
func (p *InsecureProvider) StoreSecret(secretName string, secrets map[string]string) error {
	configClient := container.ConfigClientFrom(p.dic.Get)
	if configClient == nil {
		return errors.NewCommonEdgeX(errors.KindNotAllowed, "can't store secrets. ConfigurationProvider is not in use or has not been properly initialized", nil)
	}

	// insert the top-level data about the secret name
	err := configClient.PutConfigurationValue(config.GetInsecureSecretNameFullPath(secretName), []byte(secretName))
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindCommunicationError, "error setting secretName value in the config provider", err)
	}

	// insert each secret key/value pair
	for key, value := range secrets {
		err = configClient.PutConfigurationValue(config.GetInsecureSecretDataFullPath(secretName, key), []byte(value))
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindCommunicationError, "error setting secretData key/value pair in the config provider", err)
		}
	}
	return nil
}

// SecretsUpdated resets LastUpdate time for the Insecure Secrets.
func (p *InsecureProvider) SecretsUpdated() {
	p.lastUpdated = time.Now()
}

// SecretsLastUpdated returns the last time insecure secrets were updated
func (p *InsecureProvider) SecretsLastUpdated() time.Time {
	return p.lastUpdated
}

// HasSecret returns true if the service's SecretStore contains a secret at the specified secretName.
func (p *InsecureProvider) HasSecret(secretName string) (bool, error) {
	insecureSecrets := p.configuration.GetInsecureSecrets()
	if insecureSecrets == nil {
		err := fmt.Errorf("InsecureSecret missing from configuration")
		return false, err
	}

	for _, insecureSecret := range insecureSecrets {
		if insecureSecret.SecretName == secretName {
			return true, nil
		}
	}

	return false, nil
}

// ListSecretNames returns a list of SecretName for the current service from an insecure/secure secret store.
func (p *InsecureProvider) ListSecretNames() ([]string, error) {
	var results []string

	insecureSecrets := p.configuration.GetInsecureSecrets()
	if insecureSecrets == nil {
		err := fmt.Errorf("InsecureSecrets missing from configuration")
		return nil, err
	}

	for _, insecureSecret := range insecureSecrets {
		results = append(results, insecureSecret.SecretName)
	}

	return results, nil
}

// RegisterSecretUpdatedCallback registers a callback for a secret. If you specify secret.WildcardName
// as the secretName, then the callback will be called for any updated secret. Callbacks set for a specific
// secretName are given a higher precedence over wildcard ones, and will be called instead of the wildcard one
// if both are present.
func (p *InsecureProvider) RegisterSecretUpdatedCallback(secretName string, callback func(secretName string)) error {
	if _, ok := p.registeredSecretCallbacks[secretName]; ok {
		return fmt.Errorf("there is a callback already registered for secretName '%v'", secretName)
	}

	// Register new call back for secretName.
	p.registeredSecretCallbacks[secretName] = callback

	return nil
}

// SecretUpdatedAtSecretName performs updates and callbacks for an updated secret or secretName.
func (p *InsecureProvider) SecretUpdatedAtSecretName(secretName string) {
	p.securitySecretsStored.Inc(1)

	p.lastUpdated = time.Now()
	if p.registeredSecretCallbacks == nil {
		return
	}

	// Execute Callback for provided secretName.
	if callback, ok := p.registeredSecretCallbacks[secretName]; ok {
		p.lc.Debugf("invoking callback registered for secretName: '%s'", secretName)
		callback(secretName)

		// if no callback is registered for secretName, see if wildcard callback is provided.
	} else if callback, ok = p.registeredSecretCallbacks[WildcardName]; ok {
		p.lc.Debugf("invoking wildcard callback for secretName: '%s'", secretName)
		callback(secretName)
	}
}

// DeregisterSecretUpdatedCallback removes a secret's registered callback secretName.
func (p *InsecureProvider) DeregisterSecretUpdatedCallback(secretName string) {
	// Remove secretName from map.
	delete(p.registeredSecretCallbacks, secretName)
}

// GetMetricsToRegister returns all metric objects that needs to be registered.
func (p *InsecureProvider) GetMetricsToRegister() map[string]interface{} {
	return map[string]interface{}{
		secretsRequestedMetricName: p.securitySecretsRequested,
		secretsStoredMetricName:    p.securitySecretsStored,
	}
}

// GetSelfJWT returns an encoded JWT for the current identity-based secret store token
func (p *InsecureProvider) GetSelfJWT() (string, error) {
	// If security is disabled, return an empty string
	// It is presumed HTTP invokers will not add an
	// authorization token that is empty to outbound requests.
	return "", nil
}

// IsJWTValid evaluates a given JWT and returns a true/false if the JWT is valid (i.e. belongs to us and current) or not
func (p *InsecureProvider) IsJWTValid(jwt string) (bool, error) {
	return true, nil
}

func (p *InsecureProvider) HttpTransport() http.RoundTripper {
	return http.DefaultTransport
}

func (p *InsecureProvider) SetHttpTransport(_ http.RoundTripper) {
	//empty on purpose
}

func (p *InsecureProvider) IsZeroTrustEnabled() bool {
	return false
}

func (p *InsecureProvider) EnableZeroTrust() {
	//empty on purpose
}

func (p *InsecureProvider) FallbackDialer() *net.Dialer {
	return &net.Dialer{}
}

func (p *InsecureProvider) SetFallbackDialer(_ *net.Dialer) {
	//empty on purpose
}
