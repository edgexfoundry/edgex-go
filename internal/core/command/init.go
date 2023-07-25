/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright (c) 2019-2023 Intel Corporation
 * Copyright (C) 2021 IOTech Ltd
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

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	clients "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/http"
	"github.com/gorilla/mux"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	router      *mux.Router
	serviceName string
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(router *mux.Router, serviceName string) *Bootstrap {
	return &Bootstrap{
		router:      router,
		serviceName: serviceName,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the command service.
func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	LoadRestRoutes(b.router, dic, b.serviceName)

	// DeviceServiceCommandClient is not part of the common clients handled by the NewClientsBootstrap handler
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.DeviceServiceCommandClientName: func(get di.Get) interface{} { // add API DeviceServiceCommandClient
			jwtSecretProvider := secret.NewJWTSecretProvider(container.SecretProviderExtFrom(get))
			return clients.NewDeviceServiceCommandClient(jwtSecretProvider)
		},
	})

	return true
}
