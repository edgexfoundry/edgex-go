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
	eKuiper              = "kuiper"
	rulesEngine          = eKuiper
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
	// app services
	appServiceConfigurable = "app-service-configurable"
	// management services
	systemManagementAgent = "sys-mgmt-agent"
)
