//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package registry

import (
	"github.com/edgexfoundry/go-mod-registry/v2/pkg/types"
)

type Client interface {
	// Registers the current service with Registry for discover and health check
	Register() error

	// Un-registers the current service with Registry for discover and health check
	Unregister() error

	// Registers a
	RegisterCheck(id string, name string, notes string, url string, interval string) error

	// Simply checks if Registry is up and running at the configured URL
	IsAlive() bool

	// Gets the service endpoint information for the target ID from the Registry
	GetServiceEndpoint(serviceId string) (types.ServiceEndpoint, error)

	// Gets all the service endpoints information from the Registry
	GetAllServiceEndpoints() ([]types.ServiceEndpoint, error)

	// Checks with the Registry if the target service is available, i.e. registered and healthy
	IsServiceAvailable(serviceId string) (bool, error)
}
