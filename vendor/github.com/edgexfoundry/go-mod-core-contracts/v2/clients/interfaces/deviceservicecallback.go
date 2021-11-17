//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// DeviceServiceCallbackClient defines the interface for interactions with the callback endpoint on the EdgeX Foundry device service.
type DeviceServiceCallbackClient interface {
	// AddDeviceCallback invokes device service's callback API for adding device
	AddDeviceCallback(ctx context.Context, request requests.AddDeviceRequest) (common.BaseResponse, errors.EdgeX)
	// UpdateDeviceCallback invokes device service's callback API for updating device
	UpdateDeviceCallback(ctx context.Context, request requests.UpdateDeviceRequest) (common.BaseResponse, errors.EdgeX)
	// DeleteDeviceCallback invokes device service's callback API for deleting device
	DeleteDeviceCallback(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
	// UpdateDeviceProfileCallback invokes device service's callback API for updating device profile
	UpdateDeviceProfileCallback(ctx context.Context, request requests.DeviceProfileRequest) (common.BaseResponse, errors.EdgeX)
	// AddProvisionWatcherCallback invokes device service's callback API for adding provision watcher
	AddProvisionWatcherCallback(ctx context.Context, request requests.AddProvisionWatcherRequest) (common.BaseResponse, errors.EdgeX)
	// UpdateProvisionWatcherCallback invokes device service's callback API for updating provision watcher
	UpdateProvisionWatcherCallback(ctx context.Context, request requests.UpdateProvisionWatcherRequest) (common.BaseResponse, errors.EdgeX)
	// DeleteProvisionWatcherCallback invokes device service's callback API for deleting provision watcher
	DeleteProvisionWatcherCallback(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
	// UpdateDeviceServiceCallback invokes device service's callback API for updating device service
	UpdateDeviceServiceCallback(ctx context.Context, request requests.UpdateDeviceServiceRequest) (common.BaseResponse, errors.EdgeX)
}
