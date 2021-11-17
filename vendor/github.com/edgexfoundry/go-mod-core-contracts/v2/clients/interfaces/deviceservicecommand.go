//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// DeviceServiceCommandClient defines the interface for interactions with the device command endpoints on the EdgeX Foundry device services.
type DeviceServiceCommandClient interface {
	// GetCommand invokes device service's command API for issuing get(read) command
	GetCommand(ctx context.Context, baseUrl string, deviceName string, commandName string, queryParams string) (*responses.EventResponse, errors.EdgeX)
	// SetCommand invokes device service's command API for issuing set(write) command
	SetCommand(ctx context.Context, baseUrl string, deviceName string, commandName string, queryParams string, settings map[string]string) (common.BaseResponse, errors.EdgeX)
	// SetCommandWithObject invokes device service's set command API and the settings supports object value type
	SetCommandWithObject(ctx context.Context, baseUrl string, deviceName string, commandName string, queryParams string, settings map[string]interface{}) (common.BaseResponse, errors.EdgeX)
}
