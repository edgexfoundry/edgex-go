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
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type DeviceAllExecutor interface {
	Execute() ([]contract.Device, error)
}

type deviceLoadAll struct {
	config   bootstrapConfig.ServiceInfo
	database DeviceLoader
	logger   logger.LoggingClient
}

func NewDeviceLoadAll(cfg bootstrapConfig.ServiceInfo, db DeviceLoader, log logger.LoggingClient) DeviceAllExecutor {
	return deviceLoadAll{config: cfg, database: db, logger: log}
}

func (op deviceLoadAll) Execute() (devices []contract.Device, err error) {
	devices, err = op.database.GetAllDevices()
	if err != nil {
		op.logger.Error(err.Error())
		return
	}

	if len(devices) > op.config.MaxResultCount {
		err = errors.NewErrLimitExceeded(op.config.MaxResultCount)
		return []contract.Device{}, err
	}
	return
}

// ProfileIdExecutor provides functionality for loading devices by way of the operator pattern.
type ProfileIdExecutor interface {
	Execute() ([]contract.Device, error)
}

type deviceLoadByProfileId struct {
	config   bootstrapConfig.ServiceInfo
	database DeviceLoader
	profile  string
	logger   logger.LoggingClient
}

// NewProfileIdExecutor creates a new ProfileIdExecutor
func NewProfileIdExecutor(cfg bootstrapConfig.ServiceInfo, db DeviceLoader, log logger.LoggingClient, profile string) ProfileIdExecutor {
	return deviceLoadByProfileId{config: cfg, database: db, logger: log, profile: profile}
}

// Execute retrieves the devices associated with a profile ID
func (op deviceLoadByProfileId) Execute() (devices []contract.Device, err error) {
	devices, err = op.database.GetDevicesByProfileId(op.profile)
	if err != nil {
		op.logger.Error(err.Error())
		return
	}

	if len(devices) > op.config.MaxResultCount {
		err = errors.NewErrLimitExceeded(op.config.MaxResultCount)
		return []contract.Device{}, err
	}

	return
}
