/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package device

import (
	"fmt"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type DeviceUpdateExecutor interface {
	Execute() (err error)
}

type updateDevice struct {
	database DeviceUpdater
	device   contract.Device
	events   chan DeviceEvent
	logger   logger.LoggingClient
}

func NewUpdateDevice(ch chan DeviceEvent, db DeviceUpdater, d contract.Device, logger logger.LoggingClient) DeviceUpdateExecutor {
	return updateDevice{database: db, events: ch, device: d, logger: logger}
}

func (op updateDevice) Execute() (err error) {
	var evt DeviceEvent

	// Check if the device exists
	// First try ID
	var oldDevice contract.Device
	oldDevice, err = op.database.GetDeviceById(op.device.Id)
	if err != nil {
		// Then try name
		oldDevice, err = op.database.GetDeviceByName(op.device.Name)
		if err != nil {
			if err == db.ErrNotFound {
				err = errors.NewErrItemNotFound(fmt.Sprintf("device not found: %s %s", op.device.Name, op.device.Id))
			}
			op.logger.Error(err.Error())
			evt.Error = err
			op.events <- evt
			return
		}
	}

	// Check if the name is unique by querying for a device with the given name
	// and ensuring the ID matches what's given to us as well.
	var checkD contract.Device
	checkD, err = op.database.GetDeviceByName(op.device.Name)
	if err != nil && err != db.ErrNotFound {
		op.logger.Error(err.Error())
		evt.Error = err
		op.events <- evt
		return
	}

	// Found a device, make sure its the one we're trying to update
	// Different IDs -> Name is not unique
	if checkD.Id != op.device.Id {
		err = errors.NewErrDuplicateName("Duplicate name for Device")
		op.logger.Error(err.Error())
		evt.Error = err
		op.events <- evt
		return
	}

	var service contract.DeviceService
	if (op.device.Service.String() != contract.DeviceService{}.String()) {
		// Check if the new service exists
		// Try ID first
		service, err = op.database.GetDeviceServiceById(op.device.Service.Id)
		if err != nil {
			// Then try name
			service, err = op.database.GetDeviceServiceByName(op.device.Service.Name)
			if err != nil {
				op.logger.Error(err.Error())
				err = errors.NewErrItemNotFound("Device service not found for updated device")
				op.logger.Error(err.Error())
				evt.Error = err
				op.events <- evt
				return
			}
		}
		op.device.Service = service
	}

	var profile contract.DeviceProfile
	if (op.device.Profile.String() != contract.DeviceProfile{}.String()) {
		// Check if the new profile exists
		// Try ID first
		profile, err = op.database.GetDeviceProfileById(op.device.Profile.Id)
		if err != nil {
			// Then try Name
			profile, err = op.database.GetDeviceProfileByName(op.device.Profile.Name)
			if err != nil {
				err = errors.NewErrItemNotFound("Device profile not found for updated device")
				op.logger.Error(err.Error())
				evt.Error = err
				op.events <- evt
				return
			}
		}
		op.device.Profile = profile
	}

	if err != nil {
		op.logger.Error(err.Error())
		evt.Error = err
		op.events <- evt
		return
	}

	updatedDevice := op.updateDeviceFields(oldDevice)

	if err = op.database.UpdateDevice(updatedDevice); err != nil {
		op.logger.Error(err.Error())
		evt.Error = err
		op.events <- evt
		return
	}

	// Device updated successfully.  Publish event onto supplied channel.

	evt.DeviceId = op.device.Id
	evt.DeviceName = op.device.Name
	evt.HttpMethod = http.MethodPost
	evt.ServiceId = op.device.Service.Id

	op.events <- evt

	return nil
}

func (op updateDevice) updateDeviceFields(original contract.Device) (updated contract.Device) {
	updated = original

	if (op.device.Service.String() != contract.DeviceService{}.String()) {
		updated.Service = op.device.Service
	}
	if (op.device.Profile.String() != contract.DeviceProfile{}.String()) {
		updated.Profile = op.device.Profile
	}
	if len(op.device.Protocols) > 0 {
		updated.Protocols = op.device.Protocols
	}
	if len(op.device.AutoEvents) > 0 {
		updated.AutoEvents = op.device.AutoEvents
	}
	if op.device.AdminState != "" {
		updated.AdminState = op.device.AdminState
	}
	if op.device.Description != "" {
		updated.Description = op.device.Description
	}
	if op.device.Labels != nil {
		updated.Labels = op.device.Labels
	}
	if op.device.LastConnected != 0 {
		updated.LastConnected = op.device.LastConnected
	}
	if op.device.LastReported != 0 {
		updated.LastReported = op.device.LastReported
	}
	if op.device.Location != nil {
		updated.Location = op.device.Location
	}
	if op.device.OperatingState != contract.OperatingState("") {
		updated.OperatingState = op.device.OperatingState
	}
	if op.device.Origin != 0 {
		updated.Origin = op.device.Origin
	}
	if op.device.Name != "" {
		updated.Name = op.device.Name
	}

	return
}
