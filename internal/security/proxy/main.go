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

package proxy

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/container"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"

	"github.com/gorilla/mux"
)

func Main(ctx context.Context, cancel context.CancelFunc, _ *mux.Router, _ chan<- bool) {
	startupTimer := startup.NewStartUpTimer(clients.SecurityProxySetupServiceKey)

	var initNeeded bool
	var insecureSkipVerify bool
	var resetNeeded bool
	var userTobeCreated string
	var userOfGroup string
	var userToBeDeleted string

	// All common command-line flags have been moved to bootstrap. Service specific flags are added below.
	f := flags.NewWithUsage(
		"    --insecureSkipVerify=true/false Indicates if skipping the server side SSL cert verification, similar to -k of curl\n" +
			"    --init=true/false               Indicates if security service should be initialized\n" +
			"    --reset=true/false              Indicate if security service should be reset to initialization status\n" +
			"    --useradd=<username>            Create an account and return JWT\n" +
			"    --group=<groupname>             Group name the user belongs to\n" +
			"    --userdel=<username>            Delete an account",
	)

	if len(os.Args) < 2 {
		f.Help()
	}

	f.FlagSet.BoolVar(&insecureSkipVerify, "insecureSkipVerify", false, "")
	f.FlagSet.BoolVar(&initNeeded, "init", false, "")
	f.FlagSet.BoolVar(&resetNeeded, "reset", false, "")
	f.FlagSet.StringVar(&userTobeCreated, "useradd", "", "")
	f.FlagSet.StringVar(&userOfGroup, "group", "user", "")
	f.FlagSet.StringVar(&userToBeDeleted, "userdel", "", "")
	f.Parse(os.Args[1:])

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	bootstrap.Run(
		ctx,
		cancel,
		f,
		clients.SecurityProxySetupServiceKey,
		internal.ConfigStemSecurity+internal.ConfigMajorVersion,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			secret.NewSecret().BootstrapHandler,
			NewBootstrap(
				insecureSkipVerify,
				initNeeded,
				resetNeeded,
				userTobeCreated,
				userOfGroup,
				userToBeDeleted).BootstrapHandler,
		},
	)
}
