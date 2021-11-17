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

type SubscriptionClient struct {
	baseUrl string
}

// NewSubscriptionClient creates an instance of SubscriptionClient
func NewSubscriptionClient(baseUrl string) interfaces.SubscriptionClient {
	return &SubscriptionClient{
		baseUrl: baseUrl,
	}
}

// Add adds new subscriptions.
func (client *SubscriptionClient) Add(ctx context.Context, reqs []requests.AddSubscriptionRequest) (res []dtoCommon.BaseWithIdResponse, err errors.EdgeX) {
	err = utils.PostRequestWithRawData(ctx, &res, client.baseUrl+common.ApiSubscriptionRoute, reqs)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// Update updates subscriptions.
func (client *SubscriptionClient) Update(ctx context.Context, reqs []requests.UpdateSubscriptionRequest) (res []dtoCommon.BaseResponse, err errors.EdgeX) {
	err = utils.PatchRequest(ctx, &res, client.baseUrl+common.ApiSubscriptionRoute, reqs)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// AllSubscriptions queries subscriptions with offset and limit
func (client *SubscriptionClient) AllSubscriptions(ctx context.Context, offset int, limit int) (res responses.MultiSubscriptionsResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, common.ApiAllSubscriptionRoute, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// SubscriptionsByCategory queries subscriptions with category, offset and limit
func (client *SubscriptionClient) SubscriptionsByCategory(ctx context.Context, category string, offset int, limit int) (res responses.MultiSubscriptionsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiSubscriptionRoute, common.Category, url.QueryEscape(category))
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// SubscriptionsByLabel queries subscriptions with label, offset and limit
func (client *SubscriptionClient) SubscriptionsByLabel(ctx context.Context, label string, offset int, limit int) (res responses.MultiSubscriptionsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiSubscriptionRoute, common.Label, url.QueryEscape(label))
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// SubscriptionsByReceiver queries subscriptions with receiver, offset and limit
func (client *SubscriptionClient) SubscriptionsByReceiver(ctx context.Context, receiver string, offset int, limit int) (res responses.MultiSubscriptionsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiSubscriptionRoute, common.Receiver, url.QueryEscape(receiver))
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// SubscriptionByName query subscription by name.
func (client *SubscriptionClient) SubscriptionByName(ctx context.Context, name string) (res responses.SubscriptionResponse, err errors.EdgeX) {
	path := path.Join(common.ApiSubscriptionRoute, common.Name, url.QueryEscape(name))
	err = utils.GetRequest(ctx, &res, client.baseUrl, path, nil)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// DeleteSubscriptionByName deletes a subscription by name.
func (client *SubscriptionClient) DeleteSubscriptionByName(ctx context.Context, name string) (res dtoCommon.BaseResponse, err errors.EdgeX) {
	path := path.Join(common.ApiSubscriptionRoute, common.Name, url.QueryEscape(name))
	err = utils.DeleteRequest(ctx, &res, client.baseUrl, path)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
