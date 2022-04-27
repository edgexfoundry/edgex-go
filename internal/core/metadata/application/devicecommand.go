//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// The AddDeviceProfileDeviceCommand function accepts the device profile name and device command model from the controller functions
// and invokes updateDeviceProfile function in the infrastructure layer
func AddDeviceProfileDeviceCommand(profileName string, deviceCommand models.DeviceCommand, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	profile, err := dbClient.DeviceProfileByName(profileName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	profile.DeviceCommands = append(profile.DeviceCommands, deviceCommand)

	profileDTO := dtos.FromDeviceProfileModelToDTO(profile)
	validateErr := profileDTO.Validate()
	if validateErr != nil {
		return errors.NewCommonEdgeXWrapper(validateErr)
	}

	err = dbClient.UpdateDeviceProfile(profile)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("DeviceProfile deviceCommands added on DB successfully. Correlation-id: %s ", correlation.FromContext(ctx))

	return nil
}

func PatchDeviceProfileDeviceCommand(profileName string, dto dtos.UpdateDeviceCommand, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	profile, err := dbClient.DeviceProfileByName(profileName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// Find matched deviceCommand
	index := -1
	for i := range profile.DeviceCommands {
		if profile.DeviceCommands[i].Name == *dto.Name {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device command not found", nil)
	}

	requests.ReplaceDeviceCommandModelFieldsWithDTO(&profile.DeviceCommands[index], dto)

	err = dbClient.UpdateDeviceProfile(profile)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("DeviceProfile deviceCommands patched on DB successfully. Correlation-id: %s ", correlation.FromContext(ctx))

	return nil
}

func DeleteDeviceCommandByName(profileName string, commandName string, dic *di.Container) errors.EdgeX {
	if profileName == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "profile name is empty", nil)
	}
	if commandName == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "command name is empty", nil)
	}

	strictProfileChanges := container.ConfigurationFrom(dic.Get).Writable.ProfileChange.StrictDeviceProfileChanges
	if strictProfileChanges {
		return errors.NewCommonEdgeX(errors.KindServiceLocked, "profile change is not allowed when StrictDeviceProfileChanges config is enabled", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	// Check the associated Device existence
	devices, err := dbClient.DevicesByProfileName(0, 1, profileName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	} else if len(devices) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to update the device profile when associated device exists", nil)
	}

	profile, err := dbClient.DeviceProfileByName(profileName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	index := -1
	for i := range profile.DeviceCommands {
		if profile.DeviceCommands[i].Name == commandName {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device command not found", nil)
	}

	profile.DeviceCommands = append(profile.DeviceCommands[:index], profile.DeviceCommands[index+1:]...)
	profileDTO := dtos.FromDeviceProfileModelToDTO(profile)
	e := (&profileDTO).Validate()
	if e != nil {
		return errors.NewCommonEdgeXWrapper(e)
	}

	err = dbClient.UpdateDeviceProfile(profile)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}
