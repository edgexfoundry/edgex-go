/********************************************************************************
 *  Copyright 2019 Dell Inc.
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
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/security/authtokenloader"
	"github.com/edgexfoundry/edgex-go/internal/security/fileioperformer"

	"github.com/edgexfoundry/go-mod-secrets/pkg"

	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"
)

// SecretClientBootstrapHandler creates a secretClient to be used for obtaining secrets from a secret store manager.
// NOTE: This BootstrapHandler is responsible for creating a utility that will most likely be used by other
// BootstrapHandlers to obtain sensitive data, such as database credentials. This BootstrapHandler should be processed
// before other BootstrapHandlers, possibly even first since it has not other dependencies.
func BootstrapHandler(
	wg *sync.WaitGroup,
	context context.Context,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	// check for environment variable that turns security off
	if env := os.Getenv("EDGEX_SECURITY_SECRET_STORE"); env == "false" {
		dic.Update(di.ServiceConstructorMap{
			container.SecretClientName: func(get di.Get) interface{} {
				return nil
			},
		})
		return true
	}

	// attempt to create a new secret client
	configuration := container.ConfigurationFrom(dic.Get)
	scc, err := getSecretConfig(configuration.GetBootstrap().SecretStore)
	if err != nil {
		return false
	}

	secretClient, err := vault.NewSecretClient(scc)

	var result *pkg.SecretClient
	if err == nil {
		result = &secretClient
	}
	dic.Update(di.ServiceConstructorMap{
		container.SecretClientName: func(get di.Get) interface{} {
			return result
		},
	})

	return err == nil
}

// getSecretConfig creates a SecretConfig based on the SecretStoreInfo configuration properties.
// If a tokenfile is present it will override the Authentication.AuthToken value.
func getSecretConfig(secretStoreInfo config.SecretStoreInfo) (vault.SecretConfig, error) {
	secretConfig := vault.SecretConfig{
		Host:           secretStoreInfo.Host,
		Port:           secretStoreInfo.Port,
		Path:           secretStoreInfo.Path,
		Protocol:       secretStoreInfo.Protocol,
		Namespace:      secretStoreInfo.Namespace,
		RootCaCertPath: secretStoreInfo.RootCaCertPath,
		ServerName:     secretStoreInfo.ServerName,
		Authentication: secretStoreInfo.Authentication,
	}

	if secretStoreInfo.TokenFile == "" {
		return secretConfig, nil
	}

	fileIoPerformer := fileioperformer.NewDefaultFileIoPerformer()
	authTokenLoader := authtokenloader.NewAuthTokenLoader(fileIoPerformer)

	token, err := authTokenLoader.Load(secretStoreInfo.TokenFile)
	if err != nil {
		return secretConfig, err
	}

	secretConfig.Authentication.AuthToken = token

	return secretConfig, nil
}
