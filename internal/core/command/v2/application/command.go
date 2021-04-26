//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"strings"

	commandContainer "github.com/edgexfoundry/edgex-go/internal/core/command/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	V2Container "github.com/edgexfoundry/go-mod-bootstrap/v2/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	V2Routes "github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
)

// AllCommands query commands by offset, and limit
func AllCommands(offset int, limit int, dic *di.Container) (deviceCoreCommands []dtos.DeviceCoreCommand, err errors.EdgeX) {
	// retrieve device information through Metadata DeviceClient
	dc := V2Container.MetadataDeviceClientFrom(dic.Get)
	if dc == nil {
		return deviceCoreCommands, errors.NewCommonEdgeX(errors.KindClientError, "nil MetadataDeviceClient returned", nil)
	}
	multiDevicesResponse, err := dc.AllDevices(context.Background(), nil, offset, limit)
	if err != nil {
		return deviceCoreCommands, errors.NewCommonEdgeXWrapper(err)
	}

	// retrieve device profile information through Metadata DeviceProfileClient
	dpc := V2Container.MetadataDeviceProfileClientFrom(dic.Get)
	if dpc == nil {
		return deviceCoreCommands, errors.NewCommonEdgeX(errors.KindClientError, "nil MetadataDeviceProfileClient returned", nil)
	}

	// Prepare the url for command
	configuration := commandContainer.ConfigurationFrom(dic.Get)
	serviceUrl := configuration.Service.Url()

	deviceCoreCommands = make([]dtos.DeviceCoreCommand, len(multiDevicesResponse.Devices))
	for i, device := range multiDevicesResponse.Devices {
		deviceProfileResponse, err := dpc.DeviceProfileByName(context.Background(), device.ProfileName)
		if err != nil {
			return deviceCoreCommands, errors.NewCommonEdgeXWrapper(err)
		}
		commands := buildCoreCommands(device.Name, serviceUrl, deviceProfileResponse.Profile)

		deviceCoreCommands[i] = dtos.DeviceCoreCommand{
			DeviceName:   device.Name,
			ProfileName:  device.ProfileName,
			CoreCommands: commands,
		}
	}
	return deviceCoreCommands, nil
}

// CommandsByDeviceName query coreCommands with device name
func CommandsByDeviceName(name string, dic *di.Container) (deviceCoreCommand dtos.DeviceCoreCommand, err errors.EdgeX) {
	if name == "" {
		return deviceCoreCommand, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}

	// retrieve device information through Metadata DeviceClient
	dc := V2Container.MetadataDeviceClientFrom(dic.Get)
	if dc == nil {
		return deviceCoreCommand, errors.NewCommonEdgeX(errors.KindClientError, "nil MetadataDeviceClient returned", nil)
	}
	deviceResponse, err := dc.DeviceByName(context.Background(), name)
	if err != nil {
		return deviceCoreCommand, errors.NewCommonEdgeXWrapper(err)
	}

	// retrieve device profile information through Metadata DeviceProfileClient
	dpc := V2Container.MetadataDeviceProfileClientFrom(dic.Get)
	if dpc == nil {
		return deviceCoreCommand, errors.NewCommonEdgeX(errors.KindClientError, "nil MetadataDeviceProfileClient returned", nil)
	}
	deviceProfileResponse, err := dpc.DeviceProfileByName(context.Background(), deviceResponse.Device.ProfileName)
	if err != nil {
		return deviceCoreCommand, errors.NewCommonEdgeXWrapper(err)
	}

	// Prepare the url for command
	configuration := commandContainer.ConfigurationFrom(dic.Get)
	serviceUrl := configuration.Service.Url()

	commands := buildCoreCommands(deviceResponse.Device.Name, serviceUrl, deviceProfileResponse.Profile)

	deviceCoreCommand = dtos.DeviceCoreCommand{
		DeviceName:   deviceResponse.Device.Name,
		ProfileName:  deviceResponse.Device.ProfileName,
		CoreCommands: commands,
	}
	return deviceCoreCommand, nil
}

func commandPath(deviceName, cmdName string) string {
	return fmt.Sprintf("%s/%s/%s/%s", V2Routes.ApiDeviceRoute, V2Routes.Name, deviceName, cmdName)
}
func buildCoreCommand(deviceName, serviceUrl, cmdName, readWrite string) dtos.CoreCommand {
	cmd := dtos.CoreCommand{
		Name: cmdName,
		Url:  serviceUrl,
		Path: commandPath(deviceName, cmdName),
	}
	if strings.Contains(readWrite, V2Routes.ReadWrite_R) {
		cmd.Get = true
	}
	if strings.Contains(readWrite, V2Routes.ReadWrite_W) {
		cmd.Set = true
	}
	return cmd
}

func buildCoreCommands(deviceName string, serviceUrl string, profile dtos.DeviceProfile) []dtos.CoreCommand {
	commandMap := make(map[string]dtos.CoreCommand)
	// Build commands from device commands
	for _, c := range profile.DeviceCommands {
		if c.IsHidden {
			continue
		}
		commandMap[c.Name] = buildCoreCommand(deviceName, serviceUrl, c.Name, c.ReadWrite)
	}
	// Build commands from device resource
	for _, r := range profile.DeviceResources {
		if _, ok := commandMap[r.Name]; ok || r.IsHidden {
			continue
		}
		commandMap[r.Name] = buildCoreCommand(deviceName, serviceUrl, r.Name, r.Properties.ReadWrite)
	}
	// Convert command map to slice
	var commands []dtos.CoreCommand
	for _, cmd := range commandMap {
		commands = append(commands, cmd)
	}
	return commands
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
	dc := V2Container.MetadataDeviceClientFrom(dic.Get)
	if dc == nil {
		return res, errors.NewCommonEdgeX(errors.KindClientError, "nil MetadataDeviceClient returned", nil)
	}
	deviceResponse, err := dc.DeviceByName(context.Background(), deviceName)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	// retrieve device service information through Metadata DeviceClient
	dsc := V2Container.MetadataDeviceServiceClientFrom(dic.Get)
	if dsc == nil {
		return res, errors.NewCommonEdgeX(errors.KindClientError, "nil MetadataDeviceServiceClient returned", nil)
	}
	deviceServiceResponse, err := dsc.DeviceServiceByName(context.Background(), deviceResponse.Device.ServiceName)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	// Issue command by passing the base address of device service into DeviceServiceCommandClient
	dscc := V2Container.DeviceServiceCommandClientFrom(dic.Get)
	if dscc == nil {
		return res, errors.NewCommonEdgeX(errors.KindClientError, "nil DeviceServiceCommandClient returned", nil)
	}
	res, err = dscc.GetCommand(context.Background(), deviceServiceResponse.Service.BaseAddress, deviceName, commandName, queryParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}

	return res, nil
}

// IssueSetCommandByName issues the specified set(write) command referenced by the command name to the device/sensor, also
// referenced by name.
func IssueSetCommandByName(deviceName string, commandName string, queryParams string, settings map[string]string, dic *di.Container) (response common.BaseResponse, err errors.EdgeX) {
	if deviceName == "" {
		return response, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name cannot be empty", nil)
	}

	if commandName == "" {
		return response, errors.NewCommonEdgeX(errors.KindContractInvalid, "command name cannot be empty", nil)
	}

	// retrieve device information through Metadata DeviceClient
	dc := V2Container.MetadataDeviceClientFrom(dic.Get)
	if dc == nil {
		return response, errors.NewCommonEdgeX(errors.KindClientError, "nil MetadataDeviceClient returned", nil)
	}
	deviceResponse, err := dc.DeviceByName(context.Background(), deviceName)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}

	// retrieve device service information through Metadata DeviceClient
	dsc := V2Container.MetadataDeviceServiceClientFrom(dic.Get)
	if dsc == nil {
		return response, errors.NewCommonEdgeX(errors.KindClientError, "nil MetadataDeviceServiceClient returned", nil)
	}
	deviceServiceResponse, err := dsc.DeviceServiceByName(context.Background(), deviceResponse.Device.ServiceName)
	if err != nil {
		return response, errors.NewCommonEdgeXWrapper(err)
	}

	// Issue command by passing the base address of device service into DeviceServiceCommandClient
	dscc := V2Container.DeviceServiceCommandClientFrom(dic.Get)
	if dscc == nil {
		return response, errors.NewCommonEdgeX(errors.KindClientError, "nil DeviceServiceCommandClient returned", nil)
	}
	return dscc.SetCommand(context.Background(), deviceServiceResponse.Service.BaseAddress, deviceName, commandName, queryParams, settings)
}
