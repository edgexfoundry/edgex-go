/*******************************************************************************
 * Copyright 2017 Dell Inc.
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

package agent

import (
	"context"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	bootstrap "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

var Configuration = &ConfigurationStruct{}
var GenClients *GeneralClients
var LoggingClient logger.LoggingClient
var RegistryClient registry.Client

func initializeClients(useRegistry bool) {
	GenClients = NewGeneralClients()
	for serviceKey, serviceName := range config.ListDefaultServices() {
		GenClients.Set(
			serviceKey,
			general.NewGeneralClient(
				types.EndpointParams{
					ServiceKey:  serviceKey,
					Path:        "/",
					UseRegistry: useRegistry,
					Url:         Configuration.Clients[serviceName].Url(),
					Interval:    internal.ClientMonitorDefault,
				},
				endpoint.Endpoint{RegistryClient: &RegistryClient}))
	}
}

func BootstrapHandler(
	wg *sync.WaitGroup,
	ctx context.Context,
	startupTimer startup.Timer,
	config bootstrap.Configuration,
	logging logger.LoggingClient,
	registry registry.Client) bool {

	// update global variables.
	LoggingClient = logging
	RegistryClient = registry

	// initialize clients required by service.
	initializeClients(registry != nil)

	return true
}
