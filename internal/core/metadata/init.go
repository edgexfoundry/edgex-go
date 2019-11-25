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
package metadata

import (
	"context"
	"sync"

	metadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	errorContainer "github.com/edgexfoundry/edgex-go/internal/pkg/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
)

// Global variables
var Configuration = &ConfigurationStruct{}
var vdc coredata.ValueDescriptorClient

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the metadata service.
func BootstrapHandler(wg *sync.WaitGroup, ctx context.Context, startupTimer startup.Timer, dic *di.Container) bool {

	// initialize clients required by service.
	registryClient := container.RegistryFrom(dic.Get)

	vdc = coredata.NewValueDescriptorClient(
		types.EndpointParams{
			ServiceKey:  clients.CoreDataServiceKey,
			Path:        clients.ApiValueDescriptorRoute,
			UseRegistry: registryClient != nil,
			Url:         Configuration.Clients["CoreData"].Url() + clients.ApiValueDescriptorRoute,
			Interval:    Configuration.Service.ClientMonitor,
		},
		endpoint.Endpoint{RegistryClient: &registryClient})

	// add dependencies to container
	dic.Update(di.ServiceConstructorMap{
		errorContainer.ErrorHandlerName: func(get di.Get) interface{} {
			return errorconcept.NewErrorHandler(container.LoggingClientFrom(get))
		},
		metadataContainer.NotificationsClientName: func(get di.Get) interface{} {
			return notifications.NewNotificationsClient(types.EndpointParams{
				ServiceKey:  clients.SupportNotificationsServiceKey,
				Path:        clients.ApiNotificationRoute,
				UseRegistry: registryClient != nil,
				Url:         Configuration.Clients["Notifications"].Url() + clients.ApiNotificationRoute,
				Interval:    Configuration.Service.ClientMonitor,
			},
				endpoint.Endpoint{RegistryClient: &registryClient})
		},
	})

	return true
}
