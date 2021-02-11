//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	commandContainer "github.com/edgexfoundry/edgex-go/internal/core/command/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	V2Container "github.com/edgexfoundry/go-mod-bootstrap/v2/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	V2Routes "github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
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
		commands := make([]dtos.CoreCommand, len(deviceProfileResponse.Profile.CoreCommands))
		for index, c := range deviceProfileResponse.Profile.CoreCommands {
			commands[index] = dtos.CoreCommand{
				Name: c.Name,
				Get:  c.Get,
				Set:  c.Set,
				Url:  serviceUrl,
				Path: fmt.Sprintf("%s/%s/%s/%s/%s", V2Routes.ApiDeviceRoute, V2Routes.Name, device.Name, V2Routes.Command, c.Name),
			}
		}
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

	commands := make([]dtos.CoreCommand, len(deviceProfileResponse.Profile.CoreCommands))
	for i, c := range deviceProfileResponse.Profile.CoreCommands {
		commands[i] = dtos.CoreCommand{
			Name: c.Name,
			Get:  c.Get,
			Set:  c.Set,
			Url:  serviceUrl,
			Path: fmt.Sprintf("%s/%s/%s/%s/%s", V2Routes.ApiDeviceRoute, V2Routes.Name, deviceResponse.Device.Name, V2Routes.Command, c.Name),
		}
	}
	deviceCoreCommand = dtos.DeviceCoreCommand{
		DeviceName:   deviceResponse.Device.Name,
		ProfileName:  deviceResponse.Device.ProfileName,
		CoreCommands: commands,
	}
	return deviceCoreCommand, nil
}

// IssueGetCommandByName issues the specified get(read) command referenced by the command name to the device/sensor, also
// referenced by name.
func IssueGetCommandByName(deviceName string, commandName string, queryParams string, dic *di.Container) (event dtos.Event, err errors.EdgeX) {
	if deviceName == "" {
		return event, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name cannot be empty", nil)
	}

	if commandName == "" {
		return event, errors.NewCommonEdgeX(errors.KindContractInvalid, "command name cannot be empty", nil)
	}

	// retrieve device information through Metadata DeviceClient
	dc := V2Container.MetadataDeviceClientFrom(dic.Get)
	if dc == nil {
		return event, errors.NewCommonEdgeX(errors.KindClientError, "nil MetadataDeviceClient returned", nil)
	}
	deviceResponse, err := dc.DeviceByName(context.Background(), deviceName)
	if err != nil {
		return event, errors.NewCommonEdgeXWrapper(err)
	}

	// retrieve device service information through Metadata DeviceClient
	dsc := V2Container.MetadataDeviceServiceClientFrom(dic.Get)
	if dsc == nil {
		return event, errors.NewCommonEdgeX(errors.KindClientError, "nil MetadataDeviceServiceClient returned", nil)
	}
	deviceServiceResponse, err := dsc.DeviceServiceByName(context.Background(), deviceResponse.Device.ServiceName)
	if err != nil {
		return event, errors.NewCommonEdgeXWrapper(err)
	}

	// Issue command by passing the base address of device service into DeviceServiceCommandClient
	dscc := V2Container.DeviceServiceCommandClientFrom(dic.Get)
	if dscc == nil {
		return event, errors.NewCommonEdgeX(errors.KindClientError, "nil DeviceServiceCommandClient returned", nil)
	}
	eventResponse, err := dscc.GetCommand(context.Background(), deviceServiceResponse.Service.BaseAddress, deviceName, commandName, queryParams)
	if err != nil {
		return event, errors.NewCommonEdgeXWrapper(err)
	}

	return eventResponse.Event, nil
}
