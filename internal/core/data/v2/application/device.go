//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"

	"github.com/edgexfoundry/go-mod-bootstrap/di"

	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"

	"github.com/edgexfoundry/go-mod-core-contracts/errors"
)

// This function will be updated when CheckDevice in v2 core-metadata is available
func checkDevice(deviceName string, ctx context.Context, dic *di.Container) errors.EdgeX {
	mdc := v2DataContainer.MetadataDeviceClientFrom(dic.Get)
	configuration := dataContainer.ConfigurationFrom(dic.Get)

	if configuration.Writable.MetaDataCheck {
		_, err := mdc.CheckForDevice(ctx, deviceName)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, "checking device name failed", err)
		}
	}
	return nil
}
