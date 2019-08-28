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

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
)

const (
	envKeyUrl = "edgex_registry"
)

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

func NewRetryInfo() startup.RetryInfo {
	return startup.RetryInfo{
		Count:   GetFromEnvironmentUint(internal.BootRetryCountKey, internal.BootRetryCountDefault),
		Timeout: GetFromEnvironmentUint(internal.BootRetryTimeKey, internal.BootRetryTimeoutDefault),
		Wait:    GetFromEnvironmentUint(internal.BootRetryWaitKey, internal.BootRetryWaitDefault),
	}
}

func GetFromEnvironmentUint(key string, defval int) int {
	if env := os.Getenv(key); env != "" {
		if i, err := strconv.ParseInt(env, 10, 0); err == nil && i > 0 {
			return int(i)
		}
	}
	return defval
}
