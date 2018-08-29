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
package config

// Configuration struct populated from local TOML during service bootstrap
// NOTE: I am only following the existing pattern of putting this struct into
//       a file named "const.go". I do not think this properly belongs in a
//       file whose intent by name is for constants. If I were to move this,
//       then other services should also have their struct moved in a single PR.
type ConfigurationStruct struct {
	ConfigPath                   string
	GlobalPrefix                 string
	ConsulProtocol               string
	ConsulHost                   string
	ConsulPort                   int
	IsReset                      bool
	FailLimit                    int
	FailWaitTime                 int
	AcceptablePropertyExtensions []string
	YamlExtensions               []string
	TomlExtensions               []string
	EnableRemoteLogging          bool
	LoggingFile                  string
	LoggingRemoteURL             string
}

const consulStatusPath = "/v1/agent/self"
