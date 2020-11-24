/*******************************************************************************
 * Copyright 2021 Intel Corporation
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

package bootstrapper

import (
	"context"
	"os"

	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal"
	bootstrapper "github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/container"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/handlers"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

const (
	securityBootstrapperServiceKey = "edgex-security-bootstrapper"
)

// Main function is the wrapper for the security bootstrapper main
func Main(ctx context.Context, cancel context.CancelFunc, _ *mux.Router, _ chan<- bool) {
	// service key for this bootstrapper service
	startupTimer := startup.NewStartUpTimer(securityBootstrapperServiceKey)

	// Common Command-line flags have been moved to command.CommonFlags, but this service doesn't use all
	// the common flags so we are using our own implementation of the CommonFlags interface
	f := bootstrapper.NewCommonFlags()

	f.Parse(os.Args[1:])

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	serviceHandler := handlers.NewInitialization()

	bootstrap.Run(
		ctx,
		cancel,
		f,
		securityBootstrapperServiceKey,
		internal.ConfigStemSecurity+internal.ConfigMajorVersion,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			serviceHandler.BootstrapHandler,
		},
	)

	// exit with the code specified by serviceHandler
	os.Exit(serviceHandler.GetExitStatusCode())
}
