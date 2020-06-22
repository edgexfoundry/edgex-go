/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright (c) 2019 Intel Corporation
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

package command

import (
	"context"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/local"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
	errorContainer "github.com/edgexfoundry/edgex-go/internal/pkg/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/gorilla/mux"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	router *mux.Router
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(router *mux.Router) *Bootstrap {
	return &Bootstrap{
		router: router,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the command service.
func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	loadRestRoutes(b.router, dic)

	// TODO: there is an outstanding known issue (https://github.com/edgexfoundry/edgex-go/issues/2462)
	// 		that could be seemingly be solved by moving from JIT initialization of these external clients to static
	// 		init on startup, like registryClient and configuration are initialized.
	// 		Doing so would cover over the symptoms of the bug, but the root problem of server processing taking longer
	// 		than the configured client time out would still be present.
	// 		Until that problem is addressed by larger architectural changes, if you are experiencing a bug similar to
	//		https://github.com/edgexfoundry/edgex-go/issues/2421, the correct fix is to bump up the client timeout.
	configuration := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	// initialize clients required by the service
	dic.Update(di.ServiceConstructorMap{
		container.MetadataDeviceClientName: func(get di.Get) interface{} {
			return metadata.NewDeviceClient(local.New(configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute))
		},
		errorContainer.ErrorHandlerName: func(get di.Get) interface{} {
			return errorconcept.NewErrorHandler(lc)
		},
	})

	return true
}
