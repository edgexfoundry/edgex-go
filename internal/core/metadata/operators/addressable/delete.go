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
	database   AddressDeleter
	identifier string
}

// This method adds the provided Addressable to the database.
func (op addressDelete) Execute() error {
	// Check if the addressable exists
	idAddr, idErr := op.database.GetAddressableById(op.identifier)
	nameAddr, nameErr := op.database.GetAddressableByName(op.identifier)

	// if this case hits, then it's safe to say we don't have a usable addressable
	if idErr != nil && nameErr != nil {
		if idErr == db.ErrNotFound && nameErr == db.ErrNotFound {
			// not the cleanest thing but we can't say anything for certain about the data input
			return errors.NewErrAddressableNotFound(op.identifier, op.identifier)
		}

		return nameErr
	}

	var a contract.Addressable

	// use the addressable from the operation that did not return an error
	if idErr != nil {
		a = nameAddr
	} else {
		a = idAddr
	}

	// Check device services
	ds, err := op.database.GetDeviceServicesByAddressableId(a.Id)
	if err != nil {
		return err
	}
	if len(ds) > 0 {
		return errors.NewErrAddressableInUse(a.Name)
	}

	err = op.database.DeleteAddressableById(a.Id)
	if err != nil {
		return err
	}

	return nil
}

// This factory method returns an executor used to delete an addressable.
func NewDeleteExecutor(db AddressDeleter, identifier string) DeleteExecutor {
	return addressDelete{
		database:   db,
		identifier: identifier,
	}
}
