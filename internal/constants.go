/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
package internal

import "math"

const (
	BootRetryCountDefault        = math.MaxInt64
	BootRetryCountKey            = "edgex_registry_retry_count"
	BootRetryTimeoutDefault      = 30000
	BootRetryTimeKey             = "edgex_registry_retry_timeout"
	BootRetryWaitDefault         = 1
	BootRetryWaitKey             = "edgex_registry_retry_wait"
	ClientMonitorDefault         = 15000
	ConfigFileName               = "configuration.toml"
	ConfigRegistryStem           = "edgex/core/1.0/"
	LogDurationKey               = "duration"
	SecurityProxySetupServiceKey = "edgex-security-proxy-setup"
	WritableKey                  = "/Writable"
)
