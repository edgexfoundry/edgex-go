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
	"fmt"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type DeleteByNameExecutor interface {
	Execute() error
}

type addressDeleteByName struct {
	database    AddressDeleter
	name string
}

// This method adds the provided Addressable to the database.
func (op addressDeleteByName) Execute() error {
	// Check if the addressable exists
	a, err := op.database.GetAddressableByName(op.name)
	if err != nil {
		return err
	}

	// Check device services
	ds, err := op.database.GetDeviceServicesByAddressableId(a.Id)
	if err != nil {
		return err
	}
	if len(ds) > 0 {
		return errors.NewErrAddressableInUse(op.name)
	}
}

// This factory method returns an executor used to delete an addressable.
func NewDeleteByNameExecutor(db AddressDeleter, name string) DeleteByNameExecutor {
	return addressDeleteByName{
		database:    db,
		name: name,
	}
}

type DeleteByIdExecutor interface {
	Execute() error
}

type addressDeleteById struct {
	database    AddressDeleter
	id string
}

// This method adds the provided Addressable to the database.
func (op addressDeleteById) Execute() error {
	// Check if the addressable exists
	a, err := op.database.GetAddressableById(op.id)
	if err != nil {
		return err
	}

	// Check device services
	ds, err := op.database.GetDeviceServicesByAddressableId(a.Id)
	if err != nil {
		return err
	}
	if len(ds) > 0 {
		return errors.NewErrAddressableInUse(a.Name)
	}
}

// This factory method returns an executor used to delete an addressable.
func NewDeleteByIdExecutor(db AddressDeleter, id string) DeleteByIdExecutor {
	return addressDeleteById{
		database:    db,
		id: id,
	}
}
