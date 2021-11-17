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

// CommandClient defines the interface for interactions with the command endpoints on the EdgeX Foundry core-command service.
type CommandClient interface {
	// AllDeviceCoreCommands returns a paginated list of MultiDeviceCoreCommandsResponse. The list contains all of the commands in the system associated with their respective device.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	AllDeviceCoreCommands(ctx context.Context, offset int, limit int) (responses.MultiDeviceCoreCommandsResponse, errors.EdgeX)
	// DeviceCoreCommandsByDeviceName returns all commands associated with the specified device name.
	DeviceCoreCommandsByDeviceName(ctx context.Context, deviceName string) (responses.DeviceCoreCommandResponse, errors.EdgeX)
	// IssueGetCommandByName issues the specified read command referenced by the command name to the device/sensor that is also referenced by name.
	// dsPushEvent: If set to yes, a successful GET will result in an event being pushed to the EdgeX system. Default value is no.
	// dsReturnEvent: If set to no, there will be no Event returned in the http response. Default value is yes.
	IssueGetCommandByName(ctx context.Context, deviceName string, commandName string, dsPushEvent string, dsReturnEvent string) (*responses.EventResponse, errors.EdgeX)
	// IssueSetCommandByName issues the specified write command referenced by the command name to the device/sensor that is also referenced by name.
	IssueSetCommandByName(ctx context.Context, deviceName string, commandName string, settings map[string]string) (common.BaseResponse, errors.EdgeX)
	// IssueSetCommandByNameWithObject issues the specified write command and the settings supports object value type
	IssueSetCommandByNameWithObject(ctx context.Context, deviceName string, commandName string, settings map[string]interface{}) (common.BaseResponse, errors.EdgeX)
}
