//
// Copyright (C) 2021-2025 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/url"
	"path"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type eventClient struct {
	baseUrlFunc           clients.ClientBaseUrlFunc
	authInjector          interfaces.AuthenticationInjector
	enableNameFieldEscape bool
}

// NewEventClient creates an instance of EventClient
func NewEventClient(baseUrl string, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.EventClient {
	return &eventClient{
		baseUrlFunc:           clients.GetDefaultClientBaseUrlFunc(baseUrl),
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

// NewEventClientWithUrlCallback creates an instance of EventClient with ClientBaseUrlFunc.
func NewEventClientWithUrlCallback(baseUrlFunc clients.ClientBaseUrlFunc, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.EventClient {
	return &eventClient{
		baseUrlFunc:           baseUrlFunc,
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

func (ec *eventClient) Add(ctx context.Context, serviceName string, req requests.AddEventRequest) (
	dtoCommon.BaseWithIdResponse, errors.EdgeX) {
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(ec.enableNameFieldEscape).
		SetPath(common.ApiEventRoute).SetNameFieldPath(serviceName).SetNameFieldPath(req.Event.ProfileName).SetNameFieldPath(req.Event.DeviceName).SetNameFieldPath(req.Event.SourceName).BuildPath()
	var br dtoCommon.BaseWithIdResponse

	bytes, encoding, err := req.Encode()
	if err != nil {
		return br, errors.NewCommonEdgeXWrapper(err)
	}
	baseUrl, err := clients.GetBaseUrl(ec.baseUrlFunc)
	if err != nil {
		return br, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.PostRequest(ctx, &br, baseUrl, requestPath, bytes, encoding, ec.authInjector)
	if err != nil {
		return br, errors.NewCommonEdgeXWrapper(err)
	}
	return br, nil
}

func (ec *eventClient) AllEvents(ctx context.Context, offset, limit int) (responses.MultiEventsResponse, errors.EdgeX) {
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	res := responses.MultiEventsResponse{}
	baseUrl, err := clients.GetBaseUrl(ec.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, common.ApiAllEventRoute, requestParams, ec.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (ec *eventClient) EventCount(ctx context.Context) (dtoCommon.CountResponse, errors.EdgeX) {
	res := dtoCommon.CountResponse{}
	baseUrl, err := clients.GetBaseUrl(ec.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, common.ApiEventCountRoute, nil, ec.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (ec *eventClient) EventCountByDeviceName(ctx context.Context, name string) (dtoCommon.CountResponse, errors.EdgeX) {
	requestPath := path.Join(common.ApiEventCountRoute, common.Device, common.Name, name)
	res := dtoCommon.CountResponse{}
	baseUrl, err := clients.GetBaseUrl(ec.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, nil, ec.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (ec *eventClient) EventsByDeviceName(ctx context.Context, name string, offset, limit int) (
	responses.MultiEventsResponse, errors.EdgeX) {
	requestPath := path.Join(common.ApiEventRoute, common.Device, common.Name, name)
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	res := responses.MultiEventsResponse{}
	baseUrl, err := clients.GetBaseUrl(ec.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, ec.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (ec *eventClient) DeleteByDeviceName(ctx context.Context, name string) (dtoCommon.BaseResponse, errors.EdgeX) {
	requestPath := path.Join(common.ApiEventRoute, common.Device, common.Name, name)
	res := dtoCommon.BaseResponse{}
	baseUrl, err := clients.GetBaseUrl(ec.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.DeleteRequest(ctx, &res, baseUrl, requestPath, ec.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (ec *eventClient) EventsByTimeRange(ctx context.Context, start, end int64, offset, limit int) (
	responses.MultiEventsResponse, errors.EdgeX) {
	requestPath := path.Join(common.ApiEventRoute, common.Start, strconv.FormatInt(start, 10), common.End, strconv.FormatInt(end, 10))
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	res := responses.MultiEventsResponse{}
	baseUrl, err := clients.GetBaseUrl(ec.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, ec.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (ec *eventClient) DeleteByAge(ctx context.Context, age int) (dtoCommon.BaseResponse, errors.EdgeX) {
	requestPath := path.Join(common.ApiEventRoute, common.Age, strconv.Itoa(age))
	res := dtoCommon.BaseResponse{}
	baseUrl, err := clients.GetBaseUrl(ec.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.DeleteRequest(ctx, &res, baseUrl, requestPath, ec.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (ec *eventClient) DeleteById(ctx context.Context, id string) (dtoCommon.BaseResponse, errors.EdgeX) {
	requestPath := path.Join(common.ApiEventRoute, common.Id, id)
	res := dtoCommon.BaseResponse{}
	baseUrl, err := clients.GetBaseUrl(ec.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.DeleteRequest(ctx, &res, baseUrl, requestPath, ec.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
