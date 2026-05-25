//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type SecretStoreTokenClient struct {
	baseUrlFunc  clients.ClientBaseUrlFunc
	authInjector interfaces.AuthenticationInjector
}

// NewSecretStoreTokenClient creates an instance of SecretStoreTokenClient
func NewSecretStoreTokenClient(baseUrl string, authInjector interfaces.AuthenticationInjector) interfaces.SecretStoreTokenClient {
	return &SecretStoreTokenClient{
		baseUrlFunc:  clients.GetDefaultClientBaseUrlFunc(baseUrl),
		authInjector: authInjector,
	}
}

// NewSecretStoreTokenClientWithUrlCallback creates an instance of SecretStoreTokenClient with ClientBaseUrlFunc.
func NewSecretStoreTokenClientWithUrlCallback(baseUrlFunc clients.ClientBaseUrlFunc, authInjector interfaces.AuthenticationInjector) interfaces.AuthClient {
	return &AuthClient{
		baseUrlFunc:  baseUrlFunc,
		authInjector: authInjector,
	}
}

// RegenToken regenerates the secret store client token based on the specified entity id
func (ac *SecretStoreTokenClient) RegenToken(ctx context.Context, entityId string) (dtoCommon.BaseResponse, errors.EdgeX) {
	var response dtoCommon.BaseResponse
	baseUrl, err := clients.GetBaseUrl(ac.baseUrlFunc)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}

	path := common.NewPathBuilder().SetPath(common.ApiTokenRoute).SetPath(common.EntityId).SetPath(entityId).BuildPath()
	err = utils.PutRequest(ctx, &response, baseUrl, path, nil, nil, ac.authInjector)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	return response, nil
}
