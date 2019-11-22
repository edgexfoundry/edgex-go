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
 * @author: Tingyu Zeng, Dell
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
	"github.com/edgexfoundry/edgex-go/internal/security/proxy"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/container"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

func main() {
	startupTimer := startup.NewStartUpTimer(internal.BootRetrySecondsDefault, internal.BootTimeoutSecondsDefault)

	if len(os.Args) < 2 {
		usage.HelpCallbackSecurityProxy()
	}
	var initNeeded bool
	var insecureSkipVerify bool
	var resetNeeded bool
	var configDir, profileDir string
	var userTobeCreated string
	var userOfGroup string
	var userToBeDeleted string
	var useRegistry bool

	flag.BoolVar(&insecureSkipVerify, "insecureSkipVerify", false, "skip server side SSL verification, mainly for self-signed cert")
	flag.BoolVar(&initNeeded, "init", false, "run init procedure for security service.")
	flag.BoolVar(&resetNeeded, "reset", false, "reset reverse proxy by removing all services/routes/consumers")
	flag.StringVar(&userTobeCreated, "useradd", "", "user that needs to be added to consume the edgex services")
	flag.StringVar(&userOfGroup, "group", "user", "group that the user belongs to. By default it is in user group")
	flag.StringVar(&userToBeDeleted, "userdel", "", "user that needs to be deleted from the edgex services")
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallbackSecurityProxy
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
		clients.SecurityProxySetupServiceKey,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			proxy.NewBootstrapHandler(
				insecureSkipVerify,
				initNeeded,
				resetNeeded,
				userTobeCreated,
				userOfGroup,
				userToBeDeleted).Handler,
		},
	)
}
