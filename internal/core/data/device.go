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

// Update when the device was last reported connected
func updateDeviceLastReportedConnected(
	device string,
	loggingClient logger.LoggingClient,
	mdc metadata.DeviceClient,
	configuration *config.ConfigurationStruct) {

	// Config set to skip update last reported
	if !configuration.Writable.DeviceUpdateLastConnected {
		loggingClient.Debug("Skipping update of device connected/reported times for:  " + device)
		return
	}

	d, err := mdc.CheckForDevice(device, context.Background())
	if err != nil {
		loggingClient.Error("Error getting device " + device + ": " + err.Error())
		return
	}

	// Couldn't find device
	if len(d.Name) == 0 {
		loggingClient.Error("Error updating device connected/reported times.  Unknown device with identifier of:  " + device)
		return
	}

	t := db.MakeTimestamp()
	// Found device, now update lastReported
	// Use of context.Background because this function is invoked asynchronously from a channel
	err = mdc.UpdateLastConnectedByName(d.Name, t, context.Background())
	if err != nil {
		loggingClient.Error("Problems updating last connected value for device: " + d.Name)
		return
	}
	err = mdc.UpdateLastReportedByName(d.Name, t, context.Background())
	if err != nil {
		loggingClient.Error("Problems updating last reported value for device: " + d.Name)
	}
	return
}

// Update when the device service was last reported connected
func updateDeviceServiceLastReportedConnected(
	device string,
	loggingClient logger.LoggingClient,
	mdc metadata.DeviceClient,
	msc metadata.DeviceServiceClient,
	configuration *config.ConfigurationStruct) {

	if !configuration.Writable.ServiceUpdateLastConnected {
		loggingClient.Debug("Skipping update of device service connected/reported times for:  " + device)
		return
	}

	t := db.MakeTimestamp()

	// Get the device
	d, err := mdc.CheckForDevice(device, context.Background())
	if err != nil {
		loggingClient.Error("Error getting device " + device + ": " + err.Error())
		return
	}

	// Couldn't find device
	if len(d.Name) == 0 {
		loggingClient.Error("Error updating device connected/reported times.  Unknown device with identifier of:  " + device)
		return
	}

	// Get the device service
	s := d.Service
	if &s == nil {
		loggingClient.Error("Error updating device service connected/reported times.  Unknown device service in device:  " + d.Name)
		return
	}

	// Use of context.Background because this function is invoked asynchronously from a channel
	msc.UpdateLastConnected(s.Id, t, context.Background())
	msc.UpdateLastReported(s.Id, t, context.Background())
}

func checkDevice(
	device string,
	ctx context.Context,
	mdc metadata.DeviceClient,
	configuration *config.ConfigurationStruct) error {

	if configuration.Writable.MetaDataCheck {
		_, err := mdc.CheckForDevice(device, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
