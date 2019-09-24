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

package config

import (
	"net/url"
	"os"
	"strconv"
)

const (
	envKeyUrl = "edgex_registry"
)

// deprecated
func OverrideFromEnvironment(registry RegistryInfo) RegistryInfo {
	if env := os.Getenv(envKeyUrl); env != "" {
		if u, err := url.Parse(env); err == nil {
			if p, err := strconv.ParseInt(u.Port(), 10, 0); err == nil {
				registry.Port = int(p)
				registry.Host = u.Hostname()
				registry.Type = u.Scheme
			}
		}
	}
	return registry
}
