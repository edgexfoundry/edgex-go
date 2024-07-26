//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/client"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	envPostgresClients = "Postgres_CLIENTS"
)

// GeneratePostgresTlsCert generates TLS certifications for postgres security connection
func GeneratePostgresTlsCert(lc logger.LoggingClient, config *config.ConfigurationStruct, caller internal.HttpCaller, rootToken string, secretServiceBaseURL string) error {
	lc.Info("Generating TLS certificates for the internal Postgres server")

	// Create postgres server certificates
	caCertFilePath, caKeyFilePath, err := createServerCerts(postgresHostName, postgresServerCertOutputDir, postgresServerKeyFileName, postgresServerCertFileName)
	if err != nil {
		return fmt.Errorf("failed to create Postgres server certificate: %v", err)
	}

	secretClient := client.NewSecretClient(caller, rootToken, secretServiceBaseURL, lc)
	// Create Postgres client certificates
	err = createPostgresClientCerts(secretClient, caKeyFilePath, caCertFilePath)
	if err != nil {
		return fmt.Errorf("failed to create Postgres client certificate: %v", err)
	}

	lc.Info("Successfully generated TLS certificates for the internal Postgres server")
	return nil
}

// createPostgresClientCerts create the Postgres TLS client cert/key for each service defined in the envPostgresClients var and saves the secrets to the secret store
func createPostgresClientCerts(secretClient client.SecretClient, caKeyFilePath, caCertFilePath string) error {
	//envVarClientNames := os.Getenv(envPostgresClients)
	envVarClientNames := "core-data"
	if len(envVarClientNames) == 0 {
		return fmt.Errorf("could not found any Postgres client names defined in %s environment variable", envPostgresClients)
	}
	hostnameList := strings.Split(envVarClientNames, ",")

	// the hardcoded name "postgres-client" is for any 3rd party services
	hostnameList = append(hostnameList, "postgres-client")

	for _, hostname := range hostnameList {
		hostname = strings.TrimSpace(hostname)
		path := fmt.Sprintf("%s/%s/%s", secretBasePath, hostname, postgresTlsSecretName)
		exist, err := secretClient.HasSecret(path, messaging.SecretCACert, messaging.SecretClientCert, messaging.SecretClientKey)
		if err != nil {
			return fmt.Errorf("failed to check if the server certificates reside in Secret Store, err: %v", err)
		}
		if exist {
			continue
		}

		PostgresCertsOutputDir := filepath.Join(postgresClientCertOutputDir, hostname)
		csrFilePath := filepath.Join(PostgresCertsOutputDir, postgresCsrFileName)
		keyFilePath := filepath.Join(PostgresCertsOutputDir, postgresKeyFileName)
		certFilePath := filepath.Join(PostgresCertsOutputDir, postgresCertFileName)

		// Create the output folder for client certificates
		err = os.MkdirAll(PostgresCertsOutputDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create folder for client certificates, err: %v", err)
		}

		// set the CN value to the hostname
		subj := fmt.Sprintf("/CN=%s", hostname)
		// Create a certificate signing request (CSR) for the client
		err = runOpensslCmd([]string{"req", "-newkey", "rsa:2048", "-nodes", "-keyout", keyFilePath,
			"-subj", subj, "-out", csrFilePath})
		if err != nil {
			return fmt.Errorf("failed to generate the TLS certificate signing request, err: %v", err)
		}

		// Sign the client CSR with CA key and cert
		opensslConfFilePath := filepath.Join(postgresClientCertOutputDir, opensslConfFileName)
		err = createOpensslConf(opensslConfFilePath, hostname)
		if err != nil {
			return fmt.Errorf("failed to create the openssl config, err: %v", err)
		}
		err = runOpensslCmd([]string{"x509", "-req", "-extfile", opensslConfFilePath, "-days", "36500",
			"-in", csrFilePath, "-CA", caCertFilePath, "-CAkey", caKeyFilePath, "-CAcreateserial", "-out", certFilePath})
		if err != nil {
			return fmt.Errorf("failed to sign the client CSR, err: %v", err)
		}

		if err = os.Chmod(certFilePath, 0755); err != nil {
			return fmt.Errorf("failed to change client cert file permission, err: %v", err)
		}
		if err = os.Chmod(keyFilePath, 0755); err != nil {
			return fmt.Errorf("failed to change client key file permission, err: %v", err)
		}

		// Copy the CA cert to the client output folder
		caCert, err := os.ReadFile(caCertFilePath)
		if err != nil {
			return fmt.Errorf("failed to read CA certificate from file, err: %v", err)
		}
		copiedCaCertFilePath := filepath.Join(postgresClientCertOutputDir, hostname+"/"+caCertFileName)
		err = os.WriteFile(copiedCaCertFilePath, caCert, 0755) // nolint:gosec
		if err != nil {
			return fmt.Errorf("failed to create CA certificate file for %s, err: %v", hostname, err)
		}

		clientSecretFiles := map[string]string{}
		clientSecretFiles[messaging.SecretCACert] = copiedCaCertFilePath
		clientSecretFiles[messaging.SecretClientCert] = certFilePath
		clientSecretFiles[messaging.SecretClientKey] = keyFilePath
		secrets, err := readSecretFromFile(clientSecretFiles)
		if err != nil {
			return fmt.Errorf("failed to read the client certificates from file, err: %v", err)
		}
		err = secretClient.SetSecrets(secrets, path)
		if err != nil {
			return fmt.Errorf("failed to upload the client certificates to Secret Store, err: %v", err)
		}
	}
	return nil
}
