/*******************************************************************************
* Copyright 2020 Redis Labs
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
* @author: Andre Srinivasan
*******************************************************************************/

package redis

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/redis/config"
	"github.com/edgexfoundry/edgex-go/internal/security/redis/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/gorilla/mux"
)

func Main(ctx context.Context, cancel context.CancelFunc, router *mux.Router, readyStream chan<- bool) {
	startupTimer := startup.NewStartUpTimer(clients.SecurityBootstrapRedisKey)

	// All common command-line flags have been moved to DefaultCommonFlags.
	f := flags.New()
	f.Parse(os.Args[1:])

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	handler := NewHandler()

	// bootstrap.RunAndReturnWaitGroup is needed for the underlying configuration system.
	// Conveniently, it also creates a pipeline of functions as the list of BootstrapHandler's is
	// executed in order.
	bootstrap.RunAndReturnWaitGroup(
		ctx,
		cancel,
		f,
		clients.SecurityBootstrapRedisKey,
		internal.ConfigStemCore+internal.ConfigMajorVersion,
		configuration,
		nil,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			handlers.SecureProviderBootstrapHandler,
			handler.getCredentials,
			handler.connect,
			handler.maybeSetCredentials,
		},
	)
}
