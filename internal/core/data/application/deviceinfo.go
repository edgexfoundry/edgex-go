//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

func (a *CoreDataApp) RemoveDeviceInfosByDeviceName(deviceName string, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	err := dbClient.RemoveDeviceInfosByDeviceName(deviceName)
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "delete deviceInfos by deviceName failed", err)
	}
	return nil
}
