/*
 * Copyright (C) 2022 Canonical Ltd
 * Copyright (C) 2023 Intel Corporation
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

// snapped apps
const (
	// core services
	coreData                     = "core-data"
	coreMetadata                 = "core-metadata"
	coreCommand                  = "core-command"
	consul                       = "consul" // registry and config provider
	redis                        = "redis"  // message bus and database
	coreCommonConfigBootstrapper = "core-common-config-bootstrapper"
	// support services
	supportNotifications = "support-notifications"
	supportScheduler     = "support-scheduler"
	// security apps
	nginx                      = "nginx"          // reverse proxy
	securitySecretsConfig      = "secrets-config" // CLI app
	securitySecretStoreSetup   = "security-secretstore-setup"
	securityProxyAuth          = "security-proxy-auth" // auth plugin for nginx
	securityBootstrapper       = "security-bootstrapper"
	securityBootstrapperRedis  = "security-bootstrapper-redis"
	securityBootstrapperConsul = "security-bootstrapper-consul"
	securityBootstrapperNginx  = "security-bootstrapper-nginx"
	securityFileTokenProvider  = "security-file-token-provider"
	vault                      = "vault" // secret store
)

var (
	securityServices = []string{
		vault,
		nginx,
		securityProxyAuth,
	}
	securitySetupServices = []string{
		securitySecretStoreSetup,
		securityBootstrapperConsul,
		securityBootstrapperNginx,
		securityProxyAuth,
		securityBootstrapperRedis,
	}
	coreSetupServices = []string{
		coreCommonConfigBootstrapper,
	}
	coreServices = []string{
		consul,
		redis,
		coreData,
		coreMetadata,
		coreCommand,
	}
	supportServices = []string{
		supportNotifications,
		supportScheduler,
	}
)

func allOneshotServices() (s []string) {
	return append(securitySetupServices, coreSetupServices...)
}

func allServices() (s []string) {
	allOneshotServices := allOneshotServices()
	s = make([]string, 0, len(coreServices)+len(supportServices)+len(securityServices)+len(allOneshotServices))
	s = append(s, coreServices...)
	s = append(s, supportServices...)
	s = append(s, securityServices...)
	s = append(s, allOneshotServices...)
	return s
}
