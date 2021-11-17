/********************************************************************************
 *  Copyright 2019 Dell Inc.
 *  Copyright 2021 Intel Corp.
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
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/v2/secrets"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
)

// NewSecretProvider creates a new fully initialized the Secret Provider.
func NewSecretProvider(
	configuration interfaces.Configuration,
	ctx context.Context,
	startupTimer startup.Timer,
	dic *di.Container) (interfaces.SecretProvider, error) {
	lc := container.LoggingClientFrom(dic.Get)

	var provider interfaces.SecretProvider

	switch IsSecurityEnabled() {
	case true:
		// attempt to create a new Secure client only if security is enabled.
		var err error

		lc.Info("Creating SecretClient")

		secretStoreConfig := configuration.GetBootstrap().SecretStore

		for startupTimer.HasNotElapsed() {
			var secretConfig types.SecretConfig

			lc.Info("Reading secret store configuration and authentication token")

			tokenLoader := container.AuthTokenLoaderFrom(dic.Get)
			if tokenLoader == nil {
				tokenLoader = authtokenloader.NewAuthTokenLoader(fileioperformer.NewDefaultFileIoPerformer())
			}

			secretConfig, err = getSecretConfig(secretStoreConfig, tokenLoader)
			if err == nil {
				secureProvider := NewSecureProvider(ctx, configuration, lc, tokenLoader)
				var secretClient secrets.SecretClient

				lc.Info("Attempting to create secret client")
				secretClient, err = secrets.NewSecretsClient(ctx, secretConfig, lc, secureProvider.DefaultTokenExpiredCallback)
				if err == nil {
					secureProvider.SetClient(secretClient)
					provider = secureProvider
					lc.Info("Created SecretClient")

					lc.Debugf("SecretsFile is '%s'", secretConfig.SecretsFile)

					if len(strings.TrimSpace(secretConfig.SecretsFile)) == 0 {
						lc.Infof("SecretsFile not set, skipping seeding of service secrets.")
						break
					}

					provider = secureProvider
					lc.Info("Created SecretClient")

					err = secureProvider.LoadServiceSecrets(secretStoreConfig)
					if err != nil {
						return nil, err
					}
					break
				}
			}

			lc.Warn(fmt.Sprintf("Retryable failure while creating SecretClient: %s", err.Error()))
			startupTimer.SleepForInterval()
		}

		if err != nil {
			return nil, fmt.Errorf("unable to create SecretClient: %s", err.Error())
		}

	case false:
		provider = NewInsecureProvider(configuration, lc)
	}

	dic.Update(di.ServiceConstructorMap{
		container.SecretProviderName: func(get di.Get) interface{} {
			return provider
		},
	})

	return provider, nil
}

// getSecretConfig creates a SecretConfig based on the SecretStoreInfo configuration properties.
// If a token file is present it will override the Authentication.AuthToken value.
func getSecretConfig(secretStoreInfo config.SecretStoreInfo, tokenLoader authtokenloader.AuthTokenLoader) (types.SecretConfig, error) {
	secretConfig := types.SecretConfig{
		Type:           secretStoreInfo.Type, // Type of SecretStore implementation, i.e. Vault
		Host:           secretStoreInfo.Host,
		Port:           secretStoreInfo.Port,
		Path:           addEdgeXSecretPathPrefix(secretStoreInfo.Path),
		SecretsFile:    secretStoreInfo.SecretsFile,
		Protocol:       secretStoreInfo.Protocol,
		Namespace:      secretStoreInfo.Namespace,
		RootCaCertPath: secretStoreInfo.RootCaCertPath,
		ServerName:     secretStoreInfo.ServerName,
		Authentication: secretStoreInfo.Authentication,
	}

	if !IsSecurityEnabled() || secretStoreInfo.TokenFile == "" {
		return secretConfig, nil
	}

	token, err := tokenLoader.Load(secretStoreInfo.TokenFile)
	if err != nil {
		return secretConfig, err
	}

	secretConfig.Authentication.AuthToken = token
	return secretConfig, nil
}

func addEdgeXSecretPathPrefix(secretPath string) string {
	trimmedSecretPath := strings.TrimSpace(secretPath)

	// in this case, treat it as no secret path
	if len(trimmedSecretPath) == 0 {
		return ""
	}

	return "/" + path.Join("v1", "secret", "edgex", trimmedSecretPath) + "/"
}
