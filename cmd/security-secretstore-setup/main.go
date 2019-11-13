/*******************************************************************************
 * Copyright 2019 Dell Inc.
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
 * @author: Alain Pulluelo, ForgeRock AS
 * @author: Tingyu Zeng, Dell
 * @author: Daniel Harms, Dell
 *
 *******************************************************************************/

package main

import (
	"flag"
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap"
	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/container"
)

func main() {
	startupTimer := startup.NewStartUpTimer(internal.BootRetrySecondsDefault, internal.BootTimeoutSecondsDefault)

	if len(os.Args) < 2 {
		usage.HelpCallbackSecuritySecretStore()
	}

	var initNeeded bool
	var insecureSkipVerify bool
	var vaultInterval int
	var configDir, profileDir string
	var useRegistry bool

	flag.BoolVar(&initNeeded, "init", false, "run init procedure for security service.")
	flag.BoolVar(&insecureSkipVerify, "insecureSkipVerify", false, "skip server side SSL verification, mainly for self-signed cert")
	flag.IntVar(&vaultInterval, "vaultInterval", 30, "time to wait between checking Vault status in seconds.")
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallbackSecuritySecretStore
	flag.Parse()

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
		bootstrapContainer.ConfigurationInterfaceName: func(get di.Get) interface{} {
			return get(container.ConfigurationName)
		},
	})
	bootstrap.Run(
		configDir,
		profileDir,
		internal.ConfigFileName,
		useRegistry,
		internal.SecuritySecretStoreSetupServiceKey,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			secretstore.NewBootstrapHandler(insecureSkipVerify, initNeeded, vaultInterval).Handler,
		},
	)
}
