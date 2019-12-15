//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//

package fileprovider

import (
	"errors"

	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"

	"github.com/edgexfoundry/go-mod-bootstrap/security/authtokenloader"
	"github.com/edgexfoundry/go-mod-bootstrap/security/fileioperformer"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// fileTokenProvider stores instance data
type fileTokenProvider struct {
	logger        logger.LoggingClient
	fileOpener    fileioperformer.FileIoPerformer
	tokenProvider authtokenloader.AuthTokenLoader
	vaultClient   secretstoreclient.SecretStoreClient
	secretConfig  secretstoreclient.SecretServiceInfo
	tokenConfig   config.TokenFileProviderInfo
}

// NewTokenProvider creates a new TokenProvider
func NewTokenProvider(logger logger.LoggingClient,
	fileOpener fileioperformer.FileIoPerformer,
	tokenProvider authtokenloader.AuthTokenLoader,
	vaultClient secretstoreclient.SecretStoreClient) TokenProvider {
	return &fileTokenProvider{
		logger:        logger,
		fileOpener:    fileOpener,
		tokenProvider: tokenProvider,
		vaultClient:   vaultClient,
	}
}

// Set configuration
func (p *fileTokenProvider) SetConfiguration(secretConfig secretstoreclient.SecretServiceInfo, tokenConfig config.TokenFileProviderInfo) {
	p.secretConfig = secretConfig
	p.tokenConfig = tokenConfig
}

// Do whatever is needed
func (p *fileTokenProvider) Run() error {
	p.logger.Error("Implementation to be provided later")
	return errors.New("Not implemented")
}
