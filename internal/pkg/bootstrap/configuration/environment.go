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

package configuration

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"net/url"
	"os"
	"strconv"
)

const (
	envKeyUrl             = "edgex_registry"
	envKeyStartupDuration = "edgex_registry_startup_duration"
	envKeyStartupInterval = "edgex_registry_startup_interval"
)

// OverrideFromEnvironment overrides the registryInfo values from an environment variable value (if it exists).
func OverrideFromEnvironment(registry config.RegistryInfo, startup config.StartupInfo) (config.RegistryInfo, config.StartupInfo) {
	if env := os.Getenv(envKeyUrl); env != "" {
		if u, err := url.Parse(env); err == nil {
			if p, err := strconv.ParseInt(u.Port(), 10, 0); err == nil {
				registry.Port = int(p)
				registry.Host = u.Hostname()
				registry.Type = u.Scheme
			}
		}
	}

	//	Override the startup timer configuration, if provided.
	if env := os.Getenv(envKeyStartupDuration); env != "" {
		if n, err := strconv.ParseInt(env, 10, 0); err == nil && n > 0 {
			startup.Duration = int(n)
		}
	}
	if env := os.Getenv(envKeyStartupInterval); env != "" {
		if n, err := strconv.ParseInt(env, 10, 0); err == nil && n > 0 {
			startup.Interval = int(n)
		}
	}

	return registry, startup
}
