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

const (
	BootTimeoutDefault        = BootTimeoutSecondsDefault * 1000
	BootTimeoutSecondsDefault = 30
	BootRetrySecondsDefault   = 1
	ConfigFileName            = "configuration.toml"
	// TODO: move the config stem constants in go-mod-contracts
	ConfigStemApp      = "edgex/appservices/"
	ConfigStemCore     = "edgex/core/"
	ConfigStemDevice   = "edgex/devices/"
	ConfigStemSecurity = "edgex/security/"
	LogDurationKey     = "duration"
)

const (
	ConfigProviderEnvVar = "edgex_configuration_provider"
	WritableKey          = "/Writable"
)

const (
	AuthHeaderTitle = "Authorization"
	BearerLabel     = "Bearer "
)
const (
	BootstrapMessageBusServiceKey = "security-bootstrapper-messagebus"
)
