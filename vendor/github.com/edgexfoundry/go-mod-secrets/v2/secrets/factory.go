/*******************************************************************************
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

package secrets

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/go-mod-secrets/v2/internal/pkg/vault"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const Vault = "vault"

// NewSecretsClient creates a new instance of a SecretClient based on the passed in configuration.
// The SecretClient allows access to secret(s) for the configured token.
func NewSecretsClient(ctx context.Context, config types.SecretConfig, lc logger.LoggingClient, callback pkg.TokenExpiredCallback) (SecretClient, error) {
	if ctx == nil {
		return nil, pkg.NewErrSecretStore("background ctx is required and cannot be nil")
	}

	// Currently only have a Vault implementation, so no need to have/check type.

	switch config.Type {
	// Currently only have a Vault implementation, so type isn't actual set in configuration
	case Vault:
		return vault.NewSecretsClient(ctx, config, lc, callback)
	default:
		return nil, fmt.Errorf("invalid secrets client type of '%s'", config.Type)
	}
}

// NewSecretStoreClient creates a new instance of a SecretClient based on the passed in configuration.
// The SecretStoreClient provides management functionality to manage the secret store.
func NewSecretStoreClient(config types.SecretConfig, lc logger.LoggingClient, requester pkg.Caller) (SecretStoreClient, error) {
	switch config.Type {
	case Vault:
		return vault.NewClient(config, requester, false, lc)

	default:
		return nil, fmt.Errorf("invalid secret store client type of '%s'", config.Type)
	}
}
