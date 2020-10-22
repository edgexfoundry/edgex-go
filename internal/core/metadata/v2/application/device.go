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
