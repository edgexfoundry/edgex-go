//
// Copyright (C) 2021-2025 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type ProvisionWatcherClient struct {
	baseUrlFunc           clients.ClientBaseUrlFunc
	authInjector          interfaces.AuthenticationInjector
	enableNameFieldEscape bool
}

// NewProvisionWatcherClient creates an instance of ProvisionWatcherClient
func NewProvisionWatcherClient(baseUrl string, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.ProvisionWatcherClient {
	return &ProvisionWatcherClient{
		baseUrlFunc:           clients.GetDefaultClientBaseUrlFunc(baseUrl),
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

// NewProvisionWatcherClientWithUrlCallback creates an instance of ProvisionWatcherClient with ClientBaseUrlFunc.
func NewProvisionWatcherClientWithUrlCallback(baseUrlFunc clients.ClientBaseUrlFunc, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.ProvisionWatcherClient {
	return &ProvisionWatcherClient{
		baseUrlFunc:           baseUrlFunc,
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

func (pwc ProvisionWatcherClient) Add(ctx context.Context, reqs []requests.AddProvisionWatcherRequest) (res []dtoCommon.BaseWithIdResponse, err errors.EdgeX) {
	baseUrl, goErr := clients.GetBaseUrl(pwc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.PostRequestWithRawData(ctx, &res, baseUrl, common.ApiProvisionWatcherRoute, nil, reqs, pwc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}

func (pwc ProvisionWatcherClient) Update(ctx context.Context, reqs []requests.UpdateProvisionWatcherRequest) (res []dtoCommon.BaseResponse, err errors.EdgeX) {
	baseUrl, goErr := clients.GetBaseUrl(pwc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.PatchRequest(ctx, &res, baseUrl, common.ApiProvisionWatcherRoute, nil, reqs, pwc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}

func (pwc ProvisionWatcherClient) AllProvisionWatchers(ctx context.Context, labels []string, offset int, limit int) (res responses.MultiProvisionWatchersResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	if len(labels) > 0 {
		requestParams.Set(common.Labels, strings.Join(labels, common.CommaSeparator))
	}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	baseUrl, goErr := clients.GetBaseUrl(pwc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, common.ApiAllProvisionWatcherRoute, requestParams, pwc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}

func (pwc ProvisionWatcherClient) ProvisionWatcherByName(ctx context.Context, name string) (res responses.ProvisionWatcherResponse, err errors.EdgeX) {
	path := common.NewPathBuilder().EnableNameFieldEscape(pwc.enableNameFieldEscape).
		SetPath(common.ApiProvisionWatcherRoute).SetPath(common.Name).SetNameFieldPath(name).BuildPath()
	baseUrl, goErr := clients.GetBaseUrl(pwc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, path, nil, pwc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}

func (pwc ProvisionWatcherClient) DeleteProvisionWatcherByName(ctx context.Context, name string) (res dtoCommon.BaseResponse, err errors.EdgeX) {
	path := common.NewPathBuilder().EnableNameFieldEscape(pwc.enableNameFieldEscape).
		SetPath(common.ApiProvisionWatcherRoute).SetPath(common.Name).SetNameFieldPath(name).BuildPath()
	baseUrl, goErr := clients.GetBaseUrl(pwc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.DeleteRequest(ctx, &res, baseUrl, path, pwc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}

func (pwc ProvisionWatcherClient) ProvisionWatchersByProfileName(ctx context.Context, name string, offset int, limit int) (res responses.MultiProvisionWatchersResponse, err errors.EdgeX) {
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(pwc.enableNameFieldEscape).
		SetPath(common.ApiProvisionWatcherRoute).SetPath(common.Profile).SetPath(common.Name).SetNameFieldPath(name).BuildPath()
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	baseUrl, goErr := clients.GetBaseUrl(pwc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, pwc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}

func (pwc ProvisionWatcherClient) ProvisionWatchersByServiceName(ctx context.Context, name string, offset int, limit int) (res responses.MultiProvisionWatchersResponse, err errors.EdgeX) {
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(pwc.enableNameFieldEscape).
		SetPath(common.ApiProvisionWatcherRoute).SetPath(common.Service).SetPath(common.Name).SetNameFieldPath(name).BuildPath()
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	baseUrl, goErr := clients.GetBaseUrl(pwc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, pwc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return
}
