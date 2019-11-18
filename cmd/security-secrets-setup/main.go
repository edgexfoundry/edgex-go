/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2019 Intel Corp.
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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap"
	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/cache"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/generate"
	_import "github.com/edgexfoundry/edgex-go/internal/security/secrets/command/import"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/command/legacy"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/container"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"
)

func wrappedMain() (*config.ConfigurationStruct, int) {
	startupTimer := startup.NewStartUpTimer(internal.BootRetrySecondsDefault, internal.BootTimeoutSecondsDefault)

	var configDir string

	legacyFlags := legacy.NewFlags()
	generateFlagSet := generate.NewFlags()
	cacheFlagSet := cache.NewFlags()
	importFlagSet := _import.NewFlags()
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallbackSecuritySetup
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Please specify subcommand for " + internal.SecuritySecretsSetupServiceKey)
		flag.Usage()
		return nil, contract.StatusCodeExitNormal
	}

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
		bootstrapContainer.ConfigurationInterfaceName: func(get di.Get) interface{} {
			return get(container.ConfigurationName)
		},
	})
	serviceHandler := secrets.NewBootstrapHandler(legacyFlags, generateFlagSet, cacheFlagSet, importFlagSet, flag.Args()[0])
	bootstrap.Run(
		configDir,
		bootstrap.EmptyProfileDir,
		internal.ConfigFileName,
		bootstrap.DoNotUseRegistry,
		internal.SecuritySecretsSetupServiceKey,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			serviceHandler.Handler,
		},
	)
	return configuration, serviceHandler.ExitStatusCode()
}

func main() {
	_, exitStatusCode := wrappedMain()
	os.Exit(exitStatusCode)
}
