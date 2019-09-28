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
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore"
	"os"
)

func main() {
	startupTimer := startup.NewStartUpTimer(1, internal.BootTimeoutDefault)

	if len(os.Args) < 2 {
		usage.HelpCallbackSecuritySecretStore()
	}

	var insecureSkipVerify bool
	var configDir, profileDir string
	var useRegistry bool

	flag.BoolVar(&insecureSkipVerify, "insecureSkipVerify", false, "skip server side SSL verification, mainly for self-signed cert")
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallbackSecuritySecretStore
	flag.Parse()

	bootstrap.Run(
		configDir,
		profileDir,
		internal.ConfigFileName,
		useRegistry,
		internal.SecurityProxySetupServiceKey,
		secretstore.Configuration,
		startupTimer,
		[]interfaces.BootstrapHandler{
			secretstore.NewHandler(insecureSkipVerify).BootstrapHandler,
		})
}
