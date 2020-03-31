/*******************************************************************************
 * Copyright 2020 Redis Labs
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
 *
 * @author: Diana Atanasova
 * @author: Andre Srinivasan
 *******************************************************************************/
package secretread

import (
	"context"
	"fmt"
	"os"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"
	"github.com/edgexfoundry/go-mod-secrets/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/pkg/token/fileioperformer"
)

func GetCredentials(lc logger.LoggingClient, secureConfig Configuration) (map[string]DatabaseInfo, error) {
	if os.Getenv(SecretStore) == "false" {
		return secureConfig.Databases, nil
	}

	token, err := getAccessToken(secureConfig.SecretStore.TokenFile)
	if err != nil {
		return nil, err
	}

	bkgCtx := context.Background()
	factory := vault.NewSecretClientFactory()
	secretClient, err := factory.NewSecretClient(bkgCtx, vault.SecretConfig{
		Port:                    secureConfig.SecretStore.Port,
		Host:                    secureConfig.SecretStore.Host,
		Path:                    secureConfig.SecretStore.Path,
		Protocol:                "https",
		RootCaCertPath:          secureConfig.SecretStore.RootCaCertPath,
		ServerName:              secureConfig.SecretStore.ServerName,
		Authentication:          vault.AuthenticationInfo{AuthType: VaultToken, AuthToken: token},
		AdditionalRetryAttempts: secureConfig.SecretStore.AdditionalRetryAttempts,
		RetryWaitPeriod:         secureConfig.SecretStore.RetryWaitPeriod,
	}, lc, nil)

	if err != nil {
		lc.Error(fmt.Sprintf("failed to connect to secret store: %v", err.Error()))
		return nil, err
	}

	var credentials = make(map[string]DatabaseInfo)
	for _, dbName := range getDatabaseNames(secureConfig) {
		lc.Debug(fmt.Sprintf("reading secrets from '%s/%s' path", secureConfig.SecretStore.Path, dbName))
		secrets, err := secretClient.GetSecrets("/"+dbName, "username", "password")
		if err != nil {
			lc.Error(fmt.Sprintf("failed to read secret stores data for '%s/%s' path: %s", secureConfig.SecretStore.Path, dbName, err.Error()))
			return nil, err
		}
		crInfo := DatabaseInfo{Username: secrets["username"], Password: secrets["password"]}
		credentials[dbName] = crInfo
	}
	secureConfig.Databases = credentials
	lc.Debug("Credentials successfully read from Secret Store")

	return credentials, nil
}

func getDatabaseNames(secureConfig Configuration) []string {
	databases := make([]string, len(secureConfig.Databases))
	i := 0
	for dbName := range secureConfig.Databases {
		databases[i] = dbName
		i++
	}
	return databases
}

func getAccessToken(filename string) (string, error) {
	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenLoader := authtokenloader.NewAuthTokenLoader(fileOpener)
	token, err := tokenLoader.Load(filename)
	if err != nil {
		return "", err
	}

	return token, nil
}
