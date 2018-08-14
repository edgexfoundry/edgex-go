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

import "time"

// Update when the device was last reported connected
func updateDeviceLastReportedConnected(device string) {
	// Config set to skip update last reported
	if !configuration.DeviceUpdateLastConnected {
		LoggingClient.Debug("Skipping update of device connected/reported times for:  " + device)
		return
	}

	d, err := mdc.CheckForDevice(device)
	if err != nil {
		LoggingClient.Error("Error getting device " + device + ": " + err.Error())
		return
	}

	// Couldn't find device
	if len(d.Name) == 0 {
		LoggingClient.Error("Error updating device connected/reported times.  Unknown device with identifier of:  " + device)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)
	// Found device, now update lastReported
	err = mdc.UpdateLastConnectedByName(d.Name, t)
	if err != nil {
		LoggingClient.Error("Problems updating last connected value for device: " + d.Name)
		return
	}
	err = mdc.UpdateLastReportedByName(d.Name, t)
	if err != nil {
		LoggingClient.Error("Problems updating last reported value for device: " + d.Name)
	}
	return
}

// Update when the device service was last reported connected
func updateDeviceServiceLastReportedConnected(device string) {
	if !configuration.ServiceUpdateLastConnected {
		LoggingClient.Debug("Skipping update of device service connected/reported times for:  " + device)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	// Get the device
	d, err := mdc.CheckForDevice(device)
	if err != nil {
		LoggingClient.Error("Error getting device " + device + ": " + err.Error())
		return
	}

	// Couldn't find device
	if len(d.Name) == 0 {
		LoggingClient.Error("Error updating device connected/reported times.  Unknown device with identifier of:  " + device)
		return
	}

	// Get the device service
	s := d.Service
	if &s == nil {
		LoggingClient.Error("Error updating device service connected/reported times.  Unknown device service in device:  " + d.Name)
		return
	}

	msc.UpdateLastConnected(s.Service.Id.Hex(), t)
	msc.UpdateLastReported(s.Service.Id.Hex(), t)
}
