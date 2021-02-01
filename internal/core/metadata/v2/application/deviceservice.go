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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// The AddDeviceService function accepts the new device service model from the controller function
// and then invokes AddDeviceService function of infrastructure layer to add new device service
func AddDeviceService(d models.DeviceService, ctx context.Context, dic *di.Container) (id string, err errors.EdgeX) {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	correlationId := correlation.FromContext(ctx)
	addedDeviceService, err := dbClient.AddDeviceService(d)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf(
		"DeviceService created on DB successfully. DeviceService ID: %s, Correlation-ID: %s ",
		addedDeviceService.Id,
		correlationId,
	)

	return addedDeviceService.Id, nil
}

// DeviceServiceByName query the device service by name
func DeviceServiceByName(name string, ctx context.Context, dic *di.Container) (deviceService dtos.DeviceService, err errors.EdgeX) {
	if name == "" {
		return deviceService, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	ds, err := dbClient.DeviceServiceByName(name)
	if err != nil {
		return deviceService, errors.NewCommonEdgeXWrapper(err)
	}
	deviceService = dtos.FromDeviceServiceModelToDTO(ds)
	return deviceService, nil
}

// PatchDeviceService executes the PATCH operation with the device service DTO to replace the old data
func PatchDeviceService(dto dtos.UpdateDeviceService, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	var deviceService models.DeviceService
	var edgeXerr errors.EdgeX
	if dto.Id != nil {
		deviceService, edgeXerr = dbClient.DeviceServiceById(*dto.Id)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	} else {
		deviceService, edgeXerr = dbClient.DeviceServiceByName(*dto.Name)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	}
	if dto.Name != nil && *dto.Name != deviceService.Name {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device service name '%s' not match the exsting '%s' ", *dto.Name, deviceService.Name), nil)
	}

	requests.ReplaceDeviceServiceModelFieldsWithDTO(&deviceService, dto)

	edgeXerr = dbClient.UpdateDeviceService(deviceService)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	lc.Debugf(
		"DeviceService patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	)
	go updateDeviceServiceCallback(ctx, dic, deviceService)
	return nil
}

// DeleteDeviceServiceByName delete the device service by name
func DeleteDeviceServiceByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)

	// Check the associated Device and ProvisionWatcher existence
	devices, err := dbClient.DevicesByServiceName(0, 1, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if len(devices) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to delete the device service when associated device exists", nil)
	}
	provisionWatchers, err := dbClient.ProvisionWatchersByServiceName(0, 1, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if len(provisionWatchers) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to delete the device service when associated provisionWatcher exists", nil)
	}

	err = dbClient.DeleteDeviceServiceByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// AllDeviceServices query the device services with labels, offset, and limit
func AllDeviceServices(offset int, limit int, labels []string, ctx context.Context, dic *di.Container) (deviceServices []dtos.DeviceService, err errors.EdgeX) {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	services, err := dbClient.AllDeviceServices(offset, limit, labels)
	if err != nil {
		return deviceServices, errors.NewCommonEdgeXWrapper(err)
	}
	deviceServices = make([]dtos.DeviceService, len(services))
	for i, ds := range services {
		dsDTO := dtos.FromDeviceServiceModelToDTO(ds)
		deviceServices[i] = dsDTO
	}
	return deviceServices, nil
}
