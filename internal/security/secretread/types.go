/*******************************************************************************
 * Copyright 2020 Redis Labs
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
 * @author: Diana Atanasova
 * @author: Andre Srinivasan
 *******************************************************************************/
package secretread

import "github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"

type Configuration struct {
	Writable    WritableInfo
	Service     ServiceInfo
	SecretStore SecretStoreInfo
	Databases   map[string]DatabaseInfo
}

type SecretStoreInfo struct {
	vault.SecretConfig
	TokenFile    string
	PasswordFile string
}

type ServiceInfo struct {
	// BootTimeout indicates, in milliseconds, how long the service will retry connecting to mongo database
	// before giving up. Default is 30,000.
	BootTimeout int
}

type DatabaseInfo struct {
	Username string
	Password string
}

type WritableInfo struct {
	LogLevel       string
	RequestTimeout int
}
