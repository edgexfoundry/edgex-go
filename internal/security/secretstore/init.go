/*******************************************************************************
 * Copyright 2022 Intel Corporation
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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/kdf"
	"github.com/edgexfoundry/edgex-go/internal/security/pipedhexreader"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/container"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/secretsengine"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/tokenfilewriter"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/v2/secrets"
)

const (
	addKnownSecretsEnv   = "ADD_KNOWN_SECRETS"
	redisSecretName      = "redisdb"
	messagebusSecretName = "message-bus"
	knownSecretSeparator = ","
	serviceListBegin     = "["
	serviceListEnd       = "]"
	serviceListSeparator = ";"
	secretBasePath       = "/v1/secret/edgex" // nolint:gosec
	defaultMsgBusUser    = "msgbususer"
)

var errNotFound = errors.New("credential NOT found")

type Bootstrap struct {
	insecureSkipVerify bool
	vaultInterval      int
	validKnownSecrets  map[string]bool
}

func NewBootstrap(insecureSkipVerify bool, vaultInterval int) *Bootstrap {
	return &Bootstrap{
		insecureSkipVerify: insecureSkipVerify,
		vaultInterval:      vaultInterval,
		validKnownSecrets:  map[string]bool{redisSecretName: true, messagebusSecretName: true},
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) BootstrapHandler(ctx context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	configuration := container.ConfigurationFrom(dic.Get)
	secretStoreConfig := configuration.SecretStore
	kongAdminConfig := configuration.KongAdmin
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
			<-ticker.C
			if sCode, _ := client.HealthCheck(); sCode == http.StatusOK {
				close(healthOkCh)
				ticker.Stop()
				return
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
		return false
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
				return false
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
		revokeIssuingTokenFuc, err := tokenfilewriter.NewWriter(lc, client, fileOpener).
			CreateAndWrite(rootToken, secretStoreConfig.TokenProviderAdminTokenPath, tokenMaintenance.CreateTokenIssuingToken)
		if err != nil {
			lc.Errorf("failed to create token issuing token: %s", err.Error())
			return false
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
			return false
		}
		if err := tokenProvider.Launch(); err != nil {
			lc.Errorf("token provider failed: %s", err.Error())
			return false
		}
	} else {
		lc.Info("no token provider configured")
	}

	// Enable KV secret engine
	if err := secretsengine.New(secretsengine.KVSecretsEngineMountPoint, secretsengine.KeyValue).
		Enable(&rootToken, lc, client); err != nil {
		lc.Errorf("failed to enable KV secrets engine: %s", err.Error())
		return false
	}

	knownSecretsToAdd, err := b.getKnownSecretsToAdd()
	if err != nil {
		lc.Error(err.Error())
		return false
	}

	// credential creation
	gen := NewPasswordGenerator(lc, secretStoreConfig.PasswordProvider, secretStoreConfig.PasswordProviderArgs)
	secretStore := NewCred(httpCaller, rootToken, gen, secretStoreConfig.GetBaseURL(), lc)

	// continue credential creation

	// A little note on why there are two secrets paths. For each microservice, the redis
	// username/password is uploaded to the vault on both /v1/secret/edgex/%s/redisdb and
	// /v1/secret/edgex/redisdb/%s). The go-mod-secrets client requires a Path property to prefix all
	// secrets.
	// So edgex/%s/redisdb is for the microservices (microservices are restricted to their specific
	// edgex/%s), and edgex/redisdb/* is enumerated to initialize the database.
	// Similary for secure message bus credential.

	// Redis 5.x only supports a single shared password. When Redis 6 is released, this can be updated
	// to a per service password.

	redisCredentials, err := getCredential("security-bootstrapper-redis", secretStore, redisSecretName)
	if err != nil {
		if err != errNotFound {
			lc.Error("failed to determine if Redis credentials already exist or not: %w", err)
			return false
		}

		lc.Info("Generating new password for Redis DB")
		defaultPassword, err := secretStore.GeneratePassword(ctx)
		if err != nil {
			lc.Error("failed to generate default password for redisdb")
			return false
		}

		redisCredentials = UserPasswordPair{
			User:     "default",
			Password: defaultPassword,
		}
	} else {
		lc.Info("Redis DB credentials exist, skipping generating new password")
	}

	// Add any additional services that need the known DB secret
	lc.Infof("adding any additional services using redisdb for knownSecrets...")
	services, ok := knownSecretsToAdd[redisSecretName]
	if ok {
		for _, service := range services {
			err = addServiceCredential(lc, redisSecretName, secretStore, service, redisCredentials)
			if err != nil {
				lc.Error(err.Error())
				return false
			}
		}
	}

	lc.Infof("adding redisdb secret path for internal services...")
	for _, info := range configuration.Databases {
		service := info.Service

		// add credentials to service path if specified and they're not already there
		if len(service) != 0 {
			err = addServiceCredential(lc, redisSecretName, secretStore, service, redisCredentials)
			if err != nil {
				lc.Error(err.Error())
				return false
			}
		}
	}
	// security-bootstrapper-redis uses the path /v1/secret/edgex/security-bootstrapper-redis/ and go-mod-bootstrap
	// with append the DB type (redisdb)
	err = storeCredential(lc, "security-bootstrapper-redis", secretStore, redisSecretName, redisCredentials)
	if err != nil {
		lc.Error(err.Error())
		return false
	}

	// for secure message bus creds
	var msgBusCredentials UserPasswordPair
	if configuration.SecureMessageBus.Type != redisSecureMessageBusType &&
		configuration.SecureMessageBus.Type != noneSecureMessageBusType &&
		configuration.SecureMessageBus.Type != blankSecureMessageBusType {
		msgBusCredentials, err = getCredential(internal.BootstrapMessageBusServiceKey, secretStore, messagebusSecretName)
		if err != nil {
			if err != errNotFound {
				lc.Errorf("failed to determine if %s credentials already exist or not: %w", configuration.SecureMessageBus.Type, err)
				return false
			}

			lc.Infof("Generating new password for %s bus", configuration.SecureMessageBus.Type)
			msgBusPassword, err := secretStore.GeneratePassword(ctx)
			if err != nil {
				lc.Errorf("failed to generate password for %s bus", configuration.SecureMessageBus.Type)
				return false
			}

			msgBusCredentials = UserPasswordPair{
				User:     defaultMsgBusUser,
				Password: msgBusPassword,
			}
		} else {
			lc.Infof("%s bus credentials already exist, skipping generating new password", configuration.SecureMessageBus.Type)
		}

		lc.Infof("adding any additional services using %s for knownSecrets...", messagebusSecretName)
		services, ok := knownSecretsToAdd[messagebusSecretName]
		if ok {
			for _, service := range services {
				err = addServiceCredential(lc, messagebusSecretName, secretStore, service, msgBusCredentials)
				if err != nil {
					lc.Error(err.Error())
					return false
				}
			}
		}
		err = storeCredential(lc, internal.BootstrapMessageBusServiceKey, secretStore, messagebusSecretName, msgBusCredentials)
		if err != nil {
			lc.Error(err.Error())
			return false
		}
	}

	// determine the type of message bus
	messageBusType := configuration.SecureMessageBus.Type
	var creds UserPasswordPair
	supportedSecureType := true
	var secretName string
	switch messageBusType {
	case redisSecureMessageBusType:
		creds = redisCredentials
		secretName = redisSecretName
	case mqttSecureMessageBusType:
		creds = msgBusCredentials
		secretName = messagebusSecretName
	default:
		supportedSecureType = false
		lc.Warnf("secure message bus '%s' is not supported", messageBusType)
	}

	if supportedSecureType {
		lc.Infof("adding credentials for '%s' message bus for internal services...", messageBusType)
		for _, info := range configuration.SecureMessageBus.Services {
			service := info.Service

			// add credentials to service path if specified and they're not already there
			if len(service) != 0 {
				err = addServiceCredential(lc, secretName, secretStore, service, creds)
				if err != nil {
					lc.Error(err.Error())
					return false
				}
			}
		}
	}

	err = ConfigureSecureMessageBus(configuration.SecureMessageBus, creds, lc)
	if err != nil {
		lc.Errorf("failed to configure for Secure Message Bus: %s", err.Error())
		return false
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
			return false
		}

		if existing {
			lc.Info("proxy certificate pair are in the secret store already, skip uploading")
			return false
		}

		lc.Info("proxy certificate pair are not in the secret store yet, uploading them")
		cp, err := cert.ReadFrom(secretStoreConfig.CertFilePath, secretStoreConfig.KeyFilePath)
		if err != nil {
			lc.Error("failed to get certificate pair from volume")
			return false
		}

		lc.Info("proxy certificate pair are loaded from volume successfully, will upload to secret store")

		err = cert.UploadToStore(cp)
		if err != nil {
			lc.Error("failed to upload the proxy cert pair into the secret store")
			lc.Error(err.Error())
			return false
		}

		lc.Info("proxy certificate pair are uploaded to secret store successfully")

	} else {
		lc.Info("proxy certificate pair upload was skipped because cert secretStore value(s) were blank")
	}

	// create and save a Vault token to configure
	// Consul secret engine access, role operations, and managing Consul agent tokens.
	// Enable Consul secret engine
	if err := secretsengine.New(secretsengine.ConsulSecretEngineMountPoint, secretsengine.Consul).
		Enable(&rootToken, lc, client); err != nil {
		lc.Errorf("failed to enable Consul secrets engine: %s", err.Error())
		return false
	}

	// generate a management token for Consul secrets engine operations:
	tokenFileWriter := tokenfilewriter.NewWriter(lc, client, fileOpener)
	if _, err := tokenFileWriter.CreateAndWrite(rootToken, configuration.SecretStore.ConsulSecretsAdminTokenPath,
		tokenFileWriter.CreateMgmtTokenForConsulSecretsEngine); err != nil {
		lc.Errorf("failed to create and write the token for Consul secret management: %s", err.Error())
		return false
	}

	// Configure Kong Admin API
	//
	// For context - this process doesn't actually talk to Kong, it creates the configuration
	// file and JWT necessary in order for the Kong process to bootstrap itself with a properly
	// locked down Admin API and enable security-proxy-setup with the JWT in order to setup
	// the services/routes as configured.
	//
	// The reason why this code exists in the Secret Store setup is a matter of timing and
	// file permissions. This process has to occur before Kong is started, and cannot be executed
	// by the Kong entrypoint script because that executes as the Kong user. The JWT created
	// needs to be used by the security-proxy-setup process, so needs to be created before.
	// Since Secret Store setup runs prior to both of these, it made sense to logically drop them
	// here, especially if we're going to incorporate ties in to the Secret Store at a later
	// time.
	//
	// As of now, the private key that is generated for the "admin" group in Kong never
	// gets saved to disk out of memory. This could change in the future and be placed into
	// the Secret Store if we need to regenerate the JWT on the fly after setup has occurred.
	//
	lc.Info("Starting the Kong Admin API config file creation")

	// Get an instance of KongAdminAPI and map the paths from configuration.toml
	ka := NewKongAdminAPI(kongAdminConfig)

	// Setup Kong Admin API loopback configuration
	err = ka.Setup()
	if err != nil {
		lc.Errorf("failed to configure the Kong Admin API: %s", err.Error())
		return false
	}

	lc.Info("Vault init done successfully")
	return true

}

func (b *Bootstrap) getKnownSecretsToAdd() (map[string][]string, error) {
	// Process the env var for adding known secrets to the specified services' secret stores.
	// Format of the env var value is:
	//   "<secretName>[<serviceName>;<serviceName>; ...], <secretName>[<serviceName>;<serviceName>; ...], ..."
	knownSecretsToAdd := map[string][]string{}

	addKnownSecretsValue := strings.TrimSpace(os.Getenv(addKnownSecretsEnv))
	if len(addKnownSecretsValue) == 0 {
		return knownSecretsToAdd, nil
	}

	serviceNameRegx := regexp.MustCompile(ServiceNameValidationRegx)
	knownSecrets := strings.Split(addKnownSecretsValue, knownSecretSeparator)
	for _, secretSpec := range knownSecrets {
		// each secretSpec has format of "<secretName>[<serviceName>;<serviceName>; ...]"
		secretItems := strings.Split(secretSpec, serviceListBegin)
		if len(secretItems) != 2 {
			return nil, fmt.Errorf(
				"invalid specification for %s environment vaiable: Format of value '%s' is invalid. Missing or too many '%s'",
				addKnownSecretsEnv,
				secretSpec,
				serviceListBegin)
		}

		secretName := strings.TrimSpace(secretItems[0])

		_, valid := b.validKnownSecrets[secretName]
		if !valid {
			return nil, fmt.Errorf(
				"invalid specification for %s environment vaiable: '%s' is not a known secret",
				addKnownSecretsEnv,
				secretName)
		}

		serviceNameList := secretItems[1]
		if !strings.Contains(serviceNameList, serviceListEnd) {
			return nil, fmt.Errorf(
				"invalid specification for %s environment vaiable: Service list for '%s' missing closing '%s'",
				addKnownSecretsEnv,
				secretName,
				serviceListEnd)
		}

		serviceNameList = strings.TrimSpace(strings.Replace(serviceNameList, serviceListEnd, "", 1))
		if len(serviceNameList) == 0 {
			return nil, fmt.Errorf(
				"invalid specification for %s environment vaiable: Service name list for '%s' is empty.",
				addKnownSecretsEnv,
				secretName)
		}

		serviceNames := strings.Split(serviceNameList, serviceListSeparator)
		for index := range serviceNames {
			serviceNames[index] = strings.TrimSpace(serviceNames[index])

			if !serviceNameRegx.MatchString(serviceNames[index]) {
				return nil, fmt.Errorf(
					"invalid specification for %s environment vaiable: Service name '%s' has invalid characters.",
					addKnownSecretsEnv, serviceNames[index])
			}
		}

		// This supports listing known secret multiple times.
		// Same service name listed twice is not an issue since the add logic checks if the secret is already present.
		existingServices := knownSecretsToAdd[secretName]
		knownSecretsToAdd[secretName] = append(existingServices, serviceNames...)
	}

	return knownSecretsToAdd, nil
}

// XXX Collapse addServiceCredential and addDBCredential together by passing in the path or using
// variadic functions

func addServiceCredential(lc logger.LoggingClient, secretKeyName string, secretStore Cred, service string, pair UserPasswordPair) error {
	path := fmt.Sprintf("%s/%s/%s", secretBasePath, service, secretKeyName)
	existing, err := secretStore.AlreadyInStore(path)
	if err != nil {
		return err
	}
	if !existing {
		err = secretStore.UploadToStore(&pair, path)
		if err != nil {
			lc.Errorf("failed to upload credential pair for %s on path %s", service, path)
			return err
		}
	} else {
		lc.Infof("credentials for %s already present at path %s", service, path)
	}

	return err
}

func getCredential(credBootstrapStem string, cred Cred, service string) (UserPasswordPair, error) {
	path := fmt.Sprintf("%s/%s/%s", secretBasePath, credBootstrapStem, service)

	pair, err := cred.getUserPasswordPair(path)
	if err != nil {
		return UserPasswordPair{}, err
	}

	return *pair, err

}

func storeCredential(lc logger.LoggingClient, credBootstrapStem string, cred Cred, secretKeyName string, pair UserPasswordPair) error {
	path := fmt.Sprintf("%s/%s/%s", secretBasePath, credBootstrapStem, secretKeyName)
	existing, err := cred.AlreadyInStore(path)
	if err != nil {
		lc.Error(err.Error())
		return err
	}
	if !existing {
		err = cred.UploadToStore(&pair, path)
		if err != nil {
			lc.Errorf("failed to upload credential pair for %s on path %s", secretKeyName, path)
			return err
		}
	} else {
		lc.Infof("credentials for %s already present at path %s", secretKeyName, path)
	}

	return err
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
