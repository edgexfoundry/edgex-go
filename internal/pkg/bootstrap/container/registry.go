/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

// RegistryClientInterfaceName contains the name of the registry.Client implementation in the DIC.
var RegistryClientInterfaceName = di.TypeInstanceToName((*registry.Client)(nil))

// RegistryFrom helper function queries the DIC and returns the registry.Client implementation.
func RegistryFrom(get di.Get) registry.Client {
	registryClient := get(RegistryClientInterfaceName)
	if registryClient != nil {
		return registryClient.(registry.Client)
	}
	return (registry.Client)(nil)
}
