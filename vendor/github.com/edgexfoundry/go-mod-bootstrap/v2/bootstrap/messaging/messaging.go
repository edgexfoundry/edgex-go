/*******************************************************************************
 * Copyright 2021 Intel Corp.
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

// Package messaging contains common constants and utilities functions used when setting up Secure MessageBus.
// A common bootstrap handler can not be here due to the fact that it would pull in dependency on go-mod-messaging
// which is dependent on ZMQ. This causes all services that use go-mod-boostrap to them have a dependency on ZMQ which
// breaks the security bootstrapping.
package messaging

import (
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	AuthModeKey   = "authmode"
	SecretNameKey = "secretname"

	AuthModeNone             = "none"
	AuthModeUsernamePassword = "usernamepassword"
	AuthModeCert             = "clientcert"
	AuthModeCA               = "cacert"

	SecretUsernameKey = "username"
	SecretPasswordKey = "password"
	SecretClientKey   = "clientkey"
	SecretClientCert  = AuthModeCert
	SecretCACert      = AuthModeCA

	OptionsUsernameKey     = "Username"
	OptionsPasswordKey     = "Password"
	OptionsCertPEMBlockKey = "CertPEMBlock"
	OptionsKeyPEMBlockKey  = "KeyPEMBlock"
	OptionsCaPEMBlockKey   = "CaPEMBlock"
)

type SecretDataProvider interface {
	// GetSecret retrieves secrets from the service's SecretStore at the specified path.
	GetSecret(path string, keys ...string) (map[string]string, error)
}

type SecretData struct {
	Username     string
	Password     string
	KeyPemBlock  []byte
	CertPemBlock []byte
	CaPemBlock   []byte
}

func SetOptionsAuthData(messageBusInfo *config.MessageBusInfo, lc logger.LoggingClient, dic *di.Container) error {
	lc.Infof("Setting options for secure MessageBus with AuthMode='%s' and SecretName='%s",
		messageBusInfo.AuthMode,
		messageBusInfo.SecretName)

	secretProvider := container.SecretProviderFrom(dic.Get)
	if secretProvider == nil {
		return errors.New("secret provider is missing. Make sure it is specified to be used in bootstrap.Run()")
	}

	secretData, err := GetSecretData(messageBusInfo.AuthMode, messageBusInfo.SecretName, secretProvider)
	if err != nil {
		return fmt.Errorf("Unable to get Secret Data for secure message bus: %w", err)
	}

	if err := ValidateSecretData(messageBusInfo.AuthMode, messageBusInfo.SecretName, secretData); err != nil {
		return fmt.Errorf("Secret Data for secure message bus invalid: %w", err)
	}

	if messageBusInfo.Optional == nil {
		messageBusInfo.Optional = map[string]string{}
	}

	// Since already validated, these are the only modes that can be set at this point.
	switch messageBusInfo.AuthMode {
	case AuthModeUsernamePassword:
		messageBusInfo.Optional[OptionsUsernameKey] = secretData.Username
		messageBusInfo.Optional[OptionsPasswordKey] = secretData.Password
	case AuthModeCert:
		messageBusInfo.Optional[OptionsCertPEMBlockKey] = string(secretData.CertPemBlock)
		messageBusInfo.Optional[OptionsKeyPEMBlockKey] = string(secretData.KeyPemBlock)
	case AuthModeCA:
		messageBusInfo.Optional[OptionsCaPEMBlockKey] = string(secretData.CaPemBlock)
	}

	return nil
}

func GetSecretData(authMode string, secretName string, provider SecretDataProvider) (*SecretData, error) {
	// No Auth? No Problem!...No secrets required.
	if authMode == AuthModeNone {
		return nil, nil
	}

	secrets, err := provider.GetSecret(secretName)
	if err != nil {
		return nil, err
	}
	data := &SecretData{
		Username:     secrets[SecretUsernameKey],
		Password:     secrets[SecretPasswordKey],
		KeyPemBlock:  []byte(secrets[SecretClientKey]),
		CertPemBlock: []byte(secrets[SecretClientCert]),
		CaPemBlock:   []byte(secrets[SecretCACert]),
	}

	return data, nil
}

func ValidateSecretData(authMode string, secretName string, secretData *SecretData) error {
	switch authMode {
	case AuthModeUsernamePassword:
		if secret.IsSecurityEnabled() && (secretData.Username == "" || secretData.Password == "") {
			return fmt.Errorf("AuthModeUsernamePassword selected however Username or Password was not found for secret=%s", secretName)
		}

	case AuthModeCert:
		// need both to make a successful connection
		if len(secretData.KeyPemBlock) <= 0 || len(secretData.CertPemBlock) <= 0 {
			return fmt.Errorf("AuthModeCert selected however the key or cert PEM block was not found for secret=%s", secretName)
		}

	case AuthModeCA:
		if len(secretData.CaPemBlock) <= 0 {
			return fmt.Errorf("AuthModeCA selected however no PEM Block was found for secret=%s", secretName)
		}

	case AuthModeNone:
		// Nothing to validate
	default:
		return fmt.Errorf("Invalid AuthMode of '%s' selected", authMode)
	}

	if len(secretData.CaPemBlock) > 0 {
		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(secretData.CaPemBlock)
		if !ok {
			return errors.New("Error parsing CA Certificate")
		}
	}

	return nil
}
