// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

// snapped apps
const (
	// core services
	coreData                     = "core-data"
	coreMetadata                 = "core-metadata"
	coreCommand                  = "core-command"
	consul                       = "consul"
	redis                        = "redis"
	registry                     = consul
	configProvider               = consul
	coreCommonConfigBootstrapper = "core-common-config-bootstrapper"
	// support services
	supportNotifications = "support-notifications"
	supportScheduler     = "support-scheduler"
	// security services
	securityNginx              = "nginx"
	securitySecretsConfig      = "secrets-config"
	securitySecretStore        = "security-secret-store"
	securitySecretStoreSetup   = "security-secretstore-setup"
	securityProxy              = "security-proxy"
	securityProxyAuth          = "security-proxy-auth"
	securityBootstrapper       = "security-bootstrapper"
	securityBootstrapperRedis  = "security-bootstrapper-redis"
	securityBootstrapperConsul = "security-consul-bootstrapper"
	securityBootstrapperNginx  = "security-bootstrapper-nginx"
	securityFileTokenProvider  = "security-file-token-provider"
	secretsConfig              = "secrets-config"
	vault                      = "vault"
	secretsConfigProcessor     = "secrets-config-processor"
)

var (
	securityServices = []string{
		vault,
	}
	securitySetupServices = []string{
		securitySecretStoreSetup,
		securityBootstrapperConsul,
		securityBootstrapperNginx,
		securityProxyAuth,
		securityBootstrapperRedis,
		securityNginx,
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
	s = make([]string, len(coreServices)+len(supportServices)+len(securityServices)+len(allOneshotServices()))
	s = append(s, coreServices...)
	s = append(s, supportServices...)
	s = append(s, securityServices...)
	s = append(s, allOneshotServices()...)
	return s
}
