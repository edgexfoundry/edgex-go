/********************************************************************************
 *  Copyright (C) 2020-2025 IOTech Ltd
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
	"regexp"
	"slices"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
)

// the suggested minimum duration for auto event interval
const minAutoEventInterval = 1 * time.Millisecond

// The AddDevice function accepts the new device model from the controller function
// and then invokes AddDevice function of infrastructure layer to add new device
func AddDevice(d models.Device, ctx context.Context, dic *di.Container, bypassValidation bool, force bool) (id string, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	// Check the existence of device service before device validation
	exists, edgeXerr := dbClient.DeviceServiceNameExists(d.ServiceName)
	if edgeXerr != nil {
		return id, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("device service '%s' existence check failed", d.ServiceName), edgeXerr)
	} else if !exists {
		return id, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device service '%s' does not exists", d.ServiceName), nil)
	}

	err := validateParentProfileAndAutoEvent(dic, d)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	// check if device name already exists
	exists, err = dbClient.DeviceNameExists(d.Name)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	if exists {
		if force {
			// invoke updateDevice if force flag is enabled
			return updateDevice(d, ctx, dic)
		} else {
			return "", errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("device name %s already exists", d.Name), nil)
		}
	}

	if config.Writable.MaxDevices > 0 || config.Writable.MaxResources > 0 {
		if err = checkCapacityWithNewDevice(d, dic); err != nil {
			return "", errors.NewCommonEdgeXWrapper(err)
		}
	}

	// Execute the Device Service Validation when both bypassValidation/force values are false by default
	// Skip the Device Service Validation if either bypassValidation or force is true
	if !bypassValidation && !force { // Per De Morgan's law, !(A || B) is equivalent to !A && !B
		err = validateDeviceCallback(dtos.FromDeviceModelToDTO(d), dic)
		if err != nil {
			return "", errors.NewCommonEdgeXWrapper(err)
		}
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

	// If device is successfully created, check each AutoEvent interval value and display a warning if it's smaller than the suggested 10ms value
	for _, autoEvent := range d.AutoEvents {
		utils.CheckMinInterval(autoEvent.Interval, minAutoEventInterval, lc)
	}

	deviceDTO := dtos.FromDeviceModelToDTO(addedDevice)
	go publishSystemEvent(common.DeviceSystemEventType, common.SystemEventActionAdd, d.ServiceName, deviceDTO, ctx, dic)

	return addedDevice.Id, nil
}

// updateDevice accepts the updated device model from AddDevice function if force flag is enabled
// and then invokes UpdateDevice function of infrastructure layer to update the existing device
// the "update device" system events will be published to the msg bus at last
func updateDevice(d models.Device, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	oldDevice, err := dbClient.DeviceByName(d.Name)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	// set the id and created fields from the old device
	if d.Id == "" {
		d.Id = oldDevice.Id
	}
	if d.Created == 0 {
		d.Created = oldDevice.Created
	}

	// Old service name is used for invoking callback
	var oldServiceName string
	if d.ServiceName != "" && d.ServiceName != oldDevice.ServiceName {
		oldServiceName = oldDevice.ServiceName
	}

	if container.ConfigurationFrom(dic.Get).Writable.MaxResources > 0 {
		if err = checkResourceCapacityByExistingAndNewProfile(oldDevice.ProfileName, d.ProfileName, dic); err != nil {
			return "", errors.NewCommonEdgeXWrapper(err)
		}
	}

	err = updateDeviceInDB(d, oldServiceName, ctx, dic)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	return d.Id, nil
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
	childcount, _, err := dbClient.DeviceTree(name, 1, 0, 1, nil)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if childcount != 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "cannot delete device with children", nil)
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
func DevicesByServiceName(offset int, limit int, name string, ctx context.Context, dic *di.Container) (devices []dtos.Device, totalCount int64, err errors.EdgeX) {
	if name == "" {
		return devices, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)

	totalCount, err = dbClient.DeviceCountByServiceName(name)
	if err != nil {
		return devices, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.Device{}, totalCount, err
	}

	deviceModels, err := dbClient.DevicesByServiceName(offset, limit, name)
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
func PatchDevice(dto dtos.UpdateDevice, ctx context.Context, dic *di.Container, bypassValidation bool) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)

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

	if container.ConfigurationFrom(dic.Get).Writable.MaxResources > 0 {
		if err = checkResourceCapacityByExistingAndNewProfile(device.ProfileName, *dto.ProfileName, dic); err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	requests.ReplaceDeviceModelFieldsWithDTO(&device, dto)

	err = validateParentProfileAndAutoEvent(dic, device)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	deviceDTO := dtos.FromDeviceModelToDTO(device)

	// Execute the Device Service Validation when bypassValidation is false by default
	// Skip the Device Service Validation if bypassValidation is true
	if !bypassValidation {
		err = validateDeviceCallback(deviceDTO, dic)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	return updateDeviceInDB(device, oldServiceName, ctx, dic)
}

// updateDeviceInDB calls the UpdateDevice method from the infrastructure layer and validate the device auto events
// and publish the "update device" system event at last
func updateDeviceInDB(device models.Device, oldServiceName string, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	err := dbClient.UpdateDevice(device)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// If device is successfully updated, check each AutoEvent interval value and display a warning if it's smaller than the suggested 10ms value
	for _, autoEvent := range device.AutoEvents {
		utils.CheckMinInterval(autoEvent.Interval, minAutoEventInterval, lc)
	}

	lc.Debugf(
		"Device updated on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	)

	deviceDTO := dtos.FromDeviceModelToDTO(device)
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
func AllDevices(offset int, limit int, labels []string, parent string, maxLevels int, dic *di.Container) (devices []dtos.Device, totalCount int64, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	var deviceModels []models.Device
	if parent != "" {
		totalCount, deviceModels, err = dbClient.DeviceTree(parent, maxLevels, offset, limit, labels)
		if err != nil {
			return devices, totalCount, errors.NewCommonEdgeXWrapper(err)
		}
	} else {
		totalCount, err = dbClient.DeviceCountByLabels(labels)
		if err != nil {
			return devices, totalCount, errors.NewCommonEdgeXWrapper(err)
		}
		cont, err := utils.CheckCountRange(totalCount, offset, limit)
		if !cont {
			return []dtos.Device{}, totalCount, err
		}

		deviceModels, err = dbClient.AllDevices(offset, limit, labels)
		if err != nil {
			return devices, totalCount, errors.NewCommonEdgeXWrapper(err)
		}
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
func DevicesByProfileName(offset int, limit int, profileName string, dic *di.Container) (devices []dtos.Device, totalCount int64, err errors.EdgeX) {
	if profileName == "" {
		return devices, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "profileName is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)

	totalCount, err = dbClient.DeviceCountByProfileName(profileName)
	if err != nil {
		return devices, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.Device{}, totalCount, err
	}

	deviceModels, err := dbClient.DevicesByProfileName(offset, limit, profileName)
	if err != nil {
		return devices, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	devices = make([]dtos.Device, len(deviceModels))
	for i, d := range deviceModels {
		devices[i] = dtos.FromDeviceModelToDTO(d)
	}
	return devices, totalCount, nil
}

var errNoMessagingClient = goErrors.New("MessageBus Client not available. Please update RequireMessageBus and MessageBus configuration to enable sending System Events via the EdgeX MessageBus")

func validateParentProfileAndAutoEvent(dic *di.Container, d models.Device) errors.EdgeX {
	if d.ProfileName == "" {
		// if the profile is not set, skip the validation until we have the profile
		return nil
	}
	if (d.Name == d.Parent) && (d.Name != "") {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "a device cannot be its own parent", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	dp, err := dbClient.DeviceProfileByName(d.ProfileName)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device profile '%s' not found during validating device '%s'", d.ProfileName, d.Name), err)
	}
	if len(d.AutoEvents) == 0 {
		return nil
	}
	for _, a := range d.AutoEvents {
		_, err := time.ParseDuration(a.Interval)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("auto event interval '%s' not valid in the device '%s'", a.Interval, d.Name), err)
		}

		regex, regErr := regexp.CompilePOSIX(a.SourceName)
		if regErr != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to CompilePOSIX the auto event source name: "+a.SourceName, regErr)
		}
		hasResource := slices.ContainsFunc(dp.DeviceResources, func(r models.DeviceResource) bool {
			matchedString := regex.FindString(r.Name)
			return (r.Name == regex.String()) || (r.Name == matchedString)
		})

		hasCommand := slices.ContainsFunc(dp.DeviceCommands, func(c models.DeviceCommand) bool {
			return c.Name == a.SourceName
		})
		if !hasResource && !hasCommand {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("auto event source '%s' cannot be found in the device profile '%s'", a.SourceName, dp.Name), nil)
		}
	}
	return nil
}
