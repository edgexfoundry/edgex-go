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

	err = deviceProfileUoMValidation(d, dic)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

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

	err = deviceProfileUoMValidation(d, dic)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

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
	strictProfileDeletes := container.ConfigurationFrom(dic.Get).Writable.ProfileChange.StrictDeviceProfileDeletes
	if strictProfileDeletes {
		return errors.NewCommonEdgeX(errors.KindServiceLocked, "profile deletion is not allowed when StrictDeviceProfileDeletes config is enabled", nil)
	}
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	err := dbClient.DeleteDeviceProfileByName(name)
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

func deviceProfileUoMValidation(p models.DeviceProfile, dic *di.Container) errors.EdgeX {
	if container.ConfigurationFrom(dic.Get).Writable.UoM.Validation {
		uom := container.UnitsOfMeasureFrom(dic.Get)
		for _, dr := range p.DeviceResources {
			if ok := uom.Validate(dr.Properties.Units); !ok {
				return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("DeviceResource %s units %s is invalid", dr.Name, dr.Properties.Units), nil)
			}
		}
	}

	return nil
}
