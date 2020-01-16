/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
package data

import (
	"context"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
)

// Update when the device service was last reported connected
func updateDeviceServiceLastReportedConnected(
	device string,
	lc logger.LoggingClient,
	mdc metadata.DeviceClient,
	msc metadata.DeviceServiceClient,
	configuration *config.ConfigurationStruct) {

	if !configuration.Writable.ServiceUpdateLastConnected {
		lc.Debug("Skipping update of device service connected/reported times for:  " + device)
		return
	}

	t := db.MakeTimestamp()

	// Get the device
	d, err := mdc.CheckForDevice(context.Background(), device)
	if err != nil {
		lc.Error("Error getting device " + device + ": " + err.Error())
		return
	}

	// Couldn't find device
	if len(d.Name) == 0 {
		lc.Error("Error updating device connected/reported times.  Unknown device with identifier of:  " + device)
		return
	}

	// Get the device service
	s := d.Service
	if &s == nil {
		lc.Error("Error updating device service connected/reported times.  Unknown device service in device:  " + d.Name)
		return
	}

	// Use of context.Background because this function is invoked asynchronously from a channel
	_ = msc.UpdateLastConnected(context.Background(), s.Id, t)
	_ = msc.UpdateLastReported(context.Background(), s.Id, t)
}

func checkDevice(
	device string,
	ctx context.Context,
	mdc metadata.DeviceClient,
	configuration *config.ConfigurationStruct) error {

	if configuration.Writable.MetaDataCheck {
		_, err := mdc.CheckForDevice(ctx, device)
		if err != nil {
			return err
		}
	}
	return nil
}
