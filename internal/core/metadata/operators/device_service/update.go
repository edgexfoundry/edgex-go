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

package device_service

import (
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// UpdateAdminOrOperatingStateExecutor updates a device service's AdminState or OperatingState fields.
type UpdateAdminOrOperatingStateExecutor interface {
	Execute() error
}

type deviceServiceOpStateUpdateById struct {
	id string
	os contract.OperatingState
	db DeviceServiceUpdater
}

// NewUpdateOpStateByIdExecutor updates a device service's OperatingState, referencing the DeviceService by ID.
func NewUpdateOpStateByIdExecutor(id string, os contract.OperatingState, db DeviceServiceUpdater) UpdateAdminOrOperatingStateExecutor {
	return deviceServiceOpStateUpdateById{id: id, os: os, db: db}
}

// Execute updates the device service OperatingState.
func (op deviceServiceOpStateUpdateById) Execute() error {
	// Check if the device service exists
	ds, err := op.db.GetDeviceServiceById(op.id)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrItemNotFound(op.id)
		}

		return err
	}

	ds.OperatingState = op.os
	if err = op.db.UpdateDeviceService(ds); err != nil {
		return err
	}

	return nil
}

type deviceServiceOpStateUpdateByName struct {
	name string
	os   contract.OperatingState
	db   DeviceServiceUpdater
}

// NewUpdateOpStateByNameExecutor updates a device service's OperatingState, referencing the DeviceService by name.
func NewUpdateOpStateByNameExecutor(name string, os contract.OperatingState, db DeviceServiceUpdater) UpdateAdminOrOperatingStateExecutor {
	return deviceServiceOpStateUpdateByName{name: name, os: os, db: db}
}

// Execute updates the device service OperatingState.
func (op deviceServiceOpStateUpdateByName) Execute() error {
	// Check if the device service exists
	ds, err := op.db.GetDeviceServiceByName(op.name)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrItemNotFound(op.name)
		}

		return err
	}

	ds.OperatingState = op.os
	if err = op.db.UpdateDeviceService(ds); err != nil {
		return err
	}

	return nil
}

type deviceServiceAdminStateUpdateById struct {
	id string
	as contract.AdminState
	db DeviceServiceUpdater
}

// NewUpdateAdminStateByIdExecutor updates a device service's AdminState, referencing the DeviceService by ID.
func NewUpdateAdminStateByIdExecutor(id string, as contract.AdminState, db DeviceServiceUpdater) UpdateAdminOrOperatingStateExecutor {
	return deviceServiceAdminStateUpdateById{id: id, as: as, db: db}
}

// Execute updates the device service AdminState.
func (op deviceServiceAdminStateUpdateById) Execute() error {
	// Check if the device service exists
	ds, err := op.db.GetDeviceServiceById(op.id)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrItemNotFound(op.id)
		}

		return err
	}

	ds.AdminState = op.as
	if err = op.db.UpdateDeviceService(ds); err != nil {
		return err
	}

	return nil
}

type deviceServiceAdminStateUpdateByName struct {
	name string
	as   contract.AdminState
	db   DeviceServiceUpdater
}

// NewUpdateAdminStateByNameExecutor updates a device service's AdminState, referencing the DeviceService by name.
func NewUpdateAdminStateByNameExecutor(name string, as contract.AdminState, db DeviceServiceUpdater) UpdateAdminOrOperatingStateExecutor {
	return deviceServiceAdminStateUpdateByName{name: name, as: as, db: db}
}

// Execute updates the device service AdminState.
func (op deviceServiceAdminStateUpdateByName) Execute() error {
	// Check if the device service exists
	ds, err := op.db.GetDeviceServiceByName(op.name)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrItemNotFound(op.name)
		}

		return err
	}

	ds.AdminState = op.as
	if err = op.db.UpdateDeviceService(ds); err != nil {
		return err
	}

	return nil
}
