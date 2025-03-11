//
// Copyright (C) 2020-2025 IOTech Ltd
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
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type readingClient struct {
	baseUrlFunc           clients.ClientBaseUrlFunc
	authInjector          interfaces.AuthenticationInjector
	enableNameFieldEscape bool
}

// NewReadingClient creates an instance of ReadingClient
func NewReadingClient(baseUrl string, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.ReadingClient {
	return &readingClient{
		baseUrlFunc:           clients.GetDefaultClientBaseUrlFunc(baseUrl),
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

// NewReadingClientWithUrlCallback creates an instance of ReadingClient with ClientBaseUrlFunc.
func NewReadingClientWithUrlCallback(baseUrlFunc clients.ClientBaseUrlFunc, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.ReadingClient {
	return &readingClient{
		baseUrlFunc:           baseUrlFunc,
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

func (rc readingClient) AllReadings(ctx context.Context, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX) {
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	res := responses.MultiReadingsResponse{}
	baseUrl, err := clients.GetBaseUrl(rc.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, common.ApiAllReadingRoute, requestParams, rc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (rc readingClient) ReadingCount(ctx context.Context) (dtoCommon.CountResponse, errors.EdgeX) {
	res := dtoCommon.CountResponse{}
	baseUrl, err := clients.GetBaseUrl(rc.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, common.ApiReadingCountRoute, nil, rc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (rc readingClient) ReadingCountByDeviceName(ctx context.Context, name string) (dtoCommon.CountResponse, errors.EdgeX) {
	requestPath := path.Join(common.ApiReadingCountRoute, common.Device, common.Name, name)
	res := dtoCommon.CountResponse{}
	baseUrl, err := clients.GetBaseUrl(rc.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, nil, rc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (rc readingClient) ReadingsByDeviceName(ctx context.Context, name string, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX) {
	requestPath := path.Join(common.ApiReadingRoute, common.Device, common.Name, name)
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	res := responses.MultiReadingsResponse{}
	baseUrl, err := clients.GetBaseUrl(rc.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, rc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (rc readingClient) ReadingsByResourceName(ctx context.Context, name string, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX) {
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(rc.enableNameFieldEscape).
		SetPath(common.ApiReadingRoute).SetPath(common.ResourceName).SetNameFieldPath(name).BuildPath()
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	res := responses.MultiReadingsResponse{}
	baseUrl, err := clients.GetBaseUrl(rc.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, rc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (rc readingClient) ReadingsByTimeRange(ctx context.Context, start, end int64, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX) {
	requestPath := path.Join(common.ApiReadingRoute, common.Start, strconv.FormatInt(start, 10), common.End, strconv.FormatInt(end, 10))
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	res := responses.MultiReadingsResponse{}
	baseUrl, err := clients.GetBaseUrl(rc.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, rc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// ReadingsByResourceNameAndTimeRange returns readings by resource name and specified time range. Readings are sorted in descending order of origin time.
func (rc readingClient) ReadingsByResourceNameAndTimeRange(ctx context.Context, name string, start, end int64, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX) {
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(rc.enableNameFieldEscape).
		SetPath(common.ApiReadingRoute).SetPath(common.ResourceName).SetNameFieldPath(name).SetPath(common.Start).SetPath(strconv.FormatInt(start, 10)).SetPath(common.End).SetPath(strconv.FormatInt(end, 10)).BuildPath()
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	res := responses.MultiReadingsResponse{}
	baseUrl, err := clients.GetBaseUrl(rc.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, rc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (rc readingClient) ReadingsByDeviceNameAndResourceName(ctx context.Context, deviceName, resourceName string, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX) {
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(rc.enableNameFieldEscape).
		SetPath(common.ApiReadingRoute).SetPath(common.Device).SetPath(common.Name).SetNameFieldPath(deviceName).SetPath(common.ResourceName).SetNameFieldPath(resourceName).BuildPath()
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	res := responses.MultiReadingsResponse{}
	baseUrl, err := clients.GetBaseUrl(rc.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, rc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil

}

func (rc readingClient) ReadingsByDeviceNameAndResourceNameAndTimeRange(ctx context.Context, deviceName, resourceName string, start, end int64, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX) {
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(rc.enableNameFieldEscape).
		SetPath(common.ApiReadingRoute).SetPath(common.Device).SetPath(common.Name).SetNameFieldPath(deviceName).SetPath(common.ResourceName).SetNameFieldPath(resourceName).
		SetPath(common.Start).SetPath(strconv.FormatInt(start, 10)).SetPath(common.End).SetPath(strconv.FormatInt(end, 10)).BuildPath()
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	res := responses.MultiReadingsResponse{}
	baseUrl, err := clients.GetBaseUrl(rc.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, rc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (rc readingClient) ReadingsByDeviceNameAndResourceNamesAndTimeRange(ctx context.Context, deviceName string, resourceNames []string, start, end int64, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX) {
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(rc.enableNameFieldEscape).
		SetPath(common.ApiReadingRoute).SetPath(common.Device).SetPath(common.Name).SetNameFieldPath(deviceName).
		SetPath(common.Start).SetPath(strconv.FormatInt(start, 10)).SetPath(common.End).SetPath(strconv.FormatInt(end, 10)).BuildPath()
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	var queryPayload map[string]interface{}
	if len(resourceNames) > 0 { // gosimple S1009: len(nil slice) == 0
		queryPayload = make(map[string]interface{}, 1)
		queryPayload[common.ResourceNames] = resourceNames
	}
	res := responses.MultiReadingsResponse{}
	baseUrl, err := clients.GetBaseUrl(rc.baseUrlFunc)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	err = utils.GetRequestWithBodyRawData(ctx, &res, baseUrl, requestPath, requestParams, queryPayload, rc.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
