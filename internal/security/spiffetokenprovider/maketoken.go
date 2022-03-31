//
// Copyright (c) 2022 Intel Corporation
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

package spiffetokenprovider

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v2/secrets"
)

const (
	DefaultTokenTTL = "1h"
)

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

func makeToken(serviceName string,
	privilegedToken string,
	secretStoreClient secrets.SecretStoreClient,
	lc logger.LoggingClient) (interface{}, error) {

	lc.Infof("generating policy/token defaults for service %s", serviceName)
	lc.Infof("using policy/token defaults for service %s", serviceName)
	servicePolicy := makeDefaultTokenPolicy(serviceName)
	defaultPolicyPaths := servicePolicy["path"].(map[string]interface{})
	for pathKey, policy := range defaultPolicyPaths {
		servicePolicy["path"].(map[string]interface{})[pathKey] = policy
	}
	createTokenParameters := makeDefaultTokenParameters(serviceName, DefaultTokenTTL, DefaultTokenTTL)

	// Set a meta property that consuming serices can use to automatically scope secret queries
	createTokenParameters["meta"] = map[string]interface{}{
		"edgex-service-name": serviceName,
	}

	// Always create a policy with this name
	policyName := "edgex-service-" + serviceName

	policyBytes, err := json.Marshal(servicePolicy)
	if err != nil {
		lc.Error(fmt.Sprintf("failed encode service policy for %s: %s", serviceName, err.Error()))
		return nil, err
	}

	if err := secretStoreClient.InstallPolicy(privilegedToken, policyName, string(policyBytes)); err != nil {
		lc.Error(fmt.Sprintf("failed to install policy %s: %s", policyName, err.Error()))
		return nil, err
	}

	var createTokenResponse interface{}

	if createTokenResponse, err = secretStoreClient.CreateToken(privilegedToken, createTokenParameters); err != nil {
		lc.Error(fmt.Sprintf("failed to create vault token for service %s: %s", serviceName, err.Error()))
		return nil, err
	}

	return createTokenResponse, nil
}
