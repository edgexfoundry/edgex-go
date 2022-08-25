/*******************************************************************************
* Copyright 2022 Intel Corporation
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

package messagebus

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/messagebus/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/messagebus/container"
	msgbushandlers "github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/messagebus/handlers"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// Configure is the main entry point for configuring the database redis before startup
func Configure(ctx context.Context,
	cancel context.CancelFunc,
	flags flags.Common) {
	startupTimer := startup.NewStartUpTimer("security-bootstrapper-mqtt")

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	msgbusBootstrapHdl := msgbushandlers.NewHandler()

	// bootstrap.RunAndReturnWaitGroup is needed for the underlying configuration system.
	// Conveniently, it also creates a pipeline of functions as the list of BootstrapHandler's is
	// executed in order.
	_, _, ok := bootstrap.RunAndReturnWaitGroup(
		ctx,
		cancel,
		flags,
		"security-bootstrapper-mqtt",
		internal.ConfigStemSecurity,
		configuration,
		nil,
		startupTimer,
		dic,
		true,
		[]interfaces.BootstrapHandler{
			msgbusBootstrapHdl.GetCredentials,
			msgbusBootstrapHdl.SetupPasswordFile,
			msgbusBootstrapHdl.SetupConfFile,
		},
	)

	if !ok {
		// had some issue(s) during bootstrapping message bus
		os.Exit(1)
	}
}
