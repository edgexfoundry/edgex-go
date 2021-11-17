//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/url"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

type SystemManagementClient struct {
	baseUrl string
}

func NewSystemManagementClient(baseUrl string) interfaces.SystemManagementClient {
	return &SystemManagementClient{
		baseUrl: baseUrl,
	}
}

func (smc *SystemManagementClient) GetHealth(ctx context.Context, services []string) (res []dtoCommon.BaseWithServiceNameResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	requestParams.Set(common.Services, strings.Join(services, common.CommaSeparator))
	err = utils.GetRequest(ctx, &res, smc.baseUrl, common.ApiHealthRoute, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}

func (smc *SystemManagementClient) GetMetrics(ctx context.Context, services []string) (res []dtoCommon.BaseWithMetricsResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	requestParams.Set(common.Services, strings.Join(services, common.CommaSeparator))
	err = utils.GetRequest(ctx, &res, smc.baseUrl, common.ApiMultiMetricsRoute, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}

func (smc *SystemManagementClient) GetConfig(ctx context.Context, services []string) (res []dtoCommon.BaseWithConfigResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	requestParams.Set(common.Services, strings.Join(services, common.CommaSeparator))
	err = utils.GetRequest(ctx, &res, smc.baseUrl, common.ApiMultiConfigRoute, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}

func (smc *SystemManagementClient) DoOperation(ctx context.Context, reqs []requests.OperationRequest) (res []dtoCommon.BaseResponse, err errors.EdgeX) {
	err = utils.PostRequestWithRawData(ctx, &res, smc.baseUrl+common.ApiOperationRoute, reqs)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}
