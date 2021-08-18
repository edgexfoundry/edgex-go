//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//

package fileprovider

func makeDefaultTokenPolicy(serviceName string) map[string]interface{} {
	// protected path for secret/
	protectedPath := "secret/edgex/" + serviceName + "/*"
	capabilities := []string{"create", "update", "delete", "list", "read"}
	acl := map[string]interface{}{"capabilities": capabilities}
	// path for consul tokens
	registryCredsPath := "consul/creds/" + serviceName
	registryCredsCapabilities := []string{"read"}
	registryCredsACL := map[string]interface{}{"capabilities": registryCredsCapabilities}
	pathObject := map[string]interface{}{
		protectedPath:     acl,
		registryCredsPath: registryCredsACL,
	}
	retval := map[string]interface{}{"path": pathObject}
	return retval

	/*
		{
			"path": {
			  "secret/edgex/service-name/*": {
				"capabilities": [ "create", "update", "delete", "list", "read" ]
			  },
			  "consul/creds/service-name": {
				"capabilities": [ "read" ]
			  }
			}
		}
	*/
}

func makeDefaultTokenParameters(serviceName string, defaultTTL string, defaultPeriod string) map[string]interface{} {
	return map[string]interface{}{
		"display_name": serviceName,
		"no_parent":    true,
		"ttl":          defaultTTL,
		"period":       defaultPeriod,
		"policies":     []string{"edgex-service-" + serviceName},
	}
}
