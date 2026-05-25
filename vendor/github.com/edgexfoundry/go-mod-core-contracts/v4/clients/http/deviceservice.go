//
// Copyright (C) 2020-2021 Unknown author
// Copyright (C) 2023 Intel Corporation
// Copyright (C) 2025 IOTech Ltd
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

type DeviceServiceClient struct {
	baseUrlFunc           clients.ClientBaseUrlFunc
	authInjector          interfaces.AuthenticationInjector
	enableNameFieldEscape bool
}

// NewDeviceServiceClient creates an instance of DeviceServiceClient
func NewDeviceServiceClient(baseUrl string, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.DeviceServiceClient {
	return &DeviceServiceClient{
		baseUrlFunc:           clients.GetDefaultClientBaseUrlFunc(baseUrl),
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

// NewDeviceServiceClientWithUrlCallback creates an instance of DeviceServiceClient with ClientBaseUrlFunc.
func NewDeviceServiceClientWithUrlCallback(baseUrlFunc clients.ClientBaseUrlFunc, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.DeviceServiceClient {
	return &DeviceServiceClient{
		baseUrlFunc:           baseUrlFunc,
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

func (dsc DeviceServiceClient) Add(ctx context.Context, reqs []requests.AddDeviceServiceRequest) (
	res []dtoCommon.BaseWithIdResponse, err errors.EdgeX) {
	baseUrl, goErr := clients.GetBaseUrl(dsc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.PostRequestWithRawData(ctx, &res, baseUrl, common.ApiDeviceServiceRoute, nil, reqs, dsc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (dsc DeviceServiceClient) Update(ctx context.Context, reqs []requests.UpdateDeviceServiceRequest) (
	res []dtoCommon.BaseResponse, err errors.EdgeX) {
	baseUrl, goErr := clients.GetBaseUrl(dsc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.PatchRequest(ctx, &res, baseUrl, common.ApiDeviceServiceRoute, nil, reqs, dsc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (dsc DeviceServiceClient) AllDeviceServices(ctx context.Context, labels []string, offset int, limit int) (
	res responses.MultiDeviceServicesResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	if len(labels) > 0 {
		requestParams.Set(common.Labels, strings.Join(labels, common.CommaSeparator))
	}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	baseUrl, goErr := clients.GetBaseUrl(dsc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, common.ApiAllDeviceServiceRoute, requestParams, dsc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (dsc DeviceServiceClient) DeviceServiceByName(ctx context.Context, name string) (
	res responses.DeviceServiceResponse, err errors.EdgeX) {
	path := common.NewPathBuilder().EnableNameFieldEscape(dsc.enableNameFieldEscape).
		SetPath(common.ApiDeviceServiceRoute).SetPath(common.Name).SetNameFieldPath(name).BuildPath()
	baseUrl, goErr := clients.GetBaseUrl(dsc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, path, nil, dsc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (dsc DeviceServiceClient) DeleteByName(ctx context.Context, name string) (
	res dtoCommon.BaseResponse, err errors.EdgeX) {
	path := common.NewPathBuilder().EnableNameFieldEscape(dsc.enableNameFieldEscape).
		SetPath(common.ApiDeviceServiceRoute).SetPath(common.Name).SetNameFieldPath(name).BuildPath()
	baseUrl, goErr := clients.GetBaseUrl(dsc.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.DeleteRequest(ctx, &res, baseUrl, path, dsc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
