/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright 2020-2023 Intel Corporation
 * Copyright 2024 IOTech Ltd
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
 *******************************************************************************/

package secret

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg"
	gometrics "github.com/rcrowley/go-metrics"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/hashicorp/go-multierror"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/runtimetokenprovider"
	"github.com/edgexfoundry/go-mod-secrets/v4/secrets"
)

const (
	TokenTypeKeeper      = "keeper"
	AccessTokenAuthError = "HTTP response with status code 403"
	//nolint: gosec
	SecretsAuthError                     = "Received a '403' response"
	EnvEdgeXUseCommonAppServiceSecretKey = "EDGEX_USE_COMMON_APP_SERVICE_SECRET_KEY" // nolint: gosec
)

// SecureProvider implements the SecretProvider interface
type SecureProvider struct {
	secretClient secrets.SecretClient
	lc           logger.LoggingClient
	loader       authtokenloader.AuthTokenLoader
	// runtimeTokenProvider is for delayed start services
	runtimeTokenProvider               runtimetokenprovider.RuntimeTokenProvider
	serviceKey                         string
	secretStoreInfo                    config.SecretStoreInfo
	secretsCache                       map[string]map[string]string // secret's secretName, key, value
	cacheMutex                         *sync.RWMutex
	lastUpdated                        time.Time
	ctx                                context.Context
	registeredSecretCallbacks          map[string]func(secretName string)
	securitySecretsRequested           gometrics.Counter
	securitySecretsStored              gometrics.Counter
	securityRuntimeSecretTokenDuration gometrics.Timer
	securityGetSecretDuration          gometrics.Timer
	httpRoundTripper                   http.RoundTripper
	fallbackDialer                     *net.Dialer
	zeroTrustEnabled                   bool
}

// NewSecureProvider creates & initializes Provider instance for secure secrets.
func NewSecureProvider(ctx context.Context, secretStoreInfo *config.SecretStoreInfo, lc logger.LoggingClient,
	loader authtokenloader.AuthTokenLoader, runtimeTokenLoader runtimetokenprovider.RuntimeTokenProvider,
	serviceKey string) *SecureProvider {
	provider := &SecureProvider{
		lc:                                 lc,
		loader:                             loader,
		runtimeTokenProvider:               runtimeTokenLoader,
		serviceKey:                         serviceKey,
		secretStoreInfo:                    *secretStoreInfo,
		secretsCache:                       make(map[string]map[string]string),
		cacheMutex:                         &sync.RWMutex{},
		lastUpdated:                        time.Now(),
		ctx:                                ctx,
		registeredSecretCallbacks:          make(map[string]func(secretName string)),
		securitySecretsRequested:           gometrics.NewCounter(),
		securitySecretsStored:              gometrics.NewCounter(),
		securityRuntimeSecretTokenDuration: gometrics.NewTimer(),
		securityGetSecretDuration:          gometrics.NewTimer(),
	}
	return provider
}

// SetClient sets the secret client that is used to access the secure secrets
func (p *SecureProvider) SetClient(client secrets.SecretClient) {
	p.secretClient = client
}

// GetSecret retrieves secrets from a secret store.
// secretName specifies the type or location of the secrets to retrieve.
// keys specifies the secrets which to retrieve. If no keys are provided then all the keys associated with the
// specified secretName will be returned.
func (p *SecureProvider) GetSecret(secretName string, keys ...string) (map[string]string, error) {
	p.securitySecretsRequested.Inc(1)
	started := time.Now()
	defer p.securityGetSecretDuration.UpdateSince(started)

	if cachedSecrets := p.getSecretsCache(secretName, keys...); cachedSecrets != nil {
		return cachedSecrets, nil
	}

	if p.secretClient == nil {
		return nil, errors.New("can't get secrets. Secure secret provider is not properly initialized")
	}

	secureSecrets, err := p.secretClient.GetSecret(secretName, keys...)

	retry, err := p.reloadTokenOnAuthError(err)
	if retry {
		// Retry with potential new token
		secureSecrets, err = p.secretClient.GetSecret(secretName, keys...)
	}

	if err != nil {
		return nil, err
	}

	p.updateSecretsCache(secretName, secureSecrets)
	return secureSecrets, nil
}

func (p *SecureProvider) getSecretsCache(secretName string, keys ...string) map[string]string {
	secureSecrets := make(map[string]string)

	// Synchronize cache access
	p.cacheMutex.RLock()
	defer p.cacheMutex.RUnlock()

	// check cache for keys
	allKeysExistInCache := false
	cachedSecrets, cacheExists := p.secretsCache[secretName]
	value := ""

	if cacheExists {
		for _, key := range keys {
			value, allKeysExistInCache = cachedSecrets[key]
			if !allKeysExistInCache {
				return nil
			}
			secureSecrets[key] = value
		}

		// return secureSecrets if the requested keys exist in cache
		if allKeysExistInCache {
			return secureSecrets
		}
	}

	return nil
}

func (p *SecureProvider) updateSecretsCache(secretName string, secrets map[string]string) {
	// Synchronize cache access
	p.cacheMutex.Lock()
	defer p.cacheMutex.Unlock()

	if _, cacheExists := p.secretsCache[secretName]; !cacheExists {
		p.secretsCache[secretName] = secrets
	}

	for key, value := range secrets {
		p.secretsCache[secretName][key] = value
	}
}

// StoreSecret stores the secrets to a secret store.
// it sets the values requested at provided keys
// secretName specifies the type or location of the secrets to store
// secrets map specifies the "key": "value" pairs of secrets to store
func (p *SecureProvider) StoreSecret(secretName string, secrets map[string]string) error {
	p.securitySecretsStored.Inc(1)

	if p.secretClient == nil {
		return errors.New("can't store secrets. Secure secret provider is not properly initialized")
	}

	err := p.secretClient.StoreSecret(secretName, secrets)

	retry, err := p.reloadTokenOnAuthError(err)
	if retry {
		// Retry with potential new token
		err = p.secretClient.StoreSecret(secretName, secrets)
	}

	if err != nil {
		return err
	}

	// Execute Callbacks on registered secret secretNames.
	p.SecretUpdatedAtSecretName(secretName)

	// Synchronize cache access before clearing
	p.cacheMutex.Lock()
	// Clearing cache because adding a new secret(p) possibly invalidates the previous cache
	p.secretsCache = make(map[string]map[string]string)
	p.cacheMutex.Unlock()
	//indicate to the SDK that the cache has been invalidated
	p.lastUpdated = time.Now()
	return nil
}

func (p *SecureProvider) reloadTokenOnAuthError(err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	if !strings.Contains(err.Error(), SecretsAuthError) &&
		!strings.Contains(err.Error(), AccessTokenAuthError) {
		return false, err
	}

	// Reload token in case new token was created causing the auth error
	token, err := p.loader.Load(p.secretStoreInfo.TokenFile)
	if err != nil {
		return false, err
	}

	err = p.secretClient.SetAuthToken(p.ctx, token)
	if err != nil {
		return false, err
	}

	return true, nil
}

// SecretsUpdated is not need for secure secrets as this is handled when secrets are stored.
func (p *SecureProvider) SecretsUpdated() {
	// Do nothing
}

// SecretsLastUpdated returns the last time secure secrets were updated
func (p *SecureProvider) SecretsLastUpdated() time.Time {
	return p.lastUpdated
}

// DefaultTokenExpiredCallback is the default implementation of tokenExpiredCallback function
// It utilizes the tokenFile to re-read the token and enable retry if any update from the expired token
func (p *SecureProvider) DefaultTokenExpiredCallback(expiredToken string) (replacementToken string, retry bool) {
	tokenFile := p.secretStoreInfo.TokenFile

	// during the callback, we want to re-read the token from the disk
	// specified by tokenFile and set the retry to true if a new token
	// is different from the expiredToken
	reReadToken, err := p.loader.Load(tokenFile)
	if err != nil {
		p.lc.Error(fmt.Sprintf("fail to load auth token from tokenFile %s: %v", tokenFile, err))
		return "", false
	}

	if reReadToken == expiredToken {
		p.lc.Error("No new replacement token found for the expired token")
		return reReadToken, false
	}

	return reReadToken, true
}

func (p *SecureProvider) RuntimeTokenExpiredCallback(expiredToken string) (replacementToken string, retry bool) {
	newToken, err := p.runtimeTokenProvider.GetRawToken(p.serviceKey)
	if err != nil {
		p.lc.Errorf("failed to get a new token for service: %s: %v", p.serviceKey, err)
		return "", false
	}

	return newToken, true
}

// LoadServiceSecrets loads the service secrets from the specified file and stores them in the service's SecretStore
func (p *SecureProvider) LoadServiceSecrets(secretStoreConfig *config.SecretStoreInfo) error {

	contents, err := os.ReadFile(secretStoreConfig.SecretsFile)
	if err != nil {
		return fmt.Errorf("seeding secrets failed: %s", err.Error())
	}

	data, seedingErrs := p.seedSecrets(contents)

	if secretStoreConfig.DisableScrubSecretsFile {
		p.lc.Infof("Scrubbing of secrets file disable.")
		return seedingErrs
	}

	if err := os.WriteFile(secretStoreConfig.SecretsFile, data, 0); err != nil {
		return fmt.Errorf("seeding secrets failed: unable to overwrite file with secret data removed: %s", err.Error())
	}

	p.lc.Infof("Scrubbing of secrets file complete.")

	return seedingErrs
}

func (p *SecureProvider) seedSecrets(contents []byte) ([]byte, error) {
	serviceSecrets, err := UnmarshalServiceSecretsJson(contents)
	if err != nil {
		return nil, fmt.Errorf("seeding secrets failed unmarshaling JSON: %s", err.Error())
	}

	p.lc.Infof("Seeding %d Service Secrets", len(serviceSecrets.Secrets))

	var seedingErrs error
	for index, secret := range serviceSecrets.Secrets {
		if secret.Imported {
			p.lc.Infof("Secret for '%s' already imported. Skipping...", secret.SecretName)
			continue
		}

		// At this pint the JSON validation and above check cover all the required validation, so go to store secret.
		secretName, data := prepareSecret(secret)
		err := p.StoreSecret(secretName, data)
		if err != nil {
			message := fmt.Sprintf("failed to store secret for '%s': %s", secret.SecretName, err.Error())
			p.lc.Errorf(message)
			seedingErrs = multierror.Append(seedingErrs, errors.New(message))
			continue
		}

		p.lc.Infof("Secret for '%s' successfully stored.", secret.SecretName)

		serviceSecrets.Secrets[index].Imported = true
		serviceSecrets.Secrets[index].SecretData = make([]common.SecretDataKeyValue, 0)
	}

	// Now need to write the file back over with the imported secrets' secretData removed.
	data, err := serviceSecrets.MarshalJson()
	if err != nil {
		return nil, fmt.Errorf("seeding secrets failed marshaling back to JSON to clear secrets: %s", err.Error())
	}

	return data, seedingErrs
}

func prepareSecret(secret ServiceSecret) (string, map[string]string) {
	var secretsKV = make(map[string]string)
	for _, secret := range secret.SecretData {
		secretsKV[secret.Key] = secret.Value
	}

	secretName := strings.TrimSpace(secret.SecretName)

	return secretName, secretsKV
}

// HasSecret returns true if the service's SecretStore contains a secret at the specified secretName.
func (p *SecureProvider) HasSecret(secretName string) (bool, error) {
	_, err := p.GetSecret(secretName)

	if err != nil {
		_, ok := err.(pkg.ErrSecretNameNotFound)
		if ok {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// ListSecretNames returns a list of secretNames for the current service from an insecure/secure secret store.
func (p *SecureProvider) ListSecretNames() ([]string, error) {

	if p.secretClient == nil {
		return nil, errors.New("can't get secret secretNames. Secure secret provider is not properly initialized")
	}

	secureSecrets, err := p.secretClient.GetSecretNames()

	retry, err := p.reloadTokenOnAuthError(err)
	if retry {
		// Retry with potential new token
		secureSecrets, err = p.secretClient.GetSecretNames()
	}

	if err != nil {
		return nil, fmt.Errorf("unable to get secret secretNames: %v", err)
	}

	return secureSecrets, nil
}

// RegisterSecretUpdatedCallback registers a callback for a secret. If you specify secret.WildcardName
// as the secretName, then the callback will be called for any updated secret. Callbacks set for a specific
// secretName are given a higher precedence over wildcard ones, and will be called instead of the wildcard one
// if both are present.
func (p *SecureProvider) RegisterSecretUpdatedCallback(secretName string, callback func(secretName string)) error {
	if _, ok := p.registeredSecretCallbacks[secretName]; ok {
		return fmt.Errorf("there is a callback already registered for secretName '%v'", secretName)
	}

	// Register new call back for secretName.
	p.registeredSecretCallbacks[secretName] = callback

	return nil
}

// SecretUpdatedAtSecretName performs updates and callbacks for an updated secret or secretName.
func (p *SecureProvider) SecretUpdatedAtSecretName(secretName string) {
	p.lastUpdated = time.Now()
	if p.registeredSecretCallbacks == nil {
		return
	}

	// Execute Callback for provided secretName.
	if callback, ok := p.registeredSecretCallbacks[secretName]; ok {
		p.lc.Debugf("invoking callback registered for secretName: '%s'", secretName)
		callback(secretName)

		// if no callback is registered for secretName, see if wildcard callback is provided.
	} else if callback, ok = p.registeredSecretCallbacks[WildcardName]; ok {
		p.lc.Debugf("invoking wildcard callback for secretName: '%s'", secretName)
		callback(secretName)
	}
}

// DeregisterSecretUpdatedCallback removes a secret's registered callback secretName.
func (p *SecureProvider) DeregisterSecretUpdatedCallback(secretName string) {
	// Remove secretName from map.
	delete(p.registeredSecretCallbacks, secretName)
}

// GetMetricsToRegister returns all metric objects that needs to be registered.
func (p *SecureProvider) GetMetricsToRegister() map[string]interface{} {
	return map[string]interface{}{
		secretsRequestedMetricName:             p.securitySecretsRequested,
		secretsStoredMetricName:                p.securitySecretsStored,
		securityRuntimeSecretTokenDurationName: p.securityRuntimeSecretTokenDuration,
		securityGetSecretDurationName:          p.securityGetSecretDuration,
	}
}

// GetSelfJWT returns an encoded JWT for the current identity-based secret store token
func (p *SecureProvider) GetSelfJWT() (string, error) {
	serviceKey := p.overwriteAppServiceKey(p.serviceKey)
	return p.secretClient.GetSelfJWT(serviceKey)
}

// IsJWTValid evaluates a given JWT and returns a true/false if the JWT is valid (i.e. belongs to us and current) or not
func (p *SecureProvider) IsJWTValid(jwt string) (bool, error) {
	return p.secretClient.IsJWTValid(jwt)
}

func (p *SecureProvider) HttpTransport() http.RoundTripper {
	return p.httpRoundTripper
}

func (p *SecureProvider) SetHttpTransport(rt http.RoundTripper) {
	if p.httpRoundTripper == nil {
		p.httpRoundTripper = rt
	} else {
		p.lc.Warnf("refusing to override httpRoundTripper, already set")
	}
}

func (p *SecureProvider) FallbackDialer() *net.Dialer {
	return p.fallbackDialer
}

func (p *SecureProvider) SetFallbackDialer(dialer *net.Dialer) {
	if p.fallbackDialer == nil {
		p.fallbackDialer = dialer
	} else {
		p.lc.Warnf("refusing to override fallbackDialer, already set")
	}
}

func (p *SecureProvider) IsZeroTrustEnabled() bool {
	return p.zeroTrustEnabled
}

func (p *SecureProvider) EnableZeroTrust() {
	p.zeroTrustEnabled = true
}

func (p *SecureProvider) overwriteAppServiceKey(serviceKey string) string {
	if strings.HasPrefix(serviceKey, "app-") &&
		strings.EqualFold(os.Getenv(EnvEdgeXUseCommonAppServiceSecretKey), "true") {
		p.lc.Infof("Overwrote ASC serviceKey from %s to %s", serviceKey, config.ServiceTypeApp)
		return config.ServiceTypeApp
	}
	return serviceKey
}
