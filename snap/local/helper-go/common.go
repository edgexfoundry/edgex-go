package main

// snapped apps
const (
	// core services
	coreData       = "core-data"
	coreMetadata   = "core-metadata"
	coreCommand    = "core-command"
	consul         = "consul"
	redis          = "redis"
	registry       = consul
	configProvider = consul
	// support services
	supportNotifications = "support-notifications"
	supportScheduler     = "support-scheduler"
	// security services
	securitySecretStore        = "security-secret-store"
	securitySecretStoreSetup   = "security-secretstore-setup"
	securityProxy              = "security-proxy"
	securityProxySetup         = "security-proxy-setup"
	securityBootstrapper       = "security-bootstrapper"
	securityBootstrapperRedis  = "security-bootstrapper-redis"
	securityBootstrapperConsul = "security-consul-bootstrapper"
	securityFileTokenProvider  = "security-file-token-provider"
	secretsConfig              = "secrets-config"
	kong                       = "kong-daemon"
	postgres                   = "postgres"
	vault                      = "vault"
	secretsConfigProcessor     = "secrets-config-processor"
)

var (
	securityServices = []string{
		postgres,
		kong,
		vault,
	}
	securitySetupServices = []string{
		securitySecretStoreSetup,
		securityBootstrapperConsul,
		securityProxySetup,
		securityBootstrapperRedis,
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

func allServices() (s []string) {
	s = make([]string, len(coreServices)+len(supportServices)+len(securityServices)+len(securitySetupServices))
	s = append(s, coreServices...)
	s = append(s, supportServices...)
	s = append(s, securityServices...)
	s = append(s, securitySetupServices...)
	return s
}
