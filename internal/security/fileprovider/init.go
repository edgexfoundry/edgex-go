/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2019 Intel Corporation
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
 *******************************************************************************/

package fileprovider

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/container"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-bootstrap/security/authtokenloader"
	"github.com/edgexfoundry/go-mod-bootstrap/security/fileioperformer"
)

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func Handler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	cfg := container.ConfigurationFrom(dic.Get)
	loggingClient := bootstrapContainer.LoggingClientFrom(dic.Get)

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenProvider := authtokenloader.NewAuthTokenLoader(fileOpener)

	var req internal.HttpCaller
	if caFilePath := cfg.SecretService.CaFilePath; caFilePath != "" {
		loggingClient.Info("using certificate verification for secret store connection")
		caReader, err := fileOpener.OpenFileReader(caFilePath, os.O_RDONLY, 0400)
		if err != nil {
			loggingClient.Error(fmt.Sprintf("failed to load CA certificate: %s", err.Error()))
			return false
		}
		req = secretstoreclient.NewRequestor(loggingClient).WithTLS(caReader, cfg.SecretService.ServerName)
	} else {
		loggingClient.Info("bypassing certificate verification for secret store connection")
		req = secretstoreclient.NewRequestor(loggingClient).Insecure()
	}
	vaultScheme := cfg.SecretService.Scheme
	vaultHost := fmt.Sprintf("%s:%v", cfg.SecretService.Server, cfg.SecretService.Port)
	vaultClient := secretstoreclient.NewSecretStoreClient(loggingClient, req, vaultScheme, vaultHost)

	fileProvider := NewTokenProvider(loggingClient, fileOpener, tokenProvider, vaultClient)

	fileProvider.SetConfiguration(cfg.SecretService, cfg.TokenFileProvider)
	err := fileProvider.Run()

	if err != nil {
		loggingClient.Error(fmt.Sprintf("error occurred generating tokens: %s", err.Error()))
	}

	return false // Tell bootstrap.Run() to exit wait loop and terminate
}
