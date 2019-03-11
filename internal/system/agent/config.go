/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package agent

import "github.com/edgexfoundry/edgex-go/internal/pkg/config"

type ConfigurationStruct struct {
	Writable              WritableInfo
	Clients               map[string]config.ClientInfo
	Service               config.ServiceInfo
	ExecutorPath          string
	Registry              config.RegistryInfo
	Logging               config.LoggingInfo
	FormatSpecifier       string
}

type WritableInfo struct {
	ResendLimit int
	LogLevel    string
}
