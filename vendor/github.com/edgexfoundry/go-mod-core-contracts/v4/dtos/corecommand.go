//
// Copyright (C) 2021-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type DeviceCoreCommand struct {
	DeviceName   string        `json:"deviceName" validate:"required,edgex-dto-none-empty-string"`
	ProfileName  string        `json:"profileName" validate:"required,edgex-dto-none-empty-string"`
	CoreCommands []CoreCommand `json:"coreCommands,omitempty" validate:"dive"`
}

type CoreCommand struct {
	Name       string                 `json:"name" validate:"required,edgex-dto-none-empty-string"`
	Get        bool                   `json:"get,omitempty" validate:"required_without=Set"`
	Set        bool                   `json:"set,omitempty" validate:"required_without=Get"`
	Path       string                 `json:"path,omitempty"`
	Url        string                 `json:"url,omitempty"`
	Parameters []CoreCommandParameter `json:"parameters,omitempty"`
}

type CoreCommandParameter struct {
	ResourceName string `json:"resourceName"`
	ValueType    string `json:"valueType"`
}

func BuildCoreCommands(deviceName string, serviceUrl string, profile DeviceProfile) ([]CoreCommand, errors.EdgeX) {
	commandMap := make(map[string]CoreCommand)
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
		parameters := []CoreCommandParameter{
			{ResourceName: r.Name, ValueType: r.Properties.ValueType},
		}
		commandMap[r.Name] = buildCoreCommand(deviceName, serviceUrl, r.Name, r.Properties.ReadWrite, parameters)
	}
	// Convert command map to slice
	var commands []CoreCommand
	for _, cmd := range commandMap {
		commands = append(commands, cmd)
	}
	return commands, nil
}

// coreCommandParameters creates command parameters by mapping the resourceOperation to corresponding resourceName and valueType
func coreCommandParameters(resourceOperations []ResourceOperation, resources []DeviceResource) ([]CoreCommandParameter, errors.EdgeX) {
	parameters := make([]CoreCommandParameter, len(resourceOperations))
	for i, ro := range resourceOperations {
		r, exists := deviceResourcesByName(resources, ro.DeviceResource)
		if !exists {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device command's resource %s doesn't match any deivce resource", ro.DeviceResource), nil)
		}
		parameters[i] = CoreCommandParameter{
			ResourceName: r.Name,
			ValueType:    r.Properties.ValueType,
		}
	}
	return parameters, nil
}

func commandPath(deviceName, cmdName string) string {
	return fmt.Sprintf("%s/%s/%s/%s", common.ApiDeviceRoute, common.Name, deviceName, cmdName)
}

func buildCoreCommand(deviceName, serviceUrl, cmdName, readWrite string, parameters []CoreCommandParameter) CoreCommand {
	cmd := CoreCommand{
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

func deviceResourcesByName(resources []DeviceResource, name string) (res DeviceResource, exists bool) {
	for _, resource := range resources {
		if resource.Name == name {
			exists = true
			res = resource
			break
		}
	}
	return res, exists
}
