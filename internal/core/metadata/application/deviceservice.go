//
// Copyright (C) 2020-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// The AddDeviceService function accepts the new device service model from the controller function
// and then invokes AddDeviceService function of infrastructure layer to add new device service
func AddDeviceService(d models.DeviceService, ctx context.Context, dic *di.Container) (id string, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

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
	dbClient := container.DBClientFrom(dic.Get)
	ds, err := dbClient.DeviceServiceByName(name)
	if err != nil {
		return deviceService, errors.NewCommonEdgeXWrapper(err)
	}
	deviceService = dtos.FromDeviceServiceModelToDTO(ds)
	return deviceService, nil
}

// PatchDeviceService executes the PATCH operation with the device service DTO to replace the old data
func PatchDeviceService(dto dtos.UpdateDeviceService, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	deviceService, err := deviceServiceByDTO(dbClient, dto)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	requests.ReplaceDeviceServiceModelFieldsWithDTO(&deviceService, dto)

	err = dbClient.UpdateDeviceService(deviceService)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf(
		"DeviceService patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	)
	// Don't invoke callback if only lastConnected field is updated
	if dto.LastConnected != nil && utils.OnlyOneFieldUpdated("LastConnected", dto) {
		return nil
	}
	go updateDeviceServiceCallback(ctx, dic, deviceService)
	return nil
}

func deviceServiceByDTO(dbClient interfaces.DBClient, dto dtos.UpdateDeviceService) (deviceService models.DeviceService, err errors.EdgeX) {
	// The ID or Name is required by DTO and the DTO also accepts empty string ID if the Name is provided
	if dto.Id != nil && *dto.Id != "" {
		deviceService, err = dbClient.DeviceServiceById(*dto.Id)
		if err != nil {
			return deviceService, errors.NewCommonEdgeXWrapper(err)
		}
	} else {
		deviceService, err = dbClient.DeviceServiceByName(*dto.Name)
		if err != nil {
			return deviceService, errors.NewCommonEdgeXWrapper(err)
		}
	}
	if dto.Name != nil && *dto.Name != deviceService.Name {
		return deviceService, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device service name '%s' not match the exsting '%s' ", *dto.Name, deviceService.Name), nil)
	}
	return deviceService, nil
}

// DeleteDeviceServiceByName delete the device service by name
func DeleteDeviceServiceByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	err := dbClient.DeleteDeviceServiceByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// AllDeviceServices query the device services with labels, offset, and limit
func AllDeviceServices(offset int, limit int, labels []string, ctx context.Context, dic *di.Container) (deviceServices []dtos.DeviceService, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	services, err := dbClient.AllDeviceServices(offset, limit, labels)
	if err == nil {
		totalCount, err = dbClient.DeviceServiceCountByLabels(labels)
	}
	if err != nil {
		return deviceServices, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	deviceServices = make([]dtos.DeviceService, len(services))
	for i, s := range services {
		deviceServices[i] = dtos.FromDeviceServiceModelToDTO(s)
	}
	return deviceServices, totalCount, nil
}
