/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/command"
	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	container "github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap"
	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers/database"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers/httpserver"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers/message"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/handlers/secret"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

func main() {
	startupTimer := startup.NewStartUpTimer(internal.BootRetrySecondsDefault, internal.BootTimeoutSecondsDefault)

	var useRegistry bool
	var configDir, profileDir string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")
	flag.Usage = usage.HelpCallback
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
	httpServer := httpserver.NewBootstrap(command.LoadRestRoutes(dic))

	bootstrap.Run(
		configDir,
		profileDir,
		internal.ConfigFileName,
		useRegistry,
		clients.CoreCommandServiceKey,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			secret.NewSecret().BootstrapHandler,
			database.NewDatabase(&httpServer, configuration).BootstrapHandler,
			command.BootstrapHandler,
			telemetry.BootstrapHandler,
			httpServer.BootstrapHandler,
			message.NewBootstrap(clients.CoreCommandServiceKey, edgex.Version).BootstrapHandler,
		})
}
