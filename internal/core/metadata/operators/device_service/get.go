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

// DeviceServiceGetExecutor retrieves DeviceService according to parameters defined by the implementation.
type DeviceServiceGetExecutor interface {
	Execute() ([]contract.DeviceService, error)
}

type deviceServiceLoadAll struct {
	config   config.ServiceInfo
	database DeviceServiceLoader
	logger   logger.LoggingClient
}

// NewDeviceServiceLoadAll creates a new Executor that retrieves all DeviceService registered.
func NewDeviceServiceLoadAll(cfg config.ServiceInfo, db DeviceServiceLoader, log logger.LoggingClient) DeviceServiceGetExecutor {
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

type deviceServiceLoadByAddressableId struct {
	id string
	db DeviceServiceLoader
}

// NewDeviceServiceLoadByAddressableId creates a new Executor that retrieves all DeviceService associated with a given Addressable ID.
func NewDeviceServiceLoadByAddressableId(id string, db DeviceServiceLoader) DeviceServiceGetExecutor {
	return deviceServiceLoadByAddressableId{id: id, db: db}
}

// Execute performs an operation that retrieves all DeviceService associated with a given Addressable ID.
func (op deviceServiceLoadByAddressableId) Execute() ([]contract.DeviceService, error) {
	// Check if the Addressable exists
	_, err := op.db.GetAddressableById(op.id)
	if err != nil {
		if err == db.ErrNotFound {
			return nil, errors.NewErrItemNotFound(op.id)
		} else {
			return nil, err
		}
	}

	if ds, err := op.db.GetDeviceServicesByAddressableId(op.id); err != nil {
		return nil, err
	} else {
		return ds, nil
	}
}
