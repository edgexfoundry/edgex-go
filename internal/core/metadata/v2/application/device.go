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

	"github.com/google/uuid"
)

// The AddDevice function accepts the new device model from the controller function
// and then invokes AddDevice function of infrastructure layer to add new device
func AddDevice(d models.Device, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	exists, edgeXerr := dbClient.DeviceServiceNameExists(d.ServiceName)
	if edgeXerr != nil {
		return id, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if !exists {
		return id, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device service '%s' does not exists", d.ServiceName), nil)
	}
	exists, edgeXerr = dbClient.DeviceProfileNameExists(d.ProfileName)
	if edgeXerr != nil {
		return id, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if !exists {
		return id, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile '%s' does not exists", d.ProfileName), nil)
	}

	addedDevice, err := dbClient.AddDevice(d)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debug(fmt.Sprintf(
		"Device created on DB successfully. Device ID: %s, Correlation-ID: %s ",
		addedDevice.Id,
		correlation.FromContext(ctx),
	))

	return addedDevice.Id, nil
}

// DeleteDeviceById deletes the device by Id
func DeleteDeviceById(id string, dic *di.Container) errors.EdgeX {
	if id == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "id is empty", nil)
	}
	_, err := uuid.Parse(id)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindInvalidId, "fail to parse id as an UUID", err)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	err = dbClient.DeleteDeviceById(id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// DeleteDeviceByName deletes the device by name
func DeleteDeviceByName(name string, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	err := dbClient.DeleteDeviceByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// AllDeviceByServiceName query devices with offset, limit and name
func AllDeviceByServiceName(offset int, limit int, name string, ctx context.Context, dic *di.Container) (devices []dtos.Device, err errors.EdgeX) {
	if name == "" {
		return devices, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	deviceModels, err := dbClient.AllDeviceByServiceName(offset, limit, name)
	if err != nil {
		return devices, errors.NewCommonEdgeXWrapper(err)
	}
	devices = make([]dtos.Device, len(deviceModels))
	for i, d := range deviceModels {
		devices[i] = dtos.FromDeviceModelToDTO(d)
	}
	return devices, nil
}

// DeviceIdExists checks the device existence by id
func DeviceIdExists(id string, dic *di.Container) (exists bool, edgeXerr errors.EdgeX) {
	if id == "" {
		return exists, errors.NewCommonEdgeX(errors.KindContractInvalid, "id is empty", nil)
	}
	_, err := uuid.Parse(id)
	if err != nil {
		return exists, errors.NewCommonEdgeX(errors.KindInvalidId, "fail to parse id as an UUID", err)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	exists, edgeXerr = dbClient.DeviceIdExists(id)
	if edgeXerr != nil {
		return exists, errors.NewCommonEdgeXWrapper(err)
	}
	return exists, nil
}

// DeviceNameExists checks the device existence by name
func DeviceNameExists(name string, dic *di.Container) (exists bool, err errors.EdgeX) {
	if name == "" {
		return exists, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	exists, err = dbClient.DeviceNameExists(name)
	if err != nil {
		return exists, errors.NewCommonEdgeXWrapper(err)
	}
	return exists, nil
}

// PatchDevice executes the PATCH operation with the device DTO to replace the old data
func PatchDevice(dto dtos.UpdateDevice, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	var device models.Device
	var edgeXerr errors.EdgeX
	if dto.Id != nil {
		if *dto.Id == "" {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "id is empty", nil)
		}
		_, err := uuid.Parse(*dto.Id)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindInvalidId, "fail to parse id as an UUID", err)
		}
		device, edgeXerr = dbClient.DeviceById(*dto.Id)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	} else {
		if *dto.Name == "" {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
		}
		device, edgeXerr = dbClient.DeviceByName(*dto.Name)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	}
	if dto.Name != nil && *dto.Name != device.Name {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device name '%s' not match the exsting '%s' ", *dto.Name, device.Name), nil)
	}

	requests.ReplaceDeviceModelFieldsWithDTO(&device, dto)

	edgeXerr = dbClient.DeleteDeviceById(device.Id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	_, edgeXerr = dbClient.AddDevice(device)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	lc.Debug(fmt.Sprintf(
		"Device patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	))

	return nil
}

// AllDevices query the devices with offset, limit, and labels
func AllDevices(offset int, limit int, labels []string, dic *di.Container) (devices []dtos.Device, err errors.EdgeX) {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	dps, err := dbClient.AllDevices(offset, limit, labels)
	if err != nil {
		return devices, errors.NewCommonEdgeXWrapper(err)
	}
	devices = make([]dtos.Device, len(dps))
	for i, dp := range dps {
		devices[i] = dtos.FromDeviceModelToDTO(dp)
	}
	return devices, nil
}
