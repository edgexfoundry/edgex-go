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
	"gopkg.in/yaml.v2"
)

type addProfileExecutor interface {
	Execute() (id string, err error)
}

type addProfile struct {
	adder        DeviceProfileAdder
	profileBytes []byte
}

// Execute performs the deletion of the device profile.
func (op addProfile) Execute() (id string, err error) {
	var dp contracts.DeviceProfile

	err = yaml.Unmarshal(op.profileBytes, &dp)
	if err != nil {
		return "", err
	}

	// Check if there are duplicate names in the device profile command list
	for _, c1 := range dp.CoreCommands {
		count := 0
		for _, c2 := range dp.CoreCommands {
			if c1.Name == c2.Name {
				count += 1
			}
		}
		if count > 1 {
			err = errors.NewErrDuplicateName("Error adding device profile: Duplicate names in the commands")
			return "", err
		}
	}

	id, err = op.adder.AddDeviceProfile(dp)
	if err != nil {
		if err == db.ErrNotUnique {
			return "", errors.NewErrDuplicateName("Duplicate profile name " + dp.Name )
		} else if err == db.ErrNameEmpty {
			return "", errors.NewErrEmptyDeviceProfileName()
		}

		return "", err
	}

	return id, nil
}

// NewGetModelExecutor creates a new GetProfilesExecutor for retrieving device profiles by model.
func NewAddDeviceProfileExecutor(profileBytes []byte, adder DeviceProfileAdder) addProfileExecutor {
	return addProfile{
		profileBytes:  profileBytes,
		adder: adder,
	}
}
