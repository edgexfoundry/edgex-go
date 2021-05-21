//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"strings"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
)

// AddSecret adds EdgeX Service exclusive secret to the Secret Store
func AddSecret(dic *di.Container, request common.SecretRequest) errors.EdgeX {
	path, secret := prepareSecret(dic, request)

	secretProvider := container.SecretProviderFrom(dic.Get)
	if secretProvider == nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "secret provider is missing. Make sure it is specified to be used in bootstrap.Run()", nil)
	}

	if err := secretProvider.StoreSecret(path, secret); err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "adding secret failed", err)
	}
	return nil
}

func prepareSecret(dic *di.Container, request common.SecretRequest) (string, map[string]string) {
	secretStoreInfo := container.ConfigurationFrom(dic.Get).GetBootstrap().SecretStore

	var secretsKV = make(map[string]string)
	for _, secret := range request.SecretData {
		secretsKV[secret.Key] = secret.Value
	}

	path := strings.TrimSpace(request.Path)

	// add '/' in the full URL path if it's not already at the end of the base path or sub path
	if !strings.HasSuffix(secretStoreInfo.Path, "/") && !strings.HasPrefix(path, "/") {
		path = "/" + path
	} else if strings.HasSuffix(secretStoreInfo.Path, "/") && strings.HasPrefix(path, "/") {
		// remove extra '/' in the full URL path because secret store's (Vault) APIs don't handle extra '/'.
		path = path[1:]
	}

	return path, secretsKV
}
