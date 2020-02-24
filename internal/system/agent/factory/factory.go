/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package factory

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/urlclient"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/config"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

// clientMap defines the map used to store existing clients.
type clientMap map[string]general.GeneralClient

// factory contains references to dependencies required by the general client factory.
type factory struct {
	ctx             context.Context
	wg              *sync.WaitGroup
	registryClient  registry.Client
	configuration   config.ConfigurationClients
	serviceProtocol string
	clientMap       clientMap
	m               sync.RWMutex
}

// New is a factory function that creates a general client factory.
func New(
	ctx context.Context,
	wg *sync.WaitGroup,
	registryClient registry.Client,
	configuration config.ConfigurationClients,
	serviceProtocol string) *factory {

	return &factory{
		ctx:             ctx,
		wg:              wg,
		registryClient:  registryClient,
		configuration:   configuration,
		serviceProtocol: serviceProtocol,
		clientMap:       make(clientMap),
		m:               sync.RWMutex{},
	}
}

// New is a factory method that returns a general client factory for the specified serviceName (and creates one if
// one does not already exist).
func (f *factory) New(serviceName string) (general.GeneralClient, error) {
	f.m.Lock()
	defer f.m.Unlock()

	// client exists...
	if _, ok := f.clientMap[serviceName]; ok {
		return f.clientMap[serviceName], nil
	}

	// configuration for client exists...
	if _, ok := f.configuration[serviceName]; ok {
		f.clientMap[serviceName] = general.NewGeneralClient(
			urlclient.New(
				f.ctx,
				f.wg,
				f.registryClient,
				serviceName,
				"/",
				internal.ClientMonitorDefault,
				f.configuration[serviceName].Url(),
			),
		)
		return f.clientMap[serviceName], nil
	}

	// look to registry for client...
	if f.registryClient == nil {
		return nil, fmt.Errorf("registryClient not initialized; required to handle unknown service: %s", serviceName)
	}

	// Service unknown to SMA, so ask the Registry whether `serviceName` is available.
	ok, err := f.registryClient.IsServiceAvailable(serviceName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("service %s is not available", serviceName)
	}

	// Since serviceName is unknown to SMA, ask the Registry for a ServiceEndpoint associated with `serviceName`
	ep, err := f.registryClient.GetServiceEndpoint(serviceName)
	if err != nil {
		return nil, fmt.Errorf("GetServiceEndpoint for %s got error: %v", serviceName, err.Error())
	}

	configClient := bootstrapConfig.ClientInfo{
		Protocol: f.serviceProtocol,
		Host:     ep.Host,
		Port:     ep.Port,
	}

	f.clientMap[serviceName] = general.NewGeneralClient(
		urlclient.New(
			f.ctx,
			f.wg,
			f.registryClient,
			serviceName,
			"/",
			internal.ClientMonitorDefault,
			configClient.Url(),
		),
	)
	return f.clientMap[serviceName], nil
}
