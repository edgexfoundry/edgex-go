//
// Copyright (C) 2024-2025 IOTech Ltd
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
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type AuthClient struct {
	baseUrlFunc  clients.ClientBaseUrlFunc
	authInjector interfaces.AuthenticationInjector
}

// NewAuthClient creates an instance of AuthClient
func NewAuthClient(baseUrl string, authInjector interfaces.AuthenticationInjector) interfaces.AuthClient {
	return &AuthClient{
		baseUrlFunc:  clients.GetDefaultClientBaseUrlFunc(baseUrl),
		authInjector: authInjector,
	}
}

// NewAuthClientWithUrlCallback creates an instance of AuthClient with ClientBaseUrlFunc.
func NewAuthClientWithUrlCallback(baseUrlFunc clients.ClientBaseUrlFunc, authInjector interfaces.AuthenticationInjector) interfaces.AuthClient {
	return &AuthClient{
		baseUrlFunc:  baseUrlFunc,
		authInjector: authInjector,
	}
}

// AddKey adds new key
func (ac *AuthClient) AddKey(ctx context.Context, req requests.AddKeyDataRequest) (dtoCommon.BaseResponse, errors.EdgeX) {
	var response dtoCommon.BaseResponse
	baseUrl, err := clients.GetBaseUrl(ac.baseUrlFunc)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.PostRequestWithRawData(ctx, &response, baseUrl, common.ApiKeyRoute, nil, req, ac.authInjector)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	return response, nil
}

func (ac *AuthClient) VerificationKeyByIssuer(ctx context.Context, issuer string) (res responses.KeyDataResponse, err errors.EdgeX) {
	path := common.NewPathBuilder().SetPath(common.ApiKeyRoute).SetPath(common.VerificationKeyType).SetPath(common.Issuer).SetNameFieldPath(issuer).BuildPath()
	baseUrl, goErr := clients.GetBaseUrl(ac.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, path, nil, ac.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
