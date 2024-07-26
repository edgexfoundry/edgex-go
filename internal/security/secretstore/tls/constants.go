//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package tls

// TLS file settings
const (
	// General settings
	baseSecretOutputDir = "/tmp/edgex/secrets"
	caKeyFileName       = "ca.key"
	caCertFileName      = "ca.crt"
	opensslConfFileName = "openssl.conf"
	// secretBasePath is copied from the secretstore package to avoid import cycle
	secretBasePath = "/v1/secret/edgex" // nolint:gosec

	// Redis specific settings
	redisTlsCertOutputDir    = baseSecretOutputDir + "/redis"
	redisClientCertOutputDir = redisTlsCertOutputDir + "/client"
	redisServerCertOutputDir = redisTlsCertOutputDir + "/server"
	redisServerCertFileName  = "redis-server.crt"
	redisServerKeyFileName   = "redis-server.key"
	redisTlsSecretName       = "redis-tls"
	redisKeyFileName         = "redis.key"
	redisCsrFileName         = "redis.csr"
	redisCertFileName        = "redis.crt"
	redisHostName            = "redis"

	// Postgres specific settings
	postgresTlsCertOutputDir    = baseSecretOutputDir + "/postgres"
	postgresClientCertOutputDir = postgresTlsCertOutputDir + "/client"
	postgresServerCertOutputDir = postgresTlsCertOutputDir + "/server"
	postgresServerCertFileName  = "postgres-server.crt"
	postgresServerKeyFileName   = "postgres-server.key"
	postgresTlsSecretName       = "postgres-tls"
	postgresKeyFileName         = "postgres.key"
	postgresCsrFileName         = "postgres.csr"
	postgresCertFileName        = "postgres.crt"
	//postgresHostName            = "postgres"
	postgresHostName = "localhost"
)
