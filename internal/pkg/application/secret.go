//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"strings"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
)

// AddSecret adds EdgeX Service exclusive secret to the Secret Store
func AddSecret(dic *di.Container, request common.SecretRequest) errors.EdgeX {
	secretName, secret := prepareSecret(request)

	secretProvider := container.SecretProviderFrom(dic.Get)
	if secretProvider == nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "secret provider is missing. Make sure it is specified to be used in bootstrap.Run()", nil)
	}

	if err := secretProvider.StoreSecret(secretName, secret); err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "adding secret failed", err)
	}
	return nil
}

func prepareSecret(request common.SecretRequest) (string, map[string]string) {
	var secretsKV = make(map[string]string)
	for _, secret := range request.SecretData {
		secretsKV[secret.Key] = secret.Value
	}

	secretName := strings.TrimSpace(request.SecretName)

	return secretName, secretsKV
}
