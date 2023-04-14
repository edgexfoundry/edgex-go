//
// Copyright (c) 2020-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package shared

import (
	"context"
	"crypto/sha256"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/common"
	"github.com/edgexfoundry/edgex-go/internal/security/kdf"
	"github.com/edgexfoundry/edgex-go/internal/security/pipedhexreader"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore"
	secretStoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/v3/secrets"
)

const (
	UserPassMountPoint = "userpass"
	JWTIdentityKey     = "edgex-identity"
	UserPolicyPrefix   = "edgex-user-"
)

type CredentialStruct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ProxyUserCommon struct {
	loggingClient     logger.LoggingClient
	configuration     *secretStoreConfig.ConfigurationStruct
	fileOpener        fileioperformer.FileIoPerformer
	httpCaller        internal.HttpCaller
	secretStoreClient secrets.SecretStoreClient
}

// NewProxyUserCommon has common logic for adding and deleting users from Vault
func NewProxyUserCommon(
	lc logger.LoggingClient,
	configuration *secretStoreConfig.ConfigurationStruct) (ProxyUserCommon, error) {

	vb := ProxyUserCommon{
		loggingClient: lc,
		configuration: configuration,
		fileOpener:    fileioperformer.NewDefaultFileIoPerformer(),
	}

	// Finish initializing cmd.httpCaller

	if caFilePath := vb.configuration.SecretStore.CaFilePath; caFilePath != "" {
		lc.Info("using certificate verification for secret store connection")
		caReader, err := vb.fileOpener.OpenFileReader(caFilePath, os.O_RDONLY, 0400)
		if err != nil {
			lc.Errorf("failed to load CA certificate: %s", err.Error())
			return vb, err
		}
		vb.httpCaller = pkg.NewRequester(lc).WithTLS(caReader, vb.configuration.SecretStore.ServerName)
	} else {
		lc.Info("bypassing certificate verification for secret store connection")
		vb.httpCaller = pkg.NewRequester(lc).Insecure()
	}

	// Finish initializing cmd.secretStoreClient

	clientConfig := types.SecretConfig{
		Type:     vb.configuration.SecretStore.Type,
		Protocol: vb.configuration.SecretStore.Protocol,
		Host:     vb.configuration.SecretStore.Host,
		Port:     vb.configuration.SecretStore.Port,
	}

	ssClient, err := secrets.NewSecretStoreClient(clientConfig, vb.loggingClient, vb.httpCaller)
	if err != nil {
		lc.Errorf("failed to create SecretStoreClient: %s", err.Error())
		return vb, err
	}
	vb.secretStoreClient = ssClient

	lc.Info("SecretStoreClient created")

	return vb, err
}

// LoadServiceToken loads a vault token from SecretStore.TokenFile (secrets-token.json)
func (vb *ProxyUserCommon) LoadServiceToken() (string, func(), error) {

	// This is not a root token; don't need to revoke when we're done with it
	revokeFunc := func() {}

	tokenLoader := authtokenloader.NewAuthTokenLoader(vb.fileOpener)

	tokenPath := filepath.Join(vb.configuration.SecretStore.TokenFolderPath, vb.configuration.SecretStore.TokenFile)

	// Reload token in case new token was created causing the auth error
	token, err := tokenLoader.Load(tokenPath)
	if err != nil {
		return "", revokeFunc, err
	}

	return token, revokeFunc, nil

}

// LoadRootToken regenerates a temporary root token from Vault keyshares
func (vb *ProxyUserCommon) LoadRootToken() (string, func(), error) {
	pipedHexReader := pipedhexreader.NewPipedHexReader()
	keyDeriver := kdf.NewKdf(vb.fileOpener, vb.configuration.SecretStore.TokenFolderPath, sha256.New)
	vmkEncryption := secretstore.NewVMKEncryption(vb.fileOpener, pipedHexReader, keyDeriver)

	hook := os.Getenv("EDGEX_IKM_HOOK")
	if len(hook) > 0 {
		err := vmkEncryption.LoadIKM(hook)
		defer vmkEncryption.WipeIKM() // Ensure IKM is wiped from memory
		if err != nil {
			vb.loggingClient.Errorf("failed to setup vault master key encryption: %s", err.Error())
			return "", nil, err
		}
		vb.loggingClient.Info("Enabled encryption of Vault master key")
	} else {
		vb.loggingClient.Info("vault master key encryption not enabled. EDGEX_IKM_HOOK not set.")
	}

	var initResponse types.InitResponse
	if err := secretstore.LoadInitResponse(vb.loggingClient, vb.fileOpener, vb.configuration.SecretStore, &initResponse); err != nil {
		vb.loggingClient.Errorf("unable to load init response: %s", err.Error())
		return "", nil, err
	}

	// Create a transient root token from the key shares
	var rootToken string
	rootToken, err := vb.secretStoreClient.RegenRootToken(initResponse.Keys)
	if err != nil {
		vb.loggingClient.Errorf("could not regenerate root token %s", err.Error())
		return "", nil, err
	}
	revokeFunc := func() {
		// Revoke transient root token at the end of this function
		vb.loggingClient.Info("revoking temporary root token")
		err := vb.secretStoreClient.RevokeToken(rootToken)
		if err != nil {
			vb.loggingClient.Errorf("could not revoke temporary root token %s", err.Error())
		}
	}
	vb.loggingClient.Info("generated transient root token")

	return rootToken, revokeFunc, nil
}

// DoAddUser creates an identity and a password auth binding for it
func (vb *ProxyUserCommon) DoAddUser(privilegedToken string, username string, tokenTTL string, jwtAudience string, jwtTTL string) (CredentialStruct, error) {
	credentialGenerator := secretstore.NewDefaultCredentialGenerator()

	vb.loggingClient.Infof("using policy/token defaults for user %s", username)
	userPolicy := common.MakeDefaultTokenPolicy(username)
	defaultPolicyPaths := userPolicy["path"].(map[string]interface{})
	for pathKey, policy := range defaultPolicyPaths {
		userPolicy["path"].(map[string]interface{})[pathKey] = policy
	}

	randomPassword, err := credentialGenerator.Generate(context.TODO())
	if err != nil {
		return CredentialStruct{}, err
	}

	userManager := common.NewUserManager(vb.loggingClient, vb.secretStoreClient, UserPassMountPoint, JWTIdentityKey, privilegedToken, tokenTTL, jwtAudience, jwtTTL)

	err = userManager.CreatePasswordUserWithPolicy(username, randomPassword, UserPolicyPrefix, userPolicy)
	if err != nil {
		return CredentialStruct{}, err
	}

	return CredentialStruct{
		Username: username,
		Password: randomPassword,
	}, nil
}

// DoDeleteUser performs user deletion
func (vb *ProxyUserCommon) DoDeleteUser(privilegedToken string, username string) error {

	userManager := common.NewUserManager(vb.loggingClient, vb.secretStoreClient, UserPassMountPoint, JWTIdentityKey, privilegedToken, "", "", "")

	err := userManager.DeletePasswordUser(username)
	if err != nil {
		return err
	}

	return nil
}
