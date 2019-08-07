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

package device_profile

import (
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contracts "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type addProfileExecutor interface {
	Execute() (id string, err error)
}

type addProfile struct {
	adder         DeviceProfileAdder
	deviceProfile contracts.DeviceProfile
}

// Execute performs the deletion of the device profile.
func (op addProfile) Execute() (id string, err error) {
	valid, err := op.deviceProfile.Validate()
	if !valid || err != nil {
		return "", err
	}

	for _, command := range op.deviceProfile.CoreCommands {
		valid, err = command.Validate()
		if !valid {
			return "", errors.NewErrBadRequest(command.Name)
		} else if err != nil {
			return "", err
		}
	}

	id, err = op.adder.AddDeviceProfile(op.deviceProfile)
	if err != nil {
		if err == db.ErrNotUnique {
			return "", errors.NewErrDuplicateName("Duplicate profile name " + op.deviceProfile.Name)
		} else if err == db.ErrNameEmpty {
			return "", errors.NewErrEmptyDeviceProfileName()
		}

		return "", err
	}

	return id, nil
}

// NewGetModelExecutor creates a new GetProfilesExecutor for retrieving device profiles by model.
func NewAddDeviceProfileExecutor(deviceProfile contracts.DeviceProfile, adder DeviceProfileAdder) addProfileExecutor {
	return addProfile{
		deviceProfile: deviceProfile,
		adder:         adder,
	}
}
