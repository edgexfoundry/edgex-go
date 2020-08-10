/*******************************************************************************
 * Copyright 2019 Dell Inc.
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
 *******************************************************************************/

package secretstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/container"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/pkg/token/fileioperformer"
)

type Bootstrap struct {
	insecureSkipVerify bool
	vaultInterval      int
}

func NewBootstrap(insecureSkipVerify bool, vaultInterval int) *Bootstrap {
	return &Bootstrap{
		insecureSkipVerify: insecureSkipVerify,
		vaultInterval:      vaultInterval,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) BootstrapHandler(ctx context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	configuration := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	//step 1: boot up secretstore general steps same as other EdgeX microservice

	//step 2: initialize the communications
	fileOpener := fileioperformer.NewDefaultFileIoPerformer()

	var req internal.HttpCaller
	if caFilePath := configuration.SecretService.CaFilePath; caFilePath != "" {
		lc.Info("using certificate verification for secret store connection")
		caReader, err := fileOpener.OpenFileReader(caFilePath, os.O_RDONLY, 0400)
		if err != nil {
			lc.Error(fmt.Sprintf("failed to load CA certificate: %s", err.Error()))
			return false
		}
		req = secretstoreclient.NewRequestor(lc).WithTLS(caReader, configuration.SecretService.ServerName)
	} else {
		lc.Info("bypassing certificate verification for secret store connection")
		req = secretstoreclient.NewRequestor(lc).Insecure()
	}

	vaultScheme := configuration.SecretService.Scheme
	vaultHost := fmt.Sprintf("%s:%v", configuration.SecretService.Server, configuration.SecretService.Port)
	intervalDuration := time.Duration(b.vaultInterval) * time.Second
	vc := secretstoreclient.NewSecretStoreClient(lc, req, vaultScheme, vaultHost)
	var initResponse secretstoreclient.InitResponse // reused many places in below flow

	//step 3: initialize and unseal Vault
	for shouldContinue := true; shouldContinue; {
		// Anonymous function used to prevent file handles from accumulating
		func() {
			sCode, _ := vc.HealthCheck()

			switch sCode {
			case http.StatusOK:
				// Load the init response from disk since we need it to regenerate root token later
				if err := loadInitResponse(lc, fileOpener, configuration.SecretService, &initResponse); err != nil {
					lc.Error(fmt.Sprintf("unable to load init response: %s", err.Error()))
					os.Exit(1)
				}
				lc.Info(fmt.Sprintf("vault is initialized and unsealed (status code: %d)", sCode))
				shouldContinue = false
			case http.StatusTooManyRequests:
				lc.Error(fmt.Sprintf("vault is unsealed and in standby mode (Status Code: %d)", sCode))
				shouldContinue = false
			case http.StatusNotImplemented:
				lc.Info(fmt.Sprintf("vault is not initialized (status code: %d). Starting initialization and unseal phases", sCode))
				_, err := vc.Init(configuration.SecretService.VaultSecretThreshold,
					configuration.SecretService.VaultSecretShares, &initResponse)
				if configuration.SecretService.RevokeRootTokens {
					// Never persist the root token to disk on secret store initialization if we intend to revoke it later
					initResponse.RootToken = ""
					lc.Info("Root token stripped from init response for security reasons")
				}
				if err := saveInitResponse(lc, fileOpener, configuration.SecretService, &initResponse); err != nil {
					lc.Error(fmt.Sprintf("unable to save init response: %s", err.Error()))
					os.Exit(1)
				}
				_, err = vc.Unseal(&initResponse)
				if err == nil {
					shouldContinue = false
				}
			case http.StatusServiceUnavailable:
				lc.Info(fmt.Sprintf("vault is sealed (status code: %d). Starting unseal phase", sCode))
				if err := loadInitResponse(lc, fileOpener, configuration.SecretService, &initResponse); err != nil {
					lc.Error(fmt.Sprintf("unable to load init response: %s", err.Error()))
					os.Exit(1)
				}
				_, err := vc.Unseal(&initResponse)
				if err == nil {
					shouldContinue = false
				}
			default:
				if sCode == 0 {
					lc.Error(fmt.Sprintf("vault is in an unknown state. No Status code available"))
				} else {
					lc.Error(fmt.Sprintf("vault is in an unknown state. Status code: %d", sCode))
				}
			}
		}()

		if shouldContinue {
			lc.Info(fmt.Sprintf("trying Vault init/unseal again in %d seconds", b.vaultInterval))
			time.Sleep(intervalDuration)
		}
	}

	/* After vault is init'd and unsealed, it takes a while to get ready to accept any request. During which period any request will get http 500 error.
	We need to check the status constantly until it return http StatusOK.
	*/
	ticker := time.NewTicker(time.Second)
	healthOkCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if sCode, _ := vc.HealthCheck(); sCode == http.StatusOK {
					close(healthOkCh)
					ticker.Stop()
					return
				}
			}
		}
	}()

	// Wait on a StatusOK response from vc.HealthCheck()
	<-healthOkCh

	// create new root token
	// defer revoke token
	// optional: revoke other root token
	// revoke old tokens
	// create delegate credential
	// spawn token provider
	// create db credentials
	// upload kong certificate
	tokenMaintenance := NewTokenMaintenance(lc, vc)

	// Create a transient root token from the key shares
	var rootToken string
	err := vc.RegenRootToken(&initResponse, &rootToken)
	if err != nil {
		lc.Error(fmt.Sprintf("could not regenerate root token %s", err.Error()))
		os.Exit(1)
	}
	defer func() {
		// Revoke transient root token at the end of this funciton
		lc.Info("revoking temporary root token")
		_, err := vc.RevokeSelf(rootToken)
		if err != nil {
			lc.Error(fmt.Sprintf("could not revoke temporary root token %s", err.Error()))
		}
	}()
	lc.Info("generated transient root token")

	// Revoke the other root tokens
	if configuration.SecretService.RevokeRootTokens {
		if initResponse.RootToken != "" {
			initResponse.RootToken = ""
			if err := saveInitResponse(lc, fileOpener, configuration.SecretService, &initResponse); err != nil {
				lc.Error(fmt.Sprintf("unable to save init response: %s", err.Error()))
				os.Exit(1)
			}
			lc.Info("Root token stripped from init response (on disk) for security reasons")
		}
		if err = tokenMaintenance.RevokeRootTokens(rootToken); err != nil {
			lc.Warn(fmt.Sprintf("failed to revoke non-transient root tokens %s", err.Error()))
		}
		lc.Info("completed cleanup of old root tokens")
	} else {
		lc.Info("not revoking existing root tokens")
	}

	// Revoke non-root tokens from previous runs
	err = tokenMaintenance.RevokeNonRootTokens(rootToken)
	if err != nil {
		lc.Warn("failed to revoke non-root tokens")
	}
	lc.Info("completed cleanup of old admin/service tokens")

	// If configured to do so, create a token issuing token
	if configuration.SecretService.TokenProviderAdminTokenPath != "" {
		revokeIssuingTokenFuc, err := makeTokenIssuingToken(lc, configuration, tokenMaintenance, fileOpener, rootToken)
		if err != nil {
			lc.Error(fmt.Sprintf("failed to create token issuing token %s", err.Error()))
			os.Exit(1)
		}
		if configuration.SecretService.TokenProviderType == OneShotProvider {
			// Revoke the admin token at the end of the current function if running a one-shot provider
			// otherwise assume the token provider will keep its token fresh after this point
			defer revokeIssuingTokenFuc()
		}
	}

	//Step 4: Launch token handler
	tokenProvider := NewTokenProvider(ctx, lc, NewDefaultExecRunner())
	if configuration.SecretService.TokenProvider != "" {
		if err := tokenProvider.SetConfiguration(configuration.SecretService); err != nil {
			lc.Error(fmt.Sprintf("failed to configure token provider: %s", err.Error()))
			os.Exit(1)
		}
		if err := tokenProvider.Launch(); err != nil {
			lc.Error(fmt.Sprintf("token provider failed: %s", err.Error()))
			os.Exit(1)
		}
	} else {
		lc.Info("no token provider configured")
	}

	// Enable KV secret engine
	err = enableKVSecretsEngine(lc, vc, rootToken)
	if err != nil {
		lc.Error(fmt.Sprintf("failed to enable KV secrets engine: %s", err.Error()))
		os.Exit(1)
	}

	// credential creation
	gen := NewPasswordGenerator(lc, configuration.SecretService.PasswordProvider, configuration.SecretService.PasswordProviderArgs)
	cred := NewCred(req, rootToken, gen, configuration.SecretService.GetSecretSvcBaseURL(), lc)

	// continue credential creation

	// A little note on why there are two secrets paths. For each microservice, the
	// username/password is uploaded to the vault on both /v1/secret/edgex/%s/mongodb and
	// /v1/secret/edgex/mongo/%s). The go-mod-secrets client requires a Path property to prefix all
	// secrets. docker-edgex-mongo uses that
	// (https://github.com/edgexfoundry/docker-edgex-mongo/blob/master/cmd/res/configuration.toml) in
	// order to enumerate the users and passwords when setting up the initial database authentication.
	// So edgex/%s/monodb is for the microservices (microservices are restricted to their specific
	// edgex/%s), and edgex/mongo/* is enumerated to initialize the database.

	// Redis 5.x only supports a single shared password. When Redis 6 is released, this can be updated
	// to a per service password.

	redis5Password, err := cred.GeneratePassword(ctx)
	if err != nil {
		lc.Error("failed to generate redis5 password")
		os.Exit(1)
	}
	redis5Pair := UserPasswordPair{
		User:     "redis5",
		Password: redis5Password,
	}

	for dbname, info := range configuration.Databases {
		service := info.Service
		// generate credentials
		password, err := cred.GeneratePassword(ctx)
		if err != nil {
			lc.Error(fmt.Sprintf("failed to generate credential pair for service %s", service))
			os.Exit(1)
		}
		pair := UserPasswordPair{
			User:     info.Username,
			Password: password,
		}

		// add credentials to service path if specified and they're not already there
		if len(service) != 0 {
			err = addServiceCredential(lc, "mongodb", cred, service, pair)
			if err != nil {
				lc.Error(err.Error())
				os.Exit(1)
			}

			err = addServiceCredential(lc, "redisdb", cred, service, redis5Pair)
			if err != nil {
				lc.Error(err.Error())
				os.Exit(1)
			}
		}

		err = addDBCredential(lc, "mongo", cred, dbname, pair)
		if err != nil {
			lc.Error(err.Error())
			os.Exit(1)
		}
	}

	// XXX Collapse addServiceCredential and addDBCredential together by passing in the path or using
	// variadic functions
	err = addDBCredential(lc, "redis", cred, "redis5", redis5Pair)
	if err != nil {
		lc.Error(err.Error())
		os.Exit(1)
	}

	cert := NewCerts(req, configuration.SecretService.CertPath, rootToken, configuration.SecretService.GetSecretSvcBaseURL(), lc)
	existing, err := cert.AlreadyinStore()
	if err != nil {
		lc.Error(err.Error())
		os.Exit(1)
	}

	if existing == true {
		lc.Info("proxy certificate pair are in the secret store already, skip uploading")
		return false
	}

	lc.Info("proxy certificate pair are not in the secret store yet, uploading them")
	cp, err := cert.ReadFrom(configuration.SecretService.CertFilePath, configuration.SecretService.KeyFilePath)
	if err != nil {
		lc.Error("failed to get certificate pair from volume")
		os.Exit(1)
	}

	lc.Info("proxy certificate pair are loaded from volume successfully, will upload to secret store")

	err = cert.UploadToStore(cp)
	if err != nil {
		lc.Error("failed to upload the proxy cert pair into the secret store")
		lc.Error(err.Error())
		os.Exit(1)
	}

	lc.Info("proxy certificate pair are uploaded to secret store successfully, Vault init done successfully")
	return false
}

func addServiceCredential(lc logger.LoggingClient, db string, cred Cred, service string, pair UserPasswordPair) error {
	path := fmt.Sprintf("/v1/secret/edgex/%s/%s", service, db)
	existing, err := cred.AlreadyInStore(path)
	if err != nil {
		return err
	}
	if !existing {
		err = cred.UploadToStore(&pair, path)
		if err != nil {
			lc.Error(fmt.Sprintf("failed to upload credential pair for %s on path %s", service, path))
			return err
		}
	} else {
		lc.Info(fmt.Sprintf("credentials for %s already present at path %s", service, path))
	}

	return err
}

func addDBCredential(lc logger.LoggingClient, db string, cred Cred, service string, pair UserPasswordPair) error {
	path := fmt.Sprintf("/v1/secret/edgex/%s/%s", db, service)
	existing, err := cred.AlreadyInStore(path)
	if err != nil {
		lc.Error(err.Error())
		return err
	}
	if !existing {
		err = cred.UploadToStore(&pair, path)
		if err != nil {
			lc.Error(fmt.Sprintf("failed to upload credential pair for db %s on path %s", service, path))
			return err
		}
	} else {
		lc.Info(fmt.Sprintf("credentials for %s already present at path %s", service, path))
	}

	return err
}

func makeTokenIssuingToken(
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	tokenMaintenance *TokenMaintenance,
	fileOpener fileioperformer.FileIoPerformer,
	rootToken string) (RevokeFunc, error) {

	configAdminTokenPath := configuration.SecretService.TokenProviderAdminTokenPath
	if configAdminTokenPath == "" {
		err := fmt.Errorf("TokenProviderAdminTokenPath is a required configuration setting")
		lc.Error(err.Error())
		return nil, err
	}

	// Create delegate credential for use by the token provider
	tokenIssuingToken, revokeIssuingTokenFuc, err := tokenMaintenance.CreateTokenIssuingToken(rootToken)
	if err != nil {
		lc.Error(fmt.Sprintf("failed to create token issuing token %s", err.Error()))
		return nil, err
	}
	lc.Info("created token issuing token")

	// Write the token issuing token to disk to pass it to the token provider
	adminTokenPath, err := filepath.Abs(configAdminTokenPath)
	if err != nil {
		lc.Error(fmt.Sprintf("failed to convert to absolute path %s: %s", configAdminTokenPath, err.Error()))
		revokeIssuingTokenFuc()
		return nil, err
	}
	dirOfAdminToken := filepath.Dir(adminTokenPath)
	err = fileOpener.MkdirAll(dirOfAdminToken, 0700)
	if err != nil {
		lc.Error(fmt.Sprintf("failed to create tokenpath base dir: %s", err.Error()))
		revokeIssuingTokenFuc()
		return nil, err
	}
	tokenWriter, err := fileOpener.OpenFileWriter(adminTokenPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		lc.Error(fmt.Sprintf("failed to create token issuing file %s: %s", adminTokenPath, err.Error()))
		revokeIssuingTokenFuc()
		return nil, err
	}

	encoder := json.NewEncoder(tokenWriter)
	if encoder == nil {
		err := fmt.Errorf("failed to create token encoder")
		lc.Error(err.Error())
		tokenWriter.Close()
		revokeIssuingTokenFuc()
		return nil, err
	}

	if err = encoder.Encode(tokenIssuingToken); err != nil {
		lc.Error(fmt.Sprintf("failed to write token issing token: %s", err.Error()))
		tokenWriter.Close()
		revokeIssuingTokenFuc()
		return nil, err
	}

	if err = tokenWriter.Close(); err != nil {
		lc.Error(fmt.Sprintf("failed to close token issuing file: %s", err.Error()))
		revokeIssuingTokenFuc()
		return nil, err
	}

	return revokeIssuingTokenFuc, nil
}

func enableKVSecretsEngine(
	lc logger.LoggingClient,
	vc secretstoreclient.SecretStoreClient,
	rootToken string) error {

	installed, err := vc.CheckSecretEngineInstalled(rootToken, "secret/", "kv")
	if err != nil {
		lc.Error(fmt.Sprintf("failed call to check if kv secrets engine is installed: %s", err.Error()))
		return err
	}
	if !installed {
		lc.Info("enabling KV secrets engine for the first time...")
		// Enable KV version 1 at /v1/secret path (/v1 prefix supplied by Vault)
		_, err := vc.EnableKVSecretEngine(rootToken, "secret", "1")
		if err != nil {
			lc.Error(fmt.Sprintf("failed call to enable KV secrets engine: %s", err.Error()))
			return err
		}
	} else {
		lc.Info("KV secrets engine already enabled...")
	}
	return nil
}

func loadInitResponse(
	lc logger.LoggingClient,
	fileOpener fileioperformer.FileIoPerformer,
	secretConfig secretstoreclient.SecretServiceInfo,
	initResponse *secretstoreclient.InitResponse) error {

	absPath := filepath.Join(secretConfig.TokenFolderPath, secretConfig.TokenFile)

	tokenFile, err := fileOpener.OpenFileReader(absPath, os.O_RDONLY, 0400)
	if err != nil {
		lc.Error(fmt.Sprintf("could not read master key shares file %s: %s", absPath, err.Error()))
		return err
	}
	tokenFileCloseable := fileioperformer.MakeReadCloser(tokenFile)
	defer tokenFileCloseable.Close()

	decoder := json.NewDecoder(tokenFileCloseable)
	if decoder == nil {
		err := errors.New("Failed to create JSON decoder")
		lc.Error(err.Error())
		return err
	}
	if err := decoder.Decode(initResponse); err != nil {
		lc.Error(fmt.Sprintf("unable to read token file at %s with error: %s", absPath, err.Error()))
		return err
	}

	return nil
}

func saveInitResponse(
	lc logger.LoggingClient,
	fileOpener fileioperformer.FileIoPerformer,
	secretConfig secretstoreclient.SecretServiceInfo,
	initResponse *secretstoreclient.InitResponse) error {

	absPath := filepath.Join(secretConfig.TokenFolderPath, secretConfig.TokenFile)

	tokenFile, err := fileOpener.OpenFileWriter(absPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		lc.Error(fmt.Sprintf("could not read master key shares file %s: %s", absPath, err.Error()))
		return err
	}

	encoder := json.NewEncoder(tokenFile)
	if encoder == nil {
		err := errors.New("Failed to create JSON encoder")
		lc.Error(err.Error())
		_ = tokenFile.Close()
		return err
	}
	if err := encoder.Encode(initResponse); err != nil {
		lc.Error(fmt.Sprintf("unable to write token file at %s with error: %s", absPath, err.Error()))
		_ = tokenFile.Close()
		return err
	}

	if err := tokenFile.Close(); err != nil {
		lc.Error(fmt.Sprintf("unable to close token file at %s with error: %s", absPath, err.Error()))
		_ = tokenFile.Close()
		return err
	}

	return nil
}
