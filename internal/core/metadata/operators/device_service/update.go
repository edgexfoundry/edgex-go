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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
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
	lc logger.LoggingClient
}

// NewUpdateAdminStateByIdExecutor updates a device service's AdminState, referencing the DeviceService by ID.
func NewUpdateAdminStateByIdExecutor(id string, as contract.AdminState, db DeviceServiceUpdater, lc logger.LoggingClient) UpdateAdminOrOperatingStateExecutor {
	return deviceServiceAdminStateUpdateById{id: id, as: as, db: db, lc: lc}
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

	go adminStateCallback(ds, op.lc)

	return nil
}

type deviceServiceAdminStateUpdateByName struct {
	name string
	as   contract.AdminState
	db   DeviceServiceUpdater
	lc   logger.LoggingClient
}

// NewUpdateAdminStateByNameExecutor updates a device service's AdminState, referencing the DeviceService by name.
func NewUpdateAdminStateByNameExecutor(name string, as contract.AdminState, db DeviceServiceUpdater, lc logger.LoggingClient) UpdateAdminOrOperatingStateExecutor {
	return deviceServiceAdminStateUpdateByName{name: name, as: as, db: db, lc: lc}
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

	go adminStateCallback(ds, op.lc)

	return nil
}

func adminStateCallback(
	service contract.DeviceService,
	lc logger.LoggingClient) {

	if len(service.Addressable.GetCallbackURL()) == 0 {
		return
	}

	req, err := createCallbackRequest(service)
	if err != nil {
		lc.Error(fmt.Sprintf("fail to create callback request for %s", service.Name))
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		lc.Error(fmt.Sprintf("fail to invoke callback for %s, %v", service.Name, err))
		return
	} else if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			lc.Error(fmt.Sprintf("fail to read response body, %v", err))
		}
		lc.Error(fmt.Sprintf("fail to invoke callback for %s, %s", service.Name, string(b)))
	}
	resp.Body.Close()
	resp.Close = true
}

func createCallbackRequest(service contract.DeviceService) (*http.Request, error) {
	body, err := json.Marshal(contract.CallbackAlert{ActionType: contract.SERVICE, Id: service.Id})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPut, service.Addressable.GetCallbackURL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add(clients.ContentType, clients.ContentTypeJSON)
	return req, nil
}
