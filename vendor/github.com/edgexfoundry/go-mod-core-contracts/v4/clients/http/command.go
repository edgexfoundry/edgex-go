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

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type CommandClient struct {
	baseUrlFunc           clients.ClientBaseUrlFunc
	authInjector          interfaces.AuthenticationInjector
	enableNameFieldEscape bool
}

// NewCommandClient creates an instance of CommandClient
func NewCommandClient(baseUrl string, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.CommandClient {
	return &CommandClient{
		baseUrlFunc:           clients.GetDefaultClientBaseUrlFunc(baseUrl),
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

// NewCommandClientWithUrlCallback creates an instance of CommandClient with ClientBaseUrlFunc.
func NewCommandClientWithUrlCallback(baseUrlFunc clients.ClientBaseUrlFunc, authInjector interfaces.AuthenticationInjector, enableNameFieldEscape bool) interfaces.CommandClient {
	return &CommandClient{
		baseUrlFunc:           baseUrlFunc,
		authInjector:          authInjector,
		enableNameFieldEscape: enableNameFieldEscape,
	}
}

// AllDeviceCoreCommands returns a paginated list of MultiDeviceCoreCommandsResponse. The list contains all of the commands in the system associated with their respective device.
func (client *CommandClient) AllDeviceCoreCommands(ctx context.Context, offset int, limit int) (
	res responses.MultiDeviceCoreCommandsResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	baseUrl, goErr := clients.GetBaseUrl(client.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, common.ApiAllDeviceRoute, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// DeviceCoreCommandsByDeviceName returns all commands associated with the specified device name.
func (client *CommandClient) DeviceCoreCommandsByDeviceName(ctx context.Context, name string) (
	res responses.DeviceCoreCommandResponse, err errors.EdgeX) {
	path := common.NewPathBuilder().EnableNameFieldEscape(client.enableNameFieldEscape).
		SetPath(common.ApiDeviceRoute).SetPath(common.Name).SetNameFieldPath(name).BuildPath()
	baseUrl, goErr := clients.GetBaseUrl(client.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, path, nil, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// IssueGetCommandByName issues the specified read command referenced by the command name to the device/sensor that is also referenced by name.
func (client *CommandClient) IssueGetCommandByName(ctx context.Context, deviceName string, commandName string, dsPushEvent bool, dsReturnEvent bool) (res *responses.EventResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	requestParams.Set(common.PushEvent, strconv.FormatBool(dsPushEvent))
	requestParams.Set(common.ReturnEvent, strconv.FormatBool(dsReturnEvent))
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(client.enableNameFieldEscape).
		SetPath(common.ApiDeviceRoute).SetPath(common.Name).SetNameFieldPath(deviceName).SetNameFieldPath(commandName).BuildPath()
	baseUrl, goErr := clients.GetBaseUrl(client.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (client *CommandClient) IssueGetCommandByNameWithQueryParams(ctx context.Context, deviceName string, commandName string, queryParams map[string]string) (res *responses.EventResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	for k, v := range queryParams {
		requestParams.Set(k, v)
	}
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(client.enableNameFieldEscape).
		SetPath(common.ApiDeviceRoute).SetPath(common.Name).SetNameFieldPath(deviceName).SetNameFieldPath(commandName).BuildPath()
	baseUrl, goErr := clients.GetBaseUrl(client.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.GetRequest(ctx, &res, baseUrl, requestPath, requestParams, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

// IssueSetCommandByName issues the specified write command referenced by the command name to the device/sensor that is also referenced by name.
func (client *CommandClient) IssueSetCommandByName(ctx context.Context, deviceName string, commandName string, settings map[string]any) (res dtoCommon.BaseResponse, err errors.EdgeX) {
	requestPath := common.NewPathBuilder().EnableNameFieldEscape(client.enableNameFieldEscape).
		SetPath(common.ApiDeviceRoute).SetPath(common.Name).SetNameFieldPath(deviceName).SetNameFieldPath(commandName).BuildPath()
	baseUrl, goErr := clients.GetBaseUrl(client.baseUrlFunc)
	if goErr != nil {
		return res, errors.NewCommonEdgeXWrapper(goErr)
	}
	err = utils.PutRequest(ctx, &res, baseUrl, requestPath, nil, settings, client.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
