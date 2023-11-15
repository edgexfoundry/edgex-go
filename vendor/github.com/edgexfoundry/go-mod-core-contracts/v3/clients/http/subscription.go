//
// Copyright (C) 2021 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/url"
	"path"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
)

type SubscriptionClient struct {
	baseUrl               string
	authInjector          interfaces.AuthenticationInjector
	enableNameFieldEscape bool
}

// NewSubscriptionClient creates an instance of SubscriptionClient
func NewSubscriptionClient(baseUrl string, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.SubscriptionClient {
	return &SubscriptionClient{
		baseUrl:               baseUrl,
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

// Add adds new subscriptions.
func (client *SubscriptionClient) Add(ctx context.Context, reqs []requests.AddSubscriptionRequest) (res []dtoCommon.BaseWithIdResponse, err errors.EdgeX) {
	err = utils.PostRequestWithRawData(ctx, &res, client.baseUrl, common.ApiSubscriptionRoute, nil, reqs, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// Update updates subscriptions.
func (client *SubscriptionClient) Update(ctx context.Context, reqs []requests.UpdateSubscriptionRequest) (res []dtoCommon.BaseResponse, err errors.EdgeX) {
	err = utils.PatchRequest(ctx, &res, client.baseUrl, common.ApiSubscriptionRoute, nil, reqs, client.authInjector)
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
	err = utils.GetRequest(ctx, &res, client.baseUrl, common.ApiAllSubscriptionRoute, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// SubscriptionsByCategory queries subscriptions with category, offset and limit
func (client *SubscriptionClient) SubscriptionsByCategory(ctx context.Context, category string, offset int, limit int) (res responses.MultiSubscriptionsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiSubscriptionRoute, common.Category, category)
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// SubscriptionsByLabel queries subscriptions with label, offset and limit
func (client *SubscriptionClient) SubscriptionsByLabel(ctx context.Context, label string, offset int, limit int) (res responses.MultiSubscriptionsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiSubscriptionRoute, common.Label, label)
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// SubscriptionsByReceiver queries subscriptions with receiver, offset and limit
func (client *SubscriptionClient) SubscriptionsByReceiver(ctx context.Context, receiver string, offset int, limit int) (res responses.MultiSubscriptionsResponse, err errors.EdgeX) {
	requestPath := path.Join(common.ApiSubscriptionRoute, common.Receiver, receiver)
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, client.baseUrl, requestPath, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// SubscriptionByName query subscription by name.
func (client *SubscriptionClient) SubscriptionByName(ctx context.Context, name string) (res responses.SubscriptionResponse, err errors.EdgeX) {
	path := common.NewPathBuilder().EnableNameFieldEscape(client.enableNameFieldEscape).
		SetPath(common.ApiSubscriptionRoute).SetPath(common.Name).SetNameFieldPath(name).BuildPath()
	err = utils.GetRequest(ctx, &res, client.baseUrl, path, nil, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// DeleteSubscriptionByName deletes a subscription by name.
func (client *SubscriptionClient) DeleteSubscriptionByName(ctx context.Context, name string) (res dtoCommon.BaseResponse, err errors.EdgeX) {
	path := common.NewPathBuilder().EnableNameFieldEscape(client.enableNameFieldEscape).
		SetPath(common.ApiSubscriptionRoute).SetPath(common.Name).SetNameFieldPath(name).BuildPath()
	err = utils.DeleteRequest(ctx, &res, client.baseUrl, path, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
