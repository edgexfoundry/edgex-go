//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	v2MetadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// The AddDeviceProfile function accepts the new device profile model from the controller functions
// and invokes addDeviceProfile function in the infrastructure layer
func AddDeviceProfile(d models.DeviceProfile, ctx context.Context, dic *di.Container) (id string, err errors.EdgeX) {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	correlationId := correlation.FromContext(ctx)
	addedDeviceProfile, err := dbClient.AddDeviceProfile(d)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debug(fmt.Sprintf(
		"DeviceProfile created on DB successfully. DeviceProfile-id: %s, Correlation-id: %s ",
		addedDeviceProfile.Id,
		correlationId,
	))

	return addedDeviceProfile.Id, nil
}

// The UpdateDeviceProfile function accepts the device profile model from the controller functions
// and invokes updateDeviceProfile function in the infrastructure layer
func UpdateDeviceProfile(d models.DeviceProfile, ctx context.Context, dic *di.Container) (err errors.EdgeX) {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	err = dbClient.UpdateDeviceProfile(d)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debug(fmt.Sprintf(
		"DeviceProfile updated on DB successfully. Correlation-id: %s ",
		correlation.FromContext(ctx),
	))
	go updateDeviceProfileCallback(ctx, dic, dtos.FromDeviceProfileModelToDTO(d))
	return nil
}

// DeviceProfileByName query the device profile by name
func DeviceProfileByName(name string, ctx context.Context, dic *di.Container) (deviceProfile dtos.DeviceProfile, err errors.EdgeX) {
	if name == "" {
		return deviceProfile, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
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
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)

	// Check the associated Device and ProvisionWatcher existence
	devices, err := dbClient.DevicesByProfileName(0, 1, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if len(devices) > 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to delete the device profile when associated device exists", nil)
	}
	provisionWatchers, err := dbClient.ProvisionWatchersByProfileName(0, 1, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if len(provisionWatchers) > 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to delete the device profile when associated provisionWatcher exists", nil)
	}

	err = dbClient.DeleteDeviceProfileByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// AllDeviceProfiles query the device profiles with offset, and limit
func AllDeviceProfiles(offset int, limit int, labels []string, dic *di.Container) (deviceProfiles []dtos.DeviceProfile, err errors.EdgeX) {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	dps, err := dbClient.AllDeviceProfiles(offset, limit, labels)
	if err != nil {
		return deviceProfiles, errors.NewCommonEdgeXWrapper(err)
	}
	deviceProfiles = make([]dtos.DeviceProfile, len(dps))
	for i, dp := range dps {
		deviceProfiles[i] = dtos.FromDeviceProfileModelToDTO(dp)
	}
	return deviceProfiles, nil
}

// DeviceProfilesByModel query the device profiles with offset, limit and model
func DeviceProfilesByModel(offset int, limit int, model string, dic *di.Container) (deviceProfiles []dtos.DeviceProfile, err errors.EdgeX) {
	if model == "" {
		return deviceProfiles, errors.NewCommonEdgeX(errors.KindContractInvalid, "model is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	dps, err := dbClient.DeviceProfilesByModel(offset, limit, model)
	if err != nil {
		return deviceProfiles, errors.NewCommonEdgeXWrapper(err)
	}
	deviceProfiles = make([]dtos.DeviceProfile, len(dps))
	for i, dp := range dps {
		deviceProfiles[i] = dtos.FromDeviceProfileModelToDTO(dp)
	}
	return deviceProfiles, nil
}

// DeviceProfilesByManufacturer query the device profiles with offset, limit and manufacturer
func DeviceProfilesByManufacturer(offset int, limit int, manufacturer string, dic *di.Container) (deviceProfiles []dtos.DeviceProfile, err errors.EdgeX) {
	if manufacturer == "" {
		return deviceProfiles, errors.NewCommonEdgeX(errors.KindContractInvalid, "manufacturer is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	dps, err := dbClient.DeviceProfilesByManufacturer(offset, limit, manufacturer)
	if err != nil {
		return deviceProfiles, errors.NewCommonEdgeXWrapper(err)
	}
	deviceProfiles = make([]dtos.DeviceProfile, len(dps))
	for i, dp := range dps {
		deviceProfiles[i] = dtos.FromDeviceProfileModelToDTO(dp)
	}
	return deviceProfiles, nil
}

// DeviceProfilesByManufacturerAndModel query the device profiles with offset, limit, manufacturer and model
func DeviceProfilesByManufacturerAndModel(offset int, limit int, manufacturer string, model string, dic *di.Container) (deviceProfiles []dtos.DeviceProfile, err errors.EdgeX) {
	if manufacturer == "" {
		return deviceProfiles, errors.NewCommonEdgeX(errors.KindContractInvalid, "manufacturer is empty", nil)
	}
	if model == "" {
		return deviceProfiles, errors.NewCommonEdgeX(errors.KindContractInvalid, "model is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	dps, err := dbClient.DeviceProfilesByManufacturerAndModel(offset, limit, manufacturer, model)
	if err != nil {
		return deviceProfiles, errors.NewCommonEdgeXWrapper(err)
	}
	deviceProfiles = make([]dtos.DeviceProfile, len(dps))
	for i, dp := range dps {
		deviceProfiles[i] = dtos.FromDeviceProfileModelToDTO(dp)
	}
	return deviceProfiles, nil
}
