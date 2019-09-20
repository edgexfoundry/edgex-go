//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	bootstrapinterface "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"

	"github.com/edgexfoundry/edgex-go/internal/security/authtokenloader"
	"github.com/edgexfoundry/edgex-go/internal/security/fileioperformer"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// Constants

// Dependencies

var fileProvider fileprovider.TokenProvider

func main() {
	startupTimer := startup.NewStartUpTimer(1, internal.BootTimeoutDefault)

	var useRegistry bool
	var profileDir, configDir string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallback
	flag.Parse()

	bootstrap.Run(
		configDir,
		profileDir,
		internal.ConfigFileName,
		useRegistry,
		clients.SupportLoggingServiceKey,
		fileprovider.Configuration,
		startupTimer,
		di.NewContainer(di.ServiceConstructorMap{}),
		[]bootstrapinterface.BootstrapHandler{
			func(
				wg *sync.WaitGroup,
				ctx context.Context,
				startupTimer startup.Timer,
				dic *di.Container) bool {

				logging := container.LoggingClientFrom(dic.Get)
				return runTokenProvider(logging)
			},
		})
}

func runTokenProvider(logging logger.LoggingClient) bool {

	cfg := fileprovider.Configuration

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenProvider := authtokenloader.NewAuthTokenLoader(fileOpener)

	var req internal.HttpCaller
	if caFilePath := cfg.SecretService.CaFilePath; caFilePath != "" {
		logging.Info("using certificate verification for secret store connection")
		caReader, err := fileOpener.OpenFileReader(caFilePath, os.O_RDONLY, 0400)
		if err != nil {
			logging.Error(fmt.Sprintf("failed to load CA certificate: %s", err.Error()))
			return false
		}
		req = secretstoreclient.NewRequestor(logging).WithTLS(caReader, cfg.SecretService.ServerName)
	} else {
		logging.Info("bypassing certificate verification for secret store connection")
		req = secretstoreclient.NewRequestor(logging).Insecure()
	}
	vaultScheme := cfg.SecretService.Scheme
	vaultHost := fmt.Sprintf("%s:%v", cfg.SecretService.Server, cfg.SecretService.Port)
	vaultClient := secretstoreclient.NewSecretStoreClient(logging, req, vaultScheme, vaultHost)

	if fileProvider == nil {
		// main_test.go has injected a testing mock if fileProvider != nil
		fileProvider = fileprovider.NewTokenProvider(logging, fileOpener, tokenProvider, vaultClient)
	}

	fileProvider.SetConfiguration(cfg.SecretService, cfg.TokenFileProvider)
	err := fileProvider.Run()

	if err != nil {
		logging.Error(fmt.Sprintf("error occurred generating tokens: %s", err.Error()))
	}

	return false // Tell bootstrap.Run() to exit wait loop and terminate
}
