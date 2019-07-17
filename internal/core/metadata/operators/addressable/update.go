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
package addressable

import (
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type UpdateExecutor interface {
	Execute() error
}

type addressUpdate struct {
	updater     AddressUpdater
	addressable contract.Addressable
}

// This method updates the provided Addressable in the database.
func (op addressUpdate) Execute() error {
	var dest contract.Addressable
	var err error
	// Check if the addressable exists
	if op.addressable.Id == "" {
		dest, err = op.updater.GetAddressableByName(op.addressable.Name)
	} else {
		dest, err = op.updater.GetAddressableById(op.addressable.Id)
	}

	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrAddressableNotFound(op.addressable.Id, op.addressable.Name)
		}
		return err
	}

	// If the name is changed, check if the addressable is still in use
	if op.addressable.Name != "" && op.addressable.Name != dest.Name {
		// determine if the addressable is still in use
		ds, err := op.updater.GetDeviceServicesByAddressableId(op.addressable.Id)
		if err != nil {
			return err
		}
		if len(ds) > 0 {
			err = errors.NewErrAddressableInUse(dest.Name)
			return err
		}
	}

	if op.addressable.Name != "" {
		dest.Name = op.addressable.Name
	}
	if op.addressable.Protocol != "" {
		dest.Protocol = op.addressable.Protocol
	}
	if op.addressable.Address != "" {
		dest.Address = op.addressable.Address
	}
	if op.addressable.Port != 0 {
		dest.Port = op.addressable.Port
	}
	if op.addressable.Path != "" {
		dest.Path = op.addressable.Path
	}
	if op.addressable.Publisher != "" {
		dest.Publisher = op.addressable.Publisher
	}
	if op.addressable.User != "" {
		dest.User = op.addressable.User
	}
	if op.addressable.Password != "" {
		dest.Password = op.addressable.Password
	}
	if op.addressable.Topic != "" {
		dest.Topic = op.addressable.Topic
	}

	if err := op.updater.UpdateAddressable(dest); err != nil {
		return err
	}

	return nil
}

// This factory method returns an executor used to update an addressable.
func NewUpdateExecutor(updater AddressUpdater, addressable contract.Addressable) UpdateExecutor {
	return addressUpdate{
		updater:     updater,
		addressable: addressable,
	}
}
