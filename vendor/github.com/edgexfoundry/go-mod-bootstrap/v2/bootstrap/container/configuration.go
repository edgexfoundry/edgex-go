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
	"github.com/edgexfoundry/go-mod-configuration/v2/configuration"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// ConfigurationInterfaceName contains the name of the interfaces.Configuration implementation in the DIC.
var ConfigurationInterfaceName = di.TypeInstanceToName((*interfaces.Configuration)(nil))

// ConfigurationFrom helper function queries the DIC and returns the interfaces.Configuration implementation.
func ConfigurationFrom(get di.Get) interfaces.Configuration {
	configuration, ok := get(ConfigurationInterfaceName).(interfaces.Configuration)
	if !ok {
		return nil
	}

	return configuration
}

// ConfigClientInterfaceName contains the name of the configuration.Client implementation in the DIC.
var ConfigClientInterfaceName = di.TypeInstanceToName((*configuration.Client)(nil))

// ConfigClientFrom helper function queries the DIC and returns the configuration.Client implementation.
func ConfigClientFrom(get di.Get) configuration.Client {
	client, ok := get(ConfigClientInterfaceName).(configuration.Client)
	if !ok {
		return nil
	}

	return client
}
