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
	"os"
	"strconv"
)

const (
	envKeyPrefix = "edgex_registry_"
	envKeyHost   = envKeyPrefix + "host"
	envKeyPort   = envKeyPrefix + "port"
	envKeyType   = envKeyPrefix + "type"
)

func OverrideFromEnvironment(registry RegistryInfo) RegistryInfo {
	if env := os.Getenv(envKeyHost); env != "" {
		registry.Host = env
	}
	if env := os.Getenv(envKeyPort); env != "" {
		if v, err := strconv.ParseInt(env, 10, 0); err == nil {
			registry.Port = int(v)
		}
	}
	if env := os.Getenv(envKeyType); env != "" {
		registry.Type = env
	}
	return registry
}
