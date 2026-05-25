package identity

/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

import (
	"fmt"
	"strings"
)

const (
	ConfigFieldCert           = "cert"
	ConfigFieldKey            = "key"
	ConfigFieldServerCert     = "server_cert"
	ConfigFieldServerKey      = "server_key"
	ConfigFieldAltServerCerts = "alt_server_certs"
	ConfigFieldCa             = "ca"
)

// Config represents the basic data structure for and identity configuration. A Config provides details on where the
// x509 certificates and private keys are located/stored for the identity. These values are interpreted by the
// LoadIdentity function to produce an Identity that can be used to create crypto configurations (i.e. tls.Config).
// Storage locations include files, in-memory PEM, and hardware tokens.
//
// Key, Cert, ServerCert, ServerKey, and CA are URLs with the following schemes: `file`, `pem`. Additionally,
// Key supports `engine`. If the value is not in URL format it is assumed to be `file`.
//
// Example: `file://path/to/my/cert.pem` or `path/to/my/cert.pem'
// Example: `pem://-----BEGIN CERTIFICATE-----\nMIIB/TCCAYCgAwIBAgIBATAMBggqhk...`
//
// Cert must point to a source with a single leaf-first certificate chain and have a corresponding private key in the
// Key field. ServerCert may point to a source with multiple certificate chains for the same server
// (i.e. SNI/ECH support) from different trust roots. If ServerKey is not defined, it is assumed that Key will
// be the private key for all leaf certificates found in ServerCert's source. If ServerKey is defined it is assumed
// that all leaf certificates in ServerCert's source use it as a private key.
//
// The AltServerCerts property allows an array of additional server certificate chains to be loaded
// with different private keys. It is useful for scenarios where the Cert and ServerCert fields are used for
// automated certificates from one system (i.e. Ziti) and alternate server certs are automated through another
// (i.e. Lets Encrypt)

type Config struct {
	Key            string       `json:"key" yaml:"key" mapstructure:"key"`
	Cert           string       `json:"cert" yaml:"cert" mapstructure:"cert"`
	ServerCert     string       `json:"server_cert,omitempty" yaml:"server_cert,omitempty" mapstructure:"server_cert,omitempty"`
	ServerKey      string       `json:"server_key,omitempty" yaml:"server_key,omitempty" mapstructure:"server_key,omitempty"`
	AltServerCerts []ServerPair `json:"alt_server_certs,omitempty" yaml:"alt_server_certs,omitempty" mapstructure:"alt_server_certs,omitempty"`
	CA             string       `json:"ca,omitempty" yaml:"ca,omitempty" mapstructure:"ca"`
}

type ServerPair struct {
	ServerCert string `json:"server_cert,omitempty" yaml:"server_cert,omitempty" mapstructure:"server_cert,omitempty"`
	ServerKey  string `json:"server_key,omitempty" yaml:"server_key,omitempty" mapstructure:"server_key,omitempty"`
}

// Validate validates the current IdentityConfiguration to have non-empty values all fields except
// ServerKey which assumes that Key is a suitable default.
func (config *Config) Validate() error {
	return config.ValidateWithPathContext("")
}

// ValidateWithPathContext performs the same checks as Validate but also allows a path context to be
// provided for error messages when parsing deep or complex configuration.
//
// Example:
//
//	`ValidateWithPathContext("my.path")`  errors would be formatted as "required configuration value [my.path.cert]..."`
func (config *Config) ValidateWithPathContext(pathContext string) error {
	pathContext = strings.TrimSpace(pathContext)

	if pathContext != "" {
		if !strings.HasSuffix(pathContext, ".") {
			pathContext = pathContext + "."
		}
	}

	if config.Cert == "" {
		return fmt.Errorf("required configuration value [%s%s] is missing or is blank", pathContext, ConfigFieldCert)
	}

	if config.ServerCert == "" {
		return fmt.Errorf("required configuration value [%s%s] is missing or is blank", pathContext, ConfigFieldServerCert)
	}

	if config.Key == "" {
		return fmt.Errorf("required configuration value [%s%s] is missing or is blank", pathContext, ConfigFieldKey)
	}

	if config.CA == "" {
		return fmt.Errorf("required configuration value [%s%s] is missing or is blank", pathContext, ConfigFieldCa)
	}

	for i, altServerCerts := range config.AltServerCerts {
		if altServerCerts.ServerCert == "" {
			return fmt.Errorf("required configuration value [%s%s[%d].%s] is missing or is blank", pathContext, ConfigFieldAltServerCerts, i, ConfigFieldServerCert)
		}
	}

	return nil
}

// ValidateForClient validates the current IdentityConfiguration has enough values to initiate a client connection.
// For example: a tls.Config for a client in mTLS
func (config *Config) ValidateForClient() error {
	return config.ValidateForClientWithPathContext("")
}

// ValidateForClientWithPathContext performs the same checks as ValidateForClient but also allows a path context to be
// provided for error messages when parsing deep or complex configuration.
//
// Example:
//
//	`ValidateForClientWithPathContext("my.path")`  errors would be formatted as "required configuration value [my.path.cert]..."`
func (config *Config) ValidateForClientWithPathContext(pathContext string) error {
	pathContext = strings.TrimSpace(pathContext)

	if pathContext != "" {
		if !strings.HasSuffix(pathContext, ".") {
			pathContext = pathContext + "."
		}
	}

	if config.Cert == "" {
		return fmt.Errorf("required configuration value [%s%s] is missing or is blank", pathContext, ConfigFieldCert)
	}

	if config.Key == "" {
		return fmt.Errorf("required configuration value [%s%s] is missing or is blank", pathContext, ConfigFieldKey)
	}

	return nil
}

// ValidateForServer validates the current IdentityConfiguration has enough values to a client connection.
// For example: a tls.Config for a server in mTLS
func (config *Config) ValidateForServer() error {
	return config.ValidateForServerWithPathContext("")
}

// ValidateForServerWithPathContext performs the same checks as ValidateForServer but also allows a path context to be
// provided for error messages when parsing deep or complex configuration.
//
// Example:
//
//	`ValidateWithPathContext("my.path")`  errors would be formatted as "required configuration value [my.path.cert]..."`
func (config *Config) ValidateForServerWithPathContext(pathContext string) error {
	pathContext = strings.TrimSpace(pathContext)

	if pathContext != "" {
		if !strings.HasSuffix(pathContext, ".") {
			pathContext = pathContext + "."
		}
	}

	if config.ServerCert == "" {
		return fmt.Errorf("required configuration value [%s%s] is missing or is blank", pathContext, ConfigFieldServerCert)
	}

	if config.Key == "" && config.ServerKey == "" {
		return fmt.Errorf("required configuration values [%s%s], [%s%s] are both missing or are blank", pathContext, ConfigFieldKey, pathContext, ConfigFieldServerKey)
	}

	if config.CA == "" {
		return fmt.Errorf("required configuration value [%s%s] is missing or is blank", pathContext, ConfigFieldCa)
	}

	for i, altServerCerts := range config.AltServerCerts {
		if altServerCerts.ServerCert == "" {
			return fmt.Errorf("required configuration value [%s%s[%d].%s] is missing or is blank", pathContext, ConfigFieldAltServerCerts, i, ConfigFieldServerCert)
		}
	}

	return nil
}

// NewConfigFromMap will parse a standard identity configuration section that has been loaded from JSON/YAML/etc.
// parse functions that return interface{} maps. It expects the following fields to be defined as strings if present.
// If any fields are missing they are left as empty string in the resulting Config.
func NewConfigFromMap(identityMap map[interface{}]interface{}) (*Config, error) {
	return NewConfigFromMapWithPathContext(identityMap, "")
}

// NewConfigFromMapWithPathContext performs the same checks as NewConfigFromMap but also allows a path context to be
// provided for error messages when parsing deep or complex configuration.
//
// Example:
//
//	`NewConfigFromMapWithPathContext(myMap, "my.path")` errors would be formatted as "value [my.path.cert] must be a string"`
func NewConfigFromMapWithPathContext(identityMap map[interface{}]interface{}, pathContext string) (*Config, error) {
	pathContext = strings.TrimSpace(pathContext)

	if pathContext != "" {
		if !strings.HasSuffix(pathContext, ".") {
			pathContext = pathContext + "."
		}
	}

	idConfig := &Config{}

	if certInterface, ok := identityMap[ConfigFieldCert]; ok {
		if cert, ok := certInterface.(string); ok {
			idConfig.Cert = cert
		} else {
			return nil, fmt.Errorf("value [%s%s] must be a string", pathContext, ConfigFieldCert)
		}
	}

	if serverCertInterface, ok := identityMap[ConfigFieldServerCert]; ok {
		if serverCert, ok := serverCertInterface.(string); ok {
			idConfig.ServerCert = serverCert
		} else {
			return nil, fmt.Errorf("value [%s%s] must be a string", pathContext, ConfigFieldServerCert)
		}
	}

	if keyInterface, ok := identityMap[ConfigFieldKey]; ok {
		if key, ok := keyInterface.(string); ok {
			idConfig.Key = key
		} else {
			return nil, fmt.Errorf("value [%s%s] must be a string", pathContext, ConfigFieldKey)
		}
	}

	if keyInterface, ok := identityMap[ConfigFieldServerKey]; ok {
		if serverKey, ok := keyInterface.(string); ok {
			idConfig.ServerKey = serverKey
		} else {
			return nil, fmt.Errorf("value [%s%s] must be a string", pathContext, ConfigFieldServerKey)
		}
	}

	if altServerCertsInterface, ok := identityMap[ConfigFieldAltServerCerts]; ok {
		if altServerCertsSliceInterface, ok := altServerCertsInterface.([]any); ok {
			for i, altServerCertInterface := range altServerCertsSliceInterface {
				if altServerCertMap, ok := altServerCertInterface.(map[any]any); ok {
					serverCertVal := altServerCertMap[ConfigFieldServerCert]
					serverKeyVal := altServerCertMap[ConfigFieldServerKey]

					serverCert, ok := serverCertVal.(string)

					if !ok {
						return nil, fmt.Errorf("value [%s%s[%d].%s] must be a string", pathContext, ConfigFieldAltServerCerts, i, ConfigFieldServerCert)
					}

					serverKey := ""

					if serverKeyVal != nil {
						serverKey, ok = serverKeyVal.(string)

						if !ok {
							return nil, fmt.Errorf("value [%s%s[%d].%s] must be a string", pathContext, ConfigFieldAltServerCerts, i, ConfigFieldServerKey)
						}
					}

					idConfig.AltServerCerts = append(idConfig.AltServerCerts, ServerPair{
						ServerCert: serverCert,
						ServerKey:  serverKey,
					})

				} else {
					return nil, fmt.Errorf("value [%s%s[%d]] must be an obbject", pathContext, ConfigFieldAltServerCerts, i)
				}
			}
		} else {
			return nil, fmt.Errorf("value [%s%s] must be an array", pathContext, ConfigFieldAltServerCerts)
		}
	} //not required

	if caInterface, ok := identityMap[ConfigFieldCa]; ok {
		if ca, ok := caInterface.(string); ok {
			idConfig.CA = ca
		} else {
			return nil, fmt.Errorf("value [%s%s] must be a string", pathContext, ConfigFieldCa)
		}
	}

	return idConfig, nil
}
