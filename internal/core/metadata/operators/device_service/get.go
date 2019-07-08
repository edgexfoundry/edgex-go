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
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type DeviceServiceAllExecutor interface {
	Execute() ([]contract.DeviceService, error)
}

type deviceServiceLoadAll struct {
	config   config.ServiceInfo
	database DeviceServiceLoader
	logger   logger.LoggingClient
}

func NewDeviceServiceLoadAll(cfg config.ServiceInfo, db DeviceServiceLoader, log logger.LoggingClient) DeviceServiceAllExecutor {
	return deviceServiceLoadAll{config: cfg, database: db, logger: log}
}

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
