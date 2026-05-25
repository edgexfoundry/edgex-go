//
// Copyright (C) 2021 IOTech Ltd
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

type generalClient struct {
	baseUrl      string
	authInjector interfaces.AuthenticationInjector
}

func NewGeneralClient(baseUrl string, authInjector interfaces.AuthenticationInjector) interfaces.GeneralClient {
	return &generalClient{
		baseUrl:      baseUrl,
		authInjector: authInjector,
	}
}

func (g *generalClient) FetchConfiguration(ctx context.Context) (res dtoCommon.ConfigResponse, err errors.EdgeX) {
	err = utils.GetRequest(ctx, &res, g.baseUrl, common.ApiConfigRoute, nil, g.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return res, nil
}
