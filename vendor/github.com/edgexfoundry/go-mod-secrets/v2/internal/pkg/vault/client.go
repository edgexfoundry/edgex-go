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
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

// Client defines the behavior for interacting with the Vault REST secret key/value store via HTTP(S).
type Client struct {
	Config     types.SecretConfig
	HttpCaller pkg.Caller
	lc         logger.LoggingClient
	context    context.Context
	// vaultTokenCancelFunc is an internal map with token as key and the context.cancel function as value
	vaultTokenCancelFunc vaultTokenToCancelFuncMap
	mapMutex             sync.Mutex
	tokenExpiredCallback pkg.TokenExpiredCallback
}

// NewClient constructs a Vault *Client which communicates with Vault via HTTP(S)
// lc is any logging client that implements the loggingClient interface;
// today EdgeX's logger.LoggingClient from go-mod-core-contracts satisfies this implementation
func NewClient(config types.SecretConfig, requester pkg.Caller, forSecrets bool, lc logger.LoggingClient) (*Client, error) {

	if forSecrets && config.Authentication.AuthToken == "" {
		return nil, pkg.NewErrSecretStore("AuthToken is required in config")
	}

	var err error
	if requester == nil {
		requester, err = createHTTPClient(config)
		if err != nil {
			return nil, err
		}
	}

	vaultClient := Client{
		Config:               config,
		HttpCaller:           requester,
		lc:                   lc,
		mapMutex:             sync.Mutex{},
		vaultTokenCancelFunc: make(vaultTokenToCancelFuncMap),
	}

	return &vaultClient, err
}

func createHTTPClient(config types.SecretConfig) (pkg.Caller, error) {

	if config.RootCaCertPath == "" {
		return http.DefaultClient, nil
	}

	// Read and load the CA Root certificate so the client will be able to use TLS without skipping the verification of
	// the cert received by the server.
	caCert, err := ioutil.ReadFile(config.RootCaCertPath)
	if err != nil {
		return nil, ErrCaRootCert{
			path:        config.RootCaCertPath,
			description: err.Error(),
		}
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    caCertPool,
				ServerName: config.ServerName,
			},
		},
	}, nil
}
