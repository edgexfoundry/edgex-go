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

	return nil
}
