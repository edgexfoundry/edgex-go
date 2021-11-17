//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"path"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

type deviceServiceCallbackClient struct {
	baseUrl string
}

// NewDeviceServiceCallbackClient creates an instance of deviceServiceCallbackClient
func NewDeviceServiceCallbackClient(baseUrl string) interfaces.DeviceServiceCallbackClient {
	return &deviceServiceCallbackClient{
		baseUrl: baseUrl,
	}
}

func (client *deviceServiceCallbackClient) AddDeviceCallback(ctx context.Context, request requests.AddDeviceRequest) (dtoCommon.BaseResponse, errors.EdgeX) {
	var response dtoCommon.BaseResponse
	err := utils.PostRequestWithRawData(ctx, &response, client.baseUrl+common.ApiDeviceCallbackRoute, request)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	return response, nil
}

func (client *deviceServiceCallbackClient) UpdateDeviceCallback(ctx context.Context, request requests.UpdateDeviceRequest) (dtoCommon.BaseResponse, errors.EdgeX) {
	var response dtoCommon.BaseResponse
	err := utils.PutRequest(ctx, &response, client.baseUrl+common.ApiDeviceCallbackRoute, request)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	return response, nil
}

func (client *deviceServiceCallbackClient) DeleteDeviceCallback(ctx context.Context, name string) (dtoCommon.BaseResponse, errors.EdgeX) {
	var response dtoCommon.BaseResponse
	requestPath := path.Join(common.ApiDeviceCallbackRoute, common.Name, name)
	err := utils.DeleteRequest(ctx, &response, client.baseUrl, requestPath)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	return response, nil
}

func (client *deviceServiceCallbackClient) UpdateDeviceProfileCallback(ctx context.Context, request requests.DeviceProfileRequest) (dtoCommon.BaseResponse, errors.EdgeX) {
	var response dtoCommon.BaseResponse
	err := utils.PutRequest(ctx, &response, client.baseUrl+common.ApiProfileCallbackRoute, request)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	return response, nil
}

func (client *deviceServiceCallbackClient) AddProvisionWatcherCallback(ctx context.Context, request requests.AddProvisionWatcherRequest) (dtoCommon.BaseResponse, errors.EdgeX) {
	var response dtoCommon.BaseResponse
	err := utils.PostRequestWithRawData(ctx, &response, client.baseUrl+common.ApiWatcherCallbackRoute, request)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	return response, nil
}

func (client *deviceServiceCallbackClient) UpdateProvisionWatcherCallback(ctx context.Context, request requests.UpdateProvisionWatcherRequest) (dtoCommon.BaseResponse, errors.EdgeX) {
	var response dtoCommon.BaseResponse
	err := utils.PutRequest(ctx, &response, client.baseUrl+common.ApiWatcherCallbackRoute, request)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	return response, nil
}

func (client *deviceServiceCallbackClient) DeleteProvisionWatcherCallback(ctx context.Context, name string) (dtoCommon.BaseResponse, errors.EdgeX) {
	var response dtoCommon.BaseResponse
	requestPath := path.Join(common.ApiWatcherCallbackRoute, common.Name, name)
	err := utils.DeleteRequest(ctx, &response, client.baseUrl, requestPath)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	return response, nil
}

func (client *deviceServiceCallbackClient) UpdateDeviceServiceCallback(ctx context.Context, request requests.UpdateDeviceServiceRequest) (dtoCommon.BaseResponse, errors.EdgeX) {
	var response dtoCommon.BaseResponse
	err := utils.PutRequest(ctx, &response, client.baseUrl+common.ApiServiceCallbackRoute, request)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}
	return response, nil
}
