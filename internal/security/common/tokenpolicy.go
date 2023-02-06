//
// Copyright (c) 2019-2023 Intel Corporation
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
// SPDX-License-Identifier: Apache-2.0
//

package common

func MakeDefaultTokenPolicy(serviceName string) map[string]interface{} {
	// protected path for secret/
	secretsPath := "secret/edgex/" + serviceName + "/*"
	secretsAcl := map[string]interface{}{"capabilities": []string{"create", "update", "delete", "list", "read"}}
	// path for consul tokens
	registryCredsPath := "consul/creds/" + serviceName
	registryCredsACL := map[string]interface{}{"capabilities": []string{"read"}}
	// allow request identity JWT
	jwtRequestPath := "identity/oidc/token/" + serviceName
	jwtRequestACL := map[string]interface{}{"capabilities": []string{"read"}}
	// allow introspect JWT
	jwtIntrospectPath := "identity/oidc/introspect"
	jwtIntrospectACL := map[string]interface{}{"capabilities": []string{"create", "update"}}
	// access spec
	pathObject := map[string]interface{}{
		secretsPath:       secretsAcl,
		registryCredsPath: registryCredsACL,
		jwtRequestPath:    jwtRequestACL,
		jwtIntrospectPath: jwtIntrospectACL,
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
			  },
			  "identity/oidc/token/service-name": {
				"capabilities": [ "read" ]
			  },
			  "identity/oidc/introspect": {
				"capabilities": [ "create", "update" ]
			  }
			}
		}
	*/
}
