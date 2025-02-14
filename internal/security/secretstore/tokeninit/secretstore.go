//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package tokeninit

import (
	"errors"
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/container"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/tokencreatable"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/tokenfilewriter"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/tokenmaintenance"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/v4/secrets"
)

// InitAdminTokens reads the resp-init.json file and recreate root tokens and admin token for controller API usage
func InitAdminTokens(dic *di.Container) (tokencreatable.RevokeFunc, error) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	secretStoreConfig := configuration.SecretStore

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()

	var httpCaller internal.HttpCaller
	if caFilePath := secretStoreConfig.CaFilePath; caFilePath != "" {
		lc.Info("using certificate verification for secret store connection")
		caReader, err := fileOpener.OpenFileReader(caFilePath, os.O_RDONLY, 0400)
		if err != nil {
			lc.Errorf("failed to load CA certificate: %w", err)
		}
		httpCaller = pkg.NewRequester(lc).WithTLS(caReader, secretStoreConfig.ServerName)
	} else {
		lc.Info("bypassing certificate verification for secret store connection")
		httpCaller = pkg.NewRequester(lc).Insecure()
	}

	clientConfig := types.SecretConfig{
		Type:     secretStoreConfig.Type,
		Protocol: secretStoreConfig.Protocol,
		Host:     secretStoreConfig.Host,
		Port:     secretStoreConfig.Port,
	}
	secretStoreClient, err := secrets.NewSecretStoreClient(clientConfig, lc, httpCaller)
	if err != nil {
		lc.Errorf("failed to create SecretStoreClient: %w", err)
	}

	var initResponse types.InitResponse // reused many places in below flow

	// Load the init response from disk since we need it to regenerate root token later
	if err := tokenmaintenance.LoadInitResponse(lc, fileOpener, secretStoreConfig, &initResponse); err != nil {
		lc.Errorf("unable to load init response: %w", err)

	}

	tokenMaintenance := tokenmaintenance.NewTokenMaintenance(lc, secretStoreClient)
	// Create a transient root token from the key shares
	var rootToken string
	rootToken, err = secretStoreClient.RegenRootToken(initResponse.Keys)
	if err != nil {
		lc.Errorf("could not regenerate root token %w", err)

	}
	defer func() {
		// Revoke transient root token at the end of this function
		lc.Info("revoking temporary root token")
		err := secretStoreClient.RevokeToken(rootToken)
		if err != nil {
			lc.Errorf("could not revoke temporary root token %w", err)
		}
	}()
	lc.Info("generated transient root token")

	// If configured to do so, create a token issuing token
	if secretStoreConfig.TokenProviderAdminTokenPath != "" {
		revokeIssuingTokenFuc, err := tokenfilewriter.NewWriter(lc, secretStoreClient, fileOpener).
			CreateAndWrite(rootToken, secretStoreConfig.TokenProviderAdminTokenPath, tokenMaintenance.CreateTokenIssuingToken)
		if err != nil {
			lc.Errorf("failed to create token issuing token: %w", err)
		}

		return revokeIssuingTokenFuc, nil
	}
	return nil, errors.New("no TokenProviderAdminTokenPath defined in secretStoreConfig")
}
