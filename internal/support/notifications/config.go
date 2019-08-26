/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/
package notifications

import "github.com/edgexfoundry/edgex-go/internal/pkg/config"

type ConfigurationStruct struct {
	Writable  WritableInfo
	Clients   map[string]config.ClientInfo
	Databases map[string]config.DatabaseInfo
	Logging   config.LoggingInfo
	Registry  config.RegistryInfo
	Service   config.ServiceInfo
	Smtp      SmtpInfo
}

type WritableInfo struct {
	ResendLimit int
	LogLevel    string
}

type SmtpInfo struct {
	Host                 string
	Username             string
	Password             string
	Port                 int
	Sender               string
	EnableSelfSignedCert bool
	Subject              string
}

// The earlier releases do not have Username field and are using Sender field where Usename will
// be used now, to make it backward compatible fallback to Sender, which is signified by the empty
// Username field.
func (s SmtpInfo) CheckUsername() string {
	if s.Username != "" {
		return s.Username
	}
	return s.Sender
}
