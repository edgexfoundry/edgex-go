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

	"github.com/edgexfoundry/go-mod-secrets/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/pkg/token/fileioperformer"
)

type Bootstrap struct {
	exitCode int
}

func NewBootstrap() *Bootstrap {
	return &Bootstrap{
		exitCode: 0,
	}
}

// ExitCode returns desired exit code of program
func (b *Bootstrap) ExitCode() int {
	return b.exitCode
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	cfg := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenProvider := authtokenloader.NewAuthTokenLoader(fileOpener)

	var req internal.HttpCaller
	if caFilePath := cfg.SecretService.CaFilePath; caFilePath != "" {
		lc.Info("using certificate verification for secret store connection")
		caReader, err := fileOpener.OpenFileReader(caFilePath, os.O_RDONLY, 0400)
		if err != nil {
			lc.Error(fmt.Sprintf("failed to load CA certificate: %s", err.Error()))
			return false
		}
		req = secretstoreclient.NewRequestor(lc).WithTLS(caReader, cfg.SecretService.ServerName)
	} else {
		lc.Info("bypassing certificate verification for secret store connection")
		req = secretstoreclient.NewRequestor(lc).Insecure()
	}
	vaultScheme := cfg.SecretService.Scheme
	vaultHost := fmt.Sprintf("%s:%v", cfg.SecretService.Server, cfg.SecretService.Port)
	vaultClient := secretstoreclient.NewSecretStoreClient(lc, req, vaultScheme, vaultHost)

	fileProvider := NewTokenProvider(lc, fileOpener, tokenProvider, vaultClient)

	fileProvider.SetConfiguration(cfg.SecretService, cfg.TokenFileProvider)
	err := fileProvider.Run()

	if err != nil {
		lc.Error(fmt.Sprintf("error occurred generating tokens: %s", err.Error()))
		b.exitCode = 1
	}

	return false // Tell bootstrap.Run() to exit wait loop and terminate
}
