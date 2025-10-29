//
// Copyright (C) 2021-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	commandContainer "github.com/edgexfoundry/edgex-go/internal/core/command/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// AllCommands query commands by offset, and limit
func AllCommands(offset int, limit int, dic *di.Container) (deviceCoreCommands []dtos.DeviceCoreCommand, totalCount int64, err errors.EdgeX) {
	// retrieve device information through Metadata DeviceClient
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	if dc == nil {
		return deviceCoreCommands, totalCount, errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceClient returned", nil)
	}
	multiDevicesResponse, err := dc.AllDevices(context.Background(), nil, offset, limit)
	if err != nil {
		return deviceCoreCommands, totalCount, errors.NewCommonEdgeXWrapper(err)
	}

	// retrieve device profile information through Metadata DeviceProfileClient
	dpc := bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	if dpc == nil {
		return deviceCoreCommands, totalCount, errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceProfileClient returned", nil)
	}

	// Prepare the url for command
	configuration := commandContainer.ConfigurationFrom(dic.Get)
	serviceUrl := configuration.Service.Url()

	for _, device := range multiDevicesResponse.Devices {
		if len(device.ProfileName) == 0 {
			// if the profile is not set, skip the profile query
			continue
		}
		deviceProfileResponse, err := dpc.DeviceProfileByName(context.Background(), device.ProfileName)
		if err != nil {
			return deviceCoreCommands, totalCount, errors.NewCommonEdgeXWrapper(err)
		}
		commands, err := dtos.BuildCoreCommands(device.Name, serviceUrl, deviceProfileResponse.Profile)
		if err != nil {
			return nil, totalCount, errors.NewCommonEdgeXWrapper(err)
		}
		deviceCoreCommands = append(deviceCoreCommands, dtos.DeviceCoreCommand{
			DeviceName:   device.Name,
			ProfileName:  device.ProfileName,
			CoreCommands: commands,
		})
	}
	return deviceCoreCommands, multiDevicesResponse.TotalCount, nil
}

// IssueGetCommandByName issues the specified get(read) command referenced by the command name to the device/sensor, also
// referenced by name.
func IssueGetCommandByName(deviceName string, commandName string, queryParams string, dic *di.Container) (res *responses.EventResponse, err errors.EdgeX) {
	if deviceName == "" {
		return res, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name cannot be empty", nil)
	}

	if commandName == "" {
		return res, errors.NewCommonEdgeX(errors.KindContractInvalid, "command name cannot be empty", nil)
	}

	// retrieve device information through Metadata DeviceClient
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	if dc == nil {
		return res, errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceClient returned", nil)
	}
	deviceResponse, err := dc.DeviceByName(context.Background(), deviceName)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	// retrieve device service information through Metadata DeviceClient
	dsc := bootstrapContainer.DeviceServiceClientFrom(dic.Get)
	if dsc == nil {
		return res, errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceServiceClient returned", nil)
	}
	deviceServiceResponse, err := dsc.DeviceServiceByName(context.Background(), deviceResponse.Device.ServiceName)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	// Issue command by passing the base address of device service into DeviceServiceCommandClient
	dscc := bootstrapContainer.DeviceServiceCommandClientFrom(dic.Get)
	if dscc == nil {
		return res, errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceServiceCommandClient returned", nil)
	}
	res, err = dscc.GetCommand(context.Background(), deviceServiceResponse.Service.BaseAddress, deviceName, commandName, queryParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return res, nil
}

// IssueSetCommandByName issues the specified set(write) command referenced by the command name to the device/sensor, also
// referenced by name.
func IssueSetCommandByName(deviceName string, commandName string, queryParams string, settings map[string]interface{}, dic *di.Container) (response commonDTO.BaseResponse, err errors.EdgeX) {
	if deviceName == "" {
		return response, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name cannot be empty", nil)
	}

	if commandName == "" {
		return response, errors.NewCommonEdgeX(errors.KindContractInvalid, "command name cannot be empty", nil)
	}

	// retrieve device information through Metadata DeviceClient
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	if dc == nil {
		return response, errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceClient returned", nil)
	}
	deviceResponse, err := dc.DeviceByName(context.Background(), deviceName)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}

	// retrieve device service information through Metadata DeviceClient
	dsc := bootstrapContainer.DeviceServiceClientFrom(dic.Get)
	if dsc == nil {
		return response, errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceServiceClient returned", nil)
	}
	deviceServiceResponse, err := dsc.DeviceServiceByName(context.Background(), deviceResponse.Device.ServiceName)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}

	// Issue command by passing the base address of device service into DeviceServiceCommandClient
	dscc := bootstrapContainer.DeviceServiceCommandClientFrom(dic.Get)
	if dscc == nil {
		return response, errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceServiceCommandClient returned", nil)
	}
	return dscc.SetCommandWithObject(context.Background(), deviceServiceResponse.Service.BaseAddress, deviceName, commandName, queryParams, settings)
}

// CommandsByDeviceName query coreCommands with device name
func CommandsByDeviceName(name string, dic *di.Container) (deviceCoreCommand dtos.DeviceCoreCommand, err errors.EdgeX) {
	if name == "" {
		return deviceCoreCommand, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}

	// retrieve device information through Metadata DeviceClient
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	if dc == nil {
		return deviceCoreCommand, errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceClient returned", nil)
	}
	deviceResponse, err := dc.DeviceByName(context.Background(), name)
	if err != nil {
		return deviceCoreCommand, errors.NewCommonEdgeXWrapper(err)
	}

	// retrieve device profile information through Metadata DeviceProfileClient
	dpc := bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	if dpc == nil {
		return deviceCoreCommand, errors.NewCommonEdgeX(errors.KindServerError, "nil DeviceProfileClient returned", nil)
	}
	deviceProfileResponse, err := dpc.DeviceProfileByName(context.Background(), deviceResponse.Device.ProfileName)
	if err != nil {
		return deviceCoreCommand, errors.NewCommonEdgeXWrapper(err)
	}

	// Prepare the url for command
	configuration := commandContainer.ConfigurationFrom(dic.Get)
	serviceUrl := configuration.Service.Url()

	commands, err := dtos.BuildCoreCommands(deviceResponse.Device.Name, serviceUrl, deviceProfileResponse.Profile)
	if err != nil {
		return deviceCoreCommand, errors.NewCommonEdgeXWrapper(err)
	}

	deviceCoreCommand = dtos.DeviceCoreCommand{
		DeviceName:   deviceResponse.Device.Name,
		ProfileName:  deviceResponse.Device.ProfileName,
		CoreCommands: commands,
	}
	return deviceCoreCommand, nil
}
