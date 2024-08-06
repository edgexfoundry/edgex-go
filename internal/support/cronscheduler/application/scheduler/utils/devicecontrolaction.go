//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	bootstrapUtils "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

func issueSetCommandWithDeviceControlAction(dic *di.Container, action models.DeviceControlAction) (string, errors.EdgeX) {
	if action.DeviceName == "" {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "device name cannot be empty", nil)
	}

	if action.SourceName == "" {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "source name cannot be empty", nil)
	}

	var payload map[string]any
	if err := bootstrapUtils.ConvertToMap(action.Payload, &payload); err != nil {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to convert payload to map", err)
	}

	// Retrieve device information through Metadata DeviceClient
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	if dc == nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceClient returned", nil)
	}
	deviceResponse, err := dc.DeviceByName(context.Background(), action.DeviceName)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	// Retrieve device service information through Metadata DeviceClient
	dsc := bootstrapContainer.DeviceServiceClientFrom(dic.Get)
	if dsc == nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceServiceClient returned", nil)
	}
	deviceServiceResponse, err := dsc.DeviceServiceByName(context.Background(), deviceResponse.Device.ServiceName)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	// Issue command by passing the base address of device service into DeviceServiceCommandClient
	dscc := bootstrapContainer.DeviceServiceCommandClientFrom(dic.Get)
	if dscc == nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceServiceCommandClient returned", nil)
	}

	// TODO: Decide if we need to support queryParams here
	resp, err := dscc.SetCommandWithObject(context.Background(), deviceServiceResponse.Service.BaseAddress, action.DeviceName, action.SourceName, "", payload)
	if err != nil {
		return "", err
	}

	return resp.Message, nil
}
