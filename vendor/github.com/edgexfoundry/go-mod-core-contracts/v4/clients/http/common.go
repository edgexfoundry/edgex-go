//
// Copyright (C) 2020-2021 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type commonClient struct {
	baseUrl      string
	authInjector interfaces.AuthenticationInjector
}

// NewCommonClient creates an instance of CommonClient
func NewCommonClient(baseUrl string, authInjector interfaces.AuthenticationInjector) interfaces.CommonClient {
	return &commonClient{
		baseUrl:      baseUrl,
		authInjector: authInjector,
	}
}

func (cc *commonClient) Configuration(ctx context.Context) (dtoCommon.ConfigResponse, errors.EdgeX) {
	cr := dtoCommon.ConfigResponse{}
	err := utils.GetRequest(ctx, &cr, cc.baseUrl, common.ApiConfigRoute, nil, cc.authInjector)
	if err != nil {
		return cr, errors.NewCommonEdgeXWrapper(err)
	}
	return cr, nil
}

func (cc *commonClient) Ping(ctx context.Context) (dtoCommon.PingResponse, errors.EdgeX) {
	pr := dtoCommon.PingResponse{}
	err := utils.GetRequest(ctx, &pr, cc.baseUrl, common.ApiPingRoute, nil, cc.authInjector)
	if err != nil {
		return pr, errors.NewCommonEdgeXWrapper(err)
	}
	return pr, nil
}

func (cc *commonClient) Version(ctx context.Context) (dtoCommon.VersionResponse, errors.EdgeX) {
	vr := dtoCommon.VersionResponse{}
	err := utils.GetRequest(ctx, &vr, cc.baseUrl, common.ApiVersionRoute, nil, cc.authInjector)
	if err != nil {
		return vr, errors.NewCommonEdgeXWrapper(err)
	}
	return vr, nil
}

func (cc *commonClient) AddSecret(ctx context.Context, request dtoCommon.SecretRequest) (res dtoCommon.BaseResponse, err errors.EdgeX) {
	err = utils.PostRequestWithRawData(ctx, &res, cc.baseUrl, common.ApiSecretRoute, nil, request, cc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
