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

	"github.com/edgexfoundry/edgex-go/internal/core/command/container"
	errorContainer "github.com/edgexfoundry/edgex-go/internal/pkg/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
)

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the command service.
func BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	registryClient := bootstrapContainer.RegistryFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	// initialize clients required by the service
	dic.Update(di.ServiceConstructorMap{
		container.MetadataDeviceClientName: func(get di.Get) interface{} {
			return metadata.NewDeviceClient(
				types.EndpointParams{
					ServiceKey:  clients.CoreMetaDataServiceKey,
					Path:        clients.ApiDeviceRoute,
					UseRegistry: registryClient != nil,
					Url:         configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute,
					Interval:    configuration.Service.ClientMonitor,
				},
				endpoint.Endpoint{RegistryClient: &registryClient})
		},
		errorContainer.ErrorHandlerName: func(get di.Get) interface{} {
			return errorconcept.NewErrorHandler(lc)
		},
	})

	return true
}
