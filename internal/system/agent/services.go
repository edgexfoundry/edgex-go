/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package agent

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-registry/registry"
)

func getHealth(services []string, registryClient registry.Client) map[string]interface{} {
	health := make(map[string]interface{})
	for _, service := range services {
		if registryClient == nil {
			health[service] = "registry is required to obtain service health status."
			continue
		}

		// the registry service returns nil for a healthy service
		ok, err := registryClient.IsServiceAvailable(service)
		if err != nil {
			health[service] = err.Error()
			continue
		}
		if !ok {
			health[service] = fmt.Sprintf("service %s is not available", service)
			continue
		}
		health[service] = true
	}
	return health
}
