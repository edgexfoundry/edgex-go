//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/url"
	"path"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

type IntervalClient struct {
	baseUrl string
}

// NewIntervalClient creates an instance of IntervalClient
func NewIntervalClient(baseUrl string) interfaces.IntervalClient {
	return &IntervalClient{
		baseUrl: baseUrl,
	}
}

// Add adds new intervals
func (client IntervalClient) Add(ctx context.Context, reqs []requests.AddIntervalRequest) (
	res []dtoCommon.BaseWithIdResponse, err errors.EdgeX) {
	err = utils.PostRequestWithRawData(ctx, &res, client.baseUrl+common.ApiIntervalRoute, reqs)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// Update updates intervals
func (client IntervalClient) Update(ctx context.Context, reqs []requests.UpdateIntervalRequest) (
	res []dtoCommon.BaseResponse, err errors.EdgeX) {
	err = utils.PatchRequest(ctx, &res, client.baseUrl+common.ApiIntervalRoute, reqs)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// AllIntervals query the intervals with offset, limit
func (client IntervalClient) AllIntervals(ctx context.Context, offset int, limit int) (
	res responses.MultiIntervalsResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, common.ApiAllIntervalRoute, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// IntervalByName query the interval by name
func (client IntervalClient) IntervalByName(ctx context.Context, name string) (
	res responses.IntervalResponse, err errors.EdgeX) {
	path := path.Join(common.ApiIntervalRoute, common.Name, url.QueryEscape(name))
	err = utils.GetRequest(ctx, &res, client.baseUrl, path, nil)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// DeleteIntervalByName delete the interval by name
func (client IntervalClient) DeleteIntervalByName(ctx context.Context, name string) (
	res dtoCommon.BaseResponse, err errors.EdgeX) {
	path := path.Join(common.ApiIntervalRoute, common.Name, url.QueryEscape(name))
	err = utils.DeleteRequest(ctx, &res, client.baseUrl, path)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
