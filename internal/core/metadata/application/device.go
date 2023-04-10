/********************************************************************************
 *  Copyright (C) 2020-2023 IOTech Ltd
 *  Copyright 2023 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package application

import (
	"context"
	goErrors "errors"
	"fmt"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
)

// The AddDevice function accepts the new device model from the controller function
// and then invokes AddDevice function of infrastructure layer to add new device
func AddDevice(d models.Device, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	// Check the existence of device service before device validation
	exists, edgeXerr := dbClient.DeviceServiceNameExists(d.ServiceName)
	if edgeXerr != nil {
		return id, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("device service '%s' existence check failed", d.ServiceName), edgeXerr)
	} else if !exists {
		return id, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device service '%s' does not exists", d.ServiceName), nil)
	}

	err := validateDeviceCallback(dtos.FromDeviceModelToDTO(d), dic)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	addedDevice, err := dbClient.AddDevice(d)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf(
		"Device created on DB successfully. Device ID: %s, Correlation-ID: %s ",
		addedDevice.Id,
		correlation.FromContext(ctx),
	)

	deviceDTO := dtos.FromDeviceModelToDTO(addedDevice)
	go publishSystemEvent(common.DeviceSystemEventType, common.SystemEventActionAdd, d.ServiceName, deviceDTO, ctx, dic)

	return addedDevice.Id, nil
}

// DeleteDeviceByName deletes the device by name
func DeleteDeviceByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	device, err := dbClient.DeviceByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = dbClient.DeleteDeviceByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	deviceDTO := dtos.FromDeviceModelToDTO(device)
	go publishSystemEvent(common.DeviceSystemEventType, common.SystemEventActionDelete, device.ServiceName, deviceDTO, ctx, dic)

	return nil
}

// DevicesByServiceName query devices with offset, limit and name
func DevicesByServiceName(offset int, limit int, name string, ctx context.Context, dic *di.Container) (devices []dtos.Device, totalCount uint32, err errors.EdgeX) {
	if name == "" {
		return devices, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	deviceModels, err := dbClient.DevicesByServiceName(offset, limit, name)
	if err == nil {
		totalCount, err = dbClient.DeviceCountByServiceName(name)
	}
	if err != nil {
		return devices, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	devices = make([]dtos.Device, len(deviceModels))
	for i, d := range deviceModels {
		devices[i] = dtos.FromDeviceModelToDTO(d)
	}
	return devices, totalCount, nil
}

// DeviceNameExists checks the device existence by name
func DeviceNameExists(name string, dic *di.Container) (exists bool, err errors.EdgeX) {
	if name == "" {
		return exists, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	exists, err = dbClient.DeviceNameExists(name)
	if err != nil {
		return exists, errors.NewCommonEdgeXWrapper(err)
	}
	return exists, nil
}

// PatchDevice executes the PATCH operation with the device DTO to replace the old data
func PatchDevice(dto dtos.UpdateDevice, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	// Check the existence of device service before device validation
	if dto.ServiceName != nil {
		exists, edgeXerr := dbClient.DeviceServiceNameExists(*dto.ServiceName)
		if edgeXerr != nil {
			return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("device service '%s' existence check failed", *dto.ServiceName), edgeXerr)
		} else if !exists {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device service '%s' does not exists", *dto.ServiceName), nil)
		}
	}

	device, err := deviceByDTO(dbClient, dto)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// Old service name is used for invoking callback
	var oldServiceName string
	if dto.ServiceName != nil && *dto.ServiceName != device.ServiceName {
		oldServiceName = device.ServiceName
	}

	requests.ReplaceDeviceModelFieldsWithDTO(&device, dto)

	deviceDTO := dtos.FromDeviceModelToDTO(device)
	err = validateDeviceCallback(deviceDTO, dic)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	err = dbClient.UpdateDevice(device)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf(
		"Device patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	)

	if oldServiceName != "" {
		go publishSystemEvent(common.DeviceSystemEventType, common.SystemEventActionUpdate, oldServiceName, deviceDTO, ctx, dic)
	}

	go publishSystemEvent(common.DeviceSystemEventType, common.SystemEventActionUpdate, device.ServiceName, deviceDTO, ctx, dic)

	return nil
}

func deviceByDTO(dbClient interfaces.DBClient, dto dtos.UpdateDevice) (device models.Device, edgeXerr errors.EdgeX) {
	// The ID or Name is required by DTO and the DTO also accepts empty string ID if the Name is provided
	if dto.Id != nil && *dto.Id != "" {
		device, edgeXerr = dbClient.DeviceById(*dto.Id)
		if edgeXerr != nil {
			return device, errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	} else {
		device, edgeXerr = dbClient.DeviceByName(*dto.Name)
		if edgeXerr != nil {
			return device, errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	}
	if dto.Name != nil && *dto.Name != device.Name {
		return device, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device name '%s' not match the exsting '%s' ", *dto.Name, device.Name), nil)
	}
	return device, nil
}

// AllDevices query the devices with offset, limit, and labels
func AllDevices(offset int, limit int, labels []string, dic *di.Container) (devices []dtos.Device, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	deviceModels, err := dbClient.AllDevices(offset, limit, labels)
	if err == nil {
		totalCount, err = dbClient.DeviceCountByLabels(labels)
	}
	if err != nil {
		return devices, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	devices = make([]dtos.Device, len(deviceModels))
	for i, d := range deviceModels {
		devices[i] = dtos.FromDeviceModelToDTO(d)
	}
	return devices, totalCount, nil
}

// DeviceByName query the device by name
func DeviceByName(name string, dic *di.Container) (device dtos.Device, err errors.EdgeX) {
	if name == "" {
		return device, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	d, err := dbClient.DeviceByName(name)
	if err != nil {
		return device, errors.NewCommonEdgeXWrapper(err)
	}
	device = dtos.FromDeviceModelToDTO(d)
	return device, nil
}

// DevicesByProfileName query the devices with offset, limit, and profile name
func DevicesByProfileName(offset int, limit int, profileName string, dic *di.Container) (devices []dtos.Device, totalCount uint32, err errors.EdgeX) {
	if profileName == "" {
		return devices, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "profileName is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	deviceModels, err := dbClient.DevicesByProfileName(offset, limit, profileName)
	if err == nil {
		totalCount, err = dbClient.DeviceCountByProfileName(profileName)
	}
	if err != nil {
		return devices, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	devices = make([]dtos.Device, len(deviceModels))
	for i, d := range deviceModels {
		devices[i] = dtos.FromDeviceModelToDTO(d)
	}
	return devices, totalCount, nil
}

var noMessagingClientError = goErrors.New("MessageBus Client not available. Please update RequireMessageBus and MessageBus configuration to enable sending System Events via the EdgeX MessageBus")
