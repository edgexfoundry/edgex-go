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
package data

import "github.com/edgexfoundry/edgex-go/internal/pkg/config"

type ConfigurationStruct struct {
	Writable                   WritableInfo
	MessageQueue               config.MessageQueueInfo
	Clients                    map[string]config.ClientInfo
	Databases                  map[string]config.DatabaseInfo
	Logging                    config.LoggingInfo
	Registry                   config.RegistryInfo
	Service                    config.ServiceInfo
}

type WritableInfo struct {
	DeviceUpdateLastConnected  bool
	MetaDataCheck              bool
	PersistData                bool
	ServiceUpdateLastConnected bool
	ValidateCheck              bool
	LogLevel                   string
}