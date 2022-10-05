//
// Copyright (C) 2020-2022 IOTech Ltd
// Copyright (C) 2022 Intel
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"encoding/json"
	goErrors "errors"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

// The AddDevice function accepts the new device model from the controller function
// and then invokes AddDevice function of infrastructure layer to add new device
func AddDevice(d models.Device, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	err := validateDeviceCallback(ctx, dic, dtos.FromDeviceModelToDTO(d))
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

	device := dtos.FromDeviceModelToDTO(d)
	go addDeviceCallback(ctx, dic, device)

	go publishDeviceSystemEvent(common.DeviceSystemEventActionAdd, d.ServiceName, d, ctx, lc, dic)

	return addedDevice.Id, nil
}

// DeleteDeviceByName deletes the device by name
func DeleteDeviceByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

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
	go deleteDeviceCallback(ctx, dic, device)

	go publishDeviceSystemEvent(common.DeviceSystemEventActionDelete, device.ServiceName, device, ctx, lc, dic)

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

	err = validateDeviceCallback(ctx, dic, dtos.FromDeviceModelToDTO(device))
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
		go updateDeviceCallback(ctx, dic, oldServiceName, device)
		go publishDeviceSystemEvent(common.DeviceSystemEventActionUpdate, oldServiceName, device, ctx, lc, dic)
	}

	go updateDeviceCallback(ctx, dic, device.ServiceName, device)
	go publishDeviceSystemEvent(common.DeviceSystemEventActionUpdate, device.ServiceName, device, ctx, lc, dic)

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

var noMessagingClientError = goErrors.New("MessageBus Client not available. Please update RequireMessageBus and MessageQueue configuration to enable sending System Events via the EdgeX MessageBus")

func publishDeviceSystemEvent(action string, owner string, d models.Device, ctx context.Context, lc logger.LoggingClient, dic *di.Container) {
	device := dtos.FromDeviceModelToDTO(d)
	systemEvent := dtos.NewSystemEvent(common.DeviceSystemEventType, action, common.CoreMetaDataServiceKey, owner, nil, device)

	messagingClient := bootstrapContainer.MessagingClientFrom(dic.Get)
	if messagingClient == nil {
		// For 2.x this is a warning due to backwards compatability
		// TODO: For change this to be Errorf for EdgeX 3.0
		lc.Warnf("unable to publish Device System Event: %v", noMessagingClientError)
		return
	}

	config := container.ConfigurationFrom(dic.Get)
	publishTopic := fmt.Sprintf("%s/%s/%s/%s/%s/%s",
		config.MessageQueue.PublishTopicPrefix,
		systemEvent.Source,
		systemEvent.Type,
		systemEvent.Action,
		systemEvent.Owner,
		device.ProfileName)

	payload, _ := json.Marshal(systemEvent)
	envelope := types.NewMessageEnvelope(payload, ctx)
	// Correlation ID and Content type are set by the above factory function from the context of the request that
	// triggered this System Event. We'll keep that Correlation ID, but need to make sure the Content Type is set appropriate
	// for how the payload was encoded above.
	envelope.ContentType = common.ContentTypeJSON

	if err := messagingClient.Publish(envelope, publishTopic); err != nil {
		lc.Errorf("unable to publish '%s' Device System Event for device '%s' to topic '%s': %v", action, device.Name, publishTopic, err)
		return
	}

	lc.Debugf("Published the '%s' Device System Event for device '%s' to topic '%s'", action, device.Name, publishTopic)
}
