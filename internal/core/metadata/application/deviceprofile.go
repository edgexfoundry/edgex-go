//
// Copyright (C) 2020-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// The AddDeviceProfile function accepts the new device profile model from the controller functions
// and invokes addDeviceProfile function in the infrastructure layer
func AddDeviceProfile(d models.DeviceProfile, ctx context.Context, dic *di.Container) (id string, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	correlationId := correlation.FromContext(ctx)
	addedDeviceProfile, err := dbClient.AddDeviceProfile(d)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf(
		"DeviceProfile created on DB successfully. DeviceProfile-id: %s, Correlation-id: %s ",
		addedDeviceProfile.Id,
		correlationId,
	)

	return addedDeviceProfile.Id, nil
}

// The UpdateDeviceProfile function accepts the device profile model from the controller functions
// and invokes updateDeviceProfile function in the infrastructure layer
func UpdateDeviceProfile(d models.DeviceProfile, ctx context.Context, dic *di.Container) (err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	err = dbClient.UpdateDeviceProfile(d)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf(
		"DeviceProfile updated on DB successfully. Correlation-id: %s ",
		correlation.FromContext(ctx),
	)
	go updateDeviceProfileCallback(ctx, dic, dtos.FromDeviceProfileModelToDTO(d))
	return nil
}

// DeviceProfileByName query the device profile by name
func DeviceProfileByName(name string, ctx context.Context, dic *di.Container) (deviceProfile dtos.DeviceProfile, err errors.EdgeX) {
	if name == "" {
		return deviceProfile, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	dp, err := dbClient.DeviceProfileByName(name)
	if err != nil {
		return deviceProfile, errors.NewCommonEdgeXWrapper(err)
	}
	deviceProfile = dtos.FromDeviceProfileModelToDTO(dp)
	return deviceProfile, nil
}

// DeleteDeviceProfileByName delete the device profile by name
func DeleteDeviceProfileByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)

	// Check the associated Device and ProvisionWatcher existence
	devices, err := dbClient.DevicesByProfileName(0, 1, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if len(devices) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to delete the device profile when associated device exists", nil)
	}
	provisionWatchers, err := dbClient.ProvisionWatchersByProfileName(0, 1, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if len(provisionWatchers) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to delete the device profile when associated provisionWatcher exists", nil)
	}

	err = dbClient.DeleteDeviceProfileByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// AllDeviceProfiles query the device profiles with offset, and limit
func AllDeviceProfiles(offset int, limit int, labels []string, dic *di.Container) (deviceProfiles []dtos.DeviceProfile, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	dps, err := dbClient.AllDeviceProfiles(offset, limit, labels)
	if err == nil {
		totalCount, err = dbClient.DeviceProfileCountByLabels(labels)
	}
	if err != nil {
		return deviceProfiles, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	deviceProfiles = make([]dtos.DeviceProfile, len(dps))
	for i, dp := range dps {
		deviceProfiles[i] = dtos.FromDeviceProfileModelToDTO(dp)
	}
	return deviceProfiles, totalCount, nil
}

// DeviceProfilesByModel query the device profiles with offset, limit and model
func DeviceProfilesByModel(offset int, limit int, model string, dic *di.Container) (deviceProfiles []dtos.DeviceProfile, totalCount uint32, err errors.EdgeX) {
	if model == "" {
		return deviceProfiles, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "model is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	dps, err := dbClient.DeviceProfilesByModel(offset, limit, model)
	if err == nil {
		totalCount, err = dbClient.DeviceProfileCountByModel(model)
	}
	if err != nil {
		return deviceProfiles, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	deviceProfiles = make([]dtos.DeviceProfile, len(dps))
	for i, dp := range dps {
		deviceProfiles[i] = dtos.FromDeviceProfileModelToDTO(dp)
	}
	return deviceProfiles, totalCount, nil
}

// DeviceProfilesByManufacturer query the device profiles with offset, limit and manufacturer
func DeviceProfilesByManufacturer(offset int, limit int, manufacturer string, dic *di.Container) (deviceProfiles []dtos.DeviceProfile, totalCount uint32, err errors.EdgeX) {
	if manufacturer == "" {
		return deviceProfiles, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "manufacturer is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	dps, err := dbClient.DeviceProfilesByManufacturer(offset, limit, manufacturer)
	if err == nil {
		totalCount, err = dbClient.DeviceProfileCountByManufacturer(manufacturer)
	}
	if err != nil {
		return deviceProfiles, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	deviceProfiles = make([]dtos.DeviceProfile, len(dps))
	for i, dp := range dps {
		deviceProfiles[i] = dtos.FromDeviceProfileModelToDTO(dp)
	}
	return deviceProfiles, totalCount, nil
}

// DeviceProfilesByManufacturerAndModel query the device profiles with offset, limit, manufacturer and model
func DeviceProfilesByManufacturerAndModel(offset int, limit int, manufacturer string, model string, dic *di.Container) (deviceProfiles []dtos.DeviceProfile, totalCount uint32, err errors.EdgeX) {
	if manufacturer == "" {
		return deviceProfiles, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "manufacturer is empty", nil)
	}
	if model == "" {
		return deviceProfiles, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "model is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	dps, totalCount, err := dbClient.DeviceProfilesByManufacturerAndModel(offset, limit, manufacturer, model)
	if err != nil {
		return deviceProfiles, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	deviceProfiles = make([]dtos.DeviceProfile, len(dps))
	for i, dp := range dps {
		deviceProfiles[i] = dtos.FromDeviceProfileModelToDTO(dp)
	}
	return deviceProfiles, totalCount, nil
}

func PatchDeviceProfileBasicInfo(ctx context.Context, dto dtos.UpdateDeviceProfileBasicInfo, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	deviceProfile, err := deviceProfileByDTO(dbClient, dto)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	requests.ReplaceDeviceProfileModelBasicInfoFieldsWithDTO(&deviceProfile, dto)
	err = dbClient.UpdateDeviceProfile(deviceProfile)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf(
		"DeviceProfile basic info patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	)

	return nil
}

func DeleteDeviceResourceByName(profileName string, resourceName string, dic *di.Container) errors.EdgeX {
	if profileName == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "profile name is empty", nil)
	}
	if resourceName == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
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
	for i := range profile.DeviceResources {
		if profile.DeviceResources[i].Name == resourceName {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device resource not found", nil)
	}

	profile.DeviceResources = append(profile.DeviceResources[:index], profile.DeviceResources[index+1:]...)
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

func deviceProfileByDTO(dbClient interfaces.DBClient, dto dtos.UpdateDeviceProfileBasicInfo) (deviceProfile models.DeviceProfile, err errors.EdgeX) {
	// The ID or Name is required by DTO and the DTO also accepts empty string ID if the Name is provided
	if dto.Id != nil && *dto.Id != "" {
		deviceProfile, err = dbClient.DeviceProfileById(*dto.Id)
		if err != nil {
			return deviceProfile, errors.NewCommonEdgeXWrapper(err)
		}
	} else {
		deviceProfile, err = dbClient.DeviceProfileByName(*dto.Name)
		if err != nil {
			return deviceProfile, errors.NewCommonEdgeXWrapper(err)
		}
	}
	if dto.Name != nil && *dto.Name != deviceProfile.Name {
		return deviceProfile, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device profile name '%s' not match the exsting '%s' ", *dto.Name, deviceProfile.Name), nil)
	}
	return deviceProfile, nil
}
