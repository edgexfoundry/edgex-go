/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Corp.
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

package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

// a map variable to handle the case of the same caller to have
// multiple secret clients with potentially the same tokens while renewing token
// in the background go-routine
type vaultTokenToCancelFuncMap map[string]context.CancelFunc

// NewSecretsClient constructs a Vault *Client which communicates with Vault via HTTP(S) for basic usage of secrets
func NewSecretsClient(ctx context.Context, config types.SecretConfig, lc logger.LoggingClient, callback pkg.TokenExpiredCallback) (*Client, error) {
	vaultClient, err := NewClient(config, nil, true, lc)
	if err != nil {
		return nil, err
	}

	vaultClient.tokenExpiredCallback = callback
	err = vaultClient.SetAuthToken(ctx, config.Authentication.AuthToken)

	return vaultClient, err
}

// GetSecrets retrieves the secrets at the provided sub-path that matches the specified keys.
func (c *Client) GetSecrets(subPath string, keys ...string) (map[string]string, error) {

	// no need to retry now as the secret store should be ready as the security bootstrapper starts in sequence now
	data, err := c.getAllKeys(subPath)
	if err != nil {
		return nil, err
	}

	// Do not filter any of the secrets
	if len(keys) <= 0 {
		return data, nil
	}

	values := make(map[string]string)
	var notFound []string

	for _, key := range keys {
		value, success := data[key]
		if !success {
			notFound = append(notFound, key)
			continue
		}

		values[key] = value
	}

	if len(notFound) > 0 {
		return nil, pkg.NewErrSecretsNotFound(notFound)
	}

	return values, nil
}

// StoreSecrets stores the secrets at the provided sub-path for the specified keys.
func (c *Client) StoreSecrets(subPath string, secrets map[string]string) error {
	// this interface acting as facade, just calling the internal store func on the client
	return c.store(subPath, secrets)
}

// GenerateConsulToken generates a new Consul token using serviceKey as role name to
// call secret store's consul/creds API
// the serviceKey is used in the part of secret store's URL as role name and should be accessible to the API
func (c *Client) GenerateConsulToken(serviceKey string) (string, error) {
	trimmedSrvKey := strings.TrimSpace(serviceKey)
	if len(trimmedSrvKey) == 0 {
		return emptyToken, pkg.NewErrSecretStore("serviceKey cannot be empty for generating Consul token")
	}

	if len(c.Config.Authentication.AuthToken) == 0 {
		return emptyToken, pkg.NewErrSecretStore("secrete store token from config cannot be empty for generating Consul token")
	}

	credsURL, err := c.Config.BuildURL(fmt.Sprintf(GenerateConsulTokenAPI, trimmedSrvKey))
	if err != nil {
		return emptyToken, err
	}

	req, err := http.NewRequest(http.MethodGet, credsURL, http.NoBody)
	if err != nil {
		return emptyToken, err
	}

	req.Header.Set(AuthTypeHeader, c.Config.Authentication.AuthToken)

	resp, err := c.HttpCaller.Do(req)
	if err != nil {
		return emptyToken, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	tokenResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return emptyToken, err
	}

	if resp.StatusCode != http.StatusOK {
		return emptyToken, ErrHTTPResponse{
			StatusCode: resp.StatusCode,
			ErrMsg:     fmt.Sprintf("failed to generate Consul token using [%s]: %s", trimmedSrvKey, string(tokenResp)),
		}
	}

	type TokenResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var consulTokenResp TokenResp
	if err := json.NewDecoder(bytes.NewReader(tokenResp)).Decode(&consulTokenResp); err != nil {
		return emptyToken, err
	}

	c.lc.Infof("successfully generated Consul token for service %s", serviceKey)

	return consulTokenResp.Data.Token, nil
}

func (c *Client) SetAuthToken(ctx context.Context, newToken string) error {
	// mapMutex protects the internal map cache from race conditions
	c.mapMutex.Lock()

	// if there is a context already associated with the current token then need to cancel it
	if cancel, exists := c.vaultTokenCancelFunc[c.Config.Authentication.AuthToken]; exists {
		cancel()
	}

	// if there is context already associated with the new token, then we cancel it first
	if cancel, exists := c.vaultTokenCancelFunc[newToken]; exists {
		cancel()
	}

	c.mapMutex.Unlock()

	c.Config.Authentication.AuthToken = newToken

	cCtx, cancel := context.WithCancel(ctx)
	var err error

	if err = c.refreshToken(cCtx, c.tokenExpiredCallback); err != nil {
		cancel()
		c.mapMutex.Lock()
		delete(c.vaultTokenCancelFunc, c.Config.Authentication.AuthToken)
		c.mapMutex.Unlock()
	} else {
		c.mapMutex.Lock()
		c.vaultTokenCancelFunc[c.Config.Authentication.AuthToken] = cancel
		c.mapMutex.Unlock()
	}

	return err
}

func (c *Client) getTokenDetails() (*types.TokenMetadata, error) {
	// call Vault's token self lookup API
	url, err := c.Config.BuildURL(lookupSelfVaultAPI)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(AuthTypeHeader, c.Config.Authentication.AuthToken)

	resp, err := c.HttpCaller.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrHTTPResponse{
			StatusCode: resp.StatusCode,
			ErrMsg:     "failed to lookup token",
		}
	}

	// the returned JSON structure for token self-read is TokenLookupResponse
	result := TokenLookupResponse{}
	jsonDec := json.NewDecoder(resp.Body)
	if jsonDec == nil {
		return nil, pkg.NewErrSecretStore("failed to obtain json decoder")
	}

	jsonDec.UseNumber()
	if err = jsonDec.Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) refreshToken(ctx context.Context, tokenExpiredCallback pkg.TokenExpiredCallback) error {
	tokenData, err := c.getTokenDetails()

	if err != nil {
		return err
	}

	if !tokenData.Renewable {
		// token is not renewable, log warning and return
		c.lc.Warn("token is not renewable from the secret store")
		return nil
	}

	// the renewal interval is half of period value
	tokenPeriod := time.Duration(tokenData.Period) * time.Second
	renewInterval := tokenPeriod / 2
	if renewInterval <= 0 {
		// cannot renew, as the renewal interval is non-positive
		c.lc.Warn("no token renewal since renewInterval is 0")
		return nil
	}

	ttl := time.Duration(tokenData.Ttl) * time.Second

	// if the current time-to-live is already less than the half of period
	// need to renew the token right away
	if ttl <= renewInterval {
		// call renew self api
		c.lc.Info("ttl already <= half of the renewal period")
		if err := c.renewToken(); err != nil {
			return err
		}
	}

	c.context = ctx

	// goroutine to periodically renew the service token based on renewInterval
	go c.doTokenRefreshPeriodically(renewInterval, tokenExpiredCallback)

	return nil
}

func (c *Client) doTokenRefreshPeriodically(renewInterval time.Duration,
	tokenExpiredCallback pkg.TokenExpiredCallback) {
	c.lc.Infof("kick off token renewal with interval: %v", renewInterval)

	ticker := time.NewTicker(renewInterval)
	for {
		select {

		case <-c.context.Done():
			ticker.Stop()
			c.lc.Info("context cancelled, dismiss the token renewal process")
			return

		case <-ticker.C:
			// renew token to keep it refreshed
			// if err happens then handle it according to the callback func tokenExpiredCallback
			if err := c.renewToken(); err != nil {
				if isForbidden(err) {
					// the current token is expired,
					// cannot renew, handle it based upon
					// the implementation of callback from the caller if any
					if tokenExpiredCallback == nil {
						ticker.Stop()
						return
					}
					replacementToken, retry := tokenExpiredCallback(c.Config.Authentication.AuthToken)
					if !retry {
						ticker.Stop()
						return
					}
					c.Config.Authentication.AuthToken = replacementToken
					c.lc.Info("auth token is replaced")
				} else {
					// other type of errors, cannot continue, quitting the renewal routine
					c.lc.Errorf("dismiss the renewal process as the current token cannot be renewed: %v", err)
					ticker.Stop()
					return
				}
			}
		}
	}
}

func (c *Client) renewToken() error {
	// call Vault's renew self API
	url, err := c.Config.BuildURL(renewSelfVaultAPI)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set(AuthTypeHeader, c.Config.Authentication.AuthToken)

	resp, err := c.HttpCaller.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return ErrHTTPResponse{
			StatusCode: resp.StatusCode,
			ErrMsg:     "failed to renew token",
		}
	}

	c.lc.Debug("token is successfully renewed")
	return nil
}

// getAllKeys obtains all the keys that reside at the provided sub-path.
func (c *Client) getAllKeys(subPath string) (map[string]string, error) {
	url, err := c.Config.BuildSecretsPathURL(subPath)
	if err != nil {
		return nil, err
	}

	c.lc.Debug(fmt.Sprintf("Using Secrets URL of `%s`", url))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(c.Config.Authentication.AuthType, c.Config.Authentication.AuthToken)

	if c.Config.Namespace != "" {
		req.Header.Set(NamespaceHeader, c.Config.Namespace)
	}

	resp, err := c.HttpCaller.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, pkg.NewErrSecretStore(fmt.Sprintf("Received a '%d' response from the secret store", resp.StatusCode))
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	data, success := result["data"].(map[string]interface{})
	if !success || len(data) <= 0 {
		return nil, pkg.NewErrSecretStore(fmt.Sprintf("No secretKeyValues are present at the subpath: '%s'", subPath))
	}

	// Cast the secret values to strings
	secretKeyValues := make(map[string]string)
	for k, v := range data {
		secretKeyValues[k] = v.(string)
	}

	return secretKeyValues, nil
}

func isForbidden(err error) bool {
	if httpRespErr, ok := err.(ErrHTTPResponse); ok {
		return httpRespErr.StatusCode == http.StatusForbidden
	}
	return false
}

func (c *Client) store(subPath string, secrets map[string]string) error {
	if len(secrets) == 0 {
		// nothing to store
		return nil
	}

	url, err := c.Config.BuildSecretsPathURL(subPath)
	if err != nil {
		return err
	}

	c.lc.Debug(fmt.Sprintf("Using Secrets URL of `%s`", url))

	payload, err := json.Marshal(secrets)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set(c.Config.Authentication.AuthType, c.Config.Authentication.AuthToken)

	if c.Config.Namespace != "" {
		req.Header.Set(NamespaceHeader, c.Config.Namespace)
	}

	resp, err := c.HttpCaller.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return pkg.NewErrSecretStore(fmt.Sprintf("Received a '%d' response from the secret store", resp.StatusCode))
	}

	return nil
}
