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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type DeviceCreator interface {
	Execute() (id string, err error)
}

type addDevice struct {
	database DeviceAdder
	device   contract.Device
	events   chan DeviceEvent
}

func NewAddDevice(ch chan DeviceEvent, db DeviceAdder, new contract.Device) DeviceCreator {
	return addDevice{database: db, device: new, events: ch}
}

func (op addDevice) Execute() (id string, err error) {
	evt := DeviceEvent{}
	// Lookup device service by name, then ID. Verify it exists.
	// ** TODO: Change this to CheckDeviceServiceByName **
	service, err := op.database.GetDeviceServiceByName(op.device.Service.Name)
	if err != nil {
		// Try by ID
		service, err = op.database.GetDeviceServiceById(op.device.Service.Id)
		if err != nil {
			if err == db.ErrNotFound {
				err = errors.NewErrItemNotFound(fmt.Sprintf("invalid device service: %s %s", op.device.Service.Name, op.device.Service.Id))
			}
			evt.Error = err
			op.events <- evt
			return
		}
	}
	// TODO: Can decouple DeviceService (see "check" above)
	op.device.Service = service

	// Lookup device profile by name, then ID. Verify it exists.
	profile, err := op.database.GetDeviceProfileByName(op.device.Profile.Name)
	if err != nil {
		// Try by ID
		profile, err = op.database.GetDeviceProfileById(op.device.Profile.Id)
		if err != nil {
			if err == db.ErrNotFound {
				err = errors.NewErrItemNotFound(fmt.Sprintf("invalid device profile: %s %s", op.device.Profile.Name, op.device.Profile.Id))
			}
			evt.Error = err
			op.events <- evt
			return
		}
	}
	// TODO: Will be decoupling the DeviceProfile
	op.device.Profile = profile

	// Add the device
	id, err = op.database.AddDevice(op.device, profile.CoreCommands)
	if err != nil {
		if err == db.ErrNotUnique {
			err = errors.NewErrDuplicateName(fmt.Sprintf("duplicate name for device: %s", op.device.Name))
		}
		evt.Error = err
		op.events <- evt
		return
	}

	//Device added successfully. Publish event onto the supplied channel.
	evt.DeviceId = id
	evt.DeviceName = op.device.Name
	evt.HttpMethod = http.MethodPost
	evt.ServiceId = service.Id
	op.events <- evt

	return id, nil
}
