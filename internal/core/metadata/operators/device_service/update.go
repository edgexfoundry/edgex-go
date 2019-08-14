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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type UpdateOpStateByNameExecutor interface {
	Execute() error
}

type deviceServiceOpStateUpdateByName struct {
	name string
	os contract.OperatingState
	db DeviceServiceUpdater
}

func NewUpdateOpStateByNameExecutor(name string, os contract.OperatingState, db DeviceServiceUpdater) UpdateOpStateByNameExecutor {
	return deviceServiceOpStateUpdateByName{name: name, os: os, db: db}
}

func (op deviceServiceOpStateUpdateByName) Execute() error {
	// Check if the device service exists
	ds, err := op.db.GetDeviceServiceByName(op.name)
	if err != nil {
		return err
	}

	ds.OperatingState = op.os
	if err = op.db.UpdateDeviceService(ds); err != nil {
		return err
	}

	return nil
}

type UpdateOpStateByIdExecutor interface {
	Execute() error
}

type deviceServiceOpStateUpdateById struct {
	id string
	os contract.OperatingState
	db DeviceServiceUpdater
}

func NewUpdateOpStateByIdExecutor(id string, os contract.OperatingState, db DeviceServiceUpdater) UpdateOpStateByIdExecutor {
	return deviceServiceOpStateUpdateById{id: id, os: os, db: db}
}

func (op deviceServiceOpStateUpdateById) Execute() error {
	// Check if the device service exists
	ds, err := op.db.GetDeviceServiceById(op.id)
	if err != nil {
		return err
	}

	ds.OperatingState = op.os
	if err = op.db.UpdateDeviceService(ds); err != nil {
		return err
	}

	return nil
}

type UpdateAdminStateByNameExecutor interface {
	Execute() error
}

type deviceServiceAdminStateUpdateByName struct {
	name string
	as contract.AdminState
	db DeviceServiceUpdater
}

func NewUpdateAdminStateByNameExecutor(name string, as contract.AdminState, db DeviceServiceUpdater) UpdateAdminStateByNameExecutor {
	return deviceServiceAdminStateUpdateByName{name: name, as: as, db: db}
}

func (op deviceServiceAdminStateUpdateByName) Execute() error {
	// Check if the device service exists
	ds, err := op.db.GetDeviceServiceByName(op.name)
	if err != nil {
		return err
	}

	ds.AdminState = op.as
	if err = op.db.UpdateDeviceService(ds); err != nil {
		return err
	}

	return nil
}

type UpdateAdminStateByIdExecutor interface {
	Execute() error
}

type deviceServiceAdminStateUpdateById struct {
	id string
	as contract.AdminState
	db DeviceServiceUpdater
}

func NewUpdateAdminStateByIdExecutor(id string, as contract.AdminState, db DeviceServiceUpdater) UpdateAdminStateByIdExecutor {
	return deviceServiceAdminStateUpdateById{id: id, as: as, db: db}
}

func (op deviceServiceAdminStateUpdateById) Execute() error {
	// Check if the device service exists
	ds, err := op.db.GetDeviceServiceById(op.id)
	if err != nil {
		return err
	}

	ds.AdminState = op.as
	if err = op.db.UpdateDeviceService(ds); err != nil {
		return err
	}

	return nil
}
