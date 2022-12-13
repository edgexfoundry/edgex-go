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
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/container"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/v3/secrets"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-secrets/v3/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg/token/fileioperformer"
)

type Bootstrap struct {
}

func NewBootstrap() *Bootstrap {
	return &Bootstrap{}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	cfg := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenProvider := authtokenloader.NewAuthTokenLoader(fileOpener)

	var requester internal.HttpCaller
	if caFilePath := cfg.SecretStore.CaFilePath; caFilePath != "" {
		lc.Info("using certificate verification for secret store connection")
		caReader, err := fileOpener.OpenFileReader(caFilePath, os.O_RDONLY, 0400)
		if err != nil {
			lc.Errorf("failed to load CA certificate: %s", err.Error())
			return false
		}
		requester = pkg.NewRequester(lc).WithTLS(caReader, cfg.SecretStore.ServerName)
	} else {
		lc.Info("bypassing certificate verification for secret store connection")
		requester = pkg.NewRequester(lc).Insecure()
	}

	clientConfig := types.SecretConfig{
		Type:     secrets.Vault,
		Host:     cfg.SecretStore.Host,
		Port:     cfg.SecretStore.Port,
		Protocol: cfg.SecretStore.Protocol,
	}
	client, err := secrets.NewSecretStoreClient(clientConfig, lc, requester)
	if err != nil {
		lc.Errorf("error occurred creating SecretStoreClient: %s", err.Error())
		return false
	}

	fileProvider := NewTokenProvider(lc, fileOpener, tokenProvider, client)

	fileProvider.SetConfiguration(cfg.SecretStore, cfg.TokenFileProvider)
	err = fileProvider.Run()

	if err != nil {
		lc.Errorf("error occurred generating tokens: %s", err.Error())
		return false
	}

	return true
}
