//
// Copyright (C) 2022-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package secret

import (
	"fmt"
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	clientInterfaces "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
)

type jwtSecretProvider struct {
	secretProvider interfaces.SecretProviderExt
}

func NewJWTSecretProvider(secretProvider interfaces.SecretProviderExt) clientInterfaces.AuthenticationInjector {
	return &jwtSecretProvider{
		secretProvider: secretProvider,
	}
}

func (self *jwtSecretProvider) AddAuthenticationData(req *http.Request) error {
	if self.secretProvider == nil {
		// Test cases or real code may invoke NewJWTSecretProvider(nil),
		// though this is discouraged. In that case, just do nothing.
		return nil
	}

	// Otherwise if there is a secret provider, get the JWT
	jwt, err := self.secretProvider.GetSelfJWT()
	if err != nil {
		return err
	}

	// Only add authorization header if we get non-empty token back
	if len(jwt) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))
	}

	return nil
}
