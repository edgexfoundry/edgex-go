//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	v2MetadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
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

	lc.Debug(fmt.Sprintf(
		"DeviceService created on DB successfully. DeviceService ID: %s, Correlation-ID: %s ",
		addedDeviceService.Id,
		correlationId,
	))

	return addedDeviceService.Id, nil
}

// GetDeviceServiceByName query the device service by name
func GetDeviceServiceByName(name string, ctx context.Context, dic *di.Container) (deviceService dtos.DeviceService, err errors.EdgeX) {
	if name == "" {
		return deviceService, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	ds, err := dbClient.GetDeviceServiceByName(name)
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
		deviceService, edgeXerr = dbClient.GetDeviceServiceById(*dto.Id)
		if edgeXerr != nil && dto.Name == nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		} else if edgeXerr == nil && dto.Name != nil && *dto.Name != deviceService.Name {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device service name '%s' not match the exsting '%s' ", *dto.Name, deviceService.Name), nil)
		}
	}
	// Since the DTO requires one of Id or Name, so we don't need to check whether the name is nil
	if dto.Id == nil || edgeXerr != nil {
		deviceService, edgeXerr = dbClient.GetDeviceServiceByName(*dto.Name)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	}

	requests.ReplaceDeviceServiceModelFieldsWithDTO(&deviceService, dto)

	edgeXerr = dbClient.DeleteDeviceServiceById(deviceService.Id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	_, edgeXerr = dbClient.AddDeviceService(deviceService)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	lc.Debug(fmt.Sprintf(
		"DeviceService patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	))

	return nil
}

// DeleteDeviceServiceById delete the device service by Id
func DeleteDeviceServiceById(id string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if id == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "id is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	err := dbClient.DeleteDeviceServiceById(id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// DeleteDeviceServiceByName delete the device service by name
func DeleteDeviceServiceByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	err := dbClient.DeleteDeviceServiceByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}
