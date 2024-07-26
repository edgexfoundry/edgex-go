//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package secretstore

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/tls"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/secret"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

type thirdPartyServiceSecretCreator struct {
	lc                   logger.LoggingClient
	config               *config.ConfigurationStruct
	caller               internal.HttpCaller
	rootToken            string
	secretServiceBaseURL string
}

func newThirdPartyServiceSecretCreator(lc logger.LoggingClient, config *config.ConfigurationStruct, caller internal.HttpCaller,
	rootToken string, secretServiceBaseURL string) thirdPartyServiceSecretCreator {
	return thirdPartyServiceSecretCreator{lc, config, caller, rootToken, secretServiceBaseURL}
}

func (c thirdPartyServiceSecretCreator) generateTLSCerts() error {
	return c.generatePostgresTlsCert()
}

// generatePostgresSecrets generates TLS certifications for Postgres security connection
func (c thirdPartyServiceSecretCreator) generatePostgresTlsCert() error {
	if secret.IsSecurityEnabled() {
		err := tls.GeneratePostgresTlsCert(c.lc, c.config, c.caller, c.rootToken, c.secretServiceBaseURL)
		if err != nil {
			return fmt.Errorf("failed to generate TLS certificates for the internal Postgres server: %s", err.Error())
		}
	}
	return nil
}
