/*******************************************************************************
 * Copyright 2020 Intel Inc.
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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

// InsecureProvider implements the SecretProvider interface for insecure secrets
type InsecureProvider struct {
	lc            logger.LoggingClient
	configuration interfaces.Configuration
	lastUpdated   time.Time
}

// NewInsecureProvider creates, initializes Provider for insecure secrets.
func NewInsecureProvider(config interfaces.Configuration, lc logger.LoggingClient) *InsecureProvider {
	return &InsecureProvider{
		configuration: config,
		lc:            lc,
		lastUpdated:   time.Now(),
	}
}

// GetSecret retrieves secrets from a Insecure Secrets secret store.
// path specifies the type or location of the secrets to retrieve.
// keys specifies the secrets which to retrieve. If no keys are provided then all the keys associated with the
// specified path will be returned.
func (p *InsecureProvider) GetSecret(path string, keys ...string) (map[string]string, error) {
	results := make(map[string]string)
	pathExists := false
	var missingKeys []string

	insecureSecrets := p.configuration.GetInsecureSecrets()
	if insecureSecrets == nil {
		err := fmt.Errorf("InsecureSecrets missing from configuration")
		return nil, err
	}

	for _, insecureSecret := range insecureSecrets {
		if insecureSecret.Path == path {
			if len(keys) == 0 {
				// If no keys are provided then all the keys associated with the specified path will be returned
				for k, v := range insecureSecret.Secrets {
					results[k] = v
				}
				return results, nil
			}

			pathExists = true
			for _, key := range keys {
				value, keyExists := insecureSecret.Secrets[key]
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

	if !pathExists {
		// if path is not in secret store
		err := fmt.Errorf("Error, path (%v) doesn't exist in secret store", path)
		return nil, err
	}

	return results, nil
}

// StoreSecret stores the secrets, but is not supported for Insecure Secrets
func (p *InsecureProvider) StoreSecret(_ string, _ map[string]string) error {
	return errors.New("storing secrets is not supported when running in insecure mode")
}

// SecretsUpdated resets LastUpdate time for the Insecure Secrets.
func (p *InsecureProvider) SecretsUpdated() {
	p.lastUpdated = time.Now()
}

// SecretsLastUpdated returns the last time insecure secrets were updated
func (p *InsecureProvider) SecretsLastUpdated() time.Time {
	return p.lastUpdated
}

// GetAccessToken returns the AccessToken for the specified type, which in insecure mode is not need
// so just returning an empty token.
func (p *InsecureProvider) GetAccessToken(_ string, _ string) (string, error) {
	return "", nil
}
