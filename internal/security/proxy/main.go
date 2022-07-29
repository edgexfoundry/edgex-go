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

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

func Main(ctx context.Context, cancel context.CancelFunc) {
	startupTimer := startup.NewStartUpTimer(common.SecurityProxySetupServiceKey)

	var initNeeded bool
	var insecureSkipVerify bool
	var resetNeeded bool

	// All common command-line flags have been moved to bootstrap. Service specific flags are added below.
	f := flags.NewWithUsage(
		"    --insecureSkipVerify=true/false Indicates if skipping the server side SSL cert verification, similar to -k of curl\n" +
			"    --init=true/false               Indicates if security service should be initialized\n" +
			"    --reset=true/false              Indicate if security service should be reset to initialization status\n",
	)

	if len(os.Args) < 2 {
		f.Help()
	}

	f.FlagSet.BoolVar(&insecureSkipVerify, "insecureSkipVerify", false, "")
	f.FlagSet.BoolVar(&initNeeded, "init", false, "")
	f.FlagSet.BoolVar(&resetNeeded, "reset", false, "")
	f.Parse(os.Args[1:])

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	_, _, success := bootstrap.RunAndReturnWaitGroup(
		ctx,
		cancel,
		f,
		common.SecurityProxySetupServiceKey,
		internal.ConfigStemSecurity,
		configuration,
		nil,
		startupTimer,
		dic,
		true,
		[]interfaces.BootstrapHandler{
			NewBootstrap(
				insecureSkipVerify,
				initNeeded,
				resetNeeded).BootstrapHandler,
		},
	)

	if !success {
		os.Exit(1)
	}
}
