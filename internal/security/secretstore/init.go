/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Inc.
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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/kdf"
	"github.com/edgexfoundry/edgex-go/internal/security/pipedhexreader"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/container"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/v2/secrets"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
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
	secretStoreConfig := configuration.SecretStore
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	//step 1: boot up secretstore general steps same as other EdgeX microservice

	//step 2: initialize the communications
	fileOpener := fileioperformer.NewDefaultFileIoPerformer()

	var httpCaller internal.HttpCaller
	if caFilePath := secretStoreConfig.CaFilePath; caFilePath != "" {
		lc.Info("using certificate verification for secret store connection")
		caReader, err := fileOpener.OpenFileReader(caFilePath, os.O_RDONLY, 0400)
		if err != nil {
			lc.Errorf("failed to load CA certificate: %s", err.Error())
			return false
		}
		httpCaller = pkg.NewRequester(lc).WithTLS(caReader, secretStoreConfig.ServerName)
	} else {
		lc.Info("bypassing certificate verification for secret store connection")
		httpCaller = pkg.NewRequester(lc).Insecure()
	}

	intervalDuration := time.Duration(b.vaultInterval) * time.Second
	clientConfig := types.SecretConfig{
		Type:     secretStoreConfig.Type,
		Protocol: secretStoreConfig.Protocol,
		Host:     secretStoreConfig.Host,
		Port:     secretStoreConfig.Port,
	}
	client, err := secrets.NewSecretStoreClient(clientConfig, lc, httpCaller)
	if err != nil {
		lc.Errorf("failed to create SecretStoreClient: %s", err.Error())
		return false
	}

	lc.Info("SecretStoreClient created")

	pipedHexReader := pipedhexreader.NewPipedHexReader()
	keyDeriver := kdf.NewKdf(fileOpener, secretStoreConfig.TokenFolderPath, sha256.New)
	vmkEncryption := NewVMKEncryption(fileOpener, pipedHexReader, keyDeriver)

	hook := os.Getenv("IKM_HOOK")
	if len(hook) > 0 {
		err := vmkEncryption.LoadIKM(hook)
		defer vmkEncryption.WipeIKM() // Ensure IKM is wiped from memory
		if err != nil {
			lc.Errorf("failed to setup vault master key encryption: %s", err.Error())
			return false
		}
		lc.Info("Enabled encryption of Vault master key")
	} else {
		lc.Info("vault master key encryption not enabled. IKM_HOOK not set.")
	}

	var initResponse types.InitResponse // reused many places in below flow

	//step 3: initialize and unseal Vault
	for shouldContinue := true; shouldContinue; {
		// Anonymous function used to prevent file handles from accumulating
		terminalFailure := func() bool {
			sCode, _ := client.HealthCheck()

			switch sCode {
			case http.StatusOK:
				// Load the init response from disk since we need it to regenerate root token later
				if err := loadInitResponse(lc, fileOpener, secretStoreConfig, &initResponse); err != nil {
					lc.Errorf("unable to load init response: %s", err.Error())
					return true
				}
				lc.Infof("vault is initialized and unsealed (status code: %d)", sCode)
				shouldContinue = false

			case http.StatusTooManyRequests:
				// we're done here. Will go into ready mode or reseal
				shouldContinue = false

			case http.StatusNotImplemented:
				lc.Infof("vault is not initialized (status code: %d). Starting initialization and unseal phases", sCode)
				initResponse, err = client.Init(secretStoreConfig.VaultSecretThreshold, secretStoreConfig.VaultSecretShares)
				if err != nil {
					lc.Errorf("Unable to Initialize Vault: %s. Will try again...", err.Error())
					// Not terminal failure, should continue and try again
					return false
				}

				if secretStoreConfig.RevokeRootTokens {
					// Never persist the root token to disk on secret store initialization if we intend to revoke it later
					initResponse.RootToken = ""
					lc.Info("Root token stripped from init response for security reasons")
				}

				err = client.Unseal(initResponse.KeysBase64)
				if err != nil {
					lc.Errorf("Unable to unseal Vault: %s", err.Error())
					return true
				}

				// We need the unencrypted initResponse in order to generate a temporary root token later
				// Make a copy and save the copy, possibly encrypted
				encryptedInitResponse := initResponse
				// Optionally encrypt the vault init response based on whether encryption was enabled
				if vmkEncryption.IsEncrypting() {
					if err := vmkEncryption.EncryptInitResponse(&encryptedInitResponse); err != nil {
						lc.Errorf("failed to encrypt init response from secret store: %s", err.Error())
						return true
					}
				}
				if err := saveInitResponse(lc, fileOpener, secretStoreConfig, &encryptedInitResponse); err != nil {
					lc.Errorf("unable to save init response: %s", err.Error())
					return true
				}

			case http.StatusServiceUnavailable:
				lc.Infof("vault is sealed (status code: %d). Starting unseal phase", sCode)
				if err := loadInitResponse(lc, fileOpener, secretStoreConfig, &initResponse); err != nil {
					lc.Errorf("unable to load init response: %s", err.Error())
					return true
				}
				// Optionally decrypt the vault init response based on whether encryption was enabled
				if vmkEncryption.IsEncrypting() {
					err = vmkEncryption.DecryptInitResponse(&initResponse)
					if err != nil {
						lc.Errorf("failed to decrypt key shares for secret store unsealing: %s", err.Error())
						return true
					}
				}

				err := client.Unseal(initResponse.KeysBase64)
				if err == nil {
					shouldContinue = false
				}

			default:
				if sCode == 0 {
					lc.Errorf("vault is in an unknown state. No Status code available")
				} else {
					lc.Errorf("vault is in an unknown state. Status code: %d", sCode)
				}
			}

			return false
		}()

		if terminalFailure {
			return false
		}

		if shouldContinue {
			lc.Infof("trying Vault init/unseal again in %d seconds", b.vaultInterval)
			time.Sleep(intervalDuration)
		}
	}

	/* After vault is initialized and unsealed, it takes a while to get ready to accept any request. During which period any request will get http 500 error.
	We need to check the status constantly until it return http StatusOK.
	*/
	ticker := time.NewTicker(time.Second)
	healthOkCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if sCode, _ := client.HealthCheck(); sCode == http.StatusOK {
					close(healthOkCh)
					ticker.Stop()
					return
				}
			}
		}
	}()

	// Wait on a StatusOK response from client.HealthCheck()
	<-healthOkCh

	// create new root token
	// defer revoke token
	// optional: revoke other root token
	// revoke old tokens
	// create delegate credential
	// spawn token provider
	// create db credentials
	// upload kong certificate
	tokenMaintenance := NewTokenMaintenance(lc, client)

	// Create a transient root token from the key shares
	var rootToken string
	rootToken, err = client.RegenRootToken(initResponse.Keys)
	if err != nil {
		lc.Errorf("could not regenerate root token %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		// Revoke transient root token at the end of this function
		lc.Info("revoking temporary root token")
		err := client.RevokeToken(rootToken)
		if err != nil {
			lc.Errorf("could not revoke temporary root token %s", err.Error())
		}
	}()
	lc.Info("generated transient root token")

	// Revoke the other root tokens
	if secretStoreConfig.RevokeRootTokens {
		if initResponse.RootToken != "" {
			initResponse.RootToken = ""
			if err := saveInitResponse(lc, fileOpener, secretStoreConfig, &initResponse); err != nil {
				lc.Errorf("unable to save init response: %s", err.Error())
				os.Exit(1)
			}
			lc.Info("Root token stripped from init response (on disk) for security reasons")
		}
		if err := tokenMaintenance.RevokeRootTokens(rootToken); err != nil {
			lc.Warnf("failed to revoke non-transient root tokens %s", err.Error())
		}
		lc.Info("completed cleanup of old root tokens")
	} else {
		lc.Info("not revoking existing root tokens")
	}

	// Revoke non-root tokens from previous runs
	if err := tokenMaintenance.RevokeNonRootTokens(rootToken); err != nil {
		lc.Warn("failed to revoke non-root tokens")
	}
	lc.Info("completed cleanup of old admin/service tokens")

	// If configured to do so, create a token issuing token
	if secretStoreConfig.TokenProviderAdminTokenPath != "" {
		revokeIssuingTokenFuc, err := makeTokenIssuingToken(lc, configuration, tokenMaintenance, fileOpener, rootToken)
		if err != nil {
			lc.Errorf("failed to create token issuing token %s", err.Error())
			os.Exit(1)
		}
		if secretStoreConfig.TokenProviderType == OneShotProvider {
			// Revoke the admin token at the end of the current function if running a one-shot provider
			// otherwise assume the token provider will keep its token fresh after this point
			defer revokeIssuingTokenFuc()
		}
	}

	//Step 4: Launch token handler
	tokenProvider := NewTokenProvider(ctx, lc, NewDefaultExecRunner())
	if secretStoreConfig.TokenProvider != "" {
		if err := tokenProvider.SetConfiguration(secretStoreConfig); err != nil {
			lc.Errorf("failed to configure token provider: %s", err.Error())
			os.Exit(1)
		}
		if err := tokenProvider.Launch(); err != nil {
			lc.Errorf("token provider failed: %s", err.Error())
			os.Exit(1)
		}
	} else {
		lc.Info("no token provider configured")
	}

	// Enable KV secret engine
	if err := enableKVSecretsEngine(lc, client, rootToken); err != nil {
		lc.Errorf("failed to enable KV secrets engine: %s", err.Error())
		os.Exit(1)
	}

	// credential creation
	gen := NewPasswordGenerator(lc, secretStoreConfig.PasswordProvider, secretStoreConfig.PasswordProviderArgs)
	cred := NewCred(httpCaller, rootToken, gen, secretStoreConfig.GetBaseURL(), lc)

	// continue credential creation

	// A little note on why there are two secrets paths. For each microservice, the
	// username/password is uploaded to the vault on both /v1/secret/edgex/%s/redisdb and
	// /v1/secret/edgex/redisdb/%s). The go-mod-secrets client requires a Path property to prefix all
	// secrets.
	// So edgex/%s/redisdb is for the microservices (microservices are restricted to their specific
	// edgex/%s), and edgex/redisdb/* is enumerated to initialize the database.
	//

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

	for _, info := range configuration.Databases {
		service := info.Service

		// add credentials to service path if specified and they're not already there
		if len(service) != 0 {
			err = addServiceCredential(lc, "redisdb", cred, service, redis5Pair)
			if err != nil {
				lc.Error(err.Error())
				os.Exit(1)
			}
		}
	}

	// security-bootstrap-redis uses the path /v1/secret/edgex/bootstrap-redis/ and go-mod-bootstrap
	// with append the DB type (redisdb)
	err = addDBCredential(lc, "bootstrap-redis", cred, "redisdb", redis5Pair)
	if err != nil {
		lc.Error(err.Error())
		os.Exit(1)
	}

	// Concat all cert path secretStore values together to check for empty values
	certPathCheck := secretStoreConfig.CertPath +
		secretStoreConfig.CertFilePath +
		secretStoreConfig.KeyFilePath

	// If any of the previous three proxy cert path values are present (len > 0), attempt to upload to secret store
	if len(strings.TrimSpace(certPathCheck)) != 0 {

		// Grab the certificate & check to see if it's already in the secret store
		cert := NewCerts(httpCaller, secretStoreConfig.CertPath, rootToken, secretStoreConfig.GetBaseURL(), lc)
		existing, err := cert.AlreadyInStore()
		if err != nil {
			lc.Error(err.Error())
			os.Exit(1)
		}

		if existing {
			lc.Info("proxy certificate pair are in the secret store already, skip uploading")
			return false
		}

		lc.Info("proxy certificate pair are not in the secret store yet, uploading them")
		cp, err := cert.ReadFrom(secretStoreConfig.CertFilePath, secretStoreConfig.KeyFilePath)
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

		lc.Info("proxy certificate pair are uploaded to secret store successfully")

	} else {
		lc.Info("proxy certificate pair upload was skipped because cert secretStore value(s) were blank")
	}

	lc.Info("Vault init done successfully")
	return false

}

// XXX Collapse addServiceCredential and addDBCredential together by passing in the path or using
// variadic functions

func addServiceCredential(lc logger.LoggingClient, db string, cred Cred, service string, pair UserPasswordPair) error {
	path := fmt.Sprintf("/v1/secret/edgex/%s/%s", service, db)
	existing, err := cred.AlreadyInStore(path)
	if err != nil {
		return err
	}
	if !existing {
		err = cred.UploadToStore(&pair, path)
		if err != nil {
			lc.Errorf("failed to upload credential pair for %s on path %s", service, path)
			return err
		}
	} else {
		lc.Infof("credentials for %s already present at path %s", service, path)
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
			lc.Errorf("failed to upload credential pair for db %s on path %s", service, path)
			return err
		}
	} else {
		lc.Infof("credentials for %s already present at path %s", service, path)
	}

	return err
}

func makeTokenIssuingToken(
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	tokenMaintenance *TokenMaintenance,
	fileOpener fileioperformer.FileIoPerformer,
	rootToken string) (RevokeFunc, error) {

	configAdminTokenPath := configuration.SecretStore.TokenProviderAdminTokenPath
	if configAdminTokenPath == "" {
		err := fmt.Errorf("TokenProviderAdminTokenPath is a required configuration setting")
		lc.Error(err.Error())
		return nil, err
	}

	// Create delegate credential for use by the token provider
	tokenIssuingToken, revokeIssuingTokenFuc, err := tokenMaintenance.CreateTokenIssuingToken(rootToken)
	if err != nil {
		lc.Errorf("failed to create token issuing token %s", err.Error())
		return nil, err
	}
	lc.Info("created token issuing token")

	// Write the token issuing token to disk to pass it to the token provider
	adminTokenPath, err := filepath.Abs(configAdminTokenPath)
	if err != nil {
		lc.Errorf("failed to convert to absolute path %s: %s", configAdminTokenPath, err.Error())
		revokeIssuingTokenFuc()
		return nil, err
	}
	dirOfAdminToken := filepath.Dir(adminTokenPath)
	err = fileOpener.MkdirAll(dirOfAdminToken, 0700)
	if err != nil {
		lc.Errorf("failed to create tokenpath base dir: %s", err.Error())
		revokeIssuingTokenFuc()
		return nil, err
	}
	tokenWriter, err := fileOpener.OpenFileWriter(adminTokenPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		lc.Errorf("failed to create token issuing file %s: %s", adminTokenPath, err.Error())
		revokeIssuingTokenFuc()
		return nil, err
	}

	encoder := json.NewEncoder(tokenWriter)
	if encoder == nil {
		err := fmt.Errorf("failed to create token encoder")
		lc.Error(err.Error())
		_ = tokenWriter.Close()
		revokeIssuingTokenFuc()
		return nil, err
	}

	if err = encoder.Encode(tokenIssuingToken); err != nil {
		lc.Errorf("failed to write token issuing token: %s", err.Error())
		_ = tokenWriter.Close()
		revokeIssuingTokenFuc()
		return nil, err
	}

	if err = tokenWriter.Close(); err != nil {
		lc.Errorf("failed to close token issuing file: %s", err.Error())
		revokeIssuingTokenFuc()
		return nil, err
	}

	return revokeIssuingTokenFuc, nil
}

func enableKVSecretsEngine(
	lc logger.LoggingClient,
	client secrets.SecretStoreClient,
	rootToken string) error {

	installed, err := client.CheckSecretEngineInstalled(rootToken, "secret/", "kv")
	if err != nil {
		lc.Errorf("failed call to check if kv secrets engine is installed: %s", err.Error())
		return err
	}
	if !installed {
		lc.Info("enabling KV secrets engine for the first time...")
		// Enable KV version 1 at /v1/secret path (/v1 prefix supplied by Vault)
		err := client.EnableKVSecretEngine(rootToken, "secret", "1")
		if err != nil {
			lc.Errorf("failed call to enable KV secrets engine: %s", err.Error())
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
	secretConfig config.SecretStoreInfo,
	initResponse *types.InitResponse) error {

	absPath := filepath.Join(secretConfig.TokenFolderPath, secretConfig.TokenFile)

	tokenFile, err := fileOpener.OpenFileReader(absPath, os.O_RDONLY, 0400)
	if err != nil {
		lc.Errorf("could not read master key shares file %s: %s", absPath, err.Error())
		return err
	}
	tokenFileCloseable := fileioperformer.MakeReadCloser(tokenFile)
	defer func() { _ = tokenFileCloseable.Close() }()

	decoder := json.NewDecoder(tokenFileCloseable)
	if decoder == nil {
		err := errors.New("Failed to create JSON decoder")
		lc.Error(err.Error())
		return err
	}
	if err := decoder.Decode(initResponse); err != nil {
		lc.Errorf("unable to read token file at %s with error: %s", absPath, err.Error())
		return err
	}

	return nil
}

func saveInitResponse(
	lc logger.LoggingClient,
	fileOpener fileioperformer.FileIoPerformer,
	secretConfig config.SecretStoreInfo,
	initResponse *types.InitResponse) error {

	absPath := filepath.Join(secretConfig.TokenFolderPath, secretConfig.TokenFile)

	tokenFile, err := fileOpener.OpenFileWriter(absPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		lc.Errorf("could not read master key shares file %s: %s", absPath, err.Error())
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
		lc.Errorf("unable to write token file at %s with error: %s", absPath, err.Error())
		_ = tokenFile.Close()
		return err
	}

	if err := tokenFile.Close(); err != nil {
		lc.Errorf("unable to close token file at %s with error: %s", absPath, err.Error())
		_ = tokenFile.Close()
		return err
	}

	return nil
}
