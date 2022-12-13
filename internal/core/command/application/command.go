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

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
)

// AllCommands query commands by offset, and limit
func AllCommands(offset int, limit int, dic *di.Container) (deviceCoreCommands []dtos.DeviceCoreCommand, totalCount uint32, err errors.EdgeX) {
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

	deviceCoreCommands = make([]dtos.DeviceCoreCommand, len(multiDevicesResponse.Devices))
	for i, device := range multiDevicesResponse.Devices {
		deviceProfileResponse, err := dpc.DeviceProfileByName(context.Background(), device.ProfileName)
		if err != nil {
			return deviceCoreCommands, totalCount, errors.NewCommonEdgeXWrapper(err)
		}
		commands, err := buildCoreCommands(device.Name, serviceUrl, deviceProfileResponse.Profile)
		if err != nil {
			return nil, totalCount, errors.NewCommonEdgeXWrapper(err)
		}
		deviceCoreCommands[i] = dtos.DeviceCoreCommand{
			DeviceName:   device.Name,
			ProfileName:  device.ProfileName,
			CoreCommands: commands,
		}
	}
	return deviceCoreCommands, multiDevicesResponse.TotalCount, nil
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

	commands, err := buildCoreCommands(deviceResponse.Device.Name, serviceUrl, deviceProfileResponse.Profile)
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

func commandPath(deviceName, cmdName string) string {
	return fmt.Sprintf("%s/%s/%s/%s", common.ApiDeviceRoute, common.Name, deviceName, cmdName)
}
func buildCoreCommand(deviceName, serviceUrl, cmdName, readWrite string, parameters []dtos.CoreCommandParameter) dtos.CoreCommand {
	cmd := dtos.CoreCommand{
		Name:       cmdName,
		Url:        serviceUrl,
		Path:       commandPath(deviceName, cmdName),
		Parameters: parameters,
	}
	if strings.Contains(readWrite, common.ReadWrite_R) {
		cmd.Get = true
	}
	if strings.Contains(readWrite, common.ReadWrite_W) {
		cmd.Set = true
	}
	return cmd
}

func deviceResourcesByName(resources []dtos.DeviceResource, name string) (res dtos.DeviceResource, exists bool) {
	for _, resource := range resources {
		if resource.Name == name {
			exists = true
			res = resource
			break
		}
	}
	return res, exists
}

// coreCommandParameters creates command parameters by mapping the resourceOperation to corresponding resourceName and valueType
func coreCommandParameters(resourceOperations []dtos.ResourceOperation, resources []dtos.DeviceResource) ([]dtos.CoreCommandParameter, errors.EdgeX) {
	parameters := make([]dtos.CoreCommandParameter, len(resourceOperations))
	for i, ro := range resourceOperations {
		r, exists := deviceResourcesByName(resources, ro.DeviceResource)
		if !exists {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device command's resource %s doesn't match any deivce resource", ro.DeviceResource), nil)
		}
		parameters[i] = dtos.CoreCommandParameter{
			ResourceName: r.Name,
			ValueType:    r.Properties.ValueType,
		}
	}
	return parameters, nil
}

func buildCoreCommands(deviceName string, serviceUrl string, profile dtos.DeviceProfile) ([]dtos.CoreCommand, errors.EdgeX) {
	commandMap := make(map[string]dtos.CoreCommand)
	// Build commands from device commands
	for _, c := range profile.DeviceCommands {
		if c.IsHidden {
			continue
		}
		parameters, err := coreCommandParameters(c.ResourceOperations, profile.DeviceResources)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		commandMap[c.Name] = buildCoreCommand(deviceName, serviceUrl, c.Name, c.ReadWrite, parameters)
	}
	// Build commands from device resource
	for _, r := range profile.DeviceResources {
		if _, ok := commandMap[r.Name]; ok || r.IsHidden {
			continue
		}
		parameters := []dtos.CoreCommandParameter{
			{ResourceName: r.Name, ValueType: r.Properties.ValueType},
		}
		commandMap[r.Name] = buildCoreCommand(deviceName, serviceUrl, r.Name, r.Properties.ReadWrite, parameters)
	}
	// Convert command map to slice
	var commands []dtos.CoreCommand
	for _, cmd := range commandMap {
		commands = append(commands, cmd)
	}
	return commands, nil
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
