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
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// DeviceServiceGetListExecutor retrieves DeviceServices according to parameters defined by the implementation.
type DeviceServiceGetListExecutor interface {
	Execute() ([]contract.DeviceService, error)
}

type deviceServiceLoadAll struct {
	config   config.ServiceInfo
	database DeviceServiceLoader
	logger   logger.LoggingClient
}

// NewDeviceServiceLoadAll creates a new Executor that retrieves all DeviceService registered.
func NewDeviceServiceLoadAll(cfg config.ServiceInfo, db DeviceServiceLoader, log logger.LoggingClient) DeviceServiceGetListExecutor {
	return deviceServiceLoadAll{config: cfg, database: db, logger: log}
}

// Execute performs an operation that retrieves all DeviceService registered.
func (op deviceServiceLoadAll) Execute() (services []contract.DeviceService, err error) {
	services, err = op.database.GetAllDeviceServices()
	if err != nil {
		op.logger.Error(err.Error())
		return
	}
	if len(services) > op.config.MaxResultCount {
		err = errors.NewErrLimitExceeded(op.config.MaxResultCount)
		return []contract.DeviceService{}, err
	}
	return
}

type deviceServiceLoadByAddressable struct {
	id   string
	name string
	db   DeviceServiceLoader
}

// NewDeviceServiceLoadByAddressableName creates a new Executor that retrieves all DeviceService associated with a given Addressable name.
func NewDeviceServiceLoadByAddressableName(name string, db DeviceServiceLoader) DeviceServiceGetListExecutor {
	return deviceServiceLoadByAddressable{name: name, db: db}
}

// NewDeviceServiceLoadByAddressableID creates a new Executor that retrieves all DeviceService associated with a given Addressable ID.
func NewDeviceServiceLoadByAddressableID(id string, db DeviceServiceLoader) DeviceServiceGetListExecutor {
	return deviceServiceLoadByAddressable{id: id, db: db}
}

// Execute performs an operation that retrieves all DeviceService associated with a given Addressable.
func (op deviceServiceLoadByAddressable) Execute() ([]contract.DeviceService, error) {
	var addr contract.Addressable
	var err error

	// Check if the Addressable exists
	// determine whether we're doing a lookup by ID or name
	if op.id != "" {
		addr, err = op.db.GetAddressableById(op.id)
		if err == db.ErrNotFound {
			err = errors.NewErrItemNotFound(op.id)
		}
	} else {
		addr, err = op.db.GetAddressableByName(op.name)
		if err == db.ErrNotFound {
			err = errors.NewErrItemNotFound(op.name)
		}
	}

	if err != nil {
		return nil, err
	}

	if ds, err := op.db.GetDeviceServicesByAddressableId(addr.Id); err != nil {
		return nil, err
	} else {
		return ds, nil
	}
}

// DeviceServiceGetExecutor retrieves DeviceService according to parameters defined by the implementation.
type DeviceServiceGetExecutor interface {
	Execute() (contract.DeviceService, error)
}

type deviceServiceLoadById struct {
	id string
	db DeviceServiceLoader
}

// NewDeviceServiceLoadById creates a new Executor that retrieves the DeviceService associated with a given ID.
func NewDeviceServiceLoadById(id string, db DeviceServiceLoader) DeviceServiceGetExecutor {
	return deviceServiceLoadById{id: id, db: db}
}

// Execute performs an operation that retrieves the DeviceService associated with a given ID.
func (op deviceServiceLoadById) Execute() (contract.DeviceService, error) {
	ds, err := op.db.GetDeviceServiceById(op.id)
	if err != nil {
		if err == db.ErrNotFound {
			return contract.DeviceService{}, errors.NewErrItemNotFound(op.id)
		} else {
			return contract.DeviceService{}, err
		}
	}

	return ds, nil
}

type deviceServiceLoadByName struct {
	name string
	db   DeviceServiceLoader
}

// NewDeviceServiceLoadByName creates a new Executor that retrieves the DeviceService associated with a given name.
func NewDeviceServiceLoadByName(name string, db DeviceServiceLoader) DeviceServiceGetExecutor {
	return deviceServiceLoadByName{name: name, db: db}
}

// Execute performs an operation that retrieves the DeviceService associated with a given name.
func (op deviceServiceLoadByName) Execute() (contract.DeviceService, error) {
	ds, err := op.db.GetDeviceServiceByName(op.name)
	if err != nil {
		if err == db.ErrNotFound {
			return contract.DeviceService{}, errors.NewErrItemNotFound(op.name)
		} else {
			return contract.DeviceService{}, err
		}
	}

	return ds, nil
}
