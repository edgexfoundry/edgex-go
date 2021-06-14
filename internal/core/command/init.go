/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright (c) 2019 Intel Corporation
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

	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	clients "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

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
	LoadRestRoutes(b.router, dic)

	configuration := container.ConfigurationFrom(dic.Get)

	// initialize clients required by the service
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.MetadataDeviceClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceClient
			return clients.NewDeviceClient(configuration.Clients[common.CoreMetaDataServiceKey].Url() + common.ApiDeviceRoute)
		},
		bootstrapContainer.MetadataDeviceProfileClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceProfileClient
			return clients.NewDeviceProfileClient(configuration.Clients[common.CoreMetaDataServiceKey].Url() + common.ApiDeviceProfileRoute)
		},
		bootstrapContainer.MetadataDeviceServiceClientName: func(get di.Get) interface{} { // add v2 API MetadataDeviceServiceClient
			return clients.NewDeviceServiceClient(configuration.Clients[common.CoreMetaDataServiceKey].Url() + common.ApiDeviceServiceRoute)
		},
		bootstrapContainer.DeviceServiceCommandClientName: func(get di.Get) interface{} { // add v2 API DeviceServiceCommandClient
			return clients.NewDeviceServiceCommandClient()
		},
	})
	return true
}
