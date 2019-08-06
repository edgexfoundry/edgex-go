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

type DeleteExecutor interface {
	Execute() error
}

type addressDelete struct {
	database AddressDeleter
	id       string
	name     string
}

// This method adds the provided Addressable to the database.
func (op addressDelete) Execute() error {
	var addressable contract.Addressable
	var err error

	// deleteById and deleteByName all use the deleteById database function, so abstract away the front end difference
	if op.id == "" {
		if op.name == "" {
			// short circuit a bad request
			return errors.NewErrAddressableNotFound(op.id, op.name)
		}

		addressable, err = op.database.GetAddressableByName(op.name)
	} else {
		addressable, err = op.database.GetAddressableById(op.id)
	}

	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrAddressableNotFound(op.id, op.name)
		}

		return err
	}

	// Check device services
	ds, err := op.database.GetDeviceServicesByAddressableId(addressable.Id)
	if err != nil {
		return err
	}
	if len(ds) > 0 {
		return errors.NewErrAddressableInUse(addressable.Name)
	}

	err = op.database.DeleteAddressableById(addressable.Id)
	if err != nil {
		return err
	}

	return nil
}

// This factory method returns an executor used to delete an addressable.
// Addressables will first be searched by ID.
// If the provided ID is the empty string, it will be looked up by name.
func NewDeleteExecutor(db AddressDeleter, id string, name string) DeleteExecutor {
	return addressDelete{
		database: db,
		id:       id,
		name:     name,
	}
}
