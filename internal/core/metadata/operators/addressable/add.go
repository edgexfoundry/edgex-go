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

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type AddExecutor interface {
	Execute() (string, error)
}

type addressAdd struct {
	database    AddressWriter
	addressable contract.Addressable
}

// This method adds the provided Addressable to the database.
func (op addressAdd) Execute() (id string, err error) {
	if len(op.addressable.Name) == 0 {
		err := errors.NewErrEmptyAddressableName()
		return "", err
	}
	id, err = op.database.AddAddressable(op.addressable)
	if err != nil {
		if err == db.ErrNotUnique {
			err = errors.NewErrDuplicateName(fmt.Sprintf("duplicate name for addressable: %s", op.addressable.Name))
		}
		return "", err
	}

	return id, nil
}

// This factory method returns an executor used to add an addressable.
func NewAddExecutor(db AddressWriter, addressable contract.Addressable) AddExecutor {
	return addressAdd{
		database:    db,
		addressable: addressable,
	}
}
